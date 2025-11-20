package message

import (
	"sort"
	"sync"

	"github.com/ubaidillahfaris/whatsapp.git/internal/core/domain"
	apperrors "github.com/ubaidillahfaris/whatsapp.git/internal/pkg/errors"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/logger"
)

// ProcessorRegistry manages and executes message processors
type ProcessorRegistry struct {
	processors []domain.MessageProcessor
	mu         sync.RWMutex
	logger     *logger.Logger
}

// NewProcessorRegistry creates a new message processor registry
func NewProcessorRegistry() domain.MessageProcessorRegistry {
	return &ProcessorRegistry{
		processors: make([]domain.MessageProcessor, 0),
		logger:     logger.New("MessageProcessorRegistry"),
	}
}

// Register registers a message processor
func (r *ProcessorRegistry) Register(processor domain.MessageProcessor) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.processors = append(r.processors, processor)

	// Sort processors by priority (highest first)
	sort.Slice(r.processors, func(i, j int) bool {
		return r.processors[i].Priority() > r.processors[j].Priority()
	})

	r.logger.WithFields(map[string]interface{}{
		"name":     processor.Name(),
		"priority": processor.Priority(),
	}).Success("Message processor registered")
}

// Process processes a message through all applicable processors
func (r *ProcessorRegistry) Process(message domain.IncomingMessage) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	r.logger.WithFields(map[string]interface{}{
		"device": message.DeviceName,
		"from":   message.From,
	}).Info("Processing incoming message")

	processed := false
	var processingErrors []error

	for _, processor := range r.processors {
		if !processor.CanProcess(message) {
			continue
		}

		r.logger.WithField("processor", processor.Name()).Info("Processing with processor")

		if err := processor.Process(message); err != nil {
			r.logger.WithFields(map[string]interface{}{
				"processor": processor.Name(),
				"error":     err.Error(),
			}).Error("Processor failed")
			processingErrors = append(processingErrors, err)
			continue
		}

		processed = true
		r.logger.WithField("processor", processor.Name()).Success("Message processed")
	}

	if len(processingErrors) > 0 {
		// Return first error encountered
		return processingErrors[0]
	}

	if !processed {
		r.logger.Debug("No processor handled the message")
	}

	return nil
}

// GetProcessors returns all registered processors
func (r *ProcessorRegistry) GetProcessors() []domain.MessageProcessor {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to prevent external modification
	processors := make([]domain.MessageProcessor, len(r.processors))
	copy(processors, r.processors)
	return processors
}

// GetProcessorCount returns the number of registered processors
func (r *ProcessorRegistry) GetProcessorCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.processors)
}

// ProcessMessageUseCase handles processing of incoming messages
type ProcessMessageUseCase struct {
	registry domain.MessageProcessorRegistry
	logger   *logger.Logger
}

// NewProcessMessageUseCase creates a new process message use case
func NewProcessMessageUseCase(registry domain.MessageProcessorRegistry) *ProcessMessageUseCase {
	return &ProcessMessageUseCase{
		registry: registry,
		logger:   logger.New("ProcessMessageUseCase"),
	}
}

// Execute processes an incoming message
func (uc *ProcessMessageUseCase) Execute(message domain.IncomingMessage) error {
	uc.logger.WithFields(map[string]interface{}{
		"device": message.DeviceName,
		"from":   message.From,
	}).Info("Executing message processing")

	if message.Content == "" {
		return apperrors.NewValidationError("Message content is empty")
	}

	// Process through registry
	if err := uc.registry.Process(message); err != nil {
		uc.logger.WithField("error", err.Error()).Error("Message processing failed")
		return err
	}

	uc.logger.Success("Message processing completed")
	return nil
}
