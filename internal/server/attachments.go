package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (s *server) UploadHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		
		if err := r.ParseMultipartForm(50 << 20); err != nil {
			s.logger.Errorf("Failed to parse multipart form: %v", err)
			s.error(w, r, http.StatusBadRequest, fmt.Errorf("invalid multipart form data"))
			return
		}
		
		file, handler, err := r.FormFile("file")
		if err != nil {
			s.logger.Errorf("Failed to get file from form: %v", err)
			s.error(w, r, http.StatusBadRequest, fmt.Errorf("no file provided"))
			return
		}
		defer file.Close()

		s.logger.Infof("File upload attempt: %s, Size: %d bytes, Declared MIME: %s",
			handler.Filename, handler.Size, handler.Header.Get("Content-Type"))

		buffer := make([]byte, 512)
		_, err = file.Read(buffer)
		if err != nil && err != io.EOF {
			s.logger.Errorf("Failed to read file content for type detection: %v", err)
			s.error(w, r, http.StatusInternalServerError, fmt.Errorf("failed to process file"))
			return
		}
		
		if _, err := file.Seek(0, 0); err != nil {
			s.logger.Errorf("Failed to reset file pointer: %v", err)
			s.error(w, r, http.StatusInternalServerError, fmt.Errorf("failed to process file"))
			return
		}

		detectedType := http.DetectContentType(buffer)
		s.logger.Infof("Detected file type: %s", detectedType)

		uploadsDir := "./uploads"
		if err := os.MkdirAll(uploadsDir, 0755); err != nil {
			s.logger.Errorf("Failed to create uploads directory: %v", err)
			s.error(w, r, http.StatusInternalServerError, fmt.Errorf("failed to prepare upload directory"))
			return
		}

		uniqueName := fmt.Sprintf("%d-%s", time.Now().Unix(), filepath.Base(handler.Filename))
		filePath := filepath.Join(uploadsDir, uniqueName)

		dst, err := os.Create(filePath)
		if err != nil {
			s.logger.Errorf("Failed to create destination file %s: %v", filePath, err)
			s.error(w, r, http.StatusInternalServerError, fmt.Errorf("failed to save file"))
			return
		}
		defer dst.Close()
		
		bytesWritten, err := io.Copy(dst, file)
		if err != nil {
			s.logger.Errorf("Failed to copy file content: %v", err)
			s.error(w, r, http.StatusInternalServerError, fmt.Errorf("failed to save file"))
			return
		}

		s.logger.Infof("File uploaded successfully: %s (%d bytes)", uniqueName, bytesWritten)

		s.respond(w, r, http.StatusOK, map[string]any{
			"url":      filepath.Join("uploads", uniqueName), 
			"size":     handler.Size,
			"mimetype": detectedType,
			"filename": uniqueName,
		})
	}
}


func (s *server) DownloadHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filename := r.URL.Query().Get("file")
		if filename == "" {
			s.error(w, r, http.StatusBadRequest, fmt.Errorf("file parameter is required"))
			return
		}

		if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
			s.logger.Warnf("Potential path traversal attempt: %s", filename)
			s.error(w, r, http.StatusBadRequest, fmt.Errorf("invalid filename"))
			return
		}

		filePath := filepath.Join("./uploads", filename)

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			s.logger.Warnf("Requested file not found: %s", filename)
			s.error(w, r, http.StatusNotFound, fmt.Errorf("file not found"))
			return
		}

		s.logger.Infof("Serving file: %s", filename)
		
		http.ServeFile(w, r, filePath)
	}
}
