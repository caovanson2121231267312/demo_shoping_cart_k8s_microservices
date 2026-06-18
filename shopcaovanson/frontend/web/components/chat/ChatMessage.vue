<template>
  <div
    class="d-flex mb-3"
    :class="isOwn ? 'justify-end' : 'justify-start'"
  >
    <div
      class="pa-3 rounded-lg"
      :class="isOwn ? 'bg-primary text-white' : 'bg-grey-lighten-3'"
      style="max-width: 80%"
    >
      <div class="text-body-2">{{ message.content }}</div>
      <div
        class="text-caption mt-1"
        :class="isOwn ? 'text-blue-lighten-4' : 'text-grey'"
      >
        {{ formattedTime }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ChatMessage } from '~/types'

const props = defineProps<{
  message: ChatMessage
  currentUserId?: string
}>()

const isOwn = computed(() => props.message.sender_id === props.currentUserId)

const formattedTime = computed(() => {
  const date = new Date(props.message.created_at)
  return date.toLocaleTimeString('vi-VN', { hour: '2-digit', minute: '2-digit' })
})
</script>
