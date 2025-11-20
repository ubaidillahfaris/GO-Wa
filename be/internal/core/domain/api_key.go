package domain

import (
	"context"
	"time"
)

// APIKeyStatus represents the status of an API key
type APIKeyStatus string

const (
	APIKeyStatusActive   APIKeyStatus = "active"
	APIKeyStatusRevoked  APIKeyStatus = "revoked"
	APIKeyStatusExpired  APIKeyStatus = "expired"
	APIKeyStatusInactive APIKeyStatus = "inactive"
)

// APIKeyPermission represents permissions for an API key
type APIKeyPermission struct {
	Resource string   `json:"resource" bson:"resource"` // e.g., "devices", "messages", "whatsapp"
	Actions  []string `json:"actions" bson:"actions"`   // e.g., ["read", "write", "delete"]
}

// APIKey represents an API key entity
type APIKey struct {
	ID          string             `json:"id" bson:"_id,omitempty"`
	Key         string             `json:"key" bson:"key"`                     // The actual API key
	Name        string             `json:"name" bson:"name"`                   // Human-readable name
	Owner       string             `json:"owner" bson:"owner"`                 // Username of the owner
	Permissions []APIKeyPermission `json:"permissions" bson:"permissions"`     // Granular permissions
	Status      APIKeyStatus       `json:"status" bson:"status"`               // Current status
	RateLimit   int                `json:"rate_limit" bson:"rate_limit"`       // Requests per minute (0 = unlimited)
	LastUsedAt  *time.Time         `json:"last_used_at" bson:"last_used_at"`   // Last time the key was used
	ExpiresAt   *time.Time         `json:"expires_at" bson:"expires_at"`       // Expiration time (nil = never expires)
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// IsExpired checks if the API key has expired
func (k *APIKey) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*k.ExpiresAt)
}

// IsActive checks if the API key is active and not expired
func (k *APIKey) IsActive() bool {
	return k.Status == APIKeyStatusActive && !k.IsExpired()
}

// HasPermission checks if the API key has permission for a specific resource and action
func (k *APIKey) HasPermission(resource string, action string) bool {
	// If no permissions are set, grant all permissions (backward compatibility)
	if len(k.Permissions) == 0 {
		return true
	}

	for _, perm := range k.Permissions {
		if perm.Resource == "*" || perm.Resource == resource {
			for _, act := range perm.Actions {
				if act == "*" || act == action {
					return true
				}
			}
		}
	}
	return false
}

// UpdateLastUsed updates the last used timestamp
func (k *APIKey) UpdateLastUsed() {
	now := time.Now()
	k.LastUsedAt = &now
	k.UpdatedAt = now
}

// Revoke marks the API key as revoked
func (k *APIKey) Revoke() {
	k.Status = APIKeyStatusRevoked
	k.UpdatedAt = time.Now()
}

// CreateAPIKeyRequest represents the request to create a new API key
type CreateAPIKeyRequest struct {
	Name        string             `json:"name" binding:"required,min=3,max=100"`
	Permissions []APIKeyPermission `json:"permissions"`
	RateLimit   int                `json:"rate_limit"`                              // 0 = unlimited
	ExpiresIn   int                `json:"expires_in"`                              // Days until expiration (0 = never)
}

// UpdateAPIKeyRequest represents the request to update an API key
type UpdateAPIKeyRequest struct {
	Name        *string            `json:"name" binding:"omitempty,min=3,max=100"`
	Permissions []APIKeyPermission `json:"permissions"`
	RateLimit   *int               `json:"rate_limit"`
	Status      *APIKeyStatus      `json:"status" binding:"omitempty,oneof=active inactive revoked"`
}

// APIKeyRepository defines the interface for API key storage operations
type APIKeyRepository interface {
	// Create creates a new API key
	Create(ctx context.Context, apiKey *APIKey) error

	// GetByID retrieves an API key by its ID
	GetByID(ctx context.Context, id string) (*APIKey, error)

	// GetByKey retrieves an API key by its key value
	GetByKey(ctx context.Context, key string) (*APIKey, error)

	// List retrieves all API keys for a specific owner with pagination
	List(ctx context.Context, owner string, limit, offset int) ([]*APIKey, int64, error)

	// Update updates an existing API key
	Update(ctx context.Context, apiKey *APIKey) error

	// Delete deletes an API key by ID
	Delete(ctx context.Context, id string) error

	// UpdateLastUsed updates the last used timestamp for an API key
	UpdateLastUsed(ctx context.Context, key string) error

	// CleanupExpired marks expired API keys as expired
	CleanupExpired(ctx context.Context) (int64, error)
}
