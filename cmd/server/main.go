package main

import (
	"TextMeByte/config"
	"TextMeByte/internal/chat"
	"TextMeByte/internal/logger"
	"TextMeByte/internal/server"
	"TextMeByte/internal/storage"
	"net/http"
	"os"
)

var (
	configPath                       = "config/config.toml"
	BindAddr, DatabaseURL, SecretKey string
)

func main() {
	logger.InitLogger()
	cfg := config.NewConfig()
	config.ParseConfig(cfg, configPath)

	BindAddr = os.Getenv("BIND_ADDR")
	DatabaseURL = os.Getenv("DATABASE_URL")

	store, err := storage.NewDB(DatabaseURL)
	if err != nil {
		logger.Log.Fatalf("Failed to connect to database: %v", err)
	}

	hub := chat.NewHub(store)
	go hub.Run()

	srv := server.NewServer(store, hub)

	logger.Log.Infof("HolyGoWire server started successfully on %s", BindAddr)
	logger.Log.Fatalf("Server failed to start: %v", http.ListenAndServe(BindAddr, srv))
}
