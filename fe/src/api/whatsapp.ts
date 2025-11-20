import apiClient from './client'
import type { Contact, Group, ApiResponse, MessagePayload } from '@/types'

export const whatsappApi = {
  async getQRCode(device: string): Promise<Blob> {
    const timestamp = new Date().getTime()
    const response = await apiClient.get(`/whatsapp/${device}/qrcode?t=${timestamp}`, {
      responseType: 'blob',
      headers: {
        'Cache-Control': 'no-cache, no-store, must-revalidate',
        'Pragma': 'no-cache',
        'Expires': '0'
      }
    })
    return response.data
  },

  async disconnect(device: string) {
    const response = await apiClient.get<ApiResponse>(`/whatsapp/${device}/disconnect`)
    return response.data
  },

  async getContacts(device: string) {
    const response = await apiClient.get<ApiResponse<Contact[]>>(`/whatsapp/${device}/contacts`)
    return response.data
  },

  async getGroups(device: string) {
    const response = await apiClient.get<ApiResponse<Group[]>>(`/whatsapp/${device}/groups`)
    return response.data
  },

  async sendMessage(device: string, payload: MessagePayload) {
    const formData = new FormData()
    formData.append('receiver', payload.receiver)
    formData.append('message', payload.message)
    formData.append('receiverType', payload.receiverType)

    if (payload.file) {
      formData.append('file', payload.file)
    }

    const response = await apiClient.post<ApiResponse>(
      `/send_message/${device}`,
      formData,
      {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      }
    )
    return response.data
  },
}
