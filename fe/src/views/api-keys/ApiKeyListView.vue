<script setup lang="ts">
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { apiKeysApi } from '@/api/apiKeys'
import { useToast } from '@/composables/useToast'
import { Button } from '@/components/ui/button'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Plus, Trash2, Copy, Play } from 'lucide-vue-next'
import ApiKeyCreateDialog from '@/components/api-keys/ApiKeyCreateDialog.vue'
import type { ApiKey } from '@/types'

const toast = useToast()
const apiKeys = ref<ApiKey[]>([])
const loading = ref(false)
const showCreateDialog = ref(false)
const testingKeys = ref<Set<string>>(new Set())

async function loadApiKeys() {
  try {
    loading.value = true
    const response = await apiKeysApi.list()
    if (response.data?.keys) {
      apiKeys.value = response.data.keys
    }
  } catch (error) {
    console.error('Failed to load API keys:', error)
    toast.error('Failed to load API keys')
  } finally {
    loading.value = false
  }
}

function openCreateDialog() {
  showCreateDialog.value = true
}

function handleApiKeyCreated() {
  loadApiKeys()
}

async function copyToClipboard(text: string) {
  try {
    await navigator.clipboard.writeText(text)
    toast.success('Copied to clipboard!')
  } catch (error) {
    console.error('Failed to copy:', error)
    toast.error('Failed to copy to clipboard')
  }
}

async function revokeKey(id: string) {
  if (!confirm('Are you sure you want to revoke this API key? This action cannot be undone.')) return

  try {
    await apiKeysApi.revoke(id)
    toast.success('API key revoked successfully')
    await loadApiKeys()
  } catch (error: any) {
    console.error('Failed to revoke API key:', error)
    toast.error(error.response?.data?.error || 'Failed to revoke API key')
  }
}

async function testApiKey(apiKey: ApiKey) {
  testingKeys.value.add(apiKey.id)

  try {
    const baseURL = import.meta.env.VITE_API_BASE_URL || '/api'

    // Make a test request to /api-keys/test endpoint using the API key
    const response = await axios.post(`${baseURL}/api-keys/test`, {}, {
      headers: {
        'X-API-Key': apiKey.key
      },
      timeout: 10000
    })

    if (response.status === 200 && response.data?.status === 'success') {
      const username = response.data?.data?.authenticated_as || 'unknown'
      toast.success(`API Key "${apiKey.name}" is valid! Authenticated as: ${username}`)
    } else {
      toast.error(`API Key test failed with status: ${response.status}`)
    }
  } catch (error: any) {
    console.error('Failed to test API key:', error)

    if (error.response?.status === 401) {
      toast.error('API Key is invalid or has been revoked')
    } else if (error.response?.status === 403) {
      toast.error('API Key does not have permission to access this resource')
    } else if (error.code === 'ECONNABORTED') {
      toast.error('Test request timed out')
    } else {
      toast.error(error.response?.data?.error || 'Failed to test API key')
    }
  } finally {
    testingKeys.value.delete(apiKey.id)
  }
}

onMounted(() => {
  loadApiKeys()
})
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-8">
      <div>
        <h2 class="text-3xl font-bold text-gray-900">API Keys</h2>
        <p class="text-gray-600 mt-1">Manage your API keys for programmatic access</p>
      </div>
      <Button @click="openCreateDialog">
        <Plus class="w-4 h-4 mr-2" />
        Generate API Key
      </Button>
    </div>

    <Card v-if="loading">
      <CardContent class="pt-6">
        <p class="text-center text-gray-600">Loading API keys...</p>
      </CardContent>
    </Card>

    <Card v-else-if="apiKeys.length === 0">
      <CardContent class="pt-6">
        <p class="text-center text-gray-600">No API keys found. Click "Generate API Key" to create one.</p>
      </CardContent>
    </Card>

    <div v-else class="space-y-4">
      <Card v-for="apiKey in apiKeys" :key="apiKey.id">
        <CardHeader>
          <div class="flex items-start justify-between">
            <div>
              <CardTitle>{{ apiKey.name }}</CardTitle>
              <CardDescription>
                <span :class="{
                  'text-green-600': apiKey.status === 'active',
                  'text-yellow-600': apiKey.status === 'inactive',
                  'text-red-600': apiKey.status === 'revoked',
                  'text-gray-600': apiKey.status === 'expired'
                }">
                  {{ apiKey.status }}
                </span>
              </CardDescription>
            </div>
            <Button
              @click="revokeKey(apiKey.id)"
              variant="ghost"
              size="icon"
              class="text-destructive hover:text-destructive"
            >
              <Trash2 class="w-4 h-4" />
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div class="space-y-3">
            <div class="flex items-start gap-2">
              <code class="flex-1 px-3 py-2 bg-gray-100 rounded text-sm font-mono break-all">
                {{ apiKey.key }}
              </code>
              <div class="flex gap-2">
                <Button
                  @click="testApiKey(apiKey)"
                  variant="outline"
                  size="icon"
                  :disabled="testingKeys.has(apiKey.id) || apiKey.status !== 'active'"
                  :title="apiKey.status !== 'active' ? 'Cannot test inactive or revoked keys' : 'Test API Key'"
                >
                  <Play class="w-4 h-4" />
                </Button>
                <Button
                  @click="copyToClipboard(apiKey.key)"
                  variant="outline"
                  size="icon"
                >
                  <Copy class="w-4 h-4" />
                </Button>
              </div>
            </div>

            <div class="grid grid-cols-2 gap-4 text-sm">
              <div>
                <p class="text-gray-600">Rate Limit</p>
                <p class="font-medium">{{ apiKey.rate_limit || 'Unlimited' }}</p>
              </div>
              <div>
                <p class="text-gray-600">Last Used</p>
                <p class="font-medium">
                  {{ apiKey.last_used_at ? new Date(apiKey.last_used_at).toLocaleString() : 'Never' }}
                </p>
              </div>
              <div>
                <p class="text-gray-600">Created</p>
                <p class="font-medium">{{ new Date(apiKey.created_at).toLocaleDateString() }}</p>
              </div>
              <div v-if="apiKey.expires_at">
                <p class="text-gray-600">Expires</p>
                <p class="font-medium">{{ new Date(apiKey.expires_at).toLocaleDateString() }}</p>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>

    <ApiKeyCreateDialog
      :open="showCreateDialog"
      @update:open="showCreateDialog = $event"
      @created="handleApiKeyCreated"
    />
  </div>
</template>
