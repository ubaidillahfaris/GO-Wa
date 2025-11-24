export interface User {
  username: string
  email?: string
}

export interface LoginCredentials {
  username: string
  password: string
}

export interface RegisterCredentials {
  username: string
  password: string
  confirm_password: string
  email?: string
}

export interface Device {
  _id: string
  name: string
  owner: string
  status: 'active' | 'inactive' | 'disconnected'
  createdAt: string
  updatedAt: string
}

export interface ApiKey {
  id: string
  key: string
  name: string
  owner: string
  permissions: ApiKeyPermission[]
  status: 'active' | 'inactive' | 'revoked' | 'expired'
  rate_limit: number
  last_used_at?: string
  expires_at?: string
  created_at: string
  updated_at: string
}

export interface ApiKeyPermission {
  resource: string
  actions: string[]
}

export interface CreateApiKeyRequest {
  name: string
  permissions?: ApiKeyPermission[]
  rate_limit?: number
  expires_in?: number
}

export interface UpdateApiKeyRequest {
  name?: string
  permissions?: ApiKeyPermission[]
  rate_limit?: number
  status?: 'active' | 'inactive' | 'revoked'
}

export interface MessagePayload {
  to: string
  message: string
  receiver_type: 'user' | 'group'
  message_type?: 'text' | 'file'
  typing?: boolean
  file?: File
  filename?: string
  caption?: string
}

export interface Contact {
  jid: string
  name: string
  phone: string
}

export interface Group {
  jid: string
  name: string
  participants: number
}

export interface ApiResponse<T = any> {
  message: string
  data?: T
  error?: string
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  limit: number
  offset: number
}

export interface ApiKeyPaginatedResponse {
  keys: ApiKey[]
  total: number
  limit: number
  offset: number
}
