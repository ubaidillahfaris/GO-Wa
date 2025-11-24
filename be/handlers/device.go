package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/ubaidillahfaris/whatsapp.git/db"
	"github.com/ubaidillahfaris/whatsapp.git/helpers"
	"github.com/ubaidillahfaris/whatsapp.git/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DeviceHandler struct {
	Mongo      *db.MongoService
	Collection string
}

func NewDeviceHandler(mongo *db.MongoService) *DeviceHandler {
	return &DeviceHandler{Mongo: mongo, Collection: "devices"}
}

func (h *DeviceHandler) CreateDevice(c *gin.Context) {
	var device models.Device

	// Bind JSON dulu
	if err := c.ShouldBindJSON(&device); err != nil {
		if err.Error() == "EOF" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "request body is empty"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get username from JWT
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	usernameStr := username.(string)
	device.Owner = &usernameStr

	// Validasi struct
	validate := validator.New()
	if err := validate.Struct(device); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	device.ID = primitive.NewObjectID()
	now := time.Now().Unix()
	device.CreatedAt = &now
	device.UpdatedAt = &now

	if _, err := db.Mongo.InsertOne(c.Request.Context(), h.Collection, device); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Device created", "device": device})
}

func (h *DeviceHandler) ListDevices(c *gin.Context) {
	// Get username from JWT
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	skip, limit := helpers.GetPagination(c, 20)

	// Filter by owner
	filter := map[string]interface{}{
		"owner": username.(string),
	}

	devices, err := h.Mongo.FindAllPaginate(c.Request.Context(), h.Collection, filter, &skip, &limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, devices)
}

func (h *DeviceHandler) GetDevice(c *gin.Context) {
	// Get username from JWT
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idParam := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}

	device, err := h.Mongo.FindByID(c.Request.Context(), h.Collection, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}

	// Check ownership
	deviceOwner, ok := device["owner"].(string)
	if !ok || deviceOwner != username.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"device": device})
}

func (h *DeviceHandler) UpdateDevice(c *gin.Context) {
	// Get username from JWT
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idParam := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}

	// Check ownership first
	device, err := h.Mongo.FindByID(c.Request.Context(), h.Collection, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}

	deviceOwner, ok := device["owner"].(string)
	if !ok || deviceOwner != username.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data["updated_at"] = time.Now().Unix()

	// Prevent owner change
	delete(data, "owner")

	if err := h.Mongo.Update(c.Request.Context(), h.Collection, id, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device updated"})
}

func (h *DeviceHandler) DeleteDevice(c *gin.Context) {
	// Get username from JWT
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idParam := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}

	// Check ownership first
	device, err := h.Mongo.FindByID(c.Request.Context(), h.Collection, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}

	deviceOwner, ok := device["owner"].(string)
	if !ok || deviceOwner != username.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.Mongo.Delete(c.Request.Context(), h.Collection, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device deleted"})
}
