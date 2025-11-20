<script setup lang="ts">
import { ref } from 'vue'
import { apiKeysApi } from '@/api/apiKeys'
import { useToast } from '@/composables/useToast'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

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

async function handleCreate() {
  if (!name.value.trim()) {
    toast.error('Name is required')
    return
  }

  try {
    loading.value = true
    await apiKeysApi.create({
      name: name.value,
      rate_limit: rateLimit.value,
      expires_in: expiresIn.value,
    })

    toast.success('API key generated successfully')
    emit('created')
    emit('update:open', false)

    // Reset form
    name.value = ''
    rateLimit.value = undefined
    expiresIn.value = undefined
  } catch (error: any) {
    console.error('Failed to create API key:', error)
    toast.error(error.response?.data?.error || 'Failed to create API key')
  } finally {
    loading.value = false
  }
}

function handleClose() {
  if (!loading.value) {
    emit('update:open', false)
    name.value = ''
    rateLimit.value = undefined
    expiresIn.value = undefined
  }
}
</script>

<template>
  <Dialog :open="props.open" @update:open="handleClose">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Generate New API Key</DialogTitle>
        <DialogDescription>
          Create a new API key for programmatic access to your account
        </DialogDescription>
      </DialogHeader>

      <div class="space-y-4 py-4">
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

      <div class="flex justify-end gap-3">
        <Button variant="outline" @click="handleClose" :disabled="loading">
          Cancel
        </Button>
        <Button @click="handleCreate" :disabled="loading">
          {{ loading ? 'Generating...' : 'Generate API Key' }}
        </Button>
      </div>
    </DialogContent>
  </Dialog>
</template>
