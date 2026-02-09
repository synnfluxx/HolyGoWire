package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Message struct {
	ID          int          `json:"id" db:"id"`
	IP          string       `json:"-" db:"ip"`
	HashedIP    string       `json:"-" db:"hashed_ip"`
	Username    string       `json:"username" db:"username"`
	Content     string       `json:"text" db:"content"`
	SendAt      time.Time    `json:"created_at" db:"send_at"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

func (m *Message) SetHashedIP() error {
	bcryptHash, err := bcrypt.GenerateFromPassword([]byte(m.IP), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	m.HashedIP = string(bcryptHash)
	m.IP = ""

	return nil
}
