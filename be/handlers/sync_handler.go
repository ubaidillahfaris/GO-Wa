package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ubaidillahfaris/whatsapp.git/db"
)

type SyncHandler struct {
	mongo *db.MongoService
}

func NewSyncHandler() *SyncHandler {
	return &SyncHandler{
		mongo: db.Mongo,
	}
}

func (h *SyncHandler) SyncApp(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	// contoh: simpan ke Mongo
	_, err := h.mongo.InsertOne(context.Background(), "pekarya", payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pekarya synced"})
}
