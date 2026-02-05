package models

import (
	"errors"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

var (
	UsernameRegexp = regexp.MustCompile(`^[a-zA-Z0-9]{3,20}$`)
	//PasswordRegexp = regexp.MustCompile(`^(?=.*[A-Z])(?=.*[a-z])(?=.*[0-9])(?=.*[!@#$%^&*]).{8,}$`)
	hasUpper       = regexp.MustCompile(`[A-Z]`)
	hasLower       = regexp.MustCompile(`[a-z]`)
	hasNumber      = regexp.MustCompile(`[0-9]`)
	//hasSpecial     = regexp.MustCompile(`[!@#\$%^&*]`)
)

type User struct {
	ID                uint64 `json:"id"`
	Username          string `json:"username"`
	Password          string `json:"password,omitempty"`
	EncryptedPassword string `json:"-"`
}

func (u *User) Validate() error {

	switch {
	case !UsernameRegexp.MatchString(u.Username):
		return errors.New("username must be between 3 and 20 characters")
	case len(u.Password) < 8:
		return errors.New("password must be longer than 8 characters")
	case !hasUpper.MatchString(u.Password) || !hasLower.MatchString(u.Password):
		return errors.New("password must contain uppercase and lowercase letters")
	case !hasNumber.MatchString(u.Password):
		return errors.New("password must contain at least one number")
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
