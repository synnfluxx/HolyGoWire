package storage_test

import (
	"TextMeByte/internal/models"
	"TextMeByte/internal/storage"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
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
	if err != nil {
		t.Fatal(err)
	}

	defer db.Close()

	s, err := storage.NewStore(db)
	if err != nil {
		t.Fatal(err)
	}

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
	if err != nil {
		t.Fatal(err)
	}
	s, err := storage.NewStore(db)
	if err != nil {
		t.Fatal(err)
	}

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
