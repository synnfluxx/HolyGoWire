package server

import (
	"TextMeByte/internal/chat"
	"TextMeByte/internal/models"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
)

type mockStorage struct {
	mock.Mock
}

func (s *mockStorage) User() models.UserRepository {
	args := s.Called()
	return args.Get(0).(models.UserRepository)
}

func (s *mockStorage) Messages() models.MessagesRepository {
	args := s.Called()
	return args.Get(0).(models.MessagesRepository)
}

type mockUserRepo struct {
	mock.Mock
}

func (r *mockUserRepo) Hydrate() error {
	args := r.Called()
	return args.Error(0)
}

func (r *mockUserRepo) Create(m *models.User) error {
	args := r.Called(m)
	return args.Error(0)
}

func (r *mockUserRepo) FindByUsername(username string) (*models.User, error) {
	args := r.Called(username)

	val := args.Get(0)
	if val == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.User), args.Error(1)
}

func (r *mockUserRepo) FindByUserID(userID uint64) (*models.User, error) {
	args := r.Called(userID)
	return args.Get(0).(*models.User), args.Error(1)
}

type mockMessageRepo struct {
	mock.Mock
}

func (r *mockMessageRepo) SaveMessage(m *models.Message) error {
	args := r.Called(m)
	return args.Error(0)
}

func (r *mockMessageRepo) SaveMessageWithAttachments(m *models.Message) error {
	args := r.Called(m)
	return args.Error(0)
}

func (r *mockMessageRepo) LoadMessages(before time.Time, limit int) ([]models.Message, error) {
	args := r.Called(before, limit)
	return args.Get(0).([]models.Message), args.Error(1)
}

func (r *mockMessageRepo) LoadMessagesWithAttachments(before time.Time, limit int) ([]models.Message, error) {
	args := r.Called(before, limit)
	return args.Get(0).([]models.Message), args.Error(1)
}

func TestMocks(t *testing.T) (*httptest.Server, *mockStorage, *chat.Hub, *logrus.Logger, *mockMessageRepo, *mockUserRepo, *Server) {
	t.Helper()
	s := new(mockStorage)
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	logger.SetLevel(logrus.PanicLevel)
	hub := chat.NewHub(s)
	go hub.Run()
	ur := new(mockUserRepo)
	mr := new(mockMessageRepo)

	srv := NewServer(s, hub, logger)
	srv.ConfigureRouter(hub)

	server := httptest.NewServer(srv.Router)

	return server, s, hub, logger, mr, ur, srv
}
