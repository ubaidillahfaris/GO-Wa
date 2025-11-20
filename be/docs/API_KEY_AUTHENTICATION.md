# API Key Authentication

This document explains how to use API key authentication in the WhatsApp API application.

## Overview

The API key authentication feature allows you to:
- Generate secure API keys for programmatic access
- Manage API keys (list, update, revoke)
- Use API keys instead of JWT tokens for authentication
- Set granular permissions and rate limits per key
- Track usage with last-used timestamps

## Architecture

The API key module follows Clean Architecture principles:

```
internal/
├── core/
│   ├── domain/
│   │   └── api_key.go              # Domain entities and repository interface
│   └── usecases/
│       └── apikey/
│           ├── generate_key.go     # Generate new API key
│           ├── list_keys.go        # List user's API keys
│           ├── revoke_key.go       # Revoke API key
│           ├── update_key.go       # Update API key properties
│           └── validate_key.go     # Validate API key
├── adapters/
│   └── repositories/
│       └── api_key_mongo_repository.go  # MongoDB implementation
└── app/
    └── container.go                # Dependency injection

handlers/
└── api_key_handler.go              # HTTP handlers

middlewares/
└── apikey.go                       # API key middleware
```

## API Endpoints

### 1. Generate API Key
**POST** `/api-keys`

Creates a new API key for the authenticated user.

**Authentication:** JWT required

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
    "expires_at": "2025-11-20T00:00:00Z",
    "created_at": "2024-11-20T00:00:00Z"
  }
}
```

### 2. List API Keys
**GET** `/api-keys?limit=50&offset=0`

Lists all API keys for the authenticated user.

**Authentication:** JWT required

**Query Parameters:**
- `limit` (optional): Number of keys to return (default: 50)
- `offset` (optional): Number of keys to skip (default: 0)

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
        "status": "active",
        "last_used_at": "2024-11-19T12:00:00Z",
        "created_at": "2024-11-20T00:00:00Z"
      }
    ],
    "total": 5,
    "limit": 50,
    "offset": 0
  }
}
```

**Note:** For security, API keys are masked in list responses (only last 8 characters shown).

### 3. Get API Key
**GET** `/api-keys/:id`

Retrieves a specific API key by ID.

**Authentication:** JWT required

**Response:**
```json
{
  "message": "API key retrieved successfully",
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "key": "...last8chars",
    "name": "Production Frontend",
    "owner": "username",
    "permissions": [...],
    "status": "active",
    "rate_limit": 100,
    "last_used_at": "2024-11-19T12:00:00Z",
    "expires_at": "2025-11-20T00:00:00Z",
    "created_at": "2024-11-20T00:00:00Z"
  }
}
```

### 4. Update API Key
**PUT** `/api-keys/:id`

Updates an API key's properties.

**Authentication:** JWT required

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

**Validation:**
- Cannot change expired keys
- Cannot manually set key to expired status
- Status must be one of: `active`, `inactive`, `revoked`

### 5. Revoke API Key
**DELETE** `/api-keys/:id`

Permanently deletes an API key.

**Authentication:** JWT required

**Response:**
```json
{
  "message": "API key revoked successfully"
}
```

## Using API Keys

### With HTTP Requests

Include the API key in the `X-API-Key` header:

```bash
curl -H "X-API-Key: your-api-key-here" \
  http://localhost:3000/whatsapp/device1/contacts
```

### With JavaScript/TypeScript

```typescript
const apiKey = 'your-api-key-here';

const response = await fetch('http://localhost:3000/whatsapp/device1/contacts', {
  headers: {
    'X-API-Key': apiKey
  }
});

const data = await response.json();
```

### With Axios

```typescript
import axios from 'axios';

const client = axios.create({
  baseURL: 'http://localhost:3000',
  headers: {
    'X-API-Key': 'your-api-key-here'
  }
});

const { data } = await client.get('/whatsapp/device1/contacts');
```

## Middleware Options

### 1. API Key Only
Requires a valid API key:

```go
r.GET("/endpoint", middlewares.APIKeyMiddleware(validateUC), handler)
```

### 2. API Key OR JWT
Accepts either authentication method:

```go
r.GET("/endpoint", middlewares.APIKeyOrJWTMiddleware(validateUC), handler)
```

### 3. API Key with Permissions
Requires API key with specific permissions:

```go
r.POST("/messages",
  middlewares.APIKeyWithPermissionMiddleware(validateUC, "messages", "write"),
  handler)
```

## Permissions System

### Permission Structure

```go
type APIKeyPermission struct {
    Resource string   // e.g., "devices", "messages", "whatsapp"
    Actions  []string // e.g., ["read", "write", "delete"]
}
```

### Wildcard Support

- `resource: "*"` - Access to all resources
- `actions: ["*"]` - All actions allowed
- Default (empty permissions) - Full access (backward compatibility)

### Example Permissions

**Read-only access:**
```json
{
  "permissions": [
    {
      "resource": "*",
      "actions": ["read"]
    }
  ]
}
```

**Specific resources:**
```json
{
  "permissions": [
    {
      "resource": "messages",
      "actions": ["read", "write"]
    },
    {
      "resource": "devices",
      "actions": ["read"]
    }
  ]
}
```

## Security Features

### 1. Key Generation
- 64-byte cryptographically secure random keys
- 128 hexadecimal characters (very high entropy)

### 2. Key Masking
- Keys are masked in list responses
- Only last 8 characters shown
- Full key only shown once during generation

### 3. Status Management
- Keys can be `active`, `inactive`, `revoked`, or `expired`
- Only active and non-expired keys can authenticate
- Expired status set automatically by system

### 4. Ownership Verification
- Users can only manage their own keys
- Unauthorized access attempts are logged

### 5. Usage Tracking
- Last used timestamp updated on each use
- Helps identify inactive keys
- Updated asynchronously (non-blocking)

## Rate Limiting

Rate limiting is per API key:

```json
{
  "rate_limit": 100  // Requests per minute (0 = unlimited)
}
```

**Note:** Rate limit enforcement is stored in the domain but requires additional middleware implementation (future enhancement).

## Expiration

Keys can expire after a specified number of days:

```json
{
  "expires_in": 365  // Days (0 = never expires)
}
```

Expired keys:
- Cannot authenticate
- Return "API key has expired" error
- Status automatically changed to `expired`
- Cannot be modified

## Migration from JWT to API Keys

### Phase 1: Generate API Keys
```bash
# Login with JWT
curl -X POST http://localhost:3000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user","password":"pass"}'

# Generate API key
curl -X POST http://localhost:3000/api-keys \
  -H "Authorization: Bearer <jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"My App","expires_in":365}'

# Save the returned API key
```

### Phase 2: Update Client Applications
Replace JWT authentication with API key:

```diff
- headers: { 'Authorization': 'Bearer ' + jwtToken }
+ headers: { 'X-API-Key': apiKey }
```

### Phase 3: Rotate Keys
```bash
# Revoke old key
curl -X DELETE http://localhost:3000/api-keys/{old-key-id} \
  -H "Authorization: Bearer <jwt-token>"

# Generate new key
curl -X POST http://localhost:3000/api-keys \
  -H "Authorization: Bearer <jwt-token>" \
  -d '{"name":"My App v2","expires_in":365}'
```

## Database Schema

API keys are stored in MongoDB collection `api_keys`:

```javascript
{
  _id: ObjectId("507f1f77bcf86cd799439011"),
  key: "a1b2c3d4e5f6...128chars",
  name: "Production Frontend",
  owner: "username",
  permissions: [
    { resource: "messages", actions: ["read", "write"] }
  ],
  status: "active",
  rate_limit: 100,
  last_used_at: ISODate("2024-11-19T12:00:00Z"),
  expires_at: ISODate("2025-11-20T00:00:00Z"),
  created_at: ISODate("2024-11-20T00:00:00Z"),
  updated_at: ISODate("2024-11-20T00:00:00Z")
}
```

### Indexes
- `key` (unique)
- `owner`
- `status`
- `created_at` (descending)
- `expires_at`

## Best Practices

1. **Use Descriptive Names**
   - "Production Frontend" ✅
   - "Key 1" ❌

2. **Set Appropriate Permissions**
   - Grant minimum required permissions
   - Use specific resources instead of wildcards when possible

3. **Implement Key Rotation**
   - Rotate keys periodically (e.g., every 90-365 days)
   - Revoke unused keys

4. **Monitor Usage**
   - Check `last_used_at` timestamp
   - Revoke keys that haven't been used in months

5. **Store Keys Securely**
   - Never commit keys to version control
   - Use environment variables or secret management
   - Store keys encrypted at rest

6. **Set Expiration**
   - Always set `expires_in` for production keys
   - Use shorter expiration for testing keys

7. **Handle Revocation Gracefully**
   - Implement retry logic with new key generation
   - Monitor for authentication errors

## Troubleshooting

### "API key is required"
- Check `X-API-Key` header is present
- Verify header name is correct (case-sensitive)

### "invalid API key"
- Verify key hasn't been revoked
- Check for typos in the key value
- Ensure key belongs to correct environment

### "API key has expired"
- Generate a new key
- Update application configuration

### "API key is not active"
- Check key status (may be `inactive` or `revoked`)
- Use GET `/api-keys/:id` to check status

### "insufficient permissions for this operation"
- Verify key has required permissions
- Use PUT `/api-keys/:id` to update permissions

## Future Enhancements

- [ ] Rate limit middleware enforcement
- [ ] API key usage analytics
- [ ] Automatic key rotation
- [ ] Webhook for key events
- [ ] IP whitelist per key
- [ ] Scoped keys for specific devices

## Example: Frontend Integration

See `examples/frontend_api_client.ts` (to be created) for a complete example of integrating API keys in a Vue.js frontend application.
