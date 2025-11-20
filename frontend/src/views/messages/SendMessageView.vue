<script setup lang="ts">
import { ref } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'

const device = ref('')
const receiver = ref('')
const message = ref('')
const receiverType = ref<'individual' | 'group'>('individual')
const sending = ref(false)

async function handleSend() {
  sending.value = true
  // TODO: Implement send message
  setTimeout(() => {
    sending.value = false
    alert('Message sent! (placeholder)')
  }, 1000)
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
          <div class="space-y-2">
            <Label for="device">Device</Label>
            <Input
              id="device"
              v-model="device"
              placeholder="Enter device name"
              required
            />
          </div>

          <div class="space-y-2">
            <Label for="receiver">Receiver</Label>
            <Input
              id="receiver"
              v-model="receiver"
              placeholder="Phone number or Group JID"
              required
            />
          </div>

          <div class="space-y-2">
            <Label for="message">Message</Label>
            <textarea
              id="message"
              v-model="message"
              class="flex min-h-[120px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
              placeholder="Type your message here..."
              required
            ></textarea>
          </div>

          <div class="space-y-2">
            <Label>Receiver Type</Label>
            <div class="flex gap-4">
              <label class="flex items-center gap-2">
                <input
                  type="radio"
                  v-model="receiverType"
                  value="individual"
                  class="w-4 h-4"
                />
                Individual
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

          <Button type="submit" class="w-full" :disabled="sending">
            {{ sending ? 'Sending...' : 'Send Message' }}
          </Button>
        </form>
      </CardContent>
    </Card>
  </div>
</template>
