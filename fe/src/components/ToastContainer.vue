<script setup lang="ts">
import { useToast } from '@/composables/useToast'

const { toasts, remove } = useToast()

const getToastClasses = (type: string) => {
  const baseClasses = 'px-4 py-3 rounded-lg shadow-lg text-white mb-2 flex items-center justify-between'
  const typeClasses = {
    success: 'bg-green-500',
    error: 'bg-red-500',
    info: 'bg-blue-500',
    warning: 'bg-yellow-500',
  }
  return `${baseClasses} ${typeClasses[type as keyof typeof typeClasses] || typeClasses.info}`
}
</script>

<template>
  <div class="fixed top-4 right-4 z-50 space-y-2">
    <div
      v-for="toast in toasts"
      :key="toast.id"
      :class="getToastClasses(toast.type)"
    >
      <span>{{ toast.message }}</span>
      <button
        @click="remove(toast.id)"
        class="ml-4 text-white hover:text-gray-200 font-bold"
      >
        ×
      </button>
    </div>
  </div>
</template>
