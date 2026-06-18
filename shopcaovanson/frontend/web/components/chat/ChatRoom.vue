<template>
  <div class="d-flex flex-column h-100">
    <div class="pa-3 bg-primary text-white d-flex align-center">
      <v-icon class="mr-2">mdi-headset</v-icon>
      <span class="font-weight-medium">Support Chat</span>
      <v-spacer />
      <v-chip
        v-if="chat.connected.value"
        size="x-small"
        color="success"
        variant="flat"
      >
        Online
      </v-chip>
      <v-chip v-else size="x-small" color="warning" variant="flat">Offline</v-chip>
    </div>

    <div ref="messagesEl" class="flex-grow-1 overflow-y-auto pa-3" style="min-height: 300px">
      <LoadingSpinner v-if="loading" />
      <EmptyState
        v-else-if="!chat.activeMessages.value.length"
        icon="mdi-chat-outline"
        title="Chưa có tin nhắn"
        description="Gửi tin nhắn để bắt đầu trò chuyện với hỗ trợ."
      />
      <template v-else>
        <ChatMessage
          v-for="msg in chat.activeMessages.value"
          :key="msg.id"
          :message="msg"
          :current-user-id="auth.user.value?.id"
        />
        <div v-if="typingText" class="text-caption text-grey pa-2">{{ typingText }}</div>
      </template>
    </div>

    <v-divider />

    <div class="pa-3">
      <v-text-field
        v-model="input"
        placeholder="Nhập tin nhắn..."
        density="compact"
        variant="outlined"
        hide-details
        :disabled="!auth.isLoggedIn.value || !chat.activeRoomId.value"
        @keyup.enter="send"
        @input="onTyping"
      >
        <template #append-inner>
          <v-btn
            icon
            size="small"
            color="primary"
            variant="text"
            :disabled="!input.trim()"
            @click="send"
          >
            <v-icon>mdi-send</v-icon>
          </v-btn>
        </template>
      </v-text-field>
      <div v-if="!auth.isLoggedIn.value" class="text-caption text-warning mt-2">
        Vui lòng đăng nhập để sử dụng chat hỗ trợ.
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const auth = useAuth()
const chat = useChat()

const input = ref('')
const loading = ref(false)
const messagesEl = ref<HTMLElement | null>(null)

const typingText = computed(() => {
  const roomId = chat.activeRoomId.value
  if (!roomId) {
    return ''
  }
  const users = chat.typingUsers.value[roomId] || []
  const others = users.filter((id) => id !== auth.user.value?.id)
  if (!others.length) {
    return ''
  }
  return 'Đang nhập...'
})

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesEl.value) {
      messagesEl.value.scrollTop = messagesEl.value.scrollHeight
    }
  })
}

watch(
  () => chat.activeMessages.value.length,
  () => scrollToBottom(),
)

onMounted(async () => {
  if (auth.isLoggedIn.value) {
    loading.value = true
    try {
      chat.connect()
      await chat.ensureSupportRoom()
    } finally {
      loading.value = false
      scrollToBottom()
    }
  }
})

const send = () => {
  const roomId = chat.activeRoomId.value
  if (!roomId || !input.value.trim()) {
    return
  }
  chat.sendMessage(roomId, input.value)
  input.value = ''
}

const onTyping = () => {
  const roomId = chat.activeRoomId.value
  if (roomId) {
    chat.sendTyping(roomId)
  }
}
</script>
