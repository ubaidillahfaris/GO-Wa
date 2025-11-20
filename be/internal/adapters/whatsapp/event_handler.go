package whatsapp

import (
	"github.com/ubaidillahfaris/whatsapp.git/internal/core/domain"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/logger"
)

// EventHandler handles WhatsApp events
type EventHandler struct {
	logger             *logger.Logger
	messageRegistry    domain.MessageProcessorRegistry
	messageHandlers    []MessageHandlerFunc
	connectionHandlers []ConnectionHandlerFunc
}

// MessageHandlerFunc is a function that handles incoming messages
type MessageHandlerFunc func(deviceName string, message domain.WhatsAppMessage) error

// ConnectionHandlerFunc is a function that handles connection events
type ConnectionHandlerFunc func(deviceName string, connected bool)

// NewEventHandler creates a new event handler
func NewEventHandler(messageRegistry domain.MessageProcessorRegistry) *EventHandler {
	return &EventHandler{
		logger:             logger.New("EventHandler"),
		messageRegistry:    messageRegistry,
		messageHandlers:    make([]MessageHandlerFunc, 0),
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

	// Convert to IncomingMessage for processing
	incomingMsg := domain.IncomingMessage{
		ID:         message.ID,
		DeviceName: deviceName,
		From:       message.From,
		Content:    message.Content,
		Timestamp:  message.Timestamp,
		IsGroup:    message.ReceiverType == domain.ReceiverGroup,
	}

	// Process through message registry
	if h.messageRegistry != nil {
		if err := h.messageRegistry.Process(incomingMsg); err != nil {
			h.logger.WithFields(map[string]interface{}{
				"device": deviceName,
				"error":  err.Error(),
			}).Error("Message processing failed")
		}
	}

	// Process message through all registered legacy handlers
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
