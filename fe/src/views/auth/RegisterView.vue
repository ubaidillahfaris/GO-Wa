<script setup lang="ts">
import { ref, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useToast } from '@/composables/useToast'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '@/components/ui/card'

const router = useRouter()
const authStore = useAuthStore()
const toast = useToast()

const username = ref('')
const password = ref('')
const confirmPassword = ref('')
const passwordError = ref('')

async function handleRegister() {
  passwordError.value = ''

  if (password.value !== confirmPassword.value) {
    passwordError.value = 'Passwords do not match'
    toast.error('Passwords do not match')
    return
  }

  if (password.value.length < 6) {
    passwordError.value = 'Password must be at least 6 characters'
    toast.error('Password must be at least 6 characters')
    return
  }

  const success = await authStore.register({
    username: username.value,
    password: password.value,
    confirm_password: confirmPassword.value,
  })

  if (success) {
    toast.success('Registration successful! Redirecting...')
    await nextTick()
    await router.replace({ name: 'dashboard' })
  } else {
    toast.error(authStore.error || 'Registration failed')
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
    <div class="w-full max-w-md">
      <div class="text-center mb-8">
        <h1 class="text-3xl font-bold text-gray-900">WhatsApp API</h1>
        <p class="mt-2 text-sm text-gray-600">Create your account</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Register</CardTitle>
          <CardDescription>Create a new account to get started</CardDescription>
        </CardHeader>

        <CardContent>
          <form @submit.prevent="handleRegister" class="space-y-4">
            <div class="space-y-2">
              <Label for="username">Username</Label>
              <Input
                id="username"
                v-model="username"
                type="text"
                placeholder="Choose a username"
                required
              />
            </div>

            <div class="space-y-2">
              <Label for="password">Password</Label>
              <Input
                id="password"
                v-model="password"
                type="password"
                placeholder="Choose a password"
                required
              />
            </div>

            <div class="space-y-2">
              <Label for="confirmPassword">Confirm Password</Label>
              <Input
                id="confirmPassword"
                v-model="confirmPassword"
                type="password"
                placeholder="Confirm your password"
                required
              />
            </div>

            <div v-if="passwordError" class="text-sm text-destructive">
              {{ passwordError }}
            </div>

            <div v-if="authStore.error" class="text-sm text-destructive">
              {{ authStore.error }}
            </div>

            <Button type="submit" class="w-full" :disabled="authStore.loading">
              {{ authStore.loading ? 'Creating account...' : 'Create account' }}
            </Button>
          </form>
        </CardContent>

        <CardFooter class="flex-col">
          <p class="text-sm text-gray-600">
            Already have an account?
            <router-link :to="{ name: 'login' }" class="text-primary hover:underline font-medium">
              Sign in here
            </router-link>
          </p>
        </CardFooter>
      </Card>
    </div>
  </div>
</template>
