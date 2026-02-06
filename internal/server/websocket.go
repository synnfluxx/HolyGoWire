package server

import (
	"TextMeByte/internal/chat"
	"TextMeByte/internal/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/time/rate"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type incomingMessage struct {
	Text  string `json:"text"`
	Files []struct {
		FilePath string `json:"filepath"`
		FileType string `json:"filetype"`
		FileSize int64  `json:"filesize"`
	} `json:"files,omitempty"`
}

func (s *server) HandleWS(hub *chat.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var username = ""
		var isAuthorized bool
		var userID int64 = 0

		token := r.URL.Query().Get("token")
		if token == "" {
			isAuthorized = false
		} else {
			claims, err := ValidateToken(token)
			if err != nil {
				s.logger.Warnf("WebSocket authentication failed: %v", err)
				isAuthorized = false
			} else {
				userID = claims.UserID
				username = claims.Username
				isAuthorized = true
			}
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			s.logger.Errorf("Failed to upgrade WebSocket connection: %v", err)
			s.error(w, r, http.StatusInternalServerError, fmt.Errorf("failed to upgrade connection"))
			return
		}

		limiter := rate.NewLimiter(rate.Every(500*time.Millisecond), 7)

		client := chat.NewClient(conn, isAuthorized, int(userID), username)
		hub.Mu.Lock()
		hub.Clients[client] = true
		hub.Mu.Unlock()

		go client.WritePump()

		defer func() {
			hub.Mu.Lock()
			delete(hub.Clients, client)
			hub.Mu.Unlock()
			s.logger.Infof("WebSocket client disconnected: %s (ID: %d)", username, userID)
		}()

		s.logger.Infof("WebSocket client connected: %s (ID: %d, Authorized: %t)", username, userID, isAuthorized)

		sendToClient := func(v any) {
			payload, _ := json.Marshal(v)
			select {
			case client.Send <- payload:

			case <-time.After(time.Millisecond * 50):
				s.logger.Warnf("Client %s is slow to receive messages, dropping message", client.Name)
			}
		}

		historyMessages, err := hub.Store.Messages().LoadMessagesWithAttachments(time.Now().UTC(), 30)
		if err == nil {
			for i, j := 0, len(historyMessages)-1; i < j; i, j = i+1, j-1 {
				historyMessages[i], historyMessages[j] = historyMessages[j], historyMessages[i]
			}
			for _, msg := range historyMessages {
				sendToClient(msg)
			}
		} else {
			s.logger.Errorf("Failed to load message history for client %s: %v", username, err)
		}

		for {
			_, rawMessage, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					s.logger.Errorf("WebSocket unexpected close error for client %s: %v", username, err)
				}
				return
			}

			if !limiter.Allow() {
				s.logger.Infof("RateLimit exeeded for user: %d", client.UserID)
				sendToClient(map[string]string{
					"type":    "error",
					"message": "Too many requests. Slow Down!",
				})
			}

			if !isAuthorized {
				sendToClient(map[string]string{
					"type":    "error",
					"message": "Authentication required to send messages",
				})
				continue
			}

			var incoming incomingMessage
			if err := json.Unmarshal(rawMessage, &incoming); err != nil {
				s.logger.Warnf("Failed to unmarshal message from client %s: %v", username, err)
				sendToClient(map[string]string{
					"type":    "error",
					"message": "Invalid message format",
				})
				continue
			}

			msg := models.Message{
				Username: username,
				Content:  incoming.Text,
			}

			if len(incoming.Files) > 0 {
				msg.Attachments = make([]models.Attachment, len(incoming.Files))
				for i, f := range incoming.Files {
					msg.Attachments[i] = models.Attachment{
						FilePath: f.FilePath,
						FileType: f.FileType,
						FileSize: f.FileSize,
					}
				}
			}

			if err := hub.Store.Messages().SaveMessageWithAttachments(&msg); err != nil {
				s.logger.Errorf("Failed to save message from client %s: %v", username, err)
				sendToClient(map[string]string{
					"type":    "error",
					"message": "Failed to save message",
				})
				continue
			}

			hub.Broadcast <- msg

			s.logger.Infof("Message broadcasted from user %s (ID: %d): %s", username, userID, incoming.Text)
		}
	}
}

func (s *server) sendWebSocketError(conn *websocket.Conn, err error) {
	errMessage := map[string]string{"error": err.Error()}
	if writeErr := conn.WriteJSON(errMessage); writeErr != nil {
		if strings.Contains(writeErr.Error(), "close sent") {
			return
		}
		s.logger.Errorf("Failed to send WebSocket error message: %v", writeErr)
	}
}

func (s *server) HandleHistory(store models.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		beforeStr := r.URL.Query().Get("before")
		if beforeStr == "" {
			s.error(w, r, http.StatusBadRequest, fmt.Errorf("'before' parameter is required"))
			return
		}

		beforeTime, err := time.Parse(time.RFC3339, beforeStr)
		if err != nil {
			s.error(w, r, http.StatusBadRequest, fmt.Errorf("invalid 'before' timestamp format, expected RFC3339"))
			return
		}

		messages, err := store.Messages().LoadMessages(beforeTime, 30)
		if err != nil {
			s.logger.Errorf("Failed to load message history: %v", err)
			s.error(w, r, http.StatusInternalServerError, fmt.Errorf("failed to load message history"))
			return
		}

		s.respond(w, r, http.StatusOK, messages)
	}
}
