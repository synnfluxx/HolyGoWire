package server_test

import (
	"TextMeByte/internal/models"
	"TextMeByte/internal/server"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuth_Registration(t *testing.T) {
	testCases := []struct {
		name         string
		payload      map[string]string
		expectedCode int
	}{
		{
			name: "created",
			payload: map[string]string{
				"username": "JohnDoe1212",
				"password": "QG=8?rQ8v38*",
			},
			expectedCode: http.StatusCreated,
		},
		{
			name: "without payload",
			payload: map[string]string{
				"username": "",
				"password": "",
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	_, s, _, _, _, ur, srv := server.TestMocks(t)

	s.On("User").Return(ur)
	ur.On("Create", mock.Anything).Return(nil)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b := &bytes.Buffer{}
			json.NewEncoder(b).Encode(tc.payload)
			rec := httptest.NewRecorder()
			req, err := http.NewRequest("POST", "/registration", b)
			require.NoError(t, err)
			srv.ServeHTTP(rec, req)
			assert.Equal(t, tc.expectedCode, rec.Code)
		})
	}
}

func TestAuth_Authorization(t *testing.T) {
	_, s, _, _, _, ur, srv := server.TestMocks(t)
	tu := models.TestUser(t)
	pwRaw := tu.Password
	tu.BeforeCreate()
	tu.Password = pwRaw

	s.On("User").Return(ur)
	ur.On("Create", tu).Return(nil)
	ur.On("FindByUsername", tu.Username).Return(tu, nil)
	ur.On("FindByUsername", "nonExistingUsername").Return(nil, fmt.Errorf("user with username '%s' not found", "nonExistingUsername"))

	_ = s.User().Create(tu)

	testCases := []struct {
		name         string
		payload      map[string]string
		expectedCode int
	}{
		{
			name: "authorized",
			payload: map[string]string{
				"username": tu.Username,
				"password": tu.Password,
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "without payload",
			payload: map[string]string{
				"username": "",
				"password": "",
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "unauthorized",
			payload: map[string]string{
				"username": "nonExistingUsername",
				"password": "password",
			},
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b := &bytes.Buffer{}
			json.NewEncoder(b).Encode(tc.payload)
			rec := httptest.NewRecorder()
			req, err := http.NewRequest("POST", "/login", b)
			require.NoError(t, err)
			srv.ServeHTTP(rec, req)
			assert.Equal(t, tc.expectedCode, rec.Code)
		})
	}

}
