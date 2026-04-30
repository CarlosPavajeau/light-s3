package main

import (
	"log"

	"light-s3/internal/awsclient"
	"light-s3/internal/config"
	"light-s3/internal/handlers"
	"light-s3/internal/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	presigner, err := awsclient.NewPresignClient(cfg)
	if err != nil {
		log.Fatalf("aws client: %v", err)
	}

	h := handlers.New(presigner, cfg.BucketName, cfg.PresignExpiry)

	r := gin.Default()
	r.Use(middleware.APIKey(cfg.APIKey))

	r.POST("/presign/upload", h.UploadURL)
	r.POST("/presign/download", h.DownloadURL)

	if err := r.Run(cfg.Port); err != nil {
		log.Fatalf("server: %v", err)
	}
}
