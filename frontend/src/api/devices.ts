import apiClient from './client'
import type { Device, ApiResponse } from '@/types'

export const devicesApi = {
  async list() {
    const response = await apiClient.get<ApiResponse<Device[]>>('/devices')
    return response.data
  },

  async get(id: string) {
    const response = await apiClient.get<ApiResponse<Device>>(`/devices/${id}`)
    return response.data
  },

  async create(data: { name: string }) {
    const response = await apiClient.post<ApiResponse<Device>>('/devices', data)
    return response.data
  },

  async update(id: string, data: Partial<Device>) {
    const response = await apiClient.put<ApiResponse<Device>>(`/devices/${id}`, data)
    return response.data
  },

  async delete(id: string) {
    const response = await apiClient.delete<ApiResponse>(`/devices/${id}`)
    return response.data
  },
}
