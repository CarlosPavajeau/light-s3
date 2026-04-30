package handlers

import (
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	presigner *s3.PresignClient
	bucket    string
	expiry    time.Duration
}

func New(presigner *s3.PresignClient, bucket string, expiry time.Duration) *Handler {
	return &Handler{presigner: presigner, bucket: bucket, expiry: expiry}
}

type keyRequest struct {
	Key string `json:"key" binding:"required"`
}

func (h *Handler) UploadURL(c *gin.Context) {
	var req keyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.presigner.PresignPutObject(c.Request.Context(), &s3.PutObjectInput{
		Bucket: aws.String(h.bucket),
		Key:    aws.String(req.Key),
	}, s3.WithPresignExpires(h.expiry))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate upload URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url":    result.URL,
		"method": result.Method,
	})
}

func (h *Handler) DownloadURL(c *gin.Context) {
	var req keyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.presigner.PresignGetObject(c.Request.Context(), &s3.GetObjectInput{
		Bucket: aws.String(h.bucket),
		Key:    aws.String(req.Key),
	}, s3.WithPresignExpires(h.expiry))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate download URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url":    result.URL,
		"method": result.Method,
	})
}
