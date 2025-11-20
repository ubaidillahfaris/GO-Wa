package whatsapp

import (
	"context"

	"github.com/ubaidillahfaris/whatsapp.git/internal/core/domain"
	"github.com/ubaidillahfaris/whatsapp.git/internal/core/ports"
	"github.com/ubaidillahfaris/whatsapp.git/internal/core/usecases/whatsapp"
	apperrors "github.com/ubaidillahfaris/whatsapp.git/internal/pkg/errors"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/logger"
)

// Service implements the WhatsAppService port using use cases
type Service struct {
	manager domain.WhatsAppManagerInterface
	logger  *logger.Logger

	// Use cases
	connectUC      *whatsapp.ConnectUseCase
	disconnectUC   *whatsapp.DisconnectUseCase
	getQRUC        *whatsapp.GetQRCodeUseCase
	sendMessageUC  *whatsapp.SendMessageUseCase
	listContactsUC *whatsapp.ListContactsUseCase
	listGroupsUC   *whatsapp.ListGroupsUseCase
}

// NewService creates a new WhatsApp service
func NewService(manager domain.WhatsAppManagerInterface) ports.WhatsAppService {
	return &Service{
		manager:        manager,
		logger:         logger.New("WhatsAppService"),
		connectUC:      whatsapp.NewConnectUseCase(manager),
		disconnectUC:   whatsapp.NewDisconnectUseCase(manager),
		getQRUC:        whatsapp.NewGetQRCodeUseCase(manager),
		sendMessageUC:  whatsapp.NewSendMessageUseCase(manager),
		listContactsUC: whatsapp.NewListContactsUseCase(manager),
		listGroupsUC:   whatsapp.NewListGroupsUseCase(manager),
	}
}

// ConnectDevice connects a WhatsApp device
func (s *Service) ConnectDevice(ctx context.Context, deviceName string) error {
	return s.connectUC.Execute(ctx, deviceName)
}

// DisconnectDevice disconnects a WhatsApp device
func (s *Service) DisconnectDevice(ctx context.Context, deviceName string) error {
	return s.disconnectUC.Execute(ctx, deviceName)
}

// GetQRCode generates QR code for device pairing
func (s *Service) GetQRCode(ctx context.Context, deviceName string) (*domain.QRCodeResponse, error) {
	return s.getQRUC.Execute(ctx, deviceName)
}

// IsDeviceConnected checks if a device is connected
func (s *Service) IsDeviceConnected(deviceName string) bool {
	client, exists := s.manager.GetClient(deviceName)
	if !exists {
		return false
	}
	return client.IsConnected()
}

// GetConnectionInfo returns connection info for a device
func (s *Service) GetConnectionInfo(deviceName string) (*domain.ConnectionInfo, error) {
	client, exists := s.manager.GetClient(deviceName)
	if !exists {
		return nil, apperrors.NewNotFoundError("Device '" + deviceName + "'")
	}

	return &domain.ConnectionInfo{
		DeviceName:  deviceName,
		Status:      client.GetConnectionStatus(),
		JID:         client.GetJID(),
		IsConnected: client.IsConnected(),
	}, nil
}

// GetAllConnectionInfo returns connection info for all devices
func (s *Service) GetAllConnectionInfo() []domain.ConnectionInfo {
	return s.manager.GetAllConnectionInfo()
}

// SendMessage sends a message via WhatsApp
func (s *Service) SendMessage(ctx context.Context, params domain.SendMessageParams) error {
	return s.sendMessageUC.Execute(ctx, params)
}

// SendTextMessage sends a text message
func (s *Service) SendTextMessage(ctx context.Context, deviceName, to, message string, receiverType domain.ReceiverType) error {
	params := domain.SendMessageParams{
		DeviceName:   deviceName,
		To:           to,
		Message:      message,
		ReceiverType: receiverType,
		MessageType:  domain.MessageTypeText,
	}
	return s.sendMessageUC.Execute(ctx, params)
}

// SendFileMessage sends a file message
func (s *Service) SendFileMessage(ctx context.Context, params domain.SendMessageParams) error {
	return s.sendMessageUC.Execute(ctx, params)
}

// ListContacts retrieves contacts from a WhatsApp device
func (s *Service) ListContacts(ctx context.Context, deviceName string) ([]domain.WhatsAppContact, error) {
	return s.listContactsUC.Execute(ctx, deviceName)
}

// ListGroups retrieves groups from a WhatsApp device
func (s *Service) ListGroups(ctx context.Context, deviceName string) ([]domain.WhatsAppGroup, error) {
	return s.listGroupsUC.Execute(ctx, deviceName)
}

// CreateDevice creates a new device
func (s *Service) CreateDevice(ctx context.Context, deviceName string) error {
	_, err := s.manager.CreateClient(ctx, deviceName)
	return err
}

// RemoveDevice removes a device
func (s *Service) RemoveDevice(ctx context.Context, deviceName string) error {
	return s.manager.RemoveClient(ctx, deviceName)
}

// ListDevices returns a list of all devices
func (s *Service) ListDevices() []string {
	return s.manager.ListClients()
}

// SetPresence sets the presence status
func (s *Service) SetPresence(ctx context.Context, deviceName string, available bool) error {
	client, exists := s.manager.GetClient(deviceName)
	if !exists {
		return apperrors.NewNotFoundError("Device '" + deviceName + "'")
	}
	return client.SetPresence(ctx, available)
}

// SendTyping sends typing indicator
func (s *Service) SendTyping(ctx context.Context, deviceName, to string, typing bool) error {
	client, exists := s.manager.GetClient(deviceName)
	if !exists {
		return apperrors.NewNotFoundError("Device '" + deviceName + "'")
	}
	return client.SendTyping(ctx, to, typing)
}
