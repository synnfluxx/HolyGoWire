package storage

import (
	"TextMeByte/internal/logger"
	"TextMeByte/internal/models"
	"database/sql"
	"fmt"

	"github.com/bits-and-blooms/bloom/v3"
)

type userRepository struct {
	store  *Storage
	filter *bloom.BloomFilter
}

func (r *userRepository) Hydrate() error {
	rows, err := r.store.DB.Query("SELECT username FROM users")
	if err != nil {
		return fmt.Errorf("failed to query usernames for bloom filter: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			return fmt.Errorf("failed to scan username during hydration: %v", err)
		}
		r.filter.AddString(username)
		count++
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows during hydration: %v", err)
	}

	logger.Log.Printf("Bloom filter hydrated with %d usernames", count)
	return nil
}

func (r *userRepository) Create(u *models.User) error {
	if r.filter.TestString(u.Username) {
		return fmt.Errorf("username '%s' already exists", u.Username)
	}

	if err := u.Validate(); err != nil {
		return fmt.Errorf("user validation failed: %v", err)
	}

	if err := u.BeforeCreate(); err != nil {
		return fmt.Errorf("failed to prepare user for creation: %v", err)
	}

	err := r.store.DB.QueryRow(
		"INSERT INTO users (username, encrypted_password) VALUES ($1, $2) RETURNING id",
		u.Username,
		u.EncryptedPassword,
	).Scan(&u.ID)

	if err != nil {
		return fmt.Errorf("failed to create user in database: %v", err)
	}

	r.filter.AddString(u.Username)
	return nil
}

func (r *userRepository) FindByUsername(username string) (*models.User, error) {
	u := &models.User{}

	if !r.filter.TestString(username) {
		return nil, fmt.Errorf("user with username '%s' not found", username)
	}

	if err := r.store.DB.QueryRow(
		"SELECT id, username, encrypted_password FROM users WHERE username = $1",
		username,
	).Scan(
		&u.ID, &u.Username, &u.EncryptedPassword,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with username '%s' not found", username)
		}
		return nil, fmt.Errorf("failed to query user by username: %v", err)
	}

	return u, nil
}

func (r *userRepository) FindByUserID(userID uint64) (*models.User, error) {
	u := &models.User{}

	if err := r.store.DB.QueryRow(
		"SELECT id, username, encrypted_password FROM users WHERE id = $1",
		userID,
	).Scan(
		&u.ID, &u.Username, &u.EncryptedPassword,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user wuth userID '%d' not found", userID)
		}
		return nil, fmt.Errorf("failed to query user by userID: %v", err)
	}

	return u, nil
}
