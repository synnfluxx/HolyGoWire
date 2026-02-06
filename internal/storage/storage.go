package storage

import (
	"TextMeByte/internal/models"
	"database/sql"

	"github.com/bits-and-blooms/bloom/v3"
	_ "github.com/lib/pq"
)

type Storage struct {
	DB                *sql.DB
	UserRepository    models.UserRepository
	MessageRepository models.MessagesRepository
}

func NewDB(dbURL string) (*Storage, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	return NewStore(db)
}

func NewStore(db *sql.DB) (*Storage, error) {
	s := &Storage{DB: db}

	s.UserRepository = &userRepository{
		store:  s,
		filter: bloom.New(1000*1000*20, 5),
	}
	s.MessageRepository = &messageRepository{store: s}

	s.UserRepository.Hydrate()

	return s, nil
}

func (s *Storage) User() models.UserRepository {
	if s.UserRepository != nil {
		return s.UserRepository
	}

	s.UserRepository = &userRepository{
		store:  s,
		filter: bloom.New(1000*1000*20, 5),
	}

	s.UserRepository.Hydrate()

	return s.UserRepository
}

func (s *Storage) Messages() models.MessagesRepository {
	if s.MessageRepository != nil {
		return s.MessageRepository
	}

	s.MessageRepository = &messageRepository{
		store: s,
	}

	return s.MessageRepository
}
