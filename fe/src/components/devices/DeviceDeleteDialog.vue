<script setup lang="ts">
import { ref } from 'vue'
import { devicesApi } from '@/api/devices'
import { useToast } from '@/composables/useToast'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import type { Device } from '@/types'

const props = defineProps<{
  open: boolean
  device: Device | null
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  'deleted': []
}>()

const toast = useToast()
const loading = ref(false)

async function handleDelete() {
  if (!props.device?._id) return

  try {
    loading.value = true
    await devicesApi.delete(props.device._id)

    toast.success('Device deleted successfully')
    emit('deleted')
    emit('update:open', false)
  } catch (error: any) {
    console.error('Failed to delete device:', error)
    toast.error(error.response?.data?.error || 'Failed to delete device')
  } finally {
    loading.value = false
  }
}

function handleClose() {
  if (!loading.value) {
    emit('update:open', false)
  }
}
</script>

<template>
  <Dialog :open="props.open" @update:open="handleClose">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Delete Device</DialogTitle>
        <DialogDescription>
          Are you sure you want to delete this device? This action cannot be undone.
        </DialogDescription>
      </DialogHeader>

      <div class="py-4">
        <p class="text-sm text-gray-600">
          Device: <span class="font-semibold">{{ device?.name }}</span>
        </p>
        <p class="text-sm text-gray-600">
          Owner: <span class="font-semibold">{{ device?.owner }}</span>
        </p>
      </div>

      <div class="flex justify-end gap-3">
        <Button variant="outline" @click="handleClose" :disabled="loading">
          Cancel
        </Button>
        <Button variant="destructive" @click="handleDelete" :disabled="loading">
          {{ loading ? 'Deleting...' : 'Delete Device' }}
        </Button>
      </div>
    </DialogContent>
  </Dialog>
</template>
