<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { apiKeysApi } from '@/api/apiKeys'
import { Button } from '@/components/ui/button'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Plus, Trash2, Copy } from 'lucide-vue-next'
import type { ApiKey } from '@/types'

const apiKeys = ref<ApiKey[]>([])
const loading = ref(false)

async function loadApiKeys() {
  try {
    loading.value = true
    const response = await apiKeysApi.list()
    if (response.data?.keys) {
      apiKeys.value = response.data.keys
    }
  } catch (error) {
    console.error('Failed to load API keys:', error)
  } finally {
    loading.value = false
  }
}

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text)
  alert('Copied to clipboard!')
}

async function revokeKey(id: string) {
  if (!confirm('Are you sure you want to revoke this API key?')) return

  try {
    await apiKeysApi.revoke(id)
    await loadApiKeys()
  } catch (error) {
    console.error('Failed to revoke API key:', error)
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
      <Button>
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
            <div class="flex items-center gap-2">
              <code class="flex-1 px-3 py-2 bg-gray-100 rounded text-sm font-mono">
                {{ apiKey.key }}
              </code>
              <Button
                @click="copyToClipboard(apiKey.key)"
                variant="outline"
                size="icon"
              >
                <Copy class="w-4 h-4" />
              </Button>
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
  </div>
</template>
