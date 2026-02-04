package storage_test

import (
	"TextMeByte/internal/models"
	"TextMeByte/internal/storage"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestMessages_SaveMessage(t *testing.T) {
	testCases := []struct {
		name      string
		msg       *models.Message
		fixedTime time.Time
		isValid   bool
	}{
		{
			name:      "valid test",
			msg:       &models.Message{Username: "anon", Content: "message like real message"},
			fixedTime: time.Now().UTC(),
			isValid:   true,
		},
		{
			name:      "without username",
			msg:       &models.Message{Username: "", Content: "message like real message"},
			fixedTime: time.Now().UTC(),
			isValid:   false,
		},
		{
			name:      "without message",
			msg:       &models.Message{Username: "anon", Content: ""},
			fixedTime: time.Now().UTC(),
			isValid:   false,
		},
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	defer db.Close()

	s, _ := storage.NewStore(db)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.isValid {
				rows := sqlmock.NewRows([]string{"id", "send_at"}).AddRow(1, tc.fixedTime)

				mock.ExpectQuery("INSERT INTO messages").
					WithArgs(tc.msg.Username, tc.msg.Content).
					WillReturnRows(rows)

				err := s.Messages().SaveMessage(tc.msg)
				assert.NoError(t, err)
				assert.Equal(t, tc.fixedTime, tc.msg.SendAt)
			} else {
				err := s.Messages().SaveMessageWithAttachments(tc.msg)
				assert.Error(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestMessages_LoadMessages(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	
	s, _ := storage.NewStore(db)
	query := "SELECT id, username, content, send_at FROM messages"
	
	t.Run("succes load", func(t *testing.T) {
		now := time.Now().UTC()
		
		rows := sqlmock.NewRows([]string{"id", "username", "content", "send_at"}).
			AddRow(1, "user1", "test msg", now.Add(-time.Hour)).
			AddRow(2, "user2", "another test msg", now.Add(-time.Minute*75)).
			AddRow(3, "user3", "another another msg", now.Add(-time.Hour*2)).
			AddRow(4, "user4", "msg", now.Add(-time.Hour*3)).
			AddRow(5, "user5", "msg", now.Add(-time.Hour*3))
		
		mock.ExpectQuery(query).
			WithArgs(now, 3).
			WillReturnRows(rows)
		
		msgs, err := s.Messages().LoadMessages(now, 3)
		
		
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.NotNil(t, msgs)
		assert.Equal(t, "user1", msgs[0].Username)
	})
	
	t.Run("no output", func(t *testing.T) {
		now := time.Now().UTC()
		
		rows := sqlmock.NewRows([]string{"username", "content", "send_at"})
		
		mock.ExpectQuery(query).
			WithArgs(now, 10).
			WillReturnRows(rows)
		
		msgs, err := s.Messages().LoadMessages(now, 10)
		
		assert.NoError(t, err)
		assert.Empty(t, msgs)
	})
}
