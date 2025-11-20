package device

import (
	"context"

	"github.com/ubaidillahfaris/whatsapp.git/internal/core/domain"
	"github.com/ubaidillahfaris/whatsapp.git/internal/core/ports"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/logger"
)

// DeleteDeviceUseCase handles device deletion logic
type DeleteDeviceUseCase struct {
	deviceRepo    ports.DeviceRepository
	whatsappMgr   domain.WhatsAppManagerInterface
	logger        *logger.Logger
}

// NewDeleteDeviceUseCase creates a new DeleteDeviceUseCase
func NewDeleteDeviceUseCase(deviceRepo ports.DeviceRepository, whatsappMgr domain.WhatsAppManagerInterface) *DeleteDeviceUseCase {
	return &DeleteDeviceUseCase{
		deviceRepo:  deviceRepo,
		whatsappMgr: whatsappMgr,
		logger:      logger.New("DeleteDeviceUseCase"),
	}
}

// Execute deletes a device
func (uc *DeleteDeviceUseCase) Execute(ctx context.Context, id string) error {
	uc.logger.WithField("id", id).Info("Deleting device")

	// Get device to find its name
	device, err := uc.deviceRepo.FindByID(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to find device: %v", err)
		return err
	}

	// Remove from WhatsApp manager if exists
	if uc.whatsappMgr != nil {
		if err := uc.whatsappMgr.RemoveClient(ctx, device.Name); err != nil {
			uc.logger.Warn("Failed to remove WhatsApp client: %v", err)
			// Continue with deletion even if WhatsApp removal fails
		}
	}

	// Delete from repository (soft delete)
	if err := uc.deviceRepo.Delete(ctx, id); err != nil {
		uc.logger.Error("Failed to delete device: %v", err)
		return err
	}

	uc.logger.WithFields(map[string]interface{}{
		"id":   device.ID,
		"name": device.Name,
	}).Success("Device deleted")

	return nil
}
