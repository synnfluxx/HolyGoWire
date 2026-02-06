package server_test
/*package server_test

import (
	"TextMeByte/internal/chat"
	"TextMeByte/internal/server"
	"TextMeByte/internal/storage"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestServer_HandleWS(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	s, err := storage.NewStore(db)
	require.NoError(t, err)
	hub := new(mockHub)
	loggr := new(MockLogger)

	srv := server.NewServer(s, hub, loggr)
	srv.ConfigureRouter(hub)

}

type mockHub struct {
	mock.Mock
}

func (m *mockHub) Clients() map[*chat.Client]bool {
	args := m.Called()
	return args.Get(0).(map[*chat.Client]bool)
}

func (m *mockHub) Broadcast() chan<- any {
	args := m.Called()
	return args.Get(0).(chan<- any)
}

func (m *mockHub) Store() storage.Storage {
	args := m.Called()
	return args.Get(0).(storage.Storage)
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Infof(format string, args ...any) {
	m.Called(format, args)
}

func (m *MockLogger) Warnf(format string, args ...any) {
	m.Called(format, args)
}

func (m *MockLogger) Errorf(format string, args ...any) {
	m.Called(format, args)
}
*/
