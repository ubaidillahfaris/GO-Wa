import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api/auth'
import type { User, LoginCredentials, RegisterCredentials } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const token = ref<string | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const isAuthenticated = computed(() => !!token.value && !!user.value)

  function loadFromStorage() {
    const storedToken = localStorage.getItem('auth_token')
    const storedUser = localStorage.getItem('user')

    if (storedToken && storedUser) {
      token.value = storedToken
      user.value = JSON.parse(storedUser)
    }
  }

  async function login(credentials: LoginCredentials) {
    try {
      loading.value = true
      error.value = null

      const response = await authApi.login(credentials)

      if (response.data) {
        token.value = response.data.token
        user.value = response.data.user

        localStorage.setItem('auth_token', response.data.token)
        localStorage.setItem('user', JSON.stringify(response.data.user))
      }

      return true
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Login failed'
      return false
    } finally {
      loading.value = false
    }
  }

  async function register(credentials: RegisterCredentials) {
    try {
      loading.value = true
      error.value = null

      const response = await authApi.register(credentials)

      // After registration, auto-login
      if (response.data?.user) {
        return await login({
          username: credentials.username,
          password: credentials.password,
        })
      }

      return false
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Registration failed'
      return false
    } finally {
      loading.value = false
    }
  }

  async function checkAuth() {
    if (!token.value) return false

    try {
      const response = await authApi.checkAuth()
      if (response.data?.user) {
        user.value = response.data.user
        return true
      }
      return false
    } catch {
      logout()
      return false
    }
  }

  function logout() {
    user.value = null
    token.value = null
    localStorage.removeItem('auth_token')
    localStorage.removeItem('user')
  }

  // Load from storage on init
  loadFromStorage()

  return {
    user,
    token,
    loading,
    error,
    isAuthenticated,
    login,
    register,
    checkAuth,
    logout,
  }
})
