package main

import (
	"fmt"
	"os"

	"klinik-app/internal/config"
	"klinik-app/internal/database"
	"klinik-app/internal/logger"
	"klinik-app/internal/server"
)

func main() {
	logger.Init("info")
	logger.Info.Println("Starting Klinik Management System...")

	cfg, err := config.Load()
	if err != nil {
		logger.Error.Printf("Failed to load configuration: %v", err)
		os.Exit(1)
	}

	logger.Init(cfg.LogLevel)

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Error.Printf("Database connection failed: %v", err)
		logger.Info.Println("Server running WITHOUT database. Some features will not work.")
	} else {
		defer database.Close(db)
	}

	srv := server.New(cfg, db)

	fmt.Println("╔══════════════════════════════════════╗")
	fmt.Println("║   Sistem Manajemen Klinik v1.0.0    ║")
	fmt.Println("║   http://localhost:" + fmt.Sprintf("%d", cfg.AppPort) + "             ║")
	fmt.Println("╚══════════════════════════════════════╝")

	if err := srv.Start(); err != nil {
		logger.Error.Printf("Server error: %v", err)
		os.Exit(1)
	}
}
