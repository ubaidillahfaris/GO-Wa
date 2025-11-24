import apiClient from './client'
import type { ApiKey, CreateApiKeyRequest, UpdateApiKeyRequest, ApiResponse, ApiKeyPaginatedResponse } from '@/types'

export const apiKeysApi = {
  async list(limit = 50, offset = 0) {
    const response = await apiClient.get<ApiResponse<ApiKeyPaginatedResponse>>(`/api-keys?limit=${limit}&offset=${offset}`)
    return response.data
  },

  async get(id: string) {
    const response = await apiClient.get<ApiResponse<ApiKey>>(`/api-keys/${id}`)
    return response.data
  },

  async create(data: CreateApiKeyRequest) {
    const response = await apiClient.post<ApiResponse<ApiKey>>('/api-keys', data)
    return response.data
  },

  async update(id: string, data: UpdateApiKeyRequest) {
    const response = await apiClient.put<ApiResponse<ApiKey>>(`/api-keys/${id}`, data)
    return response.data
  },

  async revoke(id: string) {
    const response = await apiClient.delete<ApiResponse>(`/api-keys/${id}`)
    return response.data
  },
}
