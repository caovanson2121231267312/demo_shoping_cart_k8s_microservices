<template>
  <div
    class="d-flex mb-3"
    :class="isOwn ? 'justify-end' : 'justify-start'"
  >
    <div
      class="pa-3 rounded-lg chat-message"
      :class="bubbleClass"
      style="max-width: 80%"
    >
      <div v-if="isBot" class="text-caption font-weight-bold mb-1 d-flex align-center ga-1">
        <v-icon size="14">mdi-robot-outline</v-icon>
        {{ BOT_DISPLAY_NAME }}
      </div>
      <div class="text-body-2" style="white-space: pre-line">{{ message.content }}</div>
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
import { BOT_DISPLAY_NAME, BOT_USER_ID } from '~/utils/chat'

const props = defineProps<{
  message: ChatMessage
  currentUserId?: string
}>()

const isBot = computed(() => props.message.sender_id === BOT_USER_ID)
const isOwn = computed(() => !isBot.value && props.message.sender_id === props.currentUserId)

const bubbleClass = computed(() => {
  if (isBot.value) {
    return 'bg-primary-subtle chat-message--bot'
  }
  if (isOwn.value) {
    return 'bg-primary text-white'
  }
  return 'bg-grey-lighten-3'
})

const formattedTime = computed(() => {
  const date = new Date(props.message.created_at)
  return date.toLocaleTimeString('vi-VN', { hour: '2-digit', minute: '2-digit' })
})
</script>

<style scoped>
.chat-message--bot {
  border: 1px solid var(--color-primary-alpha-12);
}
</style>
