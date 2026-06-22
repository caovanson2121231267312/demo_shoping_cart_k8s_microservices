<template>
  <div class="chat-room d-flex flex-column h-100">
    <div class="pa-3 bg-primary text-white d-flex align-center">
      <v-icon class="mr-2">mdi-chat-processing-outline</v-icon>
      <span class="font-weight-medium">Hỗ trợ &amp; mua hàng</span>
      <v-spacer />
      <v-chip
        v-if="activeTab === 'admin' && auth.isLoggedIn.value && chat.connected.value"
        size="x-small"
        color="success"
        variant="flat"
      >
        Online
      </v-chip>
      <v-chip
        v-else-if="activeTab === 'admin' && auth.isLoggedIn.value && !chat.connected.value"
        size="x-small"
        color="warning"
        variant="flat"
      >
        Đang kết nối...
      </v-chip>
      <v-chip v-else size="x-small" color="info" variant="flat">AI</v-chip>
    </div>

    <v-tabs
      v-model="activeTab"
      density="compact"
      grow
      color="primary"
      class="chat-room__tabs"
    >
      <v-tab value="bot" class="text-none">
        <v-icon start size="18">mdi-robot-outline</v-icon>
        Trợ lý ảo
      </v-tab>
      <v-tab value="admin" class="text-none">
        <v-icon start size="18">mdi-headset</v-icon>
        Nhân viên
      </v-tab>
    </v-tabs>

    <div ref="messagesEl" class="flex-grow-1 overflow-y-auto pa-3 chat-room__messages">
      <LoadingSpinner v-if="loading" />

      <template v-else-if="activeTab === 'bot'">
        <EmptyState
          v-if="!botMessages.length"
          icon="mdi-robot-outline"
          title="Chưa có tin nhắn"
          description="Hỏi về sản phẩm, đơn hàng, bảo hành hoặc giờ làm việc."
        />
        <template v-else>
          <ChatMessage
            v-for="msg in botMessages"
            :key="msg.id"
            :message="msg"
            :current-user-id="botUserId"
          />
          <div v-if="botLoading" class="text-caption text-grey pa-2 d-flex align-center ga-1">
            <v-progress-circular indeterminate size="14" width="2" />
            Trợ lý đang trả lời...
          </div>
        </template>
      </template>

      <template v-else>
        <div v-if="!auth.isLoggedIn.value" class="chat-room__login-prompt text-center pa-6">
          <v-icon size="48" color="primary" class="mb-3">mdi-account-tie-voice</v-icon>
          <p class="text-body-2 text-medium-emphasis mb-4">
            Đăng nhập để chat trực tiếp với nhân viên tư vấn mua hàng, hỗ trợ đơn hàng và chính sách shop.
          </p>
          <v-btn color="primary" :to="loginPath" variant="flat">
            Đăng nhập
          </v-btn>
          <p class="text-caption text-medium-emphasis mt-4 mb-0">
            Hoặc dùng tab <strong>Trợ lý ảo</strong> để được hỗ trợ ngay không cần đăng nhập.
          </p>
        </div>
        <template v-else>
          <EmptyState
            v-if="!adminMessages.length"
            icon="mdi-headset"
            title="Chưa có tin nhắn"
            description="Nhân viên sẽ phản hồi trong giờ làm việc. Bạn có thể hỏi về sản phẩm, đơn hàng hoặc đổi trả."
          />
          <template v-else>
            <ChatMessage
              v-for="msg in adminMessages"
              :key="msg.id"
              :message="msg"
              :current-user-id="auth.user.value?.id"
              show-reactions
              @react="onReact(msg.id, $event)"
            />
            <div v-if="typingText" class="text-caption text-grey pa-2">{{ typingText }}</div>
          </template>
        </template>
      </template>
    </div>

    <v-divider />

    <div v-if="showInput" class="pa-3">
      <ChatComposer
        v-if="activeTab === 'admin' && auth.isLoggedIn.value"
        v-model="input"
        :room-id="chat.activeRoomId.value"
        :disabled="inputDisabled"
        :uploading="uploadingImage"
        :placeholder="inputPlaceholder"
        :hint="footerHint"
        @send="send"
        @typing="onTyping"
        @send-image="onSendImage"
      />
      <template v-else>
        <v-text-field
          v-model="input"
          :placeholder="inputPlaceholder"
          density="compact"
          variant="outlined"
          hide-details
          :disabled="inputDisabled"
          @keyup.enter="send"
        >
          <template #append-inner>
            <v-btn
              icon
              size="small"
              color="primary"
              variant="text"
              :disabled="!input.trim() || inputDisabled"
              @click="send"
            >
              <v-icon>mdi-send</v-icon>
            </v-btn>
          </template>
        </v-text-field>
        <div class="text-caption text-medium-emphasis mt-2">
          {{ footerHint }}
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
type ChatTab = 'bot' | 'admin'

const route = useRoute()
const auth = useAuth()
const chat = useChat()
const chatbot = useChatbot()

const activeTab = ref<ChatTab>('bot')
const input = ref('')
const loading = ref(false)
const uploadingImage = ref(false)
const messagesEl = ref<HTMLElement | null>(null)

const botMessages = computed(() => chatbot.messages.value)
const botLoading = computed(() => chatbot.loading.value)
const botUserId = computed(() => chatbot.sessionId.value)
const adminMessages = computed(() => chat.activeMessages.value)

const loginPath = computed(() => ({
  path: '/auth/login',
  query: route.fullPath !== '/' ? { redirect: route.fullPath } : undefined,
}))

const showInput = computed(() => activeTab.value === 'bot' || auth.isLoggedIn.value)

const inputPlaceholder = computed(() =>
  activeTab.value === 'bot'
    ? 'Hỏi trợ lý ảo...'
    : 'Nhắn nhân viên tư vấn...',
)

const inputDisabled = computed(() =>
  activeTab.value === 'bot' ? botLoading.value : !chat.connected.value,
)

const footerHint = computed(() => {
  if (activeTab.value === 'bot') {
    return 'Trợ lý ảo trả lời 24/7 về sản phẩm, đơn hàng và chính sách shop.'
  }
  return 'Nhân viên hỗ trợ trực tuyến trong giờ làm việc (8:00–21:00).'
})

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
  return 'Nhân viên đang nhập...'
})

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesEl.value) {
      messagesEl.value.scrollTop = messagesEl.value.scrollHeight
    }
  })
}

const initBot = async () => {
  await chatbot.greetIfEmpty()
}

const initAdmin = async () => {
  if (!auth.isLoggedIn.value) {
    return
  }
  chat.connect()
  await chat.ensureSupportRoom()
}

watch(
  () => botMessages.value.length,
  () => {
    if (activeTab.value === 'bot') {
      scrollToBottom()
    }
  },
)

watch(
  () => adminMessages.value.length,
  () => {
    if (activeTab.value === 'admin') {
      scrollToBottom()
    }
  },
)

watch(activeTab, async (tab) => {
  if (tab === 'admin') {
    loading.value = true
    try {
      await initAdmin()
    } finally {
      loading.value = false
      scrollToBottom()
    }
  } else {
    scrollToBottom()
  }
})

watch(
  () => auth.isLoggedIn.value,
  async (loggedIn) => {
    if (loggedIn && activeTab.value === 'admin') {
      await initAdmin()
    }
  },
)

onMounted(async () => {
  loading.value = true
  try {
    await initBot()
    if (auth.isLoggedIn.value && activeTab.value === 'admin') {
      await initAdmin()
    }
  } finally {
    loading.value = false
    scrollToBottom()
  }
})

const send = async () => {
  const text = input.value.trim()
  if (!text) {
    return
  }
  input.value = ''

  if (activeTab.value === 'bot') {
    await chatbot.sendMessage(text)
    scrollToBottom()
    return
  }

  const roomId = chat.activeRoomId.value
  if (!roomId) {
    return
  }
  chat.sendMessage(roomId, text)
  scrollToBottom()
}

const onSendImage = async (file: File) => {
  const roomId = chat.activeRoomId.value
  if (!roomId) {
    return
  }
  uploadingImage.value = true
  try {
    await chat.sendImageMessage(roomId, file)
    scrollToBottom()
  } catch (e) {
    console.error(e)
  } finally {
    uploadingImage.value = false
  }
}

const onReact = (messageId: string, emoji: string) => {
  const roomId = chat.activeRoomId.value
  if (!roomId) {
    return
  }
  chat.toggleReaction(roomId, messageId, emoji)
}

const onTyping = () => {
  if (activeTab.value !== 'admin' || !auth.isLoggedIn.value) {
    return
  }
  const roomId = chat.activeRoomId.value
  if (roomId) {
    chat.sendTyping(roomId)
  }
}
</script>

<style scoped>
.chat-room__tabs {
  border-bottom: 1px solid var(--color-border, rgba(0, 0, 0, 0.08));
}

.chat-room__messages {
  min-height: 280px;
  background: var(--color-bg, #fff);
}

.chat-room__login-prompt {
  max-width: 280px;
  margin: 0 auto;
}
</style>
