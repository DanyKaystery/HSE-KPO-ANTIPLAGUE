package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/domain/file"
	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/infrastructure/persistence/postgres"
	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/infrastructure/storage/local"
	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/pkg/config"
)

func main() {
	godotenv.Load(".env.local")
	cfg := config.LoadConfig()

	dbCfg := postgres.Config{
		Host: cfg.DBHost, Port: cfg.DBPort, User: cfg.DBUser,
		Password: cfg.DBPassword, DBName: cfg.DBName, SSLMode: cfg.DBSSLMode,
	}
	db, err := postgres.NewConnection(dbCfg)
	if err != nil {
		log.Fatalf("Storage Service: DB Connection failed: %v", err)
	}

	fileRepo := postgres.NewFileRepository(db)

	if err := os.MkdirAll(cfg.FileStoragePath, 0755); err != nil {
		log.Fatalf("Storage Service: Failed to create storage dir: %v", err)
	}
	fileStorage := local.NewLocalFileStorage(cfg.FileStoragePath)

	r := gin.Default()

	r.POST("/internal/upload", func(c *gin.Context) {
		uploadHandler(c, fileRepo, fileStorage)
	})

	r.GET("/internal/files/:file_id/content", func(c *gin.Context) {
		downloadHandler(c, fileRepo, fileStorage)
	})

	port := ":9091"
	log.Printf("💾 Storage Service running on %s", port)
	r.Run(port)
}

func uploadHandler(c *gin.Context, repo file.Repository, storage file.Storage) {
	header, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file"})
		return
	}

	f, err := header.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer f.Close()

	tempFileEntity := file.NewFile(header.Filename, "", header.Header.Get("Content-Type"), "", header.Size)

	path, err := storage.Upload(c.Request.Context(), tempFileEntity.ID, f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload to disk"})
		return
	}

	tempFileEntity.StoragePath = path

	if err := repo.Save(c.Request.Context(), tempFileEntity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save metadata"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"file_id": tempFileEntity.ID,
		"path":    path,
	})
}

func downloadHandler(c *gin.Context, repo file.Repository, storage file.Storage) {
	idStr := c.Param("file_id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID"})
		return
	}

	fileMeta, err := repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File metadata not found"})
		return
	}

	content, err := storage.Download(c.Request.Context(), fileMeta.StoragePath)
	if err != nil {
		log.Printf("Failed to read file from disk: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "File content not found"})
		return
	}
	defer content.Close()

	c.Header("Content-Type", fileMeta.MimeType)
	c.Header("Content-Disposition", "attachment; filename="+fileMeta.OriginalName)

	if _, err := io.Copy(c.Writer, content); err != nil {
		log.Printf("Failed to stream file: %v", err)
	}
}
