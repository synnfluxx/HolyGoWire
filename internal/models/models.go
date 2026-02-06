package models

import "time"

type MessagesRepository interface {
	SaveMessage(m *Message) error
	SaveMessageWithAttachments(m *Message) error
	LoadMessages(before time.Time, limit int) ([]Message, error)
	LoadMessagesWithAttachments(before time.Time, limit int) ([]Message, error)
}

type UserRepository interface {
	Hydrate() error
	Create(u *User) (error)
	FindByUsername(username string) (*User, error)
	FindByUserID(userID uint64) (*User, error)
}


type Storage interface {
	User() UserRepository
	Messages() MessagesRepository
}

type Hub interface{}

type Logger interface{}
