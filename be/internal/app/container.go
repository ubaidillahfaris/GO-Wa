package app

import (
	"context"
	"fmt"

	"github.com/ubaidillahfaris/whatsapp.git/internal/adapters/repositories"
	"github.com/ubaidillahfaris/whatsapp.git/internal/adapters/whatsapp"
	"github.com/ubaidillahfaris/whatsapp.git/internal/core/domain"
	"github.com/ubaidillahfaris/whatsapp.git/internal/core/ports"
	"github.com/ubaidillahfaris/whatsapp.git/internal/core/usecases/apikey"
	"github.com/ubaidillahfaris/whatsapp.git/internal/core/usecases/device"
	"github.com/ubaidillahfaris/whatsapp.git/internal/core/usecases/message"
	"github.com/ubaidillahfaris/whatsapp.git/internal/modules/quickresponse"
	qrDomain "github.com/ubaidillahfaris/whatsapp.git/internal/modules/quickresponse/domain"
	qrRepo "github.com/ubaidillahfaris/whatsapp.git/internal/modules/quickresponse/repository"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/config"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Container holds all application dependencies
type Container struct {
	// Config
	Config *config.Config

	// Database
	MongoDB *mongo.Database

	// Repositories
	DeviceRepository ports.DeviceRepository
	QRRepository     qrDomain.QuickResponseRepository
	APIKeyRepository domain.APIKeyRepository

	// Message Processing
	MessageRegistry domain.MessageProcessorRegistry
	QRProcessor     domain.MessageProcessor

	// WhatsApp
	WhatsAppEventHandler *whatsapp.EventHandler
	WhatsAppManager      domain.WhatsAppManagerInterface
	WhatsAppService      ports.WhatsAppService

	// Use Cases - Device
	CreateDeviceUC *device.CreateDeviceUseCase
	GetDeviceUC    *device.GetDeviceUseCase
	ListDevicesUC  *device.ListDevicesUseCase
	UpdateDeviceUC *device.UpdateDeviceUseCase
	DeleteDeviceUC *device.DeleteDeviceUseCase

	// Use Cases - Message
	ProcessMessageUC *message.ProcessMessageUseCase

	// Use Cases - API Key
	GenerateAPIKeyUC *apikey.GenerateKeyUseCase
	ListAPIKeysUC    *apikey.ListKeysUseCase
	RevokeAPIKeyUC   *apikey.RevokeKeyUseCase
	UpdateAPIKeyUC   *apikey.UpdateKeyUseCase
	ValidateAPIKeyUC *apikey.ValidateKeyUseCase

	logger *logger.Logger
}

// NewContainer creates and initializes the application container
func NewContainer(ctx context.Context) (*Container, error) {
	log := logger.New("Container")
	log.Info("Initializing application container")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	container := &Container{
		Config: cfg,
		logger: log,
	}

	// Initialize components in order
	if err := container.initDatabase(ctx); err != nil {
		return nil, err
	}

	if err := container.initRepositories(); err != nil {
		return nil, err
	}

	if err := container.initMessageProcessing(); err != nil {
		return nil, err
	}

	if err := container.initWhatsApp(ctx); err != nil {
		return nil, err
	}

	if err := container.initUseCases(); err != nil {
		return nil, err
	}

	log.Success("Application container initialized")
	return container, nil
}

// initDatabase initializes MongoDB connection
func (c *Container) initDatabase(ctx context.Context) error {
	c.logger.Info("Connecting to MongoDB")

	clientOpts := options.Client().ApplyURI(c.Config.MongoDB.URI)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		c.logger.Error("Failed to connect to MongoDB: %v", err)
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping database
	if err := client.Ping(ctx, nil); err != nil {
		c.logger.Error("Failed to ping MongoDB: %v", err)
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	c.MongoDB = client.Database(c.Config.MongoDB.Database)
	c.logger.WithField("database", c.Config.MongoDB.Database).Success("MongoDB connected")
	return nil
}

// initRepositories initializes all repositories
func (c *Container) initRepositories() error {
	c.logger.Info("Initializing repositories")

	// Device repository
	c.DeviceRepository = repositories.NewDeviceMongoRepository(c.MongoDB)

	// Quick Response repository
	c.QRRepository = qrRepo.NewMongoRepository(c.MongoDB)

	// API Key repository
	apiKeyRepo, err := repositories.NewAPIKeyMongoRepository(c.MongoDB, c.logger)
	if err != nil {
		return fmt.Errorf("failed to create API key repository: %w", err)
	}
	c.APIKeyRepository = apiKeyRepo

	c.logger.Success("Repositories initialized")
	return nil
}

// initMessageProcessing initializes message processing components
func (c *Container) initMessageProcessing() error {
	c.logger.Info("Initializing message processing")

	// Create message processor registry
	c.MessageRegistry = message.NewProcessorRegistry()

	// Create and register Quick Response processor
	c.QRProcessor = quickresponse.NewProcessor(c.QRRepository)
	c.MessageRegistry.Register(c.QRProcessor)

	c.logger.WithField("processors", c.MessageRegistry.GetProcessors()).Success("Message processing initialized")
	return nil
}

// initWhatsApp initializes WhatsApp components
func (c *Container) initWhatsApp(ctx context.Context) error {
	c.logger.Info("Initializing WhatsApp components")

	// Create event handler with message registry
	c.WhatsAppEventHandler = whatsapp.NewEventHandler(c.MessageRegistry)

	// Create WhatsApp manager
	c.WhatsAppManager = whatsapp.NewManager(c.WhatsAppEventHandler)

	// Load existing devices
	if err := c.WhatsAppManager.LoadExistingDevices(ctx); err != nil {
		c.logger.Warn("Failed to load existing devices: %v", err)
		// Not a fatal error, continue
	}

	// Create WhatsApp service
	c.WhatsAppService = whatsapp.NewService(c.WhatsAppManager)

	c.logger.WithField("devices", c.WhatsAppManager.GetClientCount()).Success("WhatsApp components initialized")
	return nil
}

// initUseCases initializes all use cases
func (c *Container) initUseCases() error {
	c.logger.Info("Initializing use cases")

	// Device use cases
	c.CreateDeviceUC = device.NewCreateDeviceUseCase(c.DeviceRepository)
	c.GetDeviceUC = device.NewGetDeviceUseCase(c.DeviceRepository)
	c.ListDevicesUC = device.NewListDevicesUseCase(c.DeviceRepository)
	c.UpdateDeviceUC = device.NewUpdateDeviceUseCase(c.DeviceRepository)
	c.DeleteDeviceUC = device.NewDeleteDeviceUseCase(c.DeviceRepository, c.WhatsAppManager)

	// Message use cases
	c.ProcessMessageUC = message.NewProcessMessageUseCase(c.MessageRegistry)

	// API Key use cases
	c.GenerateAPIKeyUC = apikey.NewGenerateKeyUseCase(c.APIKeyRepository, c.logger)
	c.ListAPIKeysUC = apikey.NewListKeysUseCase(c.APIKeyRepository, c.logger)
	c.RevokeAPIKeyUC = apikey.NewRevokeKeyUseCase(c.APIKeyRepository, c.logger)
	c.UpdateAPIKeyUC = apikey.NewUpdateKeyUseCase(c.APIKeyRepository, c.logger)
	c.ValidateAPIKeyUC = apikey.NewValidateKeyUseCase(c.APIKeyRepository, c.logger)

	c.logger.Success("Use cases initialized")
	return nil
}

// Shutdown performs graceful shutdown of all components
func (c *Container) Shutdown(ctx context.Context) error {
	c.logger.Info("Shutting down application")

	// Disconnect all WhatsApp clients
	if c.WhatsAppManager != nil {
		if err := c.WhatsAppManager.DisconnectAll(ctx); err != nil {
			c.logger.Warn("Error disconnecting WhatsApp clients: %v", err)
		}
	}

	// Disconnect MongoDB
	if c.MongoDB != nil {
		if err := c.MongoDB.Client().Disconnect(ctx); err != nil {
			c.logger.Warn("Error disconnecting MongoDB: %v", err)
		}
	}

	c.logger.Success("Application shutdown complete")
	return nil
}
