<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { devicesApi } from '@/api/devices'
import { Button } from '@/components/ui/button'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Plus } from 'lucide-vue-next'
import type { Device } from '@/types'

const devices = ref<Device[]>([])
const loading = ref(false)

async function loadDevices() {
  try {
    loading.value = true
    const response = await devicesApi.list()
    if (response.data) {
      devices.value = response.data
    }
  } catch (error) {
    console.error('Failed to load devices:', error)
  } finally {
    loading.value = false
  }
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
      <Button>
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
          <div class="space-y-2 text-sm">
            <p><span class="text-gray-600">Owner:</span> {{ device.owner }}</p>
            <p><span class="text-gray-600">Created:</span> {{ new Date(device.createdAt).toLocaleDateString() }}</p>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
