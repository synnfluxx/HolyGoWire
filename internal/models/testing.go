package models

import "testing"

func TestUser(t *testing.T) *User {
	t.Helper()

	return &User{
		Username: "testuser",
		Password: "QG=8?rQ8v38*",
	}
}
