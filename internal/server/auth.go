package server

import (
	"TextMeByte/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/time/rate"
)

var jwtKey = []byte(os.Getenv("SECRET_KEY"))

type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {

		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return jwtKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	return claims, nil
}

func generateToken(userID int64, username string) (string, error) {
	expiresAt := time.Now().Add(10000 * time.Hour)

	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "TextMeByte",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}

	return tokenString, nil
}

func (s *Server) getVisitorLimiter(ip string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.limiters[ip]
	if !ok {
		limiter := rate.NewLimiter(rate.Every(1*time.Minute), 5)
		s.limiters[ip] = limiter
		return limiter
	}

	return v
}

func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			s.error(w, r, http.StatusUnauthorized, fmt.Errorf("authorization header required"))
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader {
			s.error(w, r, http.StatusUnauthorized, fmt.Errorf("invalid authorization header format"))
			return
		}

		claims, err := ValidateToken(tokenStr)
		if err != nil {
			s.error(w, r, http.StatusUnauthorized, fmt.Errorf("invalid token: %v", err))
			return
		}

		ctx := context.WithValue(r.Context(), ctxKeyUser, claims.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (s *Server) handleUserCreate() http.HandlerFunc {
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, fmt.Errorf("invalid JSON payload"))
			return
		}

		if req.Username == "" || req.Password == "" {
			s.error(w, r, http.StatusBadRequest, fmt.Errorf("username and password are required"))
			return
		}

		u := &models.User{
			Username: req.Username,
			Password: req.Password,
		}

		if err := s.store.User().Create(u); err != nil {
			if strings.Contains(err.Error(), "already exists") {
				s.error(w, r, http.StatusConflict, fmt.Errorf("username already exists"))
				return
			}
			s.logger.Errorf("Failed to create user: %v", err)
			s.error(w, r, http.StatusUnprocessableEntity, fmt.Errorf("failed to create user"))
			return
		}

		token, err := generateToken(int64(u.ID), u.Username)
		if err != nil {
			s.logger.Errorf("Failed to generate token for user %s: %v", u.Username, err)
			s.error(w, r, http.StatusInternalServerError, fmt.Errorf("failed to generate authentication token"))
			return
		}

		s.respond(w, r, http.StatusCreated, map[string]any{
			"user":  u,
			"token": token,
		})
	}
}

func (s *Server) handleSessionCreate() http.HandlerFunc {
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, fmt.Errorf("invalid JSON payload"))
			return
		}

		if req.Username == "" || req.Password == "" {
			s.error(w, r, http.StatusBadRequest, fmt.Errorf("username and password are required"))
			return
		}

		u, err := s.store.User().FindByUsername(req.Username)
		if err != nil {
			s.error(w, r, http.StatusUnauthorized, fmt.Errorf("invalid username or password"))
			return
		}

		if !u.ComparePasswords(req.Password) {
			s.error(w, r, http.StatusUnauthorized, fmt.Errorf("invalid username or password"))
			return
		}

		token, err := generateToken(int64(u.ID), u.Username)
		if err != nil {
			s.logger.Errorf("Failed to generate token for user %s: %v", u.Username, err)
			s.error(w, r, http.StatusInternalServerError, fmt.Errorf("failed to generate authentication token"))
			return
		}

		s.respond(w, r, http.StatusOK, map[string]string{"token": token})
	}
}

func (s *Server) handleMe() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, ok := r.Context().Value(ctxKeyUser).(string)
		if !ok {
			s.error(w, r, http.StatusUnauthorized, fmt.Errorf("user not found in request context"))
			return
		}

		u, err := s.store.User().FindByUsername(username)
		if err != nil {
			s.logger.Errorf("Failed to find user %s: %v", username, err)
			s.error(w, r, http.StatusNotFound, fmt.Errorf("user not found"))
			return
		}

		s.respond(w, r, http.StatusOK, u)
	}
}
