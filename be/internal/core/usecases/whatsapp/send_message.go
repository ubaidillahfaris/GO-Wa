package whatsapp

import (
	"context"
	"fmt"

	"github.com/ubaidillahfaris/whatsapp.git/internal/core/domain"
	apperrors "github.com/ubaidillahfaris/whatsapp.git/internal/pkg/errors"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/logger"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/validator"
)

// SendMessageUseCase handles message sending logic
type SendMessageUseCase struct {
	manager domain.WhatsAppManagerInterface
	logger  *logger.Logger
}

// NewSendMessageUseCase creates a new SendMessageUseCase
func NewSendMessageUseCase(manager domain.WhatsAppManagerInterface) *SendMessageUseCase {
	return &SendMessageUseCase{
		manager: manager,
		logger:  logger.New("SendMessageUseCase"),
	}
}

// Execute sends a message via WhatsApp
func (uc *SendMessageUseCase) Execute(ctx context.Context, params domain.SendMessageParams) error {
	uc.logger.WithFields(map[string]interface{}{
		"device": params.DeviceName,
		"to":     params.To,
		"type":   params.MessageType,
	}).Info("Sending message")

	// Validate JID
	if !validator.ValidateWhatsAppJID(params.To) {
		return apperrors.NewValidationError(fmt.Sprintf("Invalid WhatsApp JID: %s", params.To))
	}

	// Get client
	client, exists := uc.manager.GetClient(params.DeviceName)
	if !exists {
		return apperrors.NewNotFoundError(fmt.Sprintf("Device '%s'", params.DeviceName))
	}

	// Check if connected
	if !client.IsConnected() {
		return apperrors.New(apperrors.ErrorTypeConnection,
			fmt.Sprintf("Device '%s' is not connected", params.DeviceName))
	}

	// Send typing indicator if enabled
	if params.Typing {
		if err := client.SendTyping(ctx, params.To, true); err != nil {
			uc.logger.Warn("Failed to send typing indicator: %v", err)
		}
		defer func() {
			_ = client.SendTyping(ctx, params.To, false)
		}()
	}

	// Send message based on type
	var err error
	switch params.MessageType {
	case domain.MessageTypeText:
		err = client.SendTextMessage(ctx, params.To, params.Message, params.ReceiverType)
	case domain.MessageTypeFile, domain.MessageTypeImage, domain.MessageTypeVideo, domain.MessageTypeAudio:
		err = client.SendFileMessage(ctx, params)
	default:
		return apperrors.NewValidationError(fmt.Sprintf("Unsupported message type: %s", params.MessageType))
	}

	if err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"device": params.DeviceName,
			"to":     params.To,
			"error":  err.Error(),
		}).Error("Failed to send message")
		return apperrors.NewWhatsAppError("Failed to send message", err)
	}

	uc.logger.WithFields(map[string]interface{}{
		"device": params.DeviceName,
		"to":     params.To,
		"type":   params.MessageType,
	}).Success("Message sent successfully")

	return nil
}
