package ports

import (
	"context"

	"github.com/ubaidillahfaris/whatsapp.git/internal/core/domain"
)

// WhatsAppService defines the contract for WhatsApp business logic
type WhatsAppService interface {
	// Connection Management
	ConnectDevice(ctx context.Context, deviceName string) error
	DisconnectDevice(ctx context.Context, deviceName string) error
	GetQRCode(ctx context.Context, deviceName string) (*domain.QRCodeResponse, error)
	IsDeviceConnected(deviceName string) bool
	GetConnectionInfo(deviceName string) (*domain.ConnectionInfo, error)
	GetAllConnectionInfo() []domain.ConnectionInfo

	// Messaging
	SendMessage(ctx context.Context, params domain.SendMessageParams) error
	SendTextMessage(ctx context.Context, deviceName, to, message string, receiverType domain.ReceiverType) error
	SendFileMessage(ctx context.Context, params domain.SendMessageParams) error

	// Contacts & Groups
	ListContacts(ctx context.Context, deviceName string) ([]domain.WhatsAppContact, error)
	ListGroups(ctx context.Context, deviceName string) ([]domain.WhatsAppGroup, error)

	// Device Management
	CreateDevice(ctx context.Context, deviceName string) error
	RemoveDevice(ctx context.Context, deviceName string) error
	ListDevices() []string

	// Utility
	SetPresence(ctx context.Context, deviceName string, available bool) error
	SendTyping(ctx context.Context, deviceName, to string, typing bool) error
}
