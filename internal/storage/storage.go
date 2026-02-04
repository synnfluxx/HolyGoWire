package storage

import (
	"database/sql"

	"github.com/bits-and-blooms/bloom/v3"
	_ "github.com/lib/pq" 
)

type Storage struct {
	DB                *sql.DB            
	UserRepository    *UserRepository    
	MessageRepository *MessageRepository 
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
	
	s.UserRepository = &UserRepository{
		store:  s,
		filter: bloom.New(1000*1000*20, 5), 
	}
	s.MessageRepository = &MessageRepository{store: s}
	
	s.UserRepository.Hydrate()

	return s, nil
}

func (s *Storage) User() *UserRepository {
	if s.UserRepository != nil {
		return s.UserRepository
	}

	s.UserRepository = &UserRepository{
		store:  s,
		filter: bloom.New(1000*1000*20, 5),
	}
	
	s.UserRepository.Hydrate()

	return s.UserRepository
}

func (s *Storage) Messages() *MessageRepository {
	if s.MessageRepository != nil {
		return s.MessageRepository
	}

	s.MessageRepository = &MessageRepository{
		store: s,
	}

	return s.MessageRepository
}