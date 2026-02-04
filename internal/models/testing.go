package models

import "testing"

func TestUser(t *testing.T) *User {
	t.Helper()

	return &User{
		Username: "testuser",   
		Password: "JR3|DQjQ76", 
	}
}
