package storage

import (
	"TextMeByte/internal/models"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type MessageRepository struct {
	store *Storage 
}

func (r *MessageRepository) SaveMessage(m *models.Message) error {
	if m.Username == "" || m.Content == "" {
		return fmt.Errorf("message validation failed: username and content are required")
	}
	
	query := "INSERT INTO messages (username, content) VALUES ($1, $2) RETURNING id, send_at"
	err := r.store.DB.QueryRow(query, m.Username, m.Content).Scan(&m.ID, &m.SendAt)
	if err != nil {
		return fmt.Errorf("failed to save message: %v", err)
	}
	
	return nil
}

func (r *MessageRepository) SaveMessageWithAttachments(m *models.Message) error {
	tx, err := r.store.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback() 
	
	err = tx.QueryRow(
		"INSERT INTO messages (username, content) VALUES ($1, $2) RETURNING id, send_at", 
		m.Username, m.Content,
	).Scan(&m.ID, &m.SendAt)
	if err != nil {
		return fmt.Errorf("failed to insert message: %v", err)
	}
	
	if len(m.Attachments) > 0 {
		for _, att := range m.Attachments {
			_, err := tx.Exec(
				"INSERT INTO attachments (message_id, file_path, file_type, file_size) VALUES ($1, $2, $3, $4)", 
				m.ID, att.FilePath, att.FileType, att.FileSize,
			)
			if err != nil {
				return fmt.Errorf("failed to insert attachment: %v", err)
			}
		}
	}
	
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	
	return nil
}

func (r *MessageRepository) LoadMessages(before time.Time, limit int) ([]models.Message, error) {
	query := "SELECT id, username, content, send_at FROM messages WHERE send_at < $1 ORDER BY send_at DESC LIMIT $2"
	rows, err := r.store.DB.Query(query, before, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %v", err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		if err := rows.Scan(&msg.ID, &msg.Username, &msg.Content, &msg.SendAt); err != nil {
			return nil, fmt.Errorf("failed to scan message row: %v", err)
		}
		messages = append(messages, msg)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating message rows: %v", err)
	}

	return messages, nil
}



func (r *MessageRepository) LoadMessagesWithAttachments(before time.Time, limit int) ([]models.Message, error) {
	messages, err := r.LoadMessages(before, limit)
	if err != nil {
		return nil, err
	}
	
	if len(messages) == 0 {
		return nil, nil
	}
	
	messagesID := make([]int, len(messages))
	for i, msg := range messages {
		messagesID[i] = msg.ID
	}
	
	query := `SELECT id, message_id, file_path, file_type, file_size FROM attachments WHERE message_id = ANY($1)`
	rows, err := r.store.DB.Query(query, pq.Array(messagesID))
	if err != nil {
		return nil, fmt.Errorf("failed to query attachments: %v", err)
	}
	defer rows.Close()
	
	attByID := make(map[int64][]models.Attachment)
	for rows.Next() {
		var att models.Attachment
		if err := rows.Scan(&att.ID, &att.MessageID, &att.FilePath, &att.FileType, &att.FileSize); err != nil {
			return nil, fmt.Errorf("failed to scan attachment row: %v", err)
		}
		attByID[int64(att.MessageID)] = append(attByID[int64(att.MessageID)], att)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating attachment rows: %v", err)
	}
	
	for i := range messages {
		messages[i].Attachments = attByID[int64(messages[i].ID)]
	}
	
	return messages, nil
}
