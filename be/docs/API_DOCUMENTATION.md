# GO-Wa API Documentation

Complete API reference for the WhatsApp API backend.

**Base URL:** `http://localhost:3000`

---

## Authentication

### Register User
**POST** `/auth/register`

Creates a new user account.

**Request Body:**
```json
{
  "username": "string",
  "password": "string"
}
```

**Response:**
```json
{
  "message": "User registered successfully",
  "data": {
    "user": {
      "username": "string"
    }
  }
}
```

---

### Login
**POST** `/auth/login`

Authenticates a user and returns a JWT token.

**Request Body:**
```json
{
  "username": "string",
  "password": "string"
}
```

**Response:**
```json
{
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "username": "string"
    }
  }
}
```

---

### Check Auth
**GET** `/auth/check`

Verifies if the current JWT token is valid.

**Headers:**
```
Authorization: Bearer <token>
```

**Response:**
```json
{
  "message": "Authenticated",
  "data": {
    "user": {
      "username": "string"
    }
  }
}
```

---

## Devices

All device endpoints require JWT authentication.

**Headers:**
```
Authorization: Bearer <token>
```

### List Devices
**GET** `/devices`

Retrieves all devices for the authenticated user.

**Response:**
```json
{
  "message": "Devices retrieved successfully",
  "data": [
    {
      "_id": "string",
      "name": "string",
      "owner": "string",
      "status": "active|inactive|disconnected",
      "createdAt": "2024-01-01T00:00:00Z",
      "updatedAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

---

### Get Device
**GET** `/devices/:id`

Retrieves a specific device by ID.

**Response:**
```json
{
  "message": "Device retrieved successfully",
  "data": {
    "_id": "string",
    "name": "string",
    "owner": "string",
    "status": "active",
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  }
}
```

---

### Create Device
**POST** `/devices`

Creates a new WhatsApp device.

**Request Body:**
```json
{
  "name": "My Device"
}
```

**Response:**
```json
{
  "message": "Device created successfully",
  "data": {
    "_id": "string",
    "name": "My Device",
    "owner": "username",
    "status": "inactive",
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  }
}
```

---

### Update Device
**PUT** `/devices/:id`

Updates device properties.

**Request Body:**
```json
{
  "name": "Updated Name",
  "status": "active"
}
```

**Response:**
```json
{
  "message": "Device updated successfully",
  "data": {
    "_id": "string",
    "name": "Updated Name",
    "owner": "username",
    "status": "active",
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  }
}
```

---

### Delete Device
**DELETE** `/devices/:id`

Deletes a device and disconnects its WhatsApp session.

**Response:**
```json
{
  "message": "Device deleted successfully"
}
```

---

## WhatsApp Operations

### Get QR Code
**GET** `/whatsapp/:device/qrcode`

Generates a QR code for WhatsApp authentication.

**Parameters:**
- `:device` - Device name

**Response:**
```json
{
  "message": "QR code generated",
  "data": {
    "qr": "2@abc123...",
    "device": "device_name"
  }
}
```

---

### Disconnect Device
**GET** `/whatsapp/:device/disconnect`

Disconnects a WhatsApp device session.

**Parameters:**
- `:device` - Device name

**Response:**
```json
{
  "message": "Device disconnected successfully"
}
```

---

### List Contacts
**GET** `/whatsapp/:device/contacts`

Retrieves all contacts for a device.

**Parameters:**
- `:device` - Device name

**Response:**
```json
{
  "message": "Contacts retrieved successfully",
  "data": [
    {
      "jid": "1234567890@s.whatsapp.net",
      "name": "Contact Name",
      "phone": "+1234567890"
    }
  ]
}
```

---

### List Groups
**GET** `/whatsapp/:device/groups`

Retrieves all groups for a device.

**Parameters:**
- `:device` - Device name

**Response:**
```json
{
  "message": "Groups retrieved successfully",
  "data": [
    {
      "jid": "1234567890-1234567890@g.us",
      "name": "Group Name",
      "participants": 10
    }
  ]
}
```

---

## Send Message

### Send Message
**POST** `/send_message/:device`

Sends a WhatsApp message (text or media).

**Parameters:**
- `:device` - Device name

**Request Body (multipart/form-data):**
```
receiver: string (phone number or group JID)
message: string
receiverType: "individual" | "group"
file: File (optional - for media messages)
```

**Example cURL:**
```bash
curl -X POST http://localhost:3000/send_message/device1 \
  -F "receiver=1234567890" \
  -F "message=Hello World" \
  -F "receiverType=individual" \
  -F "file=@/path/to/image.jpg"
```

**Response:**
```json
{
  "message": "Message sent successfully"
}
```

---

## API Keys

All API key endpoints require JWT authentication.

**Headers:**
```
Authorization: Bearer <token>
```

### Generate API Key
**POST** `/api-keys`

Generates a new API key for programmatic access.

**Request Body:**
```json
{
  "name": "Production Frontend",
  "permissions": [
    {
      "resource": "messages",
      "actions": ["read", "write"]
    },
    {
      "resource": "devices",
      "actions": ["read"]
    }
  ],
  "rate_limit": 100,
  "expires_in": 365
}
```

**Response:**
```json
{
  "message": "API key generated successfully",
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "key": "a1b2c3d4e5f6...128chars",
    "name": "Production Frontend",
    "owner": "username",
    "permissions": [...],
    "status": "active",
    "rate_limit": 100,
    "last_used_at": null,
    "expires_at": "2025-11-20T00:00:00Z",
    "created_at": "2024-11-20T00:00:00Z",
    "updated_at": "2024-11-20T00:00:00Z"
  }
}
```

---

### List API Keys
**GET** `/api-keys?limit=50&offset=0`

Lists all API keys for the authenticated user.

**Query Parameters:**
- `limit` (optional) - Number of keys to return (default: 50)
- `offset` (optional) - Number of keys to skip (default: 0)

**Response:**
```json
{
  "message": "API keys retrieved successfully",
  "data": {
    "keys": [
      {
        "id": "507f1f77bcf86cd799439011",
        "key": "...last8chars",
        "name": "Production Frontend",
        "owner": "username",
        "permissions": [...],
        "status": "active",
        "rate_limit": 100,
        "last_used_at": "2024-11-19T12:00:00Z",
        "expires_at": "2025-11-20T00:00:00Z",
        "created_at": "2024-11-20T00:00:00Z",
        "updated_at": "2024-11-20T00:00:00Z"
      }
    ],
    "total": 5,
    "limit": 50,
    "offset": 0
  }
}
```

---

### Get API Key
**GET** `/api-keys/:id`

Retrieves a specific API key by ID.

**Response:**
```json
{
  "message": "API key retrieved successfully",
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "key": "...last8chars",
    "name": "Production Frontend",
    ...
  }
}
```

---

### Update API Key
**PUT** `/api-keys/:id`

Updates an API key's properties.

**Request Body:**
```json
{
  "name": "Updated Name",
  "rate_limit": 200,
  "status": "inactive",
  "permissions": [
    {
      "resource": "*",
      "actions": ["*"]
    }
  ]
}
```

**Response:**
```json
{
  "message": "API key updated successfully",
  "data": {
    // Updated API key object
  }
}
```

---

### Revoke API Key
**DELETE** `/api-keys/:id`

Permanently deletes an API key.

**Response:**
```json
{
  "message": "API key revoked successfully"
}
```

---

## Quick Response

### Get All Quick Responses
**GET** `/quick_response/`

Retrieves all quick response messages.

**Response:**
```json
{
  "message": "Quick responses retrieved",
  "data": [
    {
      "_id": "string",
      "message": "string",
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

---

### Delete Quick Response
**DELETE** `/quick_response/:id`

Deletes a quick response by ID.

**Response:**
```json
{
  "message": "Quick response deleted successfully"
}
```

---

## Additional Endpoints

### Health Check
**GET** `/health`

Health check endpoint for monitoring.

**Response:**
```json
{
  "status": "ok"
}
```

---

### Ping (Token Verification)
**POST** `/ping`

Verifies JWT token validity using RSA public key.

**Headers:**
```
Authorization: Bearer <token>
```

**Response:**
```json
{
  "message": "ok",
  "username": "string"
}
```

---

### Sync App
**POST** `/sync/app`

Syncs application data (JWT protected).

**Headers:**
```
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  // Sync data
}
```

**Response:**
```json
{
  "message": "Sync successful"
}
```

---

## Using API Keys for Authentication

Instead of JWT tokens, you can use API keys for authentication:

**Headers:**
```
X-API-Key: your-api-key-here
```

**Example:**
```bash
curl -H "X-API-Key: a1b2c3d4e5f6..." \
  http://localhost:3000/whatsapp/device1/contacts
```

API keys can be used as an alternative to JWT tokens for most endpoints (except API key management endpoints which require JWT).

---

## Error Responses

All endpoints follow a consistent error format:

```json
{
  "error": "Error message",
  "details": {
    "field": "Additional error details"
  }
}
```

**Common HTTP Status Codes:**
- `200 OK` - Successful request
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid request data
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

---

## Frontend Integration

The frontend uses Vite proxy configuration to avoid CORS issues:

```typescript
// vite.config.ts
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:3000',
      changeOrigin: true,
      rewrite: (path) => path.replace(/^\/api/, ''),
    },
  },
}
```

This allows frontend to call `/api/auth/login` which proxies to `http://localhost:3000/auth/login`.

---

## Rate Limiting

API keys support rate limiting configuration (requests per minute). Set `rate_limit` when generating keys:

```json
{
  "name": "My Key",
  "rate_limit": 100  // 100 requests per minute
}
```

**Note:** Rate limit enforcement requires additional middleware (future enhancement).

---

## Permissions System

API keys support granular permissions:

**Format:**
```json
{
  "resource": "messages",  // "devices", "messages", "whatsapp", "*"
  "actions": ["read", "write", "delete"]  // or ["*"] for all
}
```

**Examples:**

Read-only access:
```json
{
  "permissions": [
    { "resource": "*", "actions": ["read"] }
  ]
}
```

Specific resources:
```json
{
  "permissions": [
    { "resource": "messages", "actions": ["read", "write"] },
    { "resource": "devices", "actions": ["read"] }
  ]
}
```

Full access (default):
```json
{
  "permissions": [
    { "resource": "*", "actions": ["*"] }
  ]
}
```

---

## WebSocket Support

**Status:** Not yet implemented

Future feature for real-time updates:
- Message delivery status
- Device connection status
- Incoming messages

---

## Best Practices

1. **Use API Keys for Frontend** - More secure than storing JWT tokens
2. **Set Expiration** - Always set `expires_in` for production keys
3. **Rotate Keys** - Regularly rotate API keys (every 90-365 days)
4. **Minimum Permissions** - Grant only required permissions
5. **Monitor Usage** - Check `last_used_at` to identify inactive keys
6. **Secure Storage** - Never commit keys to version control

---

## Changelog

### v1.0.0 (Current)
- Initial API release
- JWT authentication
- API key authentication
- Device management
- WhatsApp operations (QR, contacts, groups, send message)
- Quick response system
- Health monitoring
