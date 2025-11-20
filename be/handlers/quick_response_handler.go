package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ubaidillahfaris/whatsapp.git/db"
	"github.com/ubaidillahfaris/whatsapp.git/helpers"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type QuickResponseHandler struct {
}

func NewQuickResponseHandler() *QuickResponseHandler {
	return &QuickResponseHandler{}
}

func (q *QuickResponseHandler) GetAll(c *gin.Context) {
	skip, limit := helpers.GetPagination(c, 20)
	qr, err := db.Mongo.FindAll(c.Request.Context(), "quick_responses", nil, &skip, &limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"quick_response": qr})
}

func (q *QuickResponseHandler) DeleteId(c *gin.Context) {
	idParam := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}

	if err := db.Mongo.Delete(c.Request.Context(), "quick_responses", id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Quick response deleted"})
}
