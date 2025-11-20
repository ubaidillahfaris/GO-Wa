package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ubaidillahfaris/whatsapp.git/services"
)

// WhatsAppHandler menangani semua request terkait WhatsApp instance.
type WhatsAppHandler struct {
	manager *services.WhatsAppManager
}

// NewWhatsAppHandler menginisialisasi handler dengan WhatsAppManager tunggal.
func NewWhatsAppHandler() *WhatsAppHandler {
	return &WhatsAppHandler{
		manager: services.GetWhatsAppManager(),
	}
}

// getOrCreateDevice mencari atau membuat instance WhatsApp berdasarkan nama device.
func (h *WhatsAppHandler) getOrCreateDevice(deviceName string) (*services.WhatsAppService, error) {
	ctx := context.Background()
	return h.manager.GetOrCreateDevice(ctx, deviceName)
}

// 📋 Handler: Ambil daftar semua device aktif.
func (h *WhatsAppHandler) ListDevices(c *gin.Context) {
	devices := h.manager.ListDevices()
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"devices": devices,
	})
}

// 🔑 Handler: Generate QR untuk login WhatsApp.
func (h *WhatsAppHandler) GenerateQR(c *gin.Context) {
	deviceName := c.Param("device")
	if deviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device name is required"})
		return
	}

	// Ambil atau buat instance device
	svc, err := h.getOrCreateDevice(deviceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get or create device",
			"details": err.Error(),
		})
		return
	}

	// Thread-safe check: apakah device sudah connect
	svc.ConnectedMu.Lock()
	connected := svc.IsConnected
	svc.ConnectedMu.Unlock()
	if connected {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Device already connected",
		})
		return
	}

	// Generate QR
	qr, err := svc.GenerateQR()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to generate QR",
			"details": err.Error(),
		})
		return
	}

	if qr == "" {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Device is already logged in or QR already scanned",
		})
		return
	}

	// Convert QR string ke PNG
	png, err := (&services.QrCode{Code: qr}).ToPNG()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to generate QR image",
			"details": err.Error(),
		})
		return
	}

	c.Data(http.StatusOK, "image/png", png)
}

// 📡 Handler: Cek status koneksi WhatsApp per device.
func (h *WhatsAppHandler) GetStatus(c *gin.Context) {
	deviceName := c.Param("device")
	if deviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device name is required"})
		return
	}

	svc, err := h.getOrCreateDevice(deviceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get device", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": svc.Status(),
		"device": deviceName,
	})
}

// 🔌 Handler: Disconnect dan hapus instance device dari manager.
func (h *WhatsAppHandler) Disconnect(c *gin.Context) {
	deviceName := c.Param("device")
	if deviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device name is required"})
		return
	}

	svc, err := h.getOrCreateDevice(deviceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get device", "details": err.Error()})
		return
	}

	svc.Disconnect()
	h.manager.RemoveDevice(deviceName)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Device disconnected and removed",
	})
}

// 📇 Handler: Ambil daftar kontak WhatsApp per device
func (h *WhatsAppHandler) ListContacts(c *gin.Context) {
	deviceName := c.Param("device")
	if deviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device name is required"})
		return
	}

	svc, err := h.getOrCreateDevice(deviceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	return

	contacts, err := svc.ListContacts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"device":   deviceName,
		"contacts": contacts,
	})
}

// 🏷 Handler: Ambil daftar grup WhatsApp per device
func (h *WhatsAppHandler) ListGroups(c *gin.Context) {
	deviceName := c.Param("device")
	if deviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device name is required"})
		return
	}

	svc, err := h.getOrCreateDevice(deviceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	groups, err := svc.ListGroups()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"device": deviceName,
		"groups": groups,
	})
}

func (s *WhatsAppHandler) SendMessage(c *gin.Context) {
	ctx := context.Background()
	deviceID := c.Param("device")
	to := c.PostForm("to")
	text := c.PostForm("message")
	receiver_type := c.PostForm("receiver_type")
	messageType := c.DefaultPostForm("message_type", "text")
	typing := c.DefaultPostForm("typing", "false") == "true"

	// Ambil device
	svc, err := s.manager.GetOrCreateDevice(ctx, deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal ambil device: " + err.Error()})
		return
	}

	var fileData []byte
	var filename, caption string

	if messageType == "file" {
		f, fileHeader, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File is required for type=file"})
			return
		}
		defer f.Close()

		fileData, err = io.ReadAll(f)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal baca file: " + err.Error()})
			return
		}

		filename = c.PostForm("filename")
		if filename == "" {
			filename = fileHeader.Filename
		}

		// pastikan ada ekstensi
		if !strings.Contains(filename, ".") {
			ext := path.Ext(fileHeader.Filename)
			filename += ext
		}

		caption = c.PostForm("caption")
		if caption == "" {
			caption = filename
		}

		if len(fileData) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File kosong"})
			return
		}

		os.MkdirAll("uploads/whatsapp", os.ModePerm)

		localPath := fmt.Sprintf("uploads/whatsapp/%s", filename)
		err = os.WriteFile(localPath, fileData, 0644)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal simpan file: " + err.Error()})
			return
		}
	}

	err = svc.MessageSvc.SendMessage(ctx, to, text, typing, receiver_type, messageType, fileData, filename, caption)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal kirim pesan: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Pesan berhasil dikirim"})

}
