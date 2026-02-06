package e2e_test

import (
	"TextMeByte/internal/logger"
	"TextMeByte/internal/models"
	"TextMeByte/tests/e2e"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_UploadAndSendMessageWithAttachments(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	logger.InitLogger()

	testServer, store, cleanup := e2e.SetupE2ETest(t)
	defer cleanup()

	serverURL := testServer.URL
	wsURL := "ws" + serverURL[4:]

	t.Log("Step 1: Creating test image file")
	testImageFilePath, originalSize := e2e.CreateTestImageFile(t)
	defer os.Remove(testImageFilePath)

	originalData, err := os.ReadFile(testImageFilePath)
	require.NoError(t, err)

	t.Log("Step 2: Uploading file to server")
	uploadResp := e2e.UploadFile(t, serverURL, testImageFilePath)

	filename, ok := uploadResp["url"].(string)
	require.True(t, ok, "Upload should return filename")

	filesize, ok := uploadResp["size"].(float64)
	require.True(t, ok, "Upload should return filesize")

	mimetype, ok := uploadResp["mimetype"].(string)
	require.True(t, ok, "Upload should return mimetype")

	assert.Equal(t, originalSize, int64(filesize))
	assert.Equal(t, "image/jpeg", mimetype)

	t.Log("Step 3: Logging in")
	tu := models.TestUser(t)
	token := e2e.LoginAndGetToken(t, serverURL, tu.Username, tu.Password)
	assert.NotEmpty(t, token)

	t.Log("Step 4: Connecting to WS")
	ws := e2e.ConnectWS(t, fmt.Sprintf("%s/ws?token=%s", wsURL, token))
	defer ws.Close()

	time.Sleep(100 * time.Millisecond)

	t.Log("Step 5: Sending message with attachment")
	msgToSend := map[string]any{
		"text": "Check out this img!",
		"files": []map[string]any{
			{
				"filepath": filename,
				"filesize": int64(filesize),
				"filetype": mimetype,
			},
		},
	}

	time.Sleep(500 * time.Millisecond)
	e2e.SendWSMessage(t, ws, msgToSend)

	t.Log("Step 6: Receiving broadcast message")
	received := e2e.ReceiveWSMessage(t, ws, 5*time.Second)

	assert.Equal(t, tu.Username, received["username"])
	assert.Equal(t, "Check out this img!", received["text"])

	attachments, ok := received["attachments"].([]any)
	require.True(t, ok)
	require.Len(t, attachments, 1, "Should have 1 attachment")

	att := attachments[0].(map[string]any)
	assert.Equal(t, filename, att["filepath"])
	assert.Equal(t, filesize, att["filesize"])
	assert.Equal(t, mimetype, att["filetype"])

	t.Log("Step 7: Verifying file in database")
	messages, err := store.Messages().LoadMessagesWithAttachments(time.Now().UTC().Add(time.Hour), 10)
	require.NoError(t, err)
	require.NotEmpty(t, messages)

	require.NotNil(t, messages[0], "Should be our message")
	require.Len(t, messages[0].Attachments, 1)
	assert.Equal(t, filename, messages[0].Attachments[0].FilePath)
	assert.Equal(t, originalSize, messages[0].Attachments[0].FileSize)

	t.Log("Step 8: Downloading file and compairing")
	downloadURL := fmt.Sprintf("%s/download?file=%s", serverURL, filepath.Base(filename))
	downloadedData := e2e.DownloadFile(t, downloadURL)

	assert.Equal(t, downloadedData, originalData)

	t.Log("E2E TEST COMPLETED SUCESSFULLY!")
}
