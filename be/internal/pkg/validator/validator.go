package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// Register custom validators
	_ = validate.RegisterValidation("whatsapp_jid", validateWhatsAppJID)
	_ = validate.RegisterValidation("device_name", validateDeviceName)
}

// Validate validates a struct
func Validate(s interface{}) error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	// Format validation errors
	var validationErrors []string
	for _, err := range err.(validator.ValidationErrors) {
		validationErrors = append(validationErrors, formatValidationError(err))
	}

	return fmt.Errorf("validation failed: %s", strings.Join(validationErrors, "; "))
}

// ValidateVar validates a single variable
func ValidateVar(field interface{}, tag string) error {
	return validate.Var(field, tag)
}

// formatValidationError formats a validation error to a human-readable message
func formatValidationError(err validator.FieldError) string {
	field := err.Field()
	tag := err.Tag()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, err.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, err.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, err.Param())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, err.Param())
	case "whatsapp_jid":
		return fmt.Sprintf("%s must be a valid WhatsApp JID", field)
	case "device_name":
		return fmt.Sprintf("%s must be a valid device name (alphanumeric, dash, underscore only)", field)
	default:
		return fmt.Sprintf("%s failed validation: %s", field, tag)
	}
}

// Custom Validators

// validateWhatsAppJID validates WhatsApp JID format
// Format: number@s.whatsapp.net or number@g.us
func validateWhatsAppJID(fl validator.FieldLevel) bool {
	jid := fl.Field().String()
	if jid == "" {
		return false
	}

	// WhatsApp JID patterns:
	// Individual: 1234567890@s.whatsapp.net
	// Group: 1234567890-1234567890@g.us
	patterns := []string{
		`^\d+@s\.whatsapp\.net$`,
		`^\d+-\d+@g\.us$`,
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, jid)
		if matched {
			return true
		}
	}

	return false
}

// validateDeviceName validates device name format
// Only alphanumeric, dash, and underscore allowed
func validateDeviceName(fl validator.FieldLevel) bool {
	name := fl.Field().String()
	if name == "" {
		return false
	}

	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, name)
	return matched
}

// Common validation structs

// SendMessageRequest represents a message sending request
type SendMessageRequest struct {
	To           string `json:"to" validate:"required,whatsapp_jid"`
	Message      string `json:"message" validate:"required"`
	ReceiverType string `json:"receiver_type" validate:"required,oneof=individual group"`
	MessageType  string `json:"message_type" validate:"required,oneof=text file"`
}

// CreateDeviceRequest represents a device creation request
type CreateDeviceRequest struct {
	Name        string `json:"name" validate:"required,device_name,min=3,max=50"`
	Description string `json:"description" validate:"max=200"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6"`
	Email    string `json:"email" validate:"omitempty,email"`
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page  int `form:"page" validate:"omitempty,min=1"`
	Limit int `form:"limit" validate:"omitempty,min=1,max=100"`
}

// GetPaginationParams extracts and validates pagination parameters
func GetPaginationParams(page, limit int) (int, int, error) {
	params := &PaginationParams{
		Page:  page,
		Limit: limit,
	}

	// Set defaults
	if params.Page == 0 {
		params.Page = 1
	}
	if params.Limit == 0 {
		params.Limit = 20
	}

	// Validate
	if err := Validate(params); err != nil {
		return 0, 0, err
	}

	return params.Page, params.Limit, nil
}

// ValidateWhatsAppJID validates a WhatsApp JID string
func ValidateWhatsAppJID(jid string) bool {
	patterns := []string{
		`^\d+@s\.whatsapp\.net$`,
		`^\d+-\d+@g\.us$`,
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, jid)
		if matched {
			return true
		}
	}

	return false
}

// ValidateDeviceName validates a device name string
func ValidateDeviceName(name string) bool {
	if name == "" || len(name) < 3 || len(name) > 50 {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, name)
	return matched
}
