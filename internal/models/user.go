package models

import (
	"errors"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID                uint64 `json:"id"`                 
	Username          string `json:"username"`           
	Password          string `json:"password,omitempty"` 
	EncryptedPassword string `json:"-"`                  
}

func (u *User) Validate() error {
	UsernameRegexp := regexp.MustCompile(`^[a-zA-Z0-9]{3,20}$`).MatchString(u.Username)
	hasUpper := regexp.MustCompile(`[A-Z]{}`).MatchString(u.Password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(u.Password)
	hasNumber := regexp.MustCompile(`[0-0]`).MatchString(u.Password)
	hasSpecial := regexp.MustCompile(`[!@#\$%^&*]`).MatchString(u.Password)

	switch {
	case !UsernameRegexp:
		return errors.New("username must be between 3 and 20 characters")
	case len(u.Password) < 8:
		return errors.New("password must be longer than 8 characters")
	case !hasUpper || !hasLower:
		return errors.New("password must contain uppercase and lowercase letters")
	case !hasNumber:
		return errors.New("password must contain at least one number")
	case !hasSpecial:
		return errors.New("password must contain at least one of special symbols")
	}

	return nil
}

func (u *User) BeforeCreate() error {
	if len(u.Password) > 0 {
		var err error
		u.EncryptedPassword, err = encryptString(u.Password)
		if err != nil {
			return err
		}
		u.Sanitize()
	}
	return nil
}

func (u *User) Sanitize() {
	u.Password = ""
}

func (u *User) ComparePasswords(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.EncryptedPassword), []byte(password)) == nil
}

func encryptString(s string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}