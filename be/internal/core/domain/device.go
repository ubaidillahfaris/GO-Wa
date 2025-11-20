package domain

import "time"

// Device represents a WhatsApp device configuration
type Device struct {
	ID          string
	Name        string
	Owner       string
	Description string
	Status      DeviceStatus
	JID         string // WhatsApp JID when connected
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// DeviceStatus represents the status of a device
type DeviceStatus string

const (
	DeviceStatusActive   DeviceStatus = "active"
	DeviceStatusInactive DeviceStatus = "inactive"
	DeviceStatusDeleted  DeviceStatus = "deleted"
)

// CreateDeviceRequest represents a request to create a device
type CreateDeviceRequest struct {
	Name        string
	Owner       string
	Description string
}

// UpdateDeviceRequest represents a request to update a device
type UpdateDeviceRequest struct {
	Name        *string
	Description *string
	Status      *DeviceStatus
}

// DeviceFilter represents filters for querying devices
type DeviceFilter struct {
	Owner  string
	Status DeviceStatus
}
