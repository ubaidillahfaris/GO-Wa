import apiClient from './client'
import type { LoginCredentials, RegisterCredentials, User, ApiResponse } from '@/types'

export const authApi = {
  async login(credentials: LoginCredentials) {
    const response = await apiClient.post<ApiResponse<{ token: string; user: User }>>('/auth/login', credentials)
    return response.data
  },

  async register(credentials: RegisterCredentials) {
    const response = await apiClient.post<ApiResponse<{ user: User }>>('/auth/register', credentials)
    return response.data
  },

  async checkAuth() {
    const response = await apiClient.get<ApiResponse<{ user: User }>>('/auth/check')
    return response.data
  },
}
