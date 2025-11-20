package whatsapp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ubaidillahfaris/whatsapp.git/internal/core/domain"
	apperrors "github.com/ubaidillahfaris/whatsapp.git/internal/pkg/errors"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/config"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/logger"
)

// Manager manages multiple WhatsApp clients
type Manager struct {
	clients      map[string]*Client
	mu           sync.RWMutex
	logger       *logger.Logger
	eventHandler domain.WhatsAppEventHandler
	config       *config.Config
}

// NewManager creates a new WhatsApp manager
func NewManager(eventHandler domain.WhatsAppEventHandler) *Manager {
	return &Manager{
		clients:      make(map[string]*Client),
		logger:       logger.New("WhatsAppManager"),
		eventHandler: eventHandler,
		config:       config.Get(),
	}
}

// LoadExistingDevices loads existing devices from stores directory
func (m *Manager) LoadExistingDevices(ctx context.Context) error {
	m.logger.Info("Loading existing devices from stores directory")

	storesDir := m.config.WhatsApp.StoresDir

	// Create stores directory if not exists
	if err := os.MkdirAll(storesDir, 0755); err != nil {
		m.logger.Error("Failed to create stores directory: %v", err)
		return apperrors.NewInternalError("Failed to create stores directory", err)
	}

	// Find all .db files in stores directory
	files, err := filepath.Glob(filepath.Join(storesDir, "*_store.db"))
	if err != nil {
		m.logger.Error("Failed to glob store files: %v", err)
		return apperrors.NewInternalError("Failed to glob store files", err)
	}

	m.logger.WithField("count", len(files)).Info("Found store files")

	// Load each device
	for _, file := range files {
		deviceName := strings.TrimSuffix(filepath.Base(file), "_store.db")

		m.logger.WithField("device", deviceName).Info("Loading device")

		if _, err := m.CreateClient(ctx, deviceName); err != nil {
			m.logger.WithField("device", deviceName).Warn("Failed to load device: %v", err)
			continue
		}
	}

	m.logger.Success("Existing devices loaded")
	return nil
}

// CreateClient creates a new WhatsApp client
func (m *Manager) CreateClient(ctx context.Context, deviceName string) (domain.WhatsAppClientInterface, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.WithField("device", deviceName).Info("Creating client")

	// Check if client already exists
	if _, exists := m.clients[deviceName]; exists {
		m.logger.WithField("device", deviceName).Warn("Client already exists")
		return m.clients[deviceName], nil
	}

	// Create client config
	clientConfig := ClientConfig{
		DeviceName:     deviceName,
		StoresDir:      m.config.WhatsApp.StoresDir,
		EventHandler:   m.eventHandler,
		MaxConcurrency: m.config.WhatsApp.MaxConcurrency,
		LogLevel:       "ERROR",
	}

	// Create new client
	client, err := NewClient(ctx, clientConfig)
	if err != nil {
		m.logger.WithField("device", deviceName).Error("Failed to create client: %v", err)
		return nil, err
	}

	m.clients[deviceName] = client

	m.logger.WithField("device", deviceName).Success("Client created")
	return client, nil
}

// GetClient retrieves a client by device name
func (m *Manager) GetClient(deviceName string) (domain.WhatsAppClientInterface, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, exists := m.clients[deviceName]
	return client, exists
}

// RemoveClient removes a client and cleans up resources
func (m *Manager) RemoveClient(ctx context.Context, deviceName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.WithField("device", deviceName).Info("Removing client")

	client, exists := m.clients[deviceName]
	if !exists {
		return apperrors.NewNotFoundError(fmt.Sprintf("Device '%s'", deviceName))
	}

	// Disconnect client
	if err := client.Disconnect(ctx); err != nil {
		m.logger.WithField("device", deviceName).Warn("Error disconnecting client: %v", err)
	}

	// Remove from map
	delete(m.clients, deviceName)

	// Optionally delete store file
	storeFile := filepath.Join(m.config.WhatsApp.StoresDir, deviceName+"_store.db")
	if err := os.Remove(storeFile); err != nil {
		m.logger.WithField("device", deviceName).Warn("Failed to delete store file: %v", err)
	}

	m.logger.WithField("device", deviceName).Success("Client removed")
	return nil
}

// ListClients returns a list of all device names
func (m *Manager) ListClients() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	deviceNames := make([]string, 0, len(m.clients))
	for name := range m.clients {
		deviceNames = append(deviceNames, name)
	}

	return deviceNames
}

// DisconnectAll disconnects all clients
func (m *Manager) DisconnectAll(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Info("Disconnecting all clients")

	var errors []string
	for deviceName, client := range m.clients {
		if err := client.Disconnect(ctx); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", deviceName, err))
		}
	}

	if len(errors) > 0 {
		return apperrors.New(apperrors.ErrorTypeInternal,
			fmt.Sprintf("Failed to disconnect some clients: %s", strings.Join(errors, "; ")))
	}

	m.logger.Success("All clients disconnected")
	return nil
}

// GetAllConnectionInfo returns connection info for all clients
func (m *Manager) GetAllConnectionInfo() []domain.ConnectionInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info := make([]domain.ConnectionInfo, 0, len(m.clients))
	for deviceName, client := range m.clients {
		info = append(info, domain.ConnectionInfo{
			DeviceName:  deviceName,
			Status:      client.GetConnectionStatus(),
			JID:         client.GetJID(),
			IsConnected: client.IsConnected(),
		})
	}

	return info
}

// GetClientCount returns the number of clients
func (m *Manager) GetClientCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.clients)
}

// GetConnectedCount returns the number of connected clients
func (m *Manager) GetConnectedCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, client := range m.clients {
		if client.IsConnected() {
			count++
		}
	}

	return count
}
