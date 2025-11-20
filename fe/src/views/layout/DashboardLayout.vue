<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { Button } from '@/components/ui/button'
import { LayoutDashboard, Smartphone, MessageSquare, Key, LogOut } from 'lucide-vue-next'

const router = useRouter()
const authStore = useAuthStore()

function handleLogout() {
  authStore.logout()
  router.push({ name: 'login' })
}
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Sidebar -->
    <div class="fixed inset-y-0 left-0 w-64 bg-white border-r">
      <div class="flex flex-col h-full">
        <!-- Logo/Header -->
        <div class="px-6 py-8">
          <h1 class="text-2xl font-bold text-gray-900">WhatsApp API</h1>
          <p class="text-sm text-gray-600 mt-1">{{ authStore.user?.username }}</p>
        </div>

        <!-- Navigation -->
        <nav class="flex-1 px-4 space-y-1">
          <router-link
            :to="{ name: 'dashboard' }"
            class="flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-colors hover:bg-gray-100"
            active-class="bg-primary text-primary-foreground hover:bg-primary/90"
          >
            <LayoutDashboard class="w-5 h-5" />
            Dashboard
          </router-link>

          <router-link
            :to="{ name: 'devices' }"
            class="flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-colors hover:bg-gray-100"
            active-class="bg-primary text-primary-foreground hover:bg-primary/90"
          >
            <Smartphone class="w-5 h-5" />
            Devices
          </router-link>

          <router-link
            :to="{ name: 'messages' }"
            class="flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-colors hover:bg-gray-100"
            active-class="bg-primary text-primary-foreground hover:bg-primary/90"
          >
            <MessageSquare class="w-5 h-5" />
            Send Message
          </router-link>

          <router-link
            :to="{ name: 'api-keys' }"
            class="flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-colors hover:bg-gray-100"
            active-class="bg-primary text-primary-foreground hover:bg-primary/90"
          >
            <Key class="w-5 h-5" />
            API Keys
          </router-link>
        </nav>

        <!-- Logout Button -->
        <div class="p-4 border-t">
          <Button @click="handleLogout" variant="outline" class="w-full justify-start gap-2">
            <LogOut class="w-4 h-4" />
            Logout
          </Button>
        </div>
      </div>
    </div>

    <!-- Main Content -->
    <div class="pl-64">
      <div class="p-8">
        <router-view />
      </div>
    </div>
  </div>
</template>
