package whatsapp

import (
	"context"
	"fmt"

	"github.com/ubaidillahfaris/whatsapp.git/internal/core/domain"
	apperrors "github.com/ubaidillahfaris/whatsapp.git/internal/pkg/errors"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/logger"
)

// GetQRCodeUseCase handles QR code generation logic
type GetQRCodeUseCase struct {
	manager domain.WhatsAppManagerInterface
	logger  *logger.Logger
}

// NewGetQRCodeUseCase creates a new GetQRCodeUseCase
func NewGetQRCodeUseCase(manager domain.WhatsAppManagerInterface) *GetQRCodeUseCase {
	return &GetQRCodeUseCase{
		manager: manager,
		logger:  logger.New("GetQRCodeUseCase"),
	}
}

// Execute generates QR code for device pairing
func (uc *GetQRCodeUseCase) Execute(ctx context.Context, deviceName string) (*domain.QRCodeResponse, error) {
	uc.logger.WithField("device", deviceName).Info("Generating QR code")

	// Get client
	client, exists := uc.manager.GetClient(deviceName)
	if !exists {
		// Create new client if not exists
		var err error
		client, err = uc.manager.CreateClient(ctx, deviceName)
		if err != nil {
			uc.logger.WithField("device", deviceName).Error("Failed to create client: %v", err)
			return nil, apperrors.NewInternalError("Failed to create WhatsApp client", err)
		}
	}

	// Check if already connected
	if client.IsConnected() {
		return nil, apperrors.New(apperrors.ErrorTypeConflict,
			fmt.Sprintf("Device '%s' is already connected", deviceName))
	}

	// Get QR code
	qrResponse, err := client.GetQRCode(ctx)
	if err != nil {
		uc.logger.WithField("device", deviceName).Error("Failed to get QR code: %v", err)
		return nil, apperrors.NewWhatsAppError("Failed to generate QR code", err)
	}

	uc.logger.WithField("device", deviceName).Success("QR code generated")
	return qrResponse, nil
}
