# GO-Wa Architecture Documentation

Clean Architecture implementation untuk WhatsApp API menggunakan whatsmeow library.

## 📐 Arsitektur Overview

Aplikasi ini menggunakan **Clean Architecture** dengan **Dependency Inversion Principle** untuk memastikan:
- ✅ Testability
- ✅ Maintainability
- ✅ Extensibility
- ✅ Clear Separation of Concerns

```
┌─────────────────────────────────────────────────────────────┐
│                  Presentation Layer (HTTP)                   │
│                    handlers/, routes/                        │
└────────────────────┬────────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────────┐
│              Application Layer (Use Cases)                   │
│           internal/core/usecases/{whatsapp,device,message}   │
└─────┬──────────────────────────────────────────────┬────────┘
      │                                              │
┌─────▼──────────────────┐            ┌──────────────▼────────┐
│   Domain Layer         │            │  Ports (Interfaces)   │
│  internal/core/domain  │◄───────────┤ internal/core/ports   │
│  - Entities            │            │  - Repository         │
│  - Interfaces          │            │  - Service            │
└────────────────────────┘            └───────────────────────┘
                                                  ▲
                    ┌────────────────────────────┬┘
                    │                            │
┌───────────────────▼─────────┐   ┌──────────────▼───────────┐
│  Infrastructure Layer        │   │  Adapters Layer          │
│  internal/pkg/               │   │  internal/adapters/      │
│  - errors                    │   │  - whatsapp (whatsmeow)  │
│  - logger                    │   │  - repositories (mongo)  │
│  - config                    │   │                          │
│  - validator                 │   │                          │
└──────────────────────────────┘   └──────────────────────────┘
```

---

## 📁 Struktur Direktori

```
GO-Wa/
├── cmd/
│   └── api/
│       └── main.go                 # Entry point (future)
├── internal/
│   ├── core/                       # BUSINESS LOGIC (DOMAIN)
│   │   ├── domain/                 # Entities & Core Interfaces
│   │   │   ├── whatsapp.go        # WhatsApp entities
│   │   │   ├── message.go         # Message entities
│   │   │   └── device.go          # Device entities
│   │   ├── ports/                  # Interface definitions
│   │   │   ├── whatsapp_repository.go
│   │   │   ├── whatsapp_service.go
│   │   │   └── device_repository.go
│   │   └── usecases/              # Business Use Cases
│   │       ├── whatsapp/          # WhatsApp operations
│   │       ├── message/           # Message processing
│   │       └── device/            # Device management
│   ├── adapters/                   # INFRASTRUCTURE IMPLEMENTATIONS
│   │   ├── whatsapp/              # WhatsApp adapter (whatsmeow)
│   │   │   ├── client.go          # Client implementation
│   │   │   ├── manager.go         # Multi-device manager
│   │   │   ├── event_handler.go   # Event handling
│   │   │   └── service.go         # Service implementation
│   │   └── repositories/          # Database implementations
│   │       └── device_mongo_repository.go
│   ├── modules/                    # DOMAIN-SPECIFIC MODULES
│   │   └── quickresponse/         # Field work reporting module
│   │       ├── domain/            # QR domain entities
│   │       ├── repository/        # QR repository
│   │       ├── parser.go          # Message parser
│   │       └── processor.go       # Message processor
│   ├── pkg/                        # SHARED UTILITIES
│   │   ├── errors/                # Error handling
│   │   ├── logger/                # Logging
│   │   ├── config/                # Configuration
│   │   └── validator/             # Input validation
│   └── app/                        # APPLICATION SETUP
│       └── container.go           # Dependency injection
├── handlers/                       # HTTP Handlers (legacy - to be refactored)
├── routes/                         # Route definitions (legacy)
├── models/                         # Data models (legacy)
├── services/                       # Services (legacy)
├── db/                            # Database (legacy)
└── main.go                         # Current entry point
```

---

## 🏗️ Layer Breakdown

### 1. **Domain Layer** (`internal/core/domain/`)

**Pure business logic** - tidak ada dependency ke infrastructure.

**Entities:**
- `WhatsAppSession` - Sesi WhatsApp device
- `WhatsAppMessage` - Message yang dikirim/diterima
- `WhatsAppContact` - Kontak WhatsApp
- `WhatsAppGroup` - Group WhatsApp
- `Device` - Device configuration
- `IncomingMessage` - Message yang masuk untuk processing

**Interfaces:**
- `WhatsAppClientInterface` - Contract untuk WhatsApp client operations
- `WhatsAppManagerInterface` - Contract untuk multi-device management
- `WhatsAppEventHandler` - Contract untuk event handling
- `MessageProcessor` - Contract untuk message processing
- `MessageProcessorRegistry` - Contract untuk managing processors

---

### 2. **Ports Layer** (`internal/core/ports/`)

**Interface definitions** untuk abstraksi repository dan service.

Implements **Dependency Inversion Principle**:
- Core tidak depend ke infrastructure
- Infrastructure depend ke core melalui interfaces

**Repositories:**
- `DeviceRepository` - Device persistence
- `WhatsAppSessionRepository` - Session persistence
- `WhatsAppMessageRepository` - Message persistence

**Services:**
- `WhatsAppService` - WhatsApp business operations

---

### 3. **Use Cases Layer** (`internal/core/usecases/`)

**Application business rules** - orchestrate domain entities.

**WhatsApp Use Cases:**
- `ConnectUseCase` - Connect device
- `DisconnectUseCase` - Disconnect device
- `GetQRCodeUseCase` - Generate QR code
- `SendMessageUseCase` - Send messages
- `ListContactsUseCase` - List contacts
- `ListGroupsUseCase` - List groups

**Device Use Cases:**
- `CreateDeviceUseCase` - Create new device
- `GetDeviceUseCase` - Get device by ID/name
- `ListDevicesUseCase` - List devices with pagination
- `UpdateDeviceUseCase` - Update device
- `DeleteDeviceUseCase` - Delete device

**Message Use Cases:**
- `ProcessMessageUseCase` - Process incoming messages

---

### 4. **Adapters Layer** (`internal/adapters/`)

**Infrastructure implementations** - implement interfaces dari ports.

**WhatsApp Adapter** (`internal/adapters/whatsapp/`):
- `Client` - Whatsmeow client implementation
- `Manager` - Multi-device manager
- `EventHandler` - Event handling dengan message registry
- `Service` - Service implementation

**Repository Adapters** (`internal/adapters/repositories/`):
- `DeviceMongoRepository` - MongoDB implementation untuk devices

---

### 5. **Modules Layer** (`internal/modules/`)

**Domain-specific modules** - pluggable business modules.

**Quick Response Module** (`internal/modules/quickresponse/`):
- Domain-specific untuk irrigation field work reporting
- Parser untuk structured messages
- Processor implements `MessageProcessor` interface
- MongoDB repository
- **Completely isolated** - bisa dihapus tanpa affect core

**Adding New Modules:**
1. Create directory di `internal/modules/{module-name}/`
2. Implement `MessageProcessor` interface
3. Register ke `MessageProcessorRegistry`
4. Done! ✅

---

### 6. **Infrastructure Layer** (`internal/pkg/`)

**Shared utilities** - reusable across application.

**Components:**
- **errors**: Custom error types dengan HTTP mapping
- **logger**: Structured logging dengan emoji
- **config**: Environment-based configuration
- **validator**: Input validation dengan custom rules

---

### 7. **Application Layer** (`internal/app/`)

**Dependency injection & initialization**.

**Container** (`container.go`):
- Wires all dependencies
- Initializes all components in correct order
- Provides graceful shutdown

**Initialization Order:**
1. Load configuration
2. Connect to MongoDB
3. Initialize repositories
4. Initialize message processing
5. Initialize WhatsApp components
6. Initialize use cases

---

## 🔄 Data Flow

### **Sending a Message:**

```
HTTP Request
   │
   ▼
Handler (Presentation)
   │
   ▼
WhatsAppService (Application)
   │
   ▼
SendMessageUseCase (Business Logic)
   │
   ├─► Validate input (validator)
   ├─► Get WhatsApp client (Manager)
   ├─► Check connection status
   │
   ▼
WhatsAppClient (Infrastructure/whatsmeow)
   │
   ▼
WhatsApp Servers
```

### **Processing Incoming Message:**

```
WhatsApp Servers
   │
   ▼
WhatsAppClient (whatsmeow)
   │
   ▼
EventHandler.OnMessage
   │
   ▼
MessageProcessorRegistry
   │
   ├─► Check each registered processor
   │   └─► Can this processor handle the message?
   │
   ▼
QuickResponseProcessor (if applicable)
   │
   ├─► Parse message
   ├─► Validate
   ├─► Save to MongoDB
   │
   ▼
Done
```

---

## 🎯 Key Patterns & Principles

### **1. Dependency Inversion Principle**
- Core defines interfaces
- Infrastructure implements interfaces
- Dependencies point inward (toward core)

### **2. Single Responsibility Principle**
- Each use case handles ONE business operation
- Each repository handles ONE entity persistence
- Each processor handles ONE message type

### **3. Open/Closed Principle**
- Add new message processors without modifying core
- Add new repositories without changing use cases
- Extend functionality through composition

### **4. Interface Segregation Principle**
- Small, focused interfaces
- Clients depend only on methods they use

### **5. Repository Pattern**
- Abstract data access
- Consistent interface for different storage backends
- Easy to mock for testing

### **6. Strategy Pattern**
- `MessageProcessor` interface
- Different processors for different message types
- Runtime selection based on message content

---

## 🧪 Testing Strategy

### **Unit Testing**

**Use Cases:**
```go
// Mock dependencies
mockRepo := &MockDeviceRepository{}
mockManager := &MockWhatsAppManager{}

// Test use case
useCase := device.NewCreateDeviceUseCase(mockRepo)
device, err := useCase.Execute(ctx, request)

// Assert
assert.NoError(t, err)
assert.Equal(t, "device-1", device.Name)
```

**Domain Logic:**
```go
// Pure business logic - no mocking needed
parser := quickresponse.NewParser()
qr := parser.Parse(message)

assert.True(t, parser.IsValid(qr))
```

### **Integration Testing**

**Repository Tests:**
```go
// Use testcontainers for real MongoDB
mongoContainer := startMongoContainer(t)
repo := repositories.NewDeviceMongoRepository(mongoContainer.DB)

device, err := repo.Create(ctx, testDevice)
assert.NoError(t, err)
```

### **E2E Testing**

**Full Flow:**
```go
// Start application with test container
app := setupTestApp(t)

// Test full request flow
resp := app.Post("/devices", createDeviceRequest)
assert.Equal(t, 201, resp.StatusCode)

// Verify WhatsApp client created
assert.True(t, app.WhatsAppManager.HasClient("test-device"))
```

---

## 📊 Migration Guide

### **From Legacy to New Architecture**

#### **Phase 1: Foundation** ✅
- [x] Custom error handling
- [x] Structured logger
- [x] Config management
- [x] Input validator

#### **Phase 2: WhatsApp Core** ✅
- [x] Domain entities & interfaces
- [x] Use cases
- [x] WhatsApp adapters (whatsmeow)
- [x] Event handling

#### **Phase 3: Message Processing** ✅
- [x] Message processor registry
- [x] Quick Response module
- [x] Parser & processor

#### **Phase 4: Device Management** ✅
- [x] Device domain & repository
- [x] Device use cases
- [x] MongoDB implementation

#### **Phase 5: Application Setup** ✅
- [x] Dependency injection container
- [ ] Update HTTP handlers
- [ ] Migrate main.go

#### **Phase 6: HTTP Layer** (Next)
- [ ] Create new HTTP handlers using use cases
- [ ] Update routes to use new handlers
- [ ] Add middleware for error handling
- [ ] API documentation

#### **Phase 7: Testing & Documentation** (Final)
- [ ] Unit tests for use cases
- [ ] Integration tests for repositories
- [ ] E2E tests for full flows
- [ ] API documentation (OpenAPI)
- [ ] Deployment guide

---

## 🚀 Usage Examples

### **Creating a Device**

```go
// Using dependency injection container
container, _ := app.NewContainer(context.Background())

// Execute use case
device, err := container.CreateDeviceUC.Execute(ctx, domain.CreateDeviceRequest{
    Name:        "office-wa",
    Owner:       "admin@company.com",
    Description: "Office WhatsApp device",
})
```

### **Sending a Message**

```go
// Through WhatsApp service
err := container.WhatsAppService.SendMessage(ctx, domain.SendMessageParams{
    DeviceName:   "office-wa",
    To:           "628123456789@s.whatsapp.net",
    Message:      "Hello from Clean Architecture!",
    ReceiverType: domain.ReceiverIndividual,
    MessageType:  domain.MessageTypeText,
})
```

### **Adding a Custom Message Processor**

```go
// 1. Implement MessageProcessor interface
type OrderProcessor struct {
    orderRepo OrderRepository
}

func (p *OrderProcessor) CanProcess(msg domain.IncomingMessage) bool {
    return strings.Contains(msg.Content, "ORDER:")
}

func (p *OrderProcessor) Process(msg domain.IncomingMessage) error {
    // Parse order and save to database
    order := parseOrder(msg.Content)
    return p.orderRepo.Save(order)
}

func (p *OrderProcessor) Priority() int {
    return 50 // Medium priority
}

// 2. Register to container
container.MessageRegistry.Register(orderProcessor)

// Done! Now all incoming messages with "ORDER:" will be processed
```

---

## 🔧 Configuration

Environment variables (`.env`):

```bash
# Server
PORT=3000
ENVIRONMENT=development

# MongoDB
MONGO_USER=admin
MONGO_PASS=password
MONGO_HOST=localhost:27017
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
```

---

## 🎓 Best Practices

### **DO ✅**
- Keep domain layer pure (no infrastructure dependencies)
- Use interfaces for all external dependencies
- Write tests for use cases
- Use custom error types consistently
- Log with structured fields
- Validate all inputs
- Use dependency injection

### **DON'T ❌**
- Import infrastructure packages in domain layer
- Put business logic in handlers
- Access database directly from handlers
- Use global state (except config)
- Panic in production code (use error returns)
- Skip validation
- Hardcode configuration

---

## 📚 Further Reading

- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) by Uncle Bob
- [Domain-Driven Design](https://martinfowler.com/tags/domain%20driven%20design.html)
- [Dependency Injection in Go](https://github.com/google/wire)
- [Whatsmeow Documentation](https://github.com/tulir/whatsmeow)

---

## 🤝 Contributing

When adding new features:

1. **Start with domain** - Define entities and interfaces
2. **Write use cases** - Implement business logic
3. **Create adapters** - Implement infrastructure
4. **Add tests** - Unit + integration tests
5. **Update docs** - Document your changes

---

## 📝 License

[Your License Here]
