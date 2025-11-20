package device

import (
	"context"

	"github.com/ubaidillahfaris/whatsapp.git/internal/core/domain"
	"github.com/ubaidillahfaris/whatsapp.git/internal/core/ports"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/logger"
)

// GetDeviceUseCase handles retrieving a single device
type GetDeviceUseCase struct {
	deviceRepo ports.DeviceRepository
	logger     *logger.Logger
}

// NewGetDeviceUseCase creates a new GetDeviceUseCase
func NewGetDeviceUseCase(deviceRepo ports.DeviceRepository) *GetDeviceUseCase {
	return &GetDeviceUseCase{
		deviceRepo: deviceRepo,
		logger:     logger.New("GetDeviceUseCase"),
	}
}

// Execute retrieves a device by ID
func (uc *GetDeviceUseCase) Execute(ctx context.Context, id string) (*domain.Device, error) {
	uc.logger.WithField("id", id).Info("Getting device")

	device, err := uc.deviceRepo.FindByID(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to get device: %v", err)
		return nil, err
	}

	uc.logger.WithField("name", device.Name).Success("Device retrieved")
	return device, nil
}

// ExecuteByName retrieves a device by name
func (uc *GetDeviceUseCase) ExecuteByName(ctx context.Context, name string) (*domain.Device, error) {
	uc.logger.WithField("name", name).Info("Getting device by name")

	device, err := uc.deviceRepo.FindByName(ctx, name)
	if err != nil {
		uc.logger.Error("Failed to get device: %v", err)
		return nil, err
	}

	uc.logger.WithField("id", device.ID).Success("Device retrieved")
	return device, nil
}
