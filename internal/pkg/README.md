# Internal Packages

Foundation packages untuk aplikasi WhatsApp. Packages ini menyediakan utilities yang reusable dan konsisten di seluruh aplikasi.

## Packages

### 1. errors - Custom Error Handling

Custom error types dengan HTTP status code mapping dan error wrapping.

**Usage:**

```go
import apperrors "github.com/ubaidillahfaris/whatsapp.git/internal/pkg/errors"

// Create new error
err := apperrors.NewNotFoundError("Device")
// Output: NOT_FOUND: Device not found (StatusCode: 404)

// Wrap existing error
err := apperrors.Wrap(sqlErr, apperrors.ErrorTypeDatabase, "Failed to fetch device")
// Output: DATABASE_ERROR: Failed to fetch device (caused by: <original error>)

// Add details
err := apperrors.NewValidationError("Invalid input").
    WithDetails("field", "device_name").
    WithDetails("value", "invalid@name")

// Check and extract AppError
if apperrors.IsAppError(err) {
    appErr := apperrors.GetAppError(err)
    statusCode := appErr.StatusCode
    errType := appErr.Type
}
```

**Error Types:**
- `ErrorTypeValidation` - Input validation errors (400)
- `ErrorTypeNotFound` - Resource not found (404)
- `ErrorTypeUnauthorized` - Authentication required (401)
- `ErrorTypeForbidden` - Permission denied (403)
- `ErrorTypeConflict` - Resource conflict (409)
- `ErrorTypeWhatsApp` - WhatsApp operation errors (503)
- `ErrorTypeConnection` - Connection errors (503)
- `ErrorTypeDatabase` - Database errors (500)
- `ErrorTypeInternal` - Internal server errors (500)

---

### 2. logger - Structured Logging

Structured logger dengan emoji support, fields, dan context.

**Usage:**

```go
import "github.com/ubaidillahfaris/whatsapp.git/internal/pkg/logger"

// Create logger with prefix
log := logger.New("WhatsApp")

// Simple logging
log.Info("Device connected")
log.Error("Failed to send message")
log.Warn("Connection unstable")
log.Debug("Processing event: %s", eventType)

// With fields
log.WithField("device", deviceName).
    WithField("jid", jid).
    Info("Message sent successfully")

// With multiple fields
log.WithFields(map[string]interface{}{
    "device": deviceName,
    "to": recipientJID,
    "type": messageType,
}).Info("Sending message")

// Success message
log.Success("QR code generated")

// Error with stack trace
log.ErrorWithStack(err, "Critical error occurred")

// Package-level functions (using default logger)
logger.Info("Application started")
logger.Error("Startup failed")
logger.Success("Database connected")
```

**Log Levels:**
- `DEBUG` 🔍 - Detailed debug information
- `INFO` ℹ️ - General information
- `WARN` ⚠️ - Warning messages
- `ERROR` ❌ - Error messages
- `FATAL` 💀 - Fatal errors (exits program)
- `SUCCESS` ✅ - Success messages

---

### 3. config - Configuration Management

Centralized configuration loading dari environment variables dengan validation.

**Usage:**

```go
import "github.com/ubaidillahfaris/whatsapp.git/internal/pkg/config"

// Load configuration (usually in main.go)
cfg, err := config.Load()
if err != nil {
    log.Fatal("Failed to load config: %v", err)
}

// Access configuration
serverPort := cfg.Server.Port
mongoURI := cfg.MongoDB.URI
jwtSecret := cfg.JWT.Secret
storesDir := cfg.WhatsApp.StoresDir

// Get singleton instance
cfg := config.Get()

// Helper functions
if config.IsDevelopment() {
    // Development-specific code
}

if config.IsProduction() {
    // Production-specific code
}
```

**Environment Variables:**

```bash
# Server
PORT=3000
ENVIRONMENT=development

# MongoDB
MONGO_USER=admin
MONGO_PASS=password
MONGO_HOST=localhost:27017
MONGO_PORT=27017
MONGO_DB=qr_db

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRES_MIN=60

# WhatsApp
WHATSAPP_STORES_DIR=./stores
WHATSAPP_UPLOADS_DIR=./uploads/whatsapp
WHATSAPP_MAX_CONCURRENCY=10

# CORS
CORS_ALLOWED_ORIGIN=http://localhost:5173
CORS_MAX_AGE=43200
```

---

### 4. validator - Input Validation

Input validation menggunakan struct tags dengan custom validators untuk WhatsApp.

**Usage:**

```go
import "github.com/ubaidillahfaris/whatsapp.git/internal/pkg/validator"

// Define request struct
type SendMessageRequest struct {
    To      string `json:"to" validate:"required,whatsapp_jid"`
    Message string `json:"message" validate:"required"`
}

// Validate struct
req := &SendMessageRequest{
    To: "6281234567890@s.whatsapp.net",
    Message: "Hello!",
}

if err := validator.Validate(req); err != nil {
    // Handle validation error
    // Output: "validation failed: To must be a valid WhatsApp JID"
}

// Validate single variable
jid := "6281234567890@s.whatsapp.net"
if err := validator.ValidateVar(jid, "required,whatsapp_jid"); err != nil {
    // Handle error
}

// Helper functions
if validator.ValidateWhatsAppJID(jid) {
    // Valid JID
}

if validator.ValidateDeviceName(name) {
    // Valid device name
}

// Get pagination params
page, limit, err := validator.GetPaginationParams(pageNum, limitNum)
```

**Custom Validators:**
- `whatsapp_jid` - Validates WhatsApp JID format (individual or group)
- `device_name` - Validates device name (alphanumeric, dash, underscore only)

**Built-in Validators:**
- `required` - Field is required
- `email` - Valid email address
- `min` - Minimum length/value
- `max` - Maximum length/value
- `len` - Exact length
- `url` - Valid URL
- `oneof` - Value must be one of specified options

---

## Migration Guide

### Migrating from Current Code

1. **Error Handling:**

Before:
```go
c.JSON(500, gin.H{"error": "Database error"})
```

After:
```go
import apperrors "github.com/ubaidillahfaris/whatsapp.git/internal/pkg/errors"

err := apperrors.NewDatabaseError("Failed to fetch device", dbErr)
c.JSON(err.StatusCode, gin.H{"error": err.Message, "type": err.Type})
```

2. **Logging:**

Before:
```go
fmt.Println("🟢 Device connected:", deviceName)
```

After:
```go
log := logger.New("WhatsApp")
log.WithField("device", deviceName).Success("Device connected")
```

3. **Configuration:**

Before:
```go
port := os.Getenv("PORT")
if port == "" {
    port = "3000"
}
```

After:
```go
cfg := config.Get()
port := cfg.Server.Port
```

4. **Validation:**

Before:
```go
if message == "" {
    c.JSON(400, gin.H{"error": "Message is required"})
    return
}
```

After:
```go
type Request struct {
    Message string `json:"message" validate:"required"`
}
req := &Request{}
if err := c.ShouldBindJSON(req); err != nil {
    c.JSON(400, gin.H{"error": err.Error()})
    return
}
if err := validator.Validate(req); err != nil {
    c.JSON(400, gin.H{"error": err.Error()})
    return
}
```

---

## Best Practices

1. **Always use custom error types** untuk error handling yang konsisten
2. **Use structured logging** dengan fields instead of string concatenation
3. **Load config once** di startup, use `config.Get()` di tempat lain
4. **Validate all inputs** menggunakan struct tags
5. **Use logger dengan prefix** untuk setiap module/service
6. **Add error context** dengan `WithDetails()` untuk debugging

---

## Testing

Semua packages ini dirancang untuk mudah di-test:

```go
// Mock logger for testing
testLogger := logger.New("Test")
testLogger.SetLevel(logger.ERROR) // Suppress logs in tests

// Mock config
os.Setenv("PORT", "3000")
cfg, _ := config.Load()

// Test validators
err := validator.Validate(testStruct)
assert.NoError(t, err)
```
