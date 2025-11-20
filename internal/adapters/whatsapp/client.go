package whatsapp

import (
	"context"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ubaidillahfaris/whatsapp.git/internal/core/domain"
	apperrors "github.com/ubaidillahfaris/whatsapp.git/internal/pkg/errors"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/logger"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// Client is the WhatsApp client adapter using whatsmeow
type Client struct {
	deviceName string
	client     *whatsmeow.Client
	store      *sqlstore.Container
	logger     *logger.Logger

	ctx    context.Context
	cancel context.CancelFunc

	// QR code management
	qrMu     sync.Mutex
	latestQR string

	// Connection state
	connMu      sync.RWMutex
	isConnected bool

	// Event handler
	eventHandler domain.WhatsAppEventHandler

	// Message processing semaphore
	sem chan struct{}
}

// ClientConfig holds configuration for creating a new client
type ClientConfig struct {
	DeviceName       string
	StoresDir        string
	EventHandler     domain.WhatsAppEventHandler
	MaxConcurrency   int
	LogLevel         string
}

// NewClient creates a new WhatsApp client
func NewClient(ctx context.Context, config ClientConfig) (*Client, error) {
	log := logger.New("WhatsAppClient").WithField("device", config.DeviceName)

	// Default values
	if config.MaxConcurrency == 0 {
		config.MaxConcurrency = 10
	}
	if config.LogLevel == "" {
		config.LogLevel = "ERROR"
	}
	if config.StoresDir == "" {
		config.StoresDir = "./stores"
	}

	// Create context
	clientCtx, cancel := context.WithCancel(ctx)

	// Setup SQLite store
	dbPath := fmt.Sprintf("file:%s/%s_store.db?_foreign_keys=on", config.StoresDir, config.DeviceName)
	container, err := sqlstore.New(clientCtx, "sqlite3", dbPath,
		waLog.Stdout("DB-"+config.DeviceName, config.LogLevel, true))
	if err != nil {
		cancel()
		return nil, apperrors.NewDatabaseError("Failed to create SQLite store", err)
	}

	// Get or create device
	deviceStore, err := container.GetFirstDevice(clientCtx)
	if err != nil {
		cancel()
		return nil, apperrors.NewDatabaseError("Failed to get device from store", err)
	}
	if deviceStore == nil {
		deviceStore = container.NewDevice()
	}

	// Create whatsmeow client
	clientLog := waLog.Stdout("Client-"+config.DeviceName, "INFO", true)
	waClient := whatsmeow.NewClient(deviceStore, clientLog)

	client := &Client{
		deviceName:   config.DeviceName,
		client:       waClient,
		store:        container,
		logger:       log,
		ctx:          clientCtx,
		cancel:       cancel,
		eventHandler: config.EventHandler,
		sem:          make(chan struct{}, config.MaxConcurrency),
	}

	// Register event handlers
	client.registerEventHandlers()

	log.Success("WhatsApp client created")
	return client, nil
}

// registerEventHandlers registers whatsmeow event handlers
func (c *Client) registerEventHandlers() {
	c.client.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Connected:
			c.handleConnected()

		case *events.Disconnected:
			c.handleDisconnected()

		case *events.Message:
			c.handleMessage(v)

		case *events.QR:
			c.handleQRCode(v)
		}
	})
}

// handleConnected handles connection event
func (c *Client) handleConnected() {
	c.connMu.Lock()
	c.isConnected = true
	c.connMu.Unlock()

	jid := ""
	if c.client.Store.ID != nil {
		jid = c.client.Store.ID.String()
	}

	c.logger.WithField("jid", jid).Success("Device connected")

	if c.eventHandler != nil {
		c.eventHandler.OnConnected(c.deviceName, jid)
	}
}

// handleDisconnected handles disconnection event
func (c *Client) handleDisconnected() {
	c.connMu.Lock()
	c.isConnected = false
	c.connMu.Unlock()

	c.logger.Warn("Device disconnected")

	if c.eventHandler != nil {
		c.eventHandler.OnDisconnected(c.deviceName, "Connection lost")
	}
}

// handleMessage handles incoming message event
func (c *Client) handleMessage(evt *events.Message) {
	// Skip messages from self
	if evt.Info.IsFromMe {
		return
	}

	// Extract message content
	content := ""
	if evt.Message != nil {
		content = evt.Message.GetConversation()
	}

	// Skip empty messages
	if content == "" {
		return
	}

	c.logger.WithFields(map[string]interface{}{
		"from":    evt.Info.Sender.User,
		"message": content,
	}).Info("Received message")

	// Process message with semaphore for rate limiting
	go func() {
		c.sem <- struct{}{}
		defer func() { <-c.sem }()

		if c.eventHandler != nil {
			msg := domain.WhatsAppMessage{
				ID:        evt.Info.ID,
				From:      evt.Info.Sender.String(),
				To:        c.GetJID(),
				Type:      domain.MessageTypeText,
				Content:   content,
				Timestamp: evt.Info.Timestamp,
				IsFromMe:  evt.Info.IsFromMe,
			}
			c.eventHandler.OnMessage(c.deviceName, msg)
		}
	}()
}

// handleQRCode handles QR code event
func (c *Client) handleQRCode(evt *events.QR) {
	c.qrMu.Lock()
	c.latestQR = evt.Codes[len(evt.Codes)-1]
	c.qrMu.Unlock()

	c.logger.Info("QR code received")

	if c.eventHandler != nil {
		c.eventHandler.OnQRCode(c.deviceName, c.latestQR)
	}
}

// Connect connects the client to WhatsApp
func (c *Client) Connect(ctx context.Context) error {
	c.logger.Info("Connecting to WhatsApp")

	if c.client.IsConnected() {
		return apperrors.New(apperrors.ErrorTypeConflict, "Client already connected")
	}

	if err := c.client.Connect(); err != nil {
		c.logger.Error("Failed to connect: %v", err)
		return apperrors.NewConnectionError("Failed to connect to WhatsApp", err)
	}

	return nil
}

// Disconnect disconnects the client from WhatsApp
func (c *Client) Disconnect(ctx context.Context) error {
	defer func() {
		if r := recover(); r != nil {
			c.logger.Warn("Panic during disconnect: %v", r)
		}
	}()

	c.logger.Info("Disconnecting from WhatsApp")

	if c.client != nil {
		c.client.Disconnect()
	}

	c.cancel()

	c.connMu.Lock()
	c.isConnected = false
	c.connMu.Unlock()

	c.logger.Success("Disconnected successfully")
	return nil
}

// IsConnected returns the connection status
func (c *Client) IsConnected() bool {
	c.connMu.RLock()
	defer c.connMu.RUnlock()
	return c.isConnected && c.client.IsConnected()
}

// GetConnectionStatus returns the current connection status
func (c *Client) GetConnectionStatus() domain.ConnectionStatus {
	if c.client == nil {
		return domain.StatusDisconnected
	}
	if c.client.Store.ID == nil {
		return domain.StatusDisconnected
	}
	if c.IsConnected() {
		return domain.StatusConnected
	}
	return domain.StatusDisconnected
}

// GetQRCode generates and returns a QR code for pairing
func (c *Client) GetQRCode(ctx context.Context) (*domain.QRCodeResponse, error) {
	c.qrMu.Lock()
	defer c.qrMu.Unlock()

	c.logger.Info("Generating QR code")

	// Check if already logged in
	if c.client.Store.ID != nil && c.client.IsConnected() {
		return nil, apperrors.New(apperrors.ErrorTypeConflict, "Device already logged in")
	}

	// Return cached QR if available
	if c.latestQR != "" {
		return &domain.QRCodeResponse{
			DeviceName: c.deviceName,
			QRCode:     c.latestQR,
			Timeout:    30,
			ExpiresAt:  time.Now().Add(30 * time.Second),
		}, nil
	}

	// Get QR channel
	qrChan, _ := c.client.GetQRChannel(c.ctx)

	// Connect to get QR code
	if err := c.client.Connect(); err != nil {
		c.logger.Error("Failed to connect for QR generation: %v", err)
		return nil, apperrors.NewConnectionError("Failed to connect for QR generation", err)
	}

	// Wait for QR code event
	select {
	case evt := <-qrChan:
		if evt.Event == "code" {
			c.latestQR = evt.Code
			c.logger.Success("QR code generated")
			return &domain.QRCodeResponse{
				DeviceName: c.deviceName,
				QRCode:     evt.Code,
				Timeout:    30,
				ExpiresAt:  evt.Timeout,
			}, nil
		}
		return nil, apperrors.NewWhatsAppError(fmt.Sprintf("Unknown QR event: %s", evt.Event), nil)

	case <-time.After(30 * time.Second):
		return nil, apperrors.NewWhatsAppError("Timeout waiting for QR code", nil)

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// GetJID returns the WhatsApp JID of the device
func (c *Client) GetJID() string {
	if c.client.Store.ID == nil {
		return ""
	}
	return c.client.Store.ID.String()
}

// GetDeviceName returns the device name
func (c *Client) GetDeviceName() string {
	return c.deviceName
}

// GetDeviceInfo returns device information
func (c *Client) GetDeviceInfo() *domain.DeviceInfo {
	// WhatsApp device info - this would come from whatsmeow if available
	return &domain.DeviceInfo{
		Platform:    "whatsmeow",
		DeviceModel: "Go Client",
		OSVersion:   "Linux",
		WAVersion:   "2.0",
	}
}

// SendTextMessage sends a text message
func (c *Client) SendTextMessage(ctx context.Context, to, message string, receiverType domain.ReceiverType) error {
	c.logger.WithFields(map[string]interface{}{
		"to":      to,
		"message": message,
		"type":    receiverType,
	}).Info("Sending text message")

	if !c.IsConnected() {
		return apperrors.New(apperrors.ErrorTypeConnection, "Client not connected")
	}

	// Parse JID
	jid, err := parseJID(to)
	if err != nil {
		return apperrors.NewValidationError(fmt.Sprintf("Invalid JID: %s", to))
	}

	// Send message
	msg := &waProto.Message{
		Conversation: &message,
	}

	_, err = c.client.SendMessage(ctx, jid, msg)
	if err != nil {
		c.logger.Error("Failed to send message: %v", err)
		return apperrors.NewWhatsAppError("Failed to send message", err)
	}

	c.logger.Success("Message sent")
	return nil
}

// SendFileMessage sends a file message
func (c *Client) SendFileMessage(ctx context.Context, params domain.SendMessageParams) error {
	c.logger.WithFields(map[string]interface{}{
		"to":   params.To,
		"file": params.FileName,
		"type": params.MessageType,
	}).Info("Sending file message")

	// Implementation would handle file upload and sending
	// This is a placeholder - full implementation would read file, upload, and send
	return apperrors.New(apperrors.ErrorTypeInternal, "File sending not yet implemented in adapter")
}

// GetContacts retrieves all contacts
func (c *Client) GetContacts(ctx context.Context) ([]domain.WhatsAppContact, error) {
	c.logger.Info("Retrieving contacts")

	if !c.IsConnected() {
		return nil, apperrors.New(apperrors.ErrorTypeConnection, "Client not connected")
	}

	contactsMap, err := c.client.Store.Contacts.GetAllContacts(ctx)
	if err != nil {
		c.logger.Error("Failed to get contacts: %v", err)
		return nil, apperrors.NewDatabaseError("Failed to retrieve contacts", err)
	}

	contacts := make([]domain.WhatsAppContact, 0, len(contactsMap))
	for jid, info := range contactsMap {
		name := info.PushName
		if name == "" {
			name = jid.User
		}

		contacts = append(contacts, domain.WhatsAppContact{
			JID:          jid.String(),
			Name:         name,
			BusinessName: info.BusinessName,
		})
	}

	c.logger.WithField("count", len(contacts)).Success("Contacts retrieved")
	return contacts, nil
}

// GetGroups retrieves all groups
func (c *Client) GetGroups(ctx context.Context) ([]domain.WhatsAppGroup, error) {
	c.logger.Info("Retrieving groups")

	if !c.IsConnected() {
		return nil, apperrors.New(apperrors.ErrorTypeConnection, "Client not connected")
	}

	joinedGroups, err := c.client.GetJoinedGroups(ctx)
	if err != nil {
		c.logger.Error("Failed to get groups: %v", err)
		return nil, apperrors.NewWhatsAppError("Failed to retrieve groups", err)
	}

	groups := make([]domain.WhatsAppGroup, 0, len(joinedGroups))
	for _, jid := range joinedGroups {
		// Get group info with retry logic
		groupInfo, err := c.getGroupInfoWithRetry(ctx, jid)
		if err != nil {
			c.logger.Warn("Failed to get info for group %s: %v", jid.String(), err)
			continue
		}

		if groupInfo == nil {
			continue
		}

		participants := make([]string, 0, len(groupInfo.Participants))
		for _, p := range groupInfo.Participants {
			participants = append(participants, p.JID.String())
		}

		groups = append(groups, domain.WhatsAppGroup{
			JID:          jid.String(),
			Name:         groupInfo.Name,
			Topic:        groupInfo.Topic,
			OwnerJID:     groupInfo.OwnerJID.String(),
			Participants: participants,
			IsAnnounce:   groupInfo.IsAnnounce,
			IsLocked:     groupInfo.IsLocked,
			IsEphemeral:  groupInfo.IsEphemeral,
			CreatedAt:    groupInfo.GroupCreated,
		})
	}

	c.logger.WithField("count", len(groups)).Success("Groups retrieved")
	return groups, nil
}

// getGroupInfoWithRetry retrieves group info with retry logic
func (c *Client) getGroupInfoWithRetry(ctx context.Context, jid types.JID) (*types.GroupInfo, error) {
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		info, err := c.client.GetGroupInfo(ctx, jid)
		if err == nil {
			return info, nil
		}
		if i < maxRetries-1 {
			time.Sleep(time.Second * time.Duration(i+1))
		}
	}
	return nil, fmt.Errorf("failed after %d retries", maxRetries)
}

// SetPresence sets the presence status
func (c *Client) SetPresence(ctx context.Context, available bool) error {
	if !c.IsConnected() {
		return apperrors.New(apperrors.ErrorTypeConnection, "Client not connected")
	}

	// Implement presence setting using whatsmeow
	// This is a placeholder
	return nil
}

// SendTyping sends typing indicator
func (c *Client) SendTyping(ctx context.Context, to string, typing bool) error {
	if !c.IsConnected() {
		return apperrors.New(apperrors.ErrorTypeConnection, "Client not connected")
	}

	jid, err := parseJID(to)
	if err != nil {
		return apperrors.NewValidationError(fmt.Sprintf("Invalid JID: %s", to))
	}

	var state types.ChatPresence
	if typing {
		state = types.ChatPresenceComposing
	} else {
		state = types.ChatPresencePaused
	}

	err = c.client.SendChatPresence(jid, state, types.ChatPresenceMediaText)
	if err != nil {
		return apperrors.NewWhatsAppError("Failed to send typing indicator", err)
	}

	return nil
}

// parseJID parses a string JID into types.JID
func parseJID(jidStr string) (types.JID, error) {
	jid, err := types.ParseJID(jidStr)
	if err != nil {
		return types.JID{}, err
	}
	return jid, nil
}
