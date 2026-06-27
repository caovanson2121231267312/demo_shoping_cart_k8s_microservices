<template>
  <div
    class="chat-message-wrap mb-3"
    :class="isOwn ? 'chat-message-wrap--own' : 'chat-message-wrap--other'"
  >
    <div class="chat-message-wrap__bubble-col">
      <div class="chat-message__bubble-wrap">
        <div class="pa-3 rounded-lg chat-message" :class="bubbleClass">
          <div v-if="isBot" class="text-caption font-weight-bold mb-1 d-flex align-center ga-1">
            <v-icon size="14">mdi-robot-outline</v-icon>
            {{ BOT_DISPLAY_NAME }}
          </div>

          <a
            v-if="isImage"
            :href="imageUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="chat-message__image-link d-block"
          >
            <img
              :src="imageUrl"
              alt="Ảnh đính kèm"
              class="chat-message__image"
              loading="lazy"
            >
          </a>
          <template v-else>
            <div v-if="message.content" class="text-body-2 chat-message__text">
              {{ message.content }}
            </div>
            <ChatProductList
              v-if="message.products?.length"
              :products="message.products"
            />
          </template>

          <div class="text-caption mt-1" :class="isOwn ? 'chat-message__time--own' : 'text-grey'">
            {{ formattedTime }}
          </div>
        </div>

        <div v-if="showReactions && !isBot" class="chat-message__quick-react">
          <button
            v-for="emoji in CHAT_REACTION_EMOJIS"
            :key="emoji"
            type="button"
            class="chat-message__quick-react-btn"
            :title="`Thả ${emoji}`"
            @click="emit('react', emoji)"
          >
            {{ emoji }}
          </button>
        </div>
      </div>

      <div
        v-if="showReactions && reactionEntries.length"
        class="chat-message__reactions d-flex flex-wrap ga-1 mt-1"
      >
        <button
          v-for="[emoji, users] in reactionEntries"
          :key="emoji"
          type="button"
          class="chat-message__reaction-chip"
          :class="{ 'chat-message__reaction-chip--mine': users.includes(currentUserId || '') }"
          @click="emit('react', emoji)"
        >
          <span>{{ emoji }}</span>
          <span class="chat-message__reaction-count">{{ users.length }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ChatMessage } from '~/types'
import { BOT_DISPLAY_NAME, BOT_USER_ID, CHAT_REACTION_EMOJIS, chatMediaUrl } from '~/utils/chat'

const props = defineProps<{
  message: ChatMessage
  currentUserId?: string
  showReactions?: boolean
}>()

const emit = defineEmits<{
  react: [emoji: string]
}>()

const isBot = computed(() => props.message.sender_id === BOT_USER_ID)
const isOwn = computed(() => !isBot.value && props.message.sender_id === props.currentUserId)
const isImage = computed(() => props.message.type === 'image')
const imageUrl = computed(() => chatMediaUrl(props.message.content))

const reactionEntries = computed(() => {
  const reactions = props.message.reactions || {}
  return Object.entries(reactions).filter(([, users]) => users.length > 0)
})

const bubbleClass = computed(() => {
  if (isBot.value) {
    return 'bg-primary-subtle chat-message--bot'
  }
  if (isOwn.value) {
    return 'bg-primary text-white chat-message--own'
  }
  return 'bg-grey-lighten-3'
})

const formattedTime = computed(() => {
  const date = new Date(props.message.created_at)
  return date.toLocaleTimeString('vi-VN', { hour: '2-digit', minute: '2-digit' })
})
</script>

<style scoped>
.chat-message-wrap {
  display: flex;
  contain: layout style;
}

.chat-message-wrap--own {
  justify-content: flex-end;
}

.chat-message-wrap--other {
  justify-content: flex-start;
}

.chat-message-wrap__bubble-col {
  position: relative;
  max-width: min(92%, 360px);
}

.chat-message__bubble-wrap {
  position: relative;
}

.chat-message--bot {
  border: 1px solid var(--color-primary-alpha-12);
}

.chat-message__text {
  white-space: pre-line;
  word-break: break-word;
}

.chat-message__image-link {
  text-decoration: none;
}

.chat-message__image {
  display: block;
  max-width: min(240px, 100%);
  max-height: 200px;
  border-radius: 8px;
  object-fit: cover;
}

.chat-message--own .chat-message__time--own {
  color: rgba(255, 255, 255, 0.75);
}

.chat-message__reaction-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: 999px;
  border: 1px solid var(--color-border, rgba(0, 0, 0, 0.1));
  background: var(--color-surface, #fff);
  font-size: 0.8125rem;
  cursor: pointer;
  line-height: 1.4;
}

.chat-message__reaction-chip--mine {
  border-color: rgb(var(--v-theme-primary));
  background: rgba(var(--v-theme-primary), 0.08);
}

.chat-message__reaction-count {
  font-size: 0.6875rem;
  color: var(--color-text-muted, #666);
}

.chat-message__quick-react {
  position: absolute;
  bottom: calc(100% + 6px);
  left: 0;
  display: flex;
  gap: 2px;
  padding: 4px 6px;
  border-radius: 999px;
  background: var(--color-surface, #fff);
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.14);
  opacity: 0;
  visibility: hidden;
  pointer-events: none;
  transition: opacity 0.15s ease, visibility 0.15s ease;
  z-index: 20;
  white-space: nowrap;
}

.chat-message-wrap--own .chat-message__quick-react {
  left: auto;
  right: 0;
}

.chat-message__bubble-wrap:hover .chat-message__quick-react,
.chat-message__bubble-wrap:focus-within .chat-message__quick-react {
  opacity: 1;
  visibility: visible;
  pointer-events: auto;
}

.chat-message__quick-react-btn {
  border: none;
  background: transparent;
  border-radius: 999px;
  padding: 2px 5px;
  font-size: 0.9375rem;
  cursor: pointer;
  line-height: 1;
}

.chat-message__quick-react-btn:hover {
  background: rgba(var(--v-theme-primary), 0.1);
}
</style>
