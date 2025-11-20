package device

import (
	"context"

	"github.com/ubaidillahfaris/whatsapp.git/internal/core/domain"
	"github.com/ubaidillahfaris/whatsapp.git/internal/core/ports"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/logger"
)

// ListDevicesUseCase handles listing devices
type ListDevicesUseCase struct {
	deviceRepo ports.DeviceRepository
	logger     *logger.Logger
}

// NewListDevicesUseCase creates a new ListDevicesUseCase
func NewListDevicesUseCase(deviceRepo ports.DeviceRepository) *ListDevicesUseCase {
	return &ListDevicesUseCase{
		deviceRepo: deviceRepo,
		logger:     logger.New("ListDevicesUseCase"),
	}
}

// Execute lists devices with pagination and optional filters
func (uc *ListDevicesUseCase) Execute(ctx context.Context, filter *domain.DeviceFilter, skip, limit int) ([]*domain.Device, int64, error) {
	uc.logger.Info("Listing devices")

	// Get total count
	total, err := uc.deviceRepo.Count(ctx, filter)
	if err != nil {
		uc.logger.Error("Failed to count devices: %v", err)
		return nil, 0, err
	}

	// Get devices
	devices, err := uc.deviceRepo.FindAll(ctx, filter, skip, limit)
	if err != nil {
		uc.logger.Error("Failed to list devices: %v", err)
		return nil, 0, err
	}

	uc.logger.WithFields(map[string]interface{}{
		"count": len(devices),
		"total": total,
	}).Success("Devices listed")

	return devices, total, nil
}
