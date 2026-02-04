package storage_test

import (
	"TextMeByte/internal/storage"
	"fmt"
	"os"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
)

var (
	databaseURL string
)

func TestMain(m *testing.M) {
	data := make(map[string]any)

	toml.DecodeFile("../../config/config.toml", &data)

	databaseURL = fmt.Sprintf("%v", data["database_url"])

	os.Exit(m.Run())
}

func TestStorage(t *testing.T) {
	s, err := storage.NewDB(databaseURL) // maybe add new testcases
	assert.NotNil(t, s)
	assert.NoError(t, err)
}
