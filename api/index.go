package main

import (
	"net/http"
	"sync"

	"klinik-app/internal/config"
	"klinik-app/internal/database"
	"klinik-app/internal/logger"
	"klinik-app/internal/server"
)

var (
	initOnce sync.Once
	handler  http.Handler
	initErr  error
)

func initHandler() {
	logger.Init("info")

	cfg, err := config.Load()
	if err != nil {
		initErr = err
		return
	}
	logger.Init(cfg.LogLevel)

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Error.Printf("Database connection failed: %v", err)
		db = nil
	}

	srv := server.New(cfg, db)
	handler = srv.Handler()
}

func Handler(w http.ResponseWriter, r *http.Request) {
	initOnce.Do(initHandler)
	if initErr != nil {
		http.Error(w, "Kesalahan konfigurasi server", http.StatusInternalServerError)
		return
	}
	if handler == nil {
		http.Error(w, "Server tidak tersedia", http.StatusInternalServerError)
		return
	}
	handler.ServeHTTP(w, r)
}

func main() {}
