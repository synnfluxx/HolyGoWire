package server

import (
	"TextMeByte/internal/chat"
	"TextMeByte/internal/logger"
	"TextMeByte/internal/storage"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

const (
	ctxKeyUser        ctxKey = iota 
	ctxKeyRequestedID               
)

type ctxKey int8

type server struct {
	router *mux.Router       
	store  *storage.Storage  
	logger *logrus.Logger    
}

func NewServer(store *storage.Storage, hub *chat.Hub) *server {
	s := &server{
		router: mux.NewRouter(),
		logger: logger.Log,
		store:  store,
	}

	s.ConfigureRouter(hub)

	return s
}

func (s *server) ConfigureRouter(hub *chat.Hub) {
	s.router.Use(s.setRequestID)
	s.router.Use(s.logRequest)
	s.router.Use(handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),                                      
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}), 
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),          
	))

	
	s.router.Handle("/registration", s.handleUserCreate())           
	s.router.Handle("/login", s.handleSessionCreate())               
	s.router.Handle("/ws", s.HandleWS(hub))                         
	s.router.Handle("/history", s.HandleHistory(s.store))           
	s.router.Handle("/me", s.authMiddleware(s.handleMe()))          
	s.router.Handle("/upload", s.UploadHandler())                   
	s.router.Handle("/download", s.DownloadHandler())               
}


func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}


func (s *server) setRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyRequestedID, id)))
	})
}


func (s *server) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := logrus.WithFields(logrus.Fields{
			"remote_addr": r.RemoteAddr,
			"request_id":  r.Context().Value(ctxKeyRequestedID),
		})
		logger.Infof("Started %s %s", r.Method, r.RequestURI)

		start := time.Now()
		rw := &responseWriter{w, http.StatusOK}
		next.ServeHTTP(rw, r)
	
		var level logrus.Level
		switch {
		case rw.code >= 500:
			level = logrus.ErrorLevel
		case rw.code >= 400:
			level = logrus.WarnLevel
		default:
			level = logrus.InfoLevel
		}

		logger.Logf(
			level,
			"Completed with %d %s in %v",
			rw.code,
			http.StatusText(rw.code),
			time.Since(start),
		)
	})
}

func (s *server) respond(w http.ResponseWriter, r *http.Request, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (s *server) error(w http.ResponseWriter, r *http.Request, statusCode int, err error) {
	s.respond(w, r, statusCode, map[string]string{"error": err.Error()})
}
