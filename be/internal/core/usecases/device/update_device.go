package device

import (
	"context"

	"github.com/ubaidillahfaris/whatsapp.git/internal/core/domain"
	"github.com/ubaidillahfaris/whatsapp.git/internal/core/ports"
	apperrors "github.com/ubaidillahfaris/whatsapp.git/internal/pkg/errors"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/logger"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/validator"
)

// UpdateDeviceUseCase handles device update logic
type UpdateDeviceUseCase struct {
	deviceRepo ports.DeviceRepository
	logger     *logger.Logger
}

// NewUpdateDeviceUseCase creates a new UpdateDeviceUseCase
func NewUpdateDeviceUseCase(deviceRepo ports.DeviceRepository) *UpdateDeviceUseCase {
	return &UpdateDeviceUseCase{
		deviceRepo: deviceRepo,
		logger:     logger.New("UpdateDeviceUseCase"),
	}
}

// Execute updates a device
func (uc *UpdateDeviceUseCase) Execute(ctx context.Context, id string, req domain.UpdateDeviceRequest) (*domain.Device, error) {
	uc.logger.WithField("id", id).Info("Updating device")

	// Get existing device
	device, err := uc.deviceRepo.FindByID(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to find device: %v", err)
		return nil, err
	}

	// Update fields if provided
	if req.Name != nil {
		// Validate device name
		if !validator.ValidateDeviceName(*req.Name) {
			return nil, apperrors.NewValidationError("Invalid device name: must be alphanumeric, dash, or underscore only (3-50 characters)")
		}

		// Check if name is taken by another device
		if *req.Name != device.Name {
			existing, err := uc.deviceRepo.FindByName(ctx, *req.Name)
			if err != nil && !apperrors.IsAppError(err) {
				return nil, err
			}
			if existing != nil && existing.ID != device.ID {
				return nil, apperrors.New(apperrors.ErrorTypeConflict, "Device with this name already exists")
			}
		}

		device.Name = *req.Name
	}

	if req.Description != nil {
		device.Description = *req.Description
	}

	if req.Status != nil {
		device.Status = *req.Status
	}

	// Save updated device
	if err := uc.deviceRepo.Update(ctx, device); err != nil {
		uc.logger.Error("Failed to update device: %v", err)
		return nil, err
	}

	uc.logger.WithFields(map[string]interface{}{
		"id":   device.ID,
		"name": device.Name,
	}).Success("Device updated")

	return device, nil
}
