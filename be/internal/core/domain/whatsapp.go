package domain

import (
	"context"
	"time"
)

// ConnectionStatus represents the connection state of a WhatsApp client
type ConnectionStatus string

const (
	StatusDisconnected ConnectionStatus = "disconnected"
	StatusConnecting   ConnectionStatus = "connecting"
	StatusConnected    ConnectionStatus = "connected"
	StatusFailed       ConnectionStatus = "failed"
)

// ReceiverType represents the type of message receiver
type ReceiverType string

const (
	ReceiverIndividual ReceiverType = "individual"
	ReceiverGroup      ReceiverType = "group"
)

// MessageType represents the type of message content
type MessageType string

const (
	MessageTypeText  MessageType = "text"
	MessageTypeFile  MessageType = "file"
	MessageTypeImage MessageType = "image"
	MessageTypeVideo MessageType = "video"
	MessageTypeAudio MessageType = "audio"
)

// WhatsAppSession represents a WhatsApp device session
type WhatsAppSession struct {
	DeviceName       string
	JID              string // WhatsApp JID (e.g., 6281234567890@s.whatsapp.net)
	Status           ConnectionStatus
	QRCode           string
	LastConnected    *time.Time
	LastDisconnected *time.Time
	StoreDBPath      string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// WhatsAppContact represents a WhatsApp contact
type WhatsAppContact struct {
	JID          string
	Name         string
	BusinessName string
	IsGroup      bool
	IsBroadcast  bool
}

// WhatsAppGroup represents a WhatsApp group
type WhatsAppGroup struct {
	JID           string
	Name          string
	Topic         string
	OwnerJID      string
	Participants  []string
	IsAnnounce    bool
	IsLocked      bool
	IsEphemeral   bool
	CreatedAt     time.Time
}

// WhatsAppMessage represents a message to be sent or received
type WhatsAppMessage struct {
	ID           string
	From         string
	To           string
	Type         MessageType
	Content      string
	MediaURL     string
	Caption      string
	Timestamp    time.Time
	IsFromMe     bool
	ReceiverType ReceiverType
}

// SendMessageParams represents parameters for sending a message
type SendMessageParams struct {
	DeviceName   string
	To           string
	Message      string
	ReceiverType ReceiverType
	MessageType  MessageType
	MediaPath    string
	FileName     string
	Caption      string
	Typing       bool
}

// QRCodeResponse represents QR code generation response
type QRCodeResponse struct {
	DeviceName string
	QRCode     string
	ExpiresAt  time.Time
	Timeout    int // seconds
}

// ConnectionInfo represents connection information
type ConnectionInfo struct {
	DeviceName   string
	Status       ConnectionStatus
	JID          string
	IsConnected  bool
	LastPing     *time.Time
}

// DeviceInfo represents device information from WhatsApp
type DeviceInfo struct {
	Platform    string
	DeviceModel string
	OSVersion   string
	WAVersion   string
}

// WhatsAppClientInterface defines the contract for WhatsApp client operations
type WhatsAppClientInterface interface {
	// Connection Management
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	IsConnected() bool
	GetConnectionStatus() ConnectionStatus
	GetQRCode(ctx context.Context) (*QRCodeResponse, error)

	// Device Information
	GetJID() string
	GetDeviceName() string
	GetDeviceInfo() *DeviceInfo

	// Messaging
	SendTextMessage(ctx context.Context, to, message string, receiverType ReceiverType) error
	SendFileMessage(ctx context.Context, params SendMessageParams) error

	// Contacts & Groups
	GetContacts(ctx context.Context) ([]WhatsAppContact, error)
	GetGroups(ctx context.Context) ([]WhatsAppGroup, error)

	// Status
	SetPresence(ctx context.Context, available bool) error
	SendTyping(ctx context.Context, to string, typing bool) error
}

// WhatsAppManagerInterface defines the contract for managing multiple WhatsApp clients
type WhatsAppManagerInterface interface {
	// Client Management
	CreateClient(ctx context.Context, deviceName string) (WhatsAppClientInterface, error)
	GetClient(deviceName string) (WhatsAppClientInterface, bool)
	RemoveClient(ctx context.Context, deviceName string) error
	ListClients() []string
	GetClientCount() int

	// Bulk Operations
	DisconnectAll(ctx context.Context) error
	GetAllConnectionInfo() []ConnectionInfo
	LoadExistingDevices(ctx context.Context) error
}

// WhatsAppEventHandler defines the contract for handling WhatsApp events
type WhatsAppEventHandler interface {
	OnConnected(deviceName, jid string)
	OnDisconnected(deviceName string, reason string)
	OnQRCode(deviceName, qrCode string)
	OnMessage(deviceName string, message WhatsAppMessage)
	OnError(deviceName string, err error)
}
