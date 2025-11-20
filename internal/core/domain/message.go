package domain

import "time"

// IncomingMessage represents a received WhatsApp message
type IncomingMessage struct {
	ID           string
	DeviceName   string
	From         string
	FromName     string
	Content      string
	Timestamp    time.Time
	IsGroup      bool
	IsProcessed  bool
	ProcessedAt  *time.Time
	ProcessError string
}

// MessageProcessor defines the contract for processing incoming messages
type MessageProcessor interface {
	// Name returns the processor name for identification
	Name() string

	// CanProcess checks if this processor can handle the message
	CanProcess(message IncomingMessage) bool

	// Process processes the message and returns an error if processing fails
	Process(message IncomingMessage) error

	// Priority returns the priority of this processor (higher = processed first)
	Priority() int
}

// MessageProcessorRegistry manages message processors
type MessageProcessorRegistry interface {
	// Register registers a message processor
	Register(processor MessageProcessor)

	// Process processes a message through all applicable processors
	Process(message IncomingMessage) error

	// GetProcessors returns all registered processors
	GetProcessors() []MessageProcessor
}
