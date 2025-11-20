package quickresponse

import (
	"github.com/ubaidillahfaris/whatsapp.git/internal/core/domain"
	qrDomain "github.com/ubaidillahfaris/whatsapp.git/internal/modules/quickresponse/domain"
	apperrors "github.com/ubaidillahfaris/whatsapp.git/internal/pkg/errors"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/logger"
)

// Processor processes Quick Response messages
type Processor struct {
	parser     *Parser
	repository qrDomain.QuickResponseRepository
	logger     *logger.Logger
}

// NewProcessor creates a new QuickResponse message processor
func NewProcessor(repository qrDomain.QuickResponseRepository) *Processor {
	return &Processor{
		parser:     NewParser(),
		repository: repository,
		logger:     logger.New("QuickResponseProcessor"),
	}
}

// Name returns the processor name
func (p *Processor) Name() string {
	return "QuickResponseProcessor"
}

// CanProcess checks if this processor can handle the message
func (p *Processor) CanProcess(message domain.IncomingMessage) bool {
	return p.parser.CanParse(message.Content)
}

// Process processes the Quick Response message
func (p *Processor) Process(message domain.IncomingMessage) error {
	p.logger.WithFields(map[string]interface{}{
		"device": message.DeviceName,
		"from":   message.From,
	}).Info("Processing Quick Response message")

	// Parse message
	qr := p.parser.Parse(message.Content)

	// Validate
	if !p.parser.IsValid(qr) {
		p.logger.Warn("Message skipped: no valid officer data")
		return nil // Not an error, just skip
	}

	// Save to database
	if err := p.repository.Save(qr); err != nil {
		p.logger.Error("Failed to save Quick Response: %v", err)
		return apperrors.NewDatabaseError("Failed to save Quick Response", err)
	}

	p.logger.WithFields(map[string]interface{}{
		"officer": qr.Officer.Name,
		"id":      qr.ID,
	}).Success("Quick Response saved")

	return nil
}

// Priority returns the processor priority
func (p *Processor) Priority() int {
	return 100 // High priority for Quick Response messages
}
