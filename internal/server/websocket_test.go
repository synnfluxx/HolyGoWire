package server_test

import (
	"TextMeByte/internal/models"
	"TextMeByte/internal/server"
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func receiveMessage(t *testing.T, client *websocket.Conn, recieved *map[string]string) {
	_, raw, err := client.ReadMessage()
	require.NoError(t, err)
	json.Unmarshal(raw, recieved)
}

func TestServer_HandleWS(t *testing.T) {
	srv, s, _, _, mr, ur, _ := server.TestMocks(t)
	defer srv.Close()

	tu := models.TestUser(t)
	msg := models.TestMessage(t)
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"

	s.On("User").Return(ur)
	s.On("Messages").Return(mr)
	ur.On("Create", tu).Return(nil)
	mr.On("LoadMessagesWithAttachments", mock.AnythingOfType("time.Time"), 30).Return([]models.Message{*models.TestMessage(t)}, nil)
	mr.On("SaveMessageWithAttachments", mock.AnythingOfType("*models.Message")).Return(nil)

	req := map[string]string{
		"Username": tu.Username,
		"Password": tu.Password,
	}

	jreq, err := json.Marshal(req)
	require.NoError(t, err)

	resp, err := http.Post(srv.URL+"/registration", "application/json", bytes.NewBuffer(jreq))
	require.NoError(t, err)

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	require.NotNil(t, result["token"])

	client, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer client.Close()

	t.Run("check history loading with no auth", func(t *testing.T) {
		var received map[string]string
		receiveMessage(t, client, &received)

		assert.Equal(t, received["text"], msg.Content)
		assert.Equal(t, received["username"], msg.Username)
	})

	t.Run("test sending with no auth", func(t *testing.T) {
		payload, err := json.Marshal(msg)
		require.NoError(t, err)
		err = client.WriteMessage(websocket.TextMessage, payload)
		require.NoError(t, err)

		_, resp, err := client.ReadMessage()
		require.NoError(t, err)

		var received map[string]string
		json.Unmarshal(resp, &received)

		assert.Equal(t, received["type"], "error")
		assert.Contains(t, received["message"], "Authentication required")
	})

	authClient, _, err := websocket.DefaultDialer.Dial(wsURL+"?token="+result["token"], nil)
	assert.NoError(t, err)
	defer authClient.Close()

	t.Run("check history loading with auth", func(t *testing.T) {
		var received map[string]string
		receiveMessage(t, authClient, &received)

		assert.Equal(t, received["text"], msg.Content)
		assert.Equal(t, received["username"], msg.Username)
	})

	t.Run("test sending with auth", func(t *testing.T) {
		payload, err := json.Marshal(msg)
		require.NoError(t, err)
		err = authClient.WriteMessage(websocket.TextMessage, payload)
		require.NoError(t, err)

		_, resp, err := authClient.ReadMessage()
		require.NoError(t, err)

		var received map[string]string
		json.Unmarshal(resp, &received)

		assert.Equal(t, received["text"], msg.Content)
		assert.Equal(t, received["username"], msg.Username)
	})
}
