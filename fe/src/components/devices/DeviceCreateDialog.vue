<script setup lang="ts">
import { ref } from 'vue'
import { devicesApi } from '@/api/devices'
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
const owner = ref('')

async function handleCreate() {
  if (!name.value.trim() || !owner.value.trim()) {
    toast.error('Name and owner are required')
    return
  }

  try {
    loading.value = true
    await devicesApi.create({
      name: name.value,
      owner: owner.value,
      status: 'inactive'
    })

    toast.success('Device created successfully')
    emit('created')
    emit('update:open', false)

    // Reset form
    name.value = ''
    owner.value = ''
  } catch (error: any) {
    console.error('Failed to create device:', error)
    toast.error(error.response?.data?.error || 'Failed to create device')
  } finally {
    loading.value = false
  }
}

function handleClose() {
  if (!loading.value) {
    emit('update:open', false)
    name.value = ''
    owner.value = ''
  }
}
</script>

<template>
  <Dialog :open="props.open" @update:open="handleClose">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Create New Device</DialogTitle>
        <DialogDescription>
          Add a new WhatsApp device to your account
        </DialogDescription>
      </DialogHeader>

      <div class="space-y-4 py-4">
        <div class="space-y-2">
          <Label for="device-name">Device Name</Label>
          <Input
            id="device-name"
            v-model="name"
            placeholder="e.g., My WhatsApp Device"
            :disabled="loading"
          />
        </div>

        <div class="space-y-2">
          <Label for="device-owner">Owner</Label>
          <Input
            id="device-owner"
            v-model="owner"
            placeholder="e.g., John Doe"
            :disabled="loading"
          />
        </div>
      </div>

      <div class="flex justify-end gap-3">
        <Button variant="outline" @click="handleClose" :disabled="loading">
          Cancel
        </Button>
        <Button @click="handleCreate" :disabled="loading">
          {{ loading ? 'Creating...' : 'Create Device' }}
        </Button>
      </div>
    </DialogContent>
  </Dialog>
</template>
