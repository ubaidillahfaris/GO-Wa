<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { devicesApi } from '@/api/devices'
import { whatsappApi } from '@/api/whatsapp'
import { useToast } from '@/composables/useToast'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import type { Device } from '@/types'

const toast = useToast()

// Form fields
const devices = ref<Device[]>([])
const selectedDeviceId = ref('')
const to = ref('')
const message = ref('')
const receiverType = ref<'user' | 'group'>('user')
const messageType = ref<'text' | 'file'>('text')
const typing = ref(true)
const file = ref<File | null>(null)
const filename = ref('')
const caption = ref('')
const sending = ref(false)
const loadingDevices = ref(false)

onMounted(async () => {
  await loadDevices()
})

async function loadDevices() {
  try {
    loadingDevices.value = true
    const response = await devicesApi.list()
    if (response.data) {
      devices.value = response.data.filter((d: Device) => d.status === 'active')
      if (devices.value.length > 0 && devices.value[0]) {
        selectedDeviceId.value = devices.value[0]._id
      }
    }
  } catch (error) {
    console.error('Failed to load devices:', error)
    toast.error('Failed to load devices')
  } finally {
    loadingDevices.value = false
  }
}

function handleFileChange(event: Event) {
  const target = event.target as HTMLInputElement
  if (target.files && target.files[0]) {
    file.value = target.files[0]
    if (!filename.value) {
      filename.value = target.files[0].name
    }
    messageType.value = 'file'
  } else {
    file.value = null
    messageType.value = 'text'
  }
}

function clearFile() {
  file.value = null
  filename.value = ''
  caption.value = ''
  messageType.value = 'text'
  const fileInput = document.getElementById('file') as HTMLInputElement
  if (fileInput) {
    fileInput.value = ''
  }
}

async function handleSend() {
  if (!selectedDeviceId.value) {
    toast.error('Please select a device')
    return
  }

  if (!to.value.trim()) {
    toast.error('Please enter recipient phone number or group JID')
    return
  }

  if (!message.value.trim() && !file.value) {
    toast.error('Please enter a message or select a file')
    return
  }

  if (file.value && !filename.value.trim()) {
    toast.error('Please enter a filename')
    return
  }

  try {
    sending.value = true

    await whatsappApi.sendMessage(selectedDeviceId.value, {
      to: to.value,
      message: message.value,
      receiver_type: receiverType.value,
      message_type: messageType.value,
      typing: typing.value,
      file: file.value || undefined,
      filename: filename.value || undefined,
      caption: caption.value || undefined,
    })

    toast.success('Message sent successfully!')

    // Reset form
    to.value = ''
    message.value = ''
    clearFile()
  } catch (error: any) {
    console.error('Failed to send message:', error)
    toast.error(error.response?.data?.error || 'Failed to send message')
  } finally {
    sending.value = false
  }
}
</script>

<template>
  <div>
    <div class="mb-8">
      <h2 class="text-3xl font-bold text-gray-900">Send Message</h2>
      <p class="text-gray-600 mt-1">Send WhatsApp messages to individuals or groups</p>
    </div>

    <Card class="max-w-2xl">
      <CardHeader>
        <CardTitle>New Message</CardTitle>
        <CardDescription>Fill in the details to send a message</CardDescription>
      </CardHeader>

      <CardContent>
        <form @submit.prevent="handleSend" class="space-y-4">
          <!-- Device Selection -->
          <div class="space-y-2">
            <Label for="device">Device</Label>
            <select
              id="device"
              v-model="selectedDeviceId"
              class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
              :disabled="loadingDevices || devices.length === 0"
              required
            >
              <option value="" disabled>Select a device</option>
              <option v-for="device in devices" :key="device._id" :value="device._id">
                {{ device.name }} ({{ device.owner }})
              </option>
            </select>
            <p v-if="devices.length === 0" class="text-sm text-yellow-600">
              No active devices found. Please connect a device first.
            </p>
          </div>

          <!-- Receiver -->
          <div class="space-y-2">
            <Label for="to">Recipient</Label>
            <Input
              id="to"
              v-model="to"
              placeholder="628123456789 or group-id@g.us"
              required
            />
            <p class="text-xs text-gray-500">
              Enter phone number with country code (e.g., 628123456789) or group JID
            </p>
          </div>

          <!-- Receiver Type -->
          <div class="space-y-2">
            <Label>Receiver Type</Label>
            <div class="flex gap-4">
              <label class="flex items-center gap-2">
                <input
                  type="radio"
                  v-model="receiverType"
                  value="user"
                  class="w-4 h-4"
                />
                User
              </label>
              <label class="flex items-center gap-2">
                <input
                  type="radio"
                  v-model="receiverType"
                  value="group"
                  class="w-4 h-4"
                />
                Group
              </label>
            </div>
          </div>

          <!-- Message -->
          <div class="space-y-2">
            <Label for="message">Message</Label>
            <textarea
              id="message"
              v-model="message"
              class="flex min-h-[120px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
              placeholder="Type your message here..."
              :required="!file"
            ></textarea>
          </div>

          <!-- File Upload -->
          <div class="space-y-2">
            <Label for="file">Attachment (Optional)</Label>
            <Input
              id="file"
              type="file"
              @change="handleFileChange"
              class="cursor-pointer"
            />
            <div v-if="file" class="flex items-center justify-between p-3 bg-gray-50 rounded-md">
              <span class="text-sm text-gray-700">{{ file.name }}</span>
              <Button type="button" size="sm" variant="ghost" @click="clearFile">
                Remove
              </Button>
            </div>
          </div>

          <!-- Filename (shown when file is selected) -->
          <div v-if="file" class="space-y-2">
            <Label for="filename">Filename</Label>
            <Input
              id="filename"
              v-model="filename"
              placeholder="Enter filename with extension"
              required
            />
          </div>

          <!-- Caption (shown when file is selected) -->
          <div v-if="file" class="space-y-2">
            <Label for="caption">Caption (Optional)</Label>
            <Input
              id="caption"
              v-model="caption"
              placeholder="Enter caption for the file"
            />
          </div>

          <!-- Typing Indicator -->
          <div class="flex items-center gap-2">
            <input
              id="typing"
              type="checkbox"
              v-model="typing"
              class="w-4 h-4 rounded border-gray-300 text-primary focus:ring-primary"
            />
            <Label for="typing" class="cursor-pointer">Show typing indicator</Label>
          </div>

          <!-- Submit Button -->
          <Button type="submit" class="w-full" :disabled="sending || loadingDevices || devices.length === 0">
            {{ sending ? 'Sending...' : 'Send Message' }}
          </Button>
        </form>
      </CardContent>
    </Card>
  </div>
</template>
