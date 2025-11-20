<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '@/components/ui/card'

const router = useRouter()
const authStore = useAuthStore()

const username = ref('')
const password = ref('')

async function handleLogin() {
  const success = await authStore.login({
    username: username.value,
    password: password.value,
  })

  if (success) {
    router.push({ name: 'dashboard' })
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
    <div class="w-full max-w-md">
      <div class="text-center mb-8">
        <h1 class="text-3xl font-bold text-gray-900">WhatsApp API</h1>
        <p class="mt-2 text-sm text-gray-600">Sign in to your account</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Login</CardTitle>
          <CardDescription>Enter your credentials to access your account</CardDescription>
        </CardHeader>

        <CardContent>
          <form @submit.prevent="handleLogin" class="space-y-4">
            <div class="space-y-2">
              <Label for="username">Username</Label>
              <Input
                id="username"
                v-model="username"
                type="text"
                placeholder="Enter your username"
                required
              />
            </div>

            <div class="space-y-2">
              <Label for="password">Password</Label>
              <Input
                id="password"
                v-model="password"
                type="password"
                placeholder="Enter your password"
                required
              />
            </div>

            <div v-if="authStore.error" class="text-sm text-destructive">
              {{ authStore.error }}
            </div>

            <Button type="submit" class="w-full" :disabled="authStore.loading">
              {{ authStore.loading ? 'Signing in...' : 'Sign in' }}
            </Button>
          </form>
        </CardContent>

        <CardFooter class="flex-col">
          <p class="text-sm text-gray-600">
            Don't have an account?
            <router-link :to="{ name: 'register' }" class="text-primary hover:underline font-medium">
              Register here
            </router-link>
          </p>
        </CardFooter>
      </Card>
    </div>
  </div>
</template>
