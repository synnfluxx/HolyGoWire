package e2e

import (
	"TextMeByte/internal/chat"
	"TextMeByte/internal/models"
	"TextMeByte/internal/server"
	"TextMeByte/internal/storage"
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func SetupE2ETest(t *testing.T) (*httptest.Server, models.Storage, func()) {
	testDB := "host=localhost dbname=wschat_dev sslmode=disable" // Hardcoded for test in prod mode
	tmpDir := t.TempDir()
	os.Setenv("UPLOADS_DIR", tmpDir)

	store, err := storage.NewDB(testDB)
	require.NoError(t, err, "Failed to connect to test database")

	CleanupDatabase(t, store)

	err = store.User().Create(models.TestUser(t))
	require.NoError(t, err)

	hub := chat.NewHub(store)
	go hub.Run()

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	srv := server.NewServer(store, hub, logger)
	testServer := httptest.NewServer(srv)

	cleanup := func() {
		testServer.Close()
		CleanupDatabase(t, store)
		os.Clearenv()
	}

	return testServer, store, cleanup
}

func CleanupDatabase(t *testing.T, store *storage.Storage) {
	_, err := store.DB.Exec("DELETE FROM attachments")
	require.NoError(t, err)
	_, err = store.DB.Exec("DELETE FROM messages")
	require.NoError(t, err)
	_, err = store.DB.Exec("DELETE FROM users")
	require.NoError(t, err)
}

func CreateTestImageFile(t *testing.T) (path string, size int64) {
	tmpFile, err := os.CreateTemp("", "test-image-*.jpg")
	require.NoError(t, err)

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}

	err = jpeg.Encode(tmpFile, img, &jpeg.Options{Quality: 90})
	require.NoError(t, err)

	stat, _ := tmpFile.Stat()
	tmpFile.Close()

	return tmpFile.Name(), stat.Size()
}

func UploadFile(t *testing.T, serverURL, filePath string) map[string]any {
	file, err := os.Open(filePath)
	require.NoError(t, err)
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	require.NoError(t, err)

	_, err = io.Copy(part, file)
	require.NoError(t, err)

	writer.Close()

	req, err := http.NewRequest("POST", serverURL+"/upload", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	return result
}

func LoginAndGetToken(t *testing.T, serverURL, username, password string) string {
	loginData := map[string]string{
		"username": username,
		"password": password,
	}

	jsonData, _ := json.Marshal(loginData)

	resp, err := http.Post(serverURL+"/login", "application/json", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	token, ok := result["token"].(string)
	require.True(t, ok, "Token should be string")

	return token
}

func ConnectWS(t *testing.T, wsURL string) *websocket.Conn {
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	return ws
}

func SendWSMessage(t *testing.T, ws *websocket.Conn, msg any) {
	var mu sync.Mutex

	mu.Lock()
	err := ws.WriteJSON(msg)
	mu.Unlock()
	require.NoError(t, err)
}

func ReceiveWSMessage(t *testing.T, ws *websocket.Conn, timeout time.Duration) map[string]any {
	ws.SetReadDeadline(time.Now().Add(timeout))

	var result map[string]any
	err := ws.ReadJSON(&result)
	require.NoError(t, err)

	return result
}

func DownloadFile(t *testing.T, url string) []byte {
	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return data
}
