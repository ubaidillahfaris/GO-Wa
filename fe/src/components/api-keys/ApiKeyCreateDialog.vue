<script setup lang="ts">
import { ref } from 'vue'
import { apiKeysApi } from '@/api/apiKeys'
import { useToast } from '@/composables/useToast'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Copy, AlertTriangle } from 'lucide-vue-next'
import type { ApiKey } from '@/types'

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  'created': []
}>()

const toast = useToast()
const loading = ref(false)
const name = ref('')
const rateLimit = ref<number | undefined>(undefined)
const expiresIn = ref<number | undefined>(undefined)
const createdKey = ref<ApiKey | null>(null)
const showSuccess = ref(false)

async function handleCreate() {
  if (!name.value.trim()) {
    toast.error('Name is required')
    return
  }

  try {
    loading.value = true
    const response = await apiKeysApi.create({
      name: name.value,
      rate_limit: rateLimit.value,
      expires_in: expiresIn.value,
    })

    // Store the created key to display
    if (response.data) {
      createdKey.value = response.data
      showSuccess.value = true
      toast.success('API key generated successfully')
    } else {
      toast.error('Failed to retrieve API key from response')
    }
  } catch (error: any) {
    console.error('Failed to create API key:', error)
    toast.error(error.response?.data?.error || 'Failed to create API key')
  } finally {
    loading.value = false
  }
}

async function copyApiKey() {
  if (!createdKey.value) return

  try {
    await navigator.clipboard.writeText(createdKey.value.key)
    toast.success('API key copied to clipboard!')
  } catch (error) {
    console.error('Failed to copy:', error)
    toast.error('Failed to copy to clipboard')
  }
}

function handleClose(open: boolean) {
  // Prevent closing if showing success screen or loading
  if (loading.value || showSuccess.value) {
    return
  }

  if (!open) {
    emit('update:open', false)
    resetForm()
  }
}

function handleDone() {
  emit('created')
  emit('update:open', false)
  resetForm()
}

function resetForm() {
  name.value = ''
  rateLimit.value = undefined
  expiresIn.value = undefined
  createdKey.value = null
  showSuccess.value = false
}
</script>

<template>
  <Dialog :open="props.open" @update:open="handleClose">
    <DialogContent
      :class="showSuccess ? 'max-w-2xl' : ''"
      @interactOutside="(e: Event) => { if (showSuccess) e.preventDefault() }"
      @escapeKeyDown="(e: Event) => { if (showSuccess) e.preventDefault() }"
    >
      <DialogHeader>
        <DialogTitle>
          {{ showSuccess ? 'API Key Created Successfully!' : 'Generate New API Key' }}
        </DialogTitle>
        <DialogDescription v-if="!showSuccess">
          Create a new API key for programmatic access to your account
        </DialogDescription>
      </DialogHeader>

      <!-- Form -->
      <div v-if="!showSuccess" class="space-y-4 py-4">
        <div class="space-y-2">
          <Label for="api-key-name">Name</Label>
          <Input
            id="api-key-name"
            v-model="name"
            placeholder="e.g., Production API Key"
            :disabled="loading"
            required
          />
          <p class="text-xs text-gray-500">A descriptive name for this API key</p>
        </div>

        <div class="space-y-2">
          <Label for="rate-limit">Rate Limit (optional)</Label>
          <Input
            id="rate-limit"
            v-model.number="rateLimit"
            type="number"
            placeholder="e.g., 1000"
            :disabled="loading"
            min="0"
          />
          <p class="text-xs text-gray-500">Maximum requests per hour (leave empty for unlimited)</p>
        </div>

        <div class="space-y-2">
          <Label for="expires-in">Expires In (optional)</Label>
          <Input
            id="expires-in"
            v-model.number="expiresIn"
            type="number"
            placeholder="e.g., 30"
            :disabled="loading"
            min="1"
          />
          <p class="text-xs text-gray-500">Number of days until expiration (leave empty for no expiration)</p>
        </div>

        <div class="bg-yellow-50 border border-yellow-200 rounded-md p-3">
          <p class="text-sm text-yellow-800">
            <strong>Warning:</strong> Make sure to copy your API key after creation.
            You won't be able to see it again!
          </p>
        </div>
      </div>

      <!-- Success View with Full API Key -->
      <div v-if="showSuccess && createdKey" class="space-y-4 py-4">
        <div class="bg-red-50 border-2 border-red-300 rounded-lg p-4">
          <div class="flex items-start gap-3">
            <AlertTriangle class="w-6 h-6 text-red-600 flex-shrink-0 mt-0.5" />
            <div class="flex-1">
              <h4 class="font-semibold text-red-900 mb-2">Important: Save Your API Key Now</h4>
              <p class="text-sm text-red-800 mb-3">
                This is the <strong>only time</strong> you'll be able to see the full API key.
                Copy it now and store it securely. Once you close this dialog, the key will be masked for security.
              </p>
            </div>
          </div>
        </div>

        <div class="space-y-2">
          <Label>API Key Name</Label>
          <p class="font-medium">{{ createdKey.name }}</p>
        </div>

        <div class="space-y-2">
          <Label>Your API Key</Label>
          <div class="flex items-start gap-2">
            <code class="flex-1 px-4 py-3 bg-gray-900 text-green-400 rounded font-mono text-sm break-all select-all">{{ createdKey.key }}</code>
            <Button
              @click="copyApiKey"
              variant="outline"
              size="icon"
              class="flex-shrink-0"
            >
              <Copy class="w-4 h-4" />
            </Button>
          </div>
          <p class="text-xs text-gray-500">Click the copy button or select and copy the key above</p>
        </div>

        <div class="bg-blue-50 border border-blue-200 rounded-md p-3">
          <p class="text-sm text-blue-800">
            <strong>Usage:</strong> Include this key in your API requests using the <code class="bg-blue-100 px-1 rounded">X-API-Key</code> header.
          </p>
        </div>
      </div>

      <!-- Actions -->
      <div class="flex justify-end gap-3">
        <Button v-if="!showSuccess" variant="outline" @click="handleClose" :disabled="loading">
          Cancel
        </Button>
        <Button v-if="!showSuccess" @click="handleCreate" :disabled="loading">
          {{ loading ? 'Generating...' : 'Generate API Key' }}
        </Button>
        <Button v-if="showSuccess" @click="handleDone" variant="default">
          I've Saved My Key
        </Button>
      </div>
    </DialogContent>
  </Dialog>
</template>
