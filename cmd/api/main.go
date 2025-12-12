package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/application/service"

	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/domain/plagiarism"

	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/infrastructure/persistence/postgres"
	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/infrastructure/storage/local"
	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/infrastructure/text"
	httplayer "github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/interfaces/http"
	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/pkg/config"
)

func main() {
	if err := godotenv.Load(".env.local"); err != nil {
		log.Println("Warning: .env.local file not found, relying on environment variables")
	}

	cfg := config.LoadConfig()

	dbConfig := postgres.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
		SSLMode:  cfg.DBSSLMode,
	}

	db, err := postgres.NewConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Connected to PostgreSQL")

	workRepo := postgres.NewWorkRepository(db)
	fileRepo := postgres.NewFileRepository(db)
	plagRepo := postgres.NewPlagiarismRepository(db)

	if err := os.MkdirAll(cfg.FileStoragePath, 0755); err != nil {
		log.Fatalf("Failed to create storage directory: %v", err)
	}
	fileStorage := local.NewLocalFileStorage(cfg.FileStoragePath)

	textExtractor := text.NewSimpleExtractor()

	detector := plagiarism.NewShingleDetector()

	submissionSvc := service.NewSubmissionService(
		workRepo,
		fileRepo,
		plagRepo,
		fileStorage,
		textExtractor,
		detector,
	)

	reportSvc := service.NewReportService(plagRepo, workRepo)

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	engine.Use(gin.Recovery())

	const maxFileSize = 50 * 1024 * 1024

	httplayer.SetupRoutes(
		engine,
		db,
		submissionSvc,
		reportSvc,
		int64(maxFileSize),
	)

	serverAddr := fmt.Sprintf("%s:%s", cfg.ServerHost, cfg.ServerPort)
	log.Printf("Server starting on %s", serverAddr)

	server := &http.Server{
		Addr:           serverAddr,
		Handler:        engine,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}
