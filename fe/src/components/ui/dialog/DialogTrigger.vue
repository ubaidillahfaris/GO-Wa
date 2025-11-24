<script setup lang="ts">
import { DialogTrigger, type DialogTriggerProps } from 'radix-vue'
import { computed } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps<DialogTriggerProps & { class?: string }>()

const delegatedProps = computed(() => {
  const { class: _, ...delegated } = props
  return delegated
})
</script>

<template>
  <DialogTrigger v-bind="delegatedProps" :class="cn(props.class)">
    <slot />
  </DialogTrigger>
</template>
