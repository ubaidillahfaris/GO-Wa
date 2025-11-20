package ports

import (
	"context"

	"github.com/ubaidillahfaris/whatsapp.git/internal/core/domain"
)

// DeviceRepository defines the contract for device persistence
type DeviceRepository interface {
	// Create creates a new device
	Create(ctx context.Context, device *domain.Device) error

	// FindByID retrieves a device by ID
	FindByID(ctx context.Context, id string) (*domain.Device, error)

	// FindByName retrieves a device by name
	FindByName(ctx context.Context, name string) (*domain.Device, error)

	// FindAll retrieves all devices with optional filters
	FindAll(ctx context.Context, filter *domain.DeviceFilter, skip, limit int) ([]*domain.Device, error)

	// Update updates a device
	Update(ctx context.Context, device *domain.Device) error

	// Delete deletes a device (soft delete)
	Delete(ctx context.Context, id string) error

	// Count counts devices with optional filter
	Count(ctx context.Context, filter *domain.DeviceFilter) (int64, error)

	// UpdateJID updates the JID of a device
	UpdateJID(ctx context.Context, id, jid string) error

	// UpdateStatus updates the status of a device
	UpdateStatus(ctx context.Context, id string, status domain.DeviceStatus) error
}
