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

	g := gin.Default()
	g.Use(middleware.APIKey(cfg.APIKey))

	g.POST("/presign/upload", h.UploadURL)
	g.POST("/presign/download", h.DownloadURL)

	if err := g.Run(cfg.Port); err != nil {
		log.Fatalf("server: %v", err)
	}
}
