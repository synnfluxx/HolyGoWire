package storage_test

import (
	"TextMeByte/internal/models"
	"TextMeByte/internal/storage"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_Create(t *testing.T) {
	testCases := []struct {
		name    string
		u       *models.User
		isValid bool
	}{
		{
			name:    "uniq user",
			u:       &models.User{Username: "testuser", Password: "QG=8?rQ8v38*"},
			isValid: true,
		},
		{
			name:    "not uniq user",
			u:       &models.User{Username: "testuser", Password: "QG=8?rQ8v38*"},
			isValid: false,
		},
	}

	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	defer db.Close()

	s, err := storage.NewStore(db)
	require.NoError(t, err)

	rows := mock.NewRows([]string{"id"}).AddRow(1)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.isValid {
				mock.ExpectQuery("INSERT INTO users").
					WithArgs(tc.u.Username, sqlmock.AnyArg()).
					WillReturnRows(rows)

				err = s.User().Create(tc.u)

				assert.NoError(t, err)
				assert.NotNil(t, tc.u.ID)
				assert.NoError(t, mock.ExpectationsWereMet())
			} else {
				err := s.User().Create(tc.u)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "already exists")
			}
		})
	}

}

func TestUserRepository_FindByUsername(t *testing.T) {
	testUsers := []models.User{
		{ID: 1, Username: "alice", Password: "QG=8?rQ8v38*"},
	}

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	s, err := storage.NewStore(db)
	require.NoError(t, err)

	for _, tu := range testUsers {
		rows := sqlmock.NewRows([]string{"id"}).AddRow(tu.ID)

		mock.ExpectQuery("INSERT INTO users").
			WithArgs(tu.Username, sqlmock.AnyArg()).
			WillReturnRows(rows)

		err := s.User().Create(&tu)
		assert.NoError(t, err)
	}

	t.Run("search for existing user", func(t *testing.T) {
		expectedUser := testUsers[0]
		encPass := "veryENCpass"
		rows := sqlmock.NewRows([]string{"id", "username", "encryptedPassword"}).AddRow(expectedUser.ID, expectedUser.Username, encPass)

		mock.ExpectQuery("SELECT id, username, encrypted_password FROM users WHERE username =").
			WithArgs(expectedUser.Username).
			WillReturnRows(rows)

		u, err := s.User().FindByUsername(expectedUser.Username)
		assert.NoError(t, err)
		assert.NotNil(t, u)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("search for a non existing user", func(t *testing.T) {
		u, err := s.User().FindByUsername("lox")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.Nil(t, u)
	})
}

func TestUserRepository_FindByUserID(t *testing.T) {
	testusers := []*models.User{{ID: 1, Username: "Michael", Password: "QG=8?rQ8v38*"}, {ID: 2, Username: "Boba fett", Password: "QG=8?rQ8v38*"}}

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	store, err := storage.NewStore(db)
	require.NoError(t, err)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(testusers[0].ID)

	mock.ExpectQuery("INSERT INTO users").
		WithArgs(testusers[0].Username, sqlmock.AnyArg()).
		WillReturnRows(rows)

	err = store.User().Create(testusers[0])
	require.NoError(t, err)

	t.Run("search for existing user", func(t *testing.T) {
		expectedUser := testusers[0]
		rows := mock.NewRows([]string{"id", "username", "encrypted_password"}).
			AddRow(expectedUser.ID, expectedUser.Username, expectedUser.EncryptedPassword)

		mock.ExpectQuery("SELECT id, username, encrypted_password FROM users WHERE id =").
			WithArgs(expectedUser.ID).
			WillReturnRows(rows)

		u, err := store.User().FindByUserID(expectedUser.ID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.NotNil(t, u)
		assert.Equal(t, u.Username, expectedUser.Username)
		assert.Equal(t, u.Password, expectedUser.Password)
		assert.Equal(t, u.ID, expectedUser.ID)
	})

	t.Run("search for non-existing user", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, username, encrypted_password FROM users WHERE id =").
			WithArgs(testusers[1].ID). // ID = 2
			WillReturnError(sql.ErrNoRows)

		u, err := store.UserRepository.FindByUserID(testusers[1].ID)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Contains(t, err.Error(), "not found")
		assert.Nil(t, u)
	})
}
