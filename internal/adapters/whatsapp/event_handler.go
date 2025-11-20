package whatsapp

import (
	"github.com/ubaidillahfaris/whatsapp.git/internal/core/domain"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/logger"
)

// EventHandler handles WhatsApp events
type EventHandler struct {
	logger            *logger.Logger
	messageHandlers   []MessageHandlerFunc
	connectionHandlers []ConnectionHandlerFunc
}

// MessageHandlerFunc is a function that handles incoming messages
type MessageHandlerFunc func(deviceName string, message domain.WhatsAppMessage) error

// ConnectionHandlerFunc is a function that handles connection events
type ConnectionHandlerFunc func(deviceName string, connected bool)

// NewEventHandler creates a new event handler
func NewEventHandler() *EventHandler {
	return &EventHandler{
		logger:            logger.New("EventHandler"),
		messageHandlers:   make([]MessageHandlerFunc, 0),
		connectionHandlers: make([]ConnectionHandlerFunc, 0),
	}
}

// RegisterMessageHandler registers a message handler
func (h *EventHandler) RegisterMessageHandler(handler MessageHandlerFunc) {
	h.messageHandlers = append(h.messageHandlers, handler)
}

// RegisterConnectionHandler registers a connection handler
func (h *EventHandler) RegisterConnectionHandler(handler ConnectionHandlerFunc) {
	h.connectionHandlers = append(h.connectionHandlers, handler)
}

// OnConnected handles connection event
func (h *EventHandler) OnConnected(deviceName, jid string) {
	h.logger.WithFields(map[string]interface{}{
		"device": deviceName,
		"jid":    jid,
	}).Success("Device connected")

	// Notify connection handlers
	for _, handler := range h.connectionHandlers {
		handler(deviceName, true)
	}
}

// OnDisconnected handles disconnection event
func (h *EventHandler) OnDisconnected(deviceName string, reason string) {
	h.logger.WithFields(map[string]interface{}{
		"device": deviceName,
		"reason": reason,
	}).Warn("Device disconnected")

	// Notify connection handlers
	for _, handler := range h.connectionHandlers {
		handler(deviceName, false)
	}
}

// OnQRCode handles QR code event
func (h *EventHandler) OnQRCode(deviceName, qrCode string) {
	h.logger.WithField("device", deviceName).Info("QR code received")
	// QR code can be handled by specific use case or handler
}

// OnMessage handles incoming message event
func (h *EventHandler) OnMessage(deviceName string, message domain.WhatsAppMessage) {
	h.logger.WithFields(map[string]interface{}{
		"device": deviceName,
		"from":   message.From,
		"type":   message.Type,
	}).Info("Message received")

	// Process message through all registered handlers
	for _, handler := range h.messageHandlers {
		if err := handler(deviceName, message); err != nil {
			h.logger.WithFields(map[string]interface{}{
				"device": deviceName,
				"error":  err.Error(),
			}).Error("Message handler failed")
		}
	}
}

// OnError handles error event
func (h *EventHandler) OnError(deviceName string, err error) {
	h.logger.WithFields(map[string]interface{}{
		"device": deviceName,
		"error":  err.Error(),
	}).Error("WhatsApp error occurred")
}
