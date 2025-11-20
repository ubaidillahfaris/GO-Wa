package whatsapp

import (
	"context"
	"fmt"

	"github.com/ubaidillahfaris/whatsapp.git/internal/core/domain"
	apperrors "github.com/ubaidillahfaris/whatsapp.git/internal/pkg/errors"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/logger"
)

// ListContactsUseCase handles listing contacts logic
type ListContactsUseCase struct {
	manager domain.WhatsAppManagerInterface
	logger  *logger.Logger
}

// NewListContactsUseCase creates a new ListContactsUseCase
func NewListContactsUseCase(manager domain.WhatsAppManagerInterface) *ListContactsUseCase {
	return &ListContactsUseCase{
		manager: manager,
		logger:  logger.New("ListContactsUseCase"),
	}
}

// Execute retrieves contacts from a WhatsApp device
func (uc *ListContactsUseCase) Execute(ctx context.Context, deviceName string) ([]domain.WhatsAppContact, error) {
	uc.logger.WithField("device", deviceName).Info("Listing contacts")

	// Get client
	client, exists := uc.manager.GetClient(deviceName)
	if !exists {
		return nil, apperrors.NewNotFoundError(fmt.Sprintf("Device '%s'", deviceName))
	}

	// Check if connected
	if !client.IsConnected() {
		return nil, apperrors.New(apperrors.ErrorTypeConnection,
			fmt.Sprintf("Device '%s' is not connected", deviceName))
	}

	// Get contacts
	contacts, err := client.GetContacts(ctx)
	if err != nil {
		uc.logger.WithField("device", deviceName).Error("Failed to get contacts: %v", err)
		return nil, apperrors.NewWhatsAppError("Failed to retrieve contacts", err)
	}

	uc.logger.WithFields(map[string]interface{}{
		"device": deviceName,
		"count":  len(contacts),
	}).Success("Contacts retrieved")

	return contacts, nil
}
