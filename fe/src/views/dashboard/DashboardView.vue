<script setup lang="ts">
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Smartphone, MessageSquare, Key, Activity } from 'lucide-vue-next'
</script>

<template>
  <div>
    <div class="mb-8">
      <h2 class="text-3xl font-bold text-gray-900">Dashboard</h2>
      <p class="text-gray-600 mt-1">Welcome to your WhatsApp API dashboard</p>
    </div>

    <!-- Stats Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
      <Card>
        <CardContent class="pt-6">
          <div class="flex items-center gap-4">
            <div class="p-3 bg-blue-100 rounded-lg">
              <Smartphone class="w-6 h-6 text-blue-600" />
            </div>
            <div>
              <p class="text-sm text-gray-600">Total Devices</p>
              <p class="text-2xl font-bold">0</p>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent class="pt-6">
          <div class="flex items-center gap-4">
            <div class="p-3 bg-green-100 rounded-lg">
              <Activity class="w-6 h-6 text-green-600" />
            </div>
            <div>
              <p class="text-sm text-gray-600">Active Devices</p>
              <p class="text-2xl font-bold">0</p>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent class="pt-6">
          <div class="flex items-center gap-4">
            <div class="p-3 bg-purple-100 rounded-lg">
              <MessageSquare class="w-6 h-6 text-purple-600" />
            </div>
            <div>
              <p class="text-sm text-gray-600">Messages Sent</p>
              <p class="text-2xl font-bold">0</p>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent class="pt-6">
          <div class="flex items-center gap-4">
            <div class="p-3 bg-orange-100 rounded-lg">
              <Key class="w-6 h-6 text-orange-600" />
            </div>
            <div>
              <p class="text-sm text-gray-600">API Keys</p>
              <p class="text-2xl font-bold">0</p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- Quick Actions -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <Card>
        <CardHeader>
          <CardTitle>Getting Started</CardTitle>
          <CardDescription>Quick steps to start using the WhatsApp API</CardDescription>
        </CardHeader>
        <CardContent>
          <ol class="space-y-3 text-sm">
            <li class="flex items-start gap-2">
              <span class="font-bold text-primary">1.</span>
              <span>Create a new device in the Devices section</span>
            </li>
            <li class="flex items-start gap-2">
              <span class="font-bold text-primary">2.</span>
              <span>Scan the QR code with your WhatsApp mobile app</span>
            </li>
            <li class="flex items-start gap-2">
              <span class="font-bold text-primary">3.</span>
              <span>Generate an API key for programmatic access</span>
            </li>
            <li class="flex items-start gap-2">
              <span class="font-bold text-primary">4.</span>
              <span>Start sending messages via the API or web interface</span>
            </li>
          </ol>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Recent Activity</CardTitle>
          <CardDescription>Latest events from your devices</CardDescription>
        </CardHeader>
        <CardContent>
          <p class="text-sm text-gray-600">No recent activity</p>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
