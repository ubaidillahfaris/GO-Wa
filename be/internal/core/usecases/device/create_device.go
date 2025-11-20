package device

import (
	"context"
	"time"

	"github.com/ubaidillahfaris/whatsapp.git/internal/core/domain"
	"github.com/ubaidillahfaris/whatsapp.git/internal/core/ports"
	apperrors "github.com/ubaidillahfaris/whatsapp.git/internal/pkg/errors"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/logger"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/validator"
)

// CreateDeviceUseCase handles device creation logic
type CreateDeviceUseCase struct {
	deviceRepo ports.DeviceRepository
	logger     *logger.Logger
}

// NewCreateDeviceUseCase creates a new CreateDeviceUseCase
func NewCreateDeviceUseCase(deviceRepo ports.DeviceRepository) *CreateDeviceUseCase {
	return &CreateDeviceUseCase{
		deviceRepo: deviceRepo,
		logger:     logger.New("CreateDeviceUseCase"),
	}
}

// Execute creates a new device
func (uc *CreateDeviceUseCase) Execute(ctx context.Context, req domain.CreateDeviceRequest) (*domain.Device, error) {
	uc.logger.WithField("name", req.Name).Info("Creating device")

	// Validate device name
	if !validator.ValidateDeviceName(req.Name) {
		return nil, apperrors.NewValidationError("Invalid device name: must be alphanumeric, dash, or underscore only (3-50 characters)")
	}

	// Validate owner
	if req.Owner == "" {
		return nil, apperrors.NewValidationError("Owner is required")
	}

	// Check if device already exists
	existing, err := uc.deviceRepo.FindByName(ctx, req.Name)
	if err != nil && !apperrors.IsAppError(err) {
		return nil, err
	}
	if existing != nil {
		return nil, apperrors.New(apperrors.ErrorTypeConflict, "Device with this name already exists")
	}

	// Create device entity
	device := &domain.Device{
		Name:        req.Name,
		Owner:       req.Owner,
		Description: req.Description,
		Status:      domain.DeviceStatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save to repository
	if err := uc.deviceRepo.Create(ctx, device); err != nil {
		uc.logger.Error("Failed to create device: %v", err)
		return nil, err
	}

	uc.logger.WithFields(map[string]interface{}{
		"id":   device.ID,
		"name": device.Name,
	}).Success("Device created")

	return device, nil
}
