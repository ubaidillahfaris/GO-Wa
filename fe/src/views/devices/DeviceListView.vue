<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { devicesApi } from '@/api/devices'
import { whatsappApi } from '@/api/whatsapp'
import { useToast } from '@/composables/useToast'
import { Button } from '@/components/ui/button'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Plus, Trash2, QrCode, Power, RefreshCw } from 'lucide-vue-next'
import DeviceCreateDialog from '@/components/devices/DeviceCreateDialog.vue'
import DeviceDeleteDialog from '@/components/devices/DeviceDeleteDialog.vue'
import QRCodeDialog from '@/components/devices/QRCodeDialog.vue'
import type { Device } from '@/types'

const toast = useToast()
const devices = ref<Device[]>([])
const loading = ref(false)
const autoRefresh = ref(true)
let refreshInterval: number | null = null

// Dialog states
const showCreateDialog = ref(false)
const showDeleteDialog = ref(false)
const showQRDialog = ref(false)
const selectedDevice = ref<Device | null>(null)

async function loadDevices(silent = false) {
  try {
    if (!silent) {
      loading.value = true
    }
    const response = await devicesApi.list()
    if (response.data) {
      // Check real WhatsApp connection status for each device
      const devicesWithStatus = await Promise.all(
        response.data.map(async (device: Device) => {
          try {
            // Check actual WhatsApp connection status
            const statusResponse = await whatsappApi.getStatus(device._id)
            const actualStatus = statusResponse.status || device.status

            // Map backend status to frontend status
            let mappedStatus: 'active' | 'inactive' | 'disconnected' = device.status
            if (actualStatus === 'connected') {
              mappedStatus = 'active'
            } else if (actualStatus === 'disconnected' || actualStatus === 'not connected') {
              mappedStatus = 'disconnected'
            } else {
              mappedStatus = 'inactive'
            }

            return {
              ...device,
              status: mappedStatus
            }
          } catch (error) {
            // If status check fails, use database status
            console.error(`Failed to check status for device ${device._id}:`, error)
            return device
          }
        })
      )
      devices.value = devicesWithStatus
    }
  } catch (error) {
    console.error('Failed to load devices:', error)
    if (!silent) {
      toast.error('Failed to load devices')
    }
  } finally {
    if (!silent) {
      loading.value = false
    }
  }
}

function startAutoRefresh() {
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
  // Refresh every 5 seconds
  refreshInterval = window.setInterval(() => {
    if (autoRefresh.value) {
      loadDevices(true) // Silent refresh
    }
  }, 5000)
}

function stopAutoRefresh() {
  if (refreshInterval) {
    clearInterval(refreshInterval)
    refreshInterval = null
  }
}

function toggleAutoRefresh() {
  autoRefresh.value = !autoRefresh.value
  if (autoRefresh.value) {
    startAutoRefresh()
    toast.info('Auto-refresh enabled')
  } else {
    stopAutoRefresh()
    toast.info('Auto-refresh disabled')
  }
}

function openCreateDialog() {
  showCreateDialog.value = true
}

function openDeleteDialog(device: Device) {
  selectedDevice.value = device
  showDeleteDialog.value = true
}

function openQRDialog(device: Device) {
  selectedDevice.value = device
  showQRDialog.value = true
}

async function handleDisconnect(device: Device) {
  try {
    await whatsappApi.disconnect(device._id)
    toast.success('Device disconnected successfully')
    await loadDevices()
  } catch (error: any) {
    console.error('Failed to disconnect device:', error)
    toast.error(error.response?.data?.error || 'Failed to disconnect device')
  }
}

function handleDeviceCreated() {
  loadDevices()
}

function handleDeviceDeleted() {
  loadDevices()
}

function handleDeviceConnected() {
  loadDevices()
}

async function handleManualRefresh() {
  await loadDevices()
  toast.success('Devices refreshed')
}

onMounted(() => {
  loadDevices()
  startAutoRefresh()
})

onUnmounted(() => {
  stopAutoRefresh()
})
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-8">
      <div>
        <h2 class="text-3xl font-bold text-gray-900">Devices</h2>
        <p class="text-gray-600 mt-1">Manage your WhatsApp devices</p>
      </div>
      <div class="flex gap-2">
        <Button variant="outline" size="sm" @click="handleManualRefresh" :disabled="loading">
          <RefreshCw class="w-4 h-4 mr-2" :class="{ 'animate-spin': loading }" />
          Refresh
        </Button>
        <Button
          variant="outline"
          size="sm"
          @click="toggleAutoRefresh"
          :class="{ 'bg-green-50 text-green-600 border-green-600': autoRefresh }"
        >
          <RefreshCw class="w-4 h-4 mr-2" :class="{ 'animate-spin': autoRefresh }" />
          Auto ({{ autoRefresh ? 'ON' : 'OFF' }})
        </Button>
        <Button @click="openCreateDialog">
          <Plus class="w-4 h-4 mr-2" />
          Add Device
        </Button>
      </div>
    </div>

    <Card v-if="loading">
      <CardContent class="pt-6">
        <p class="text-center text-gray-600">Loading devices...</p>
      </CardContent>
    </Card>

    <Card v-else-if="devices.length === 0">
      <CardContent class="pt-6">
        <p class="text-center text-gray-600">No devices found. Click "Add Device" to create one.</p>
      </CardContent>
    </Card>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <Card v-for="device in devices" :key="device._id">
        <CardHeader>
          <CardTitle>{{ device.name }}</CardTitle>
          <CardDescription>
            <span :class="{
              'text-green-600': device.status === 'active',
              'text-yellow-600': device.status === 'inactive',
              'text-red-600': device.status === 'disconnected'
            }">
              {{ device.status }}
            </span>
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div class="space-y-2 text-sm mb-4">
            <p><span class="text-gray-600">Owner:</span> {{ device.owner }}</p>
            <p><span class="text-gray-600">Created:</span> {{ new Date(device.createdAt).toLocaleDateString() }}</p>
          </div>

          <div class="flex flex-wrap gap-2">
            <Button
              size="sm"
              variant="outline"
              @click="openQRDialog(device)"
              :disabled="device.status === 'active'"
            >
              <QrCode class="w-4 h-4 mr-1" />
              Connect
            </Button>

            <Button
              size="sm"
              variant="outline"
              @click="handleDisconnect(device)"
              :disabled="device.status !== 'active'"
            >
              <Power class="w-4 h-4 mr-1" />
              Disconnect
            </Button>

            <Button
              size="sm"
              variant="destructive"
              @click="openDeleteDialog(device)"
            >
              <Trash2 class="w-4 h-4 mr-1" />
              Delete
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- Dialogs -->
    <DeviceCreateDialog
      :open="showCreateDialog"
      @update:open="showCreateDialog = $event"
      @created="handleDeviceCreated"
    />

    <DeviceDeleteDialog
      :open="showDeleteDialog"
      :device="selectedDevice"
      @update:open="showDeleteDialog = $event"
      @deleted="handleDeviceDeleted"
    />

    <QRCodeDialog
      :open="showQRDialog"
      :device="selectedDevice"
      @update:open="showQRDialog = $event"
      @connected="handleDeviceConnected"
    />
  </div>
</template>
