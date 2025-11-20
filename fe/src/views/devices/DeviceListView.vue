<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { devicesApi } from '@/api/devices'
import { whatsappApi } from '@/api/whatsapp'
import { useToast } from '@/composables/useToast'
import { Button } from '@/components/ui/button'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Plus, Trash2, QrCode, Power } from 'lucide-vue-next'
import DeviceCreateDialog from '@/components/devices/DeviceCreateDialog.vue'
import DeviceDeleteDialog from '@/components/devices/DeviceDeleteDialog.vue'
import QRCodeDialog from '@/components/devices/QRCodeDialog.vue'
import type { Device } from '@/types'

const toast = useToast()
const devices = ref<Device[]>([])
const loading = ref(false)

// Dialog states
const showCreateDialog = ref(false)
const showDeleteDialog = ref(false)
const showQRDialog = ref(false)
const selectedDevice = ref<Device | null>(null)

async function loadDevices() {
  try {
    loading.value = true
    const response = await devicesApi.list()
    if (response.data) {
      devices.value = response.data
    }
  } catch (error) {
    console.error('Failed to load devices:', error)
    toast.error('Failed to load devices')
  } finally {
    loading.value = false
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

onMounted(() => {
  loadDevices()
})
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-8">
      <div>
        <h2 class="text-3xl font-bold text-gray-900">Devices</h2>
        <p class="text-gray-600 mt-1">Manage your WhatsApp devices</p>
      </div>
      <Button @click="openCreateDialog">
        <Plus class="w-4 h-4 mr-2" />
        Add Device
      </Button>
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
