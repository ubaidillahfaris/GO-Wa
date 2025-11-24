<script setup lang="ts">
import { ref, watch } from 'vue'
import { whatsappApi } from '@/api/whatsapp'
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
  'connected': []
}>()

const toast = useToast()
const loading = ref(false)
const qrCodeUrl = ref<string | null>(null)
let checkConnectionInterval: number | null = null

watch(() => props.open, (isOpen) => {
  if (isOpen && props.device) {
    loadQRCode()
    startCheckingConnection()
  } else {
    // Cleanup
    stopCheckingConnection()
    if (qrCodeUrl.value) {
      URL.revokeObjectURL(qrCodeUrl.value)
      qrCodeUrl.value = null
    }
  }
})

function startCheckingConnection() {
  stopCheckingConnection()

  // Check connection status every 3 seconds
  checkConnectionInterval = window.setInterval(async () => {
    if (!props.device?._id || !qrCodeUrl.value) return

    try {
      const blob = await whatsappApi.getQRCode(props.device._id)

      // If response is JSON (already connected), stop checking
      if (blob.type === 'application/json') {
        const text = await blob.text()
        const response = JSON.parse(text)

        if (response.status === 'success' && response.message?.includes('already connected')) {
          stopCheckingConnection()
          toast.success('Device connected successfully!')
          emit('connected')
          emit('update:open', false)
        }
      }
    } catch (error) {
      // Silently fail - continue checking
    }
  }, 3000)
}

function stopCheckingConnection() {
  if (checkConnectionInterval) {
    clearInterval(checkConnectionInterval)
    checkConnectionInterval = null
  }
}

async function loadQRCode() {
  if (!props.device?._id) return

  try {
    loading.value = true

    // Cleanup previous QR code
    if (qrCodeUrl.value) {
      URL.revokeObjectURL(qrCodeUrl.value)
      qrCodeUrl.value = null
    }

    const blob = await whatsappApi.getQRCode(props.device._id)

    // Check if response is JSON (device already connected)
    if (blob.type === 'application/json') {
      const text = await blob.text()
      const response = JSON.parse(text)

      if (response.status === 'success' && response.message?.includes('already connected')) {
        toast.success('Device is already connected!')
        emit('connected')
        emit('update:open', false)
        return
      }
    }

    // Create object URL from blob
    qrCodeUrl.value = URL.createObjectURL(blob)

    toast.success('QR Code generated. Please scan to connect.')
  } catch (error: any) {
    console.error('Failed to load QR code:', error)

    // Check if error response indicates device is already connected
    if (error.response?.data?.message?.includes('already connected')) {
      toast.success('Device is already connected!')
      emit('connected')
      emit('update:open', false)
      return
    }

    toast.error(error.response?.data?.error || 'Failed to load QR code')
    emit('update:open', false)
  } finally {
    loading.value = false
  }
}

function handleClose() {
  stopCheckingConnection()
  if (qrCodeUrl.value) {
    URL.revokeObjectURL(qrCodeUrl.value)
    qrCodeUrl.value = null
  }
  emit('update:open', false)
}

function handleRefresh() {
  loadQRCode()
  startCheckingConnection()
}
</script>

<template>
  <Dialog :open="props.open" @update:open="handleClose">
    <DialogContent class="max-w-md">
      <DialogHeader>
        <DialogTitle>Connect WhatsApp Device</DialogTitle>
        <DialogDescription>
          Scan the QR code with your WhatsApp mobile app to connect
        </DialogDescription>
      </DialogHeader>

      <div class="py-4">
        <div v-if="loading" class="flex items-center justify-center h-64">
          <div class="text-center">
            <div class="inline-block h-8 w-8 animate-spin rounded-full border-4 border-solid border-current border-r-transparent align-[-0.125em] motion-reduce:animate-[spin_1.5s_linear_infinite]" />
            <p class="mt-4 text-sm text-gray-600">Generating QR Code...</p>
          </div>
        </div>

        <div v-else-if="qrCodeUrl" class="flex flex-col items-center">
          <div class="bg-white p-4 rounded-lg border">
            <img
              :src="qrCodeUrl"
              :key="qrCodeUrl"
              alt="WhatsApp QR Code"
              class="w-64 h-64 object-contain"
            />
          </div>
          <p class="mt-4 text-sm text-gray-600 text-center">
            Open WhatsApp on your phone and scan this QR code
          </p>
          <div class="mt-3 flex items-center gap-2 text-xs text-green-600">
            <div class="w-2 h-2 bg-green-600 rounded-full animate-pulse"></div>
            <span>Auto-detecting connection...</span>
          </div>
        </div>

        <div v-else class="flex items-center justify-center h-64">
          <p class="text-sm text-gray-600">Failed to load QR code</p>
        </div>
      </div>

      <div class="flex justify-end gap-3">
        <Button variant="outline" @click="handleClose">
          Close
        </Button>
        <Button @click="handleRefresh" :disabled="loading">
          Refresh QR Code
        </Button>
      </div>
    </DialogContent>
  </Dialog>
</template>
