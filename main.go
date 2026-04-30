package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	srv := &http.Server{
		Addr:    cfg.Port,
		Handler: g,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
}
