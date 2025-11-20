package ports

import (
	"context"

	"github.com/ubaidillahfaris/whatsapp.git/internal/core/domain"
)

// WhatsAppSessionRepository defines the contract for WhatsApp session persistence
type WhatsAppSessionRepository interface {
	// Save saves or updates a WhatsApp session
	Save(ctx context.Context, session *domain.WhatsAppSession) error

	// FindByDeviceName retrieves a session by device name
	FindByDeviceName(ctx context.Context, deviceName string) (*domain.WhatsAppSession, error)

	// FindAll retrieves all sessions
	FindAll(ctx context.Context) ([]*domain.WhatsAppSession, error)

	// Delete removes a session
	Delete(ctx context.Context, deviceName string) error

	// UpdateStatus updates the connection status of a session
	UpdateStatus(ctx context.Context, deviceName string, status domain.ConnectionStatus) error

	// UpdateJID updates the JID of a session
	UpdateJID(ctx context.Context, deviceName string, jid string) error
}

// WhatsAppMessageRepository defines the contract for message persistence
type WhatsAppMessageRepository interface {
	// Save saves a message
	Save(ctx context.Context, message *domain.WhatsAppMessage) error

	// FindByDeviceName retrieves messages by device name
	FindByDeviceName(ctx context.Context, deviceName string, limit, offset int) ([]*domain.WhatsAppMessage, error)

	// FindByJID retrieves messages by JID
	FindByJID(ctx context.Context, jid string, limit, offset int) ([]*domain.WhatsAppMessage, error)

	// Count counts messages by filter
	Count(ctx context.Context, filter map[string]interface{}) (int64, error)
}
