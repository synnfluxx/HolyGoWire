package models

import (
	"testing"
	"time"
)

func TestUser(t *testing.T) *User {
	t.Helper()

	return &User{
		Username: "JohnDoe1337",
		Password: "QG=8?rQ8v38*",
	}
}

func TestMessage(t *testing.T) *Message {
	t.Helper()

	return &Message{
		ID: 0,
		Username: "JohnDoe1337",
		Content: "text message",
		SendAt: time.Now().UTC(),
		Attachments: []Attachment{},
	}
}
