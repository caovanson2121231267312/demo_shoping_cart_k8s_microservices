<template>
  <div class="admin-chat">
    <AdminPageHeader
      title="Chat hỗ trợ khách hàng"
      subtitle="Đăng nhập nhân viên: support@shop.com / Support@123"
    >
      <template #actions>
        <v-chip
          :color="chat.connected.value ? 'success' : 'warning'"
          size="small"
          variant="flat"
          class="mr-2"
        >
          {{ chat.connected.value ? 'WebSocket online' : 'Đang kết nối...' }}
        </v-chip>
        <v-btn to="/admin" variant="tonal" prepend-icon="mdi-arrow-left">Dashboard</v-btn>
      </template>
    </AdminPageHeader>

    <div class="admin-chat__layout">
      <aside class="admin-chat__sidebar">
        <v-tabs v-model="sidebarTab" density="compact" grow color="primary">
          <v-tab value="online" class="text-none">Online</v-tab>
          <v-tab value="rooms" class="text-none">Hội thoại</v-tab>
        </v-tabs>

        <div class="admin-chat__sidebar-body">
          <template v-if="sidebarTab === 'online'">
            <div class="admin-chat__section-title">
              Khách đang online
              <v-chip size="x-small" color="success" variant="flat">{{ onlineCustomers.length }}</v-chip>
            </div>
            <LoadingSpinner v-if="loadingOnline" />
            <div v-else-if="!onlineCustomers.length" class="admin-chat__empty">
              Chưa có khách online
            </div>
            <v-list v-else density="compact" class="pa-0">
              <v-list-item
                v-for="user in onlineCustomers"
                :key="user.user_id"
                :active="activeCustomerId === user.user_id"
                rounded="lg"
                class="mb-1"
                @click="openChatWithCustomer(user.user_id)"
              >
                <template #prepend>
                  <v-avatar color="success" size="32">
                    <v-icon color="white" size="16">mdi-account</v-icon>
                  </v-avatar>
                </template>
                <v-list-item-title class="text-body-2 font-weight-medium">
                  {{ user.email || shortId(user.user_id) }}
                </v-list-item-title>
                <v-list-item-subtitle class="text-caption">
                  {{ user.page }}
                </v-list-item-subtitle>
                <template #append>
                  <v-icon size="18" color="success">mdi-chat-outline</v-icon>
                </template>
              </v-list-item>
            </v-list>
          </template>

          <template v-else>
            <div class="admin-chat__section-title">
              Cuộc hội thoại
              <v-btn icon size="x-small" variant="text" :loading="loadingRooms" @click="loadRooms">
                <v-icon>mdi-refresh</v-icon>
              </v-btn>
            </div>
            <LoadingSpinner v-if="loadingRooms" />
            <div v-else-if="!supportRooms.length" class="admin-chat__empty">
              Chưa có hội thoại nào
            </div>
            <v-list v-else density="compact" class="pa-0">
              <v-list-item
                v-for="room in supportRooms"
                :key="room.id"
                :active="chat.activeRoomId.value === room.id"
                rounded="lg"
                class="mb-1"
                @click="selectRoom(room)"
              >
                <template #prepend>
                  <v-avatar color="primary" size="32" variant="tonal">
                    <v-icon size="16">mdi-headset</v-icon>
                  </v-avatar>
                </template>
                <v-list-item-title class="text-body-2 font-weight-medium">
                  {{ customerLabel(room.customer_id) }}
                </v-list-item-title>
                <v-list-item-subtitle class="text-caption">
                  {{ formatRoomDate(room.created_at) }}
                </v-list-item-subtitle>
              </v-list-item>
            </v-list>
          </template>
        </div>
      </aside>

      <section class="admin-chat__panel">
        <div v-if="!chat.activeRoomId.value" class="admin-chat__placeholder">
          <v-icon size="64" color="primary" class="mb-4">mdi-chat-processing-outline</v-icon>
          <p class="text-h6 mb-2">Chọn khách để bắt đầu chat</p>
          <p class="text-body-2 text-medium-emphasis">
            Danh sách bên trái hiển thị khách đang online hoặc các hội thoại đã có.
          </p>
        </div>

        <template v-else>
          <div class="admin-chat__panel-header">
            <div>
              <div class="font-weight-bold">{{ activeCustomerLabel }}</div>
              <div class="text-caption text-medium-emphasis">Hỗ trợ trực tuyến</div>
            </div>
          </div>

          <div ref="messagesEl" class="admin-chat__messages">
            <EmptyState
              v-if="!messages.length"
              icon="mdi-headset"
              title="Chưa có tin nhắn"
              description="Gửi lời chào để hỗ trợ khách hàng."
            />
            <template v-else>
              <ChatMessage
                v-for="msg in messages"
                :key="msg.id"
                :message="msg"
                :current-user-id="auth.user.value?.id"
                show-reactions
                @react="onReact(msg.id, $event)"
              />
              <div v-if="typingText" class="text-caption text-grey pa-2">{{ typingText }}</div>
            </template>
          </div>

          <div class="admin-chat__composer">
            <ChatComposer
              v-model="input"
              :room-id="chat.activeRoomId.value"
              :disabled="!chat.connected.value"
              :uploading="uploadingImage"
              placeholder="Nhắn tin cho khách..."
              hint="Gửi ảnh (tối đa 5MB) hoặc thả cảm xúc lên tin nhắn"
              @send="send"
              @typing="onTyping"
              @send-image="onSendImage"
            />
          </div>
        </template>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { OnlineUser, SupportChatRoom } from '~/types'

definePageMeta({
  layout: 'admin',
  adminSubtitle: 'Chat trực tuyến với khách hàng',
})

const auth = useAuth()
const chat = useChat()
const analytics = useAnalytics()

const sidebarTab = ref<'online' | 'rooms'>('online')
const supportRooms = ref<SupportChatRoom[]>([])
const onlineUsers = ref<OnlineUser[]>([])
const loadingOnline = ref(false)
const loadingRooms = ref(false)
const input = ref('')
const uploadingImage = ref(false)
const messagesEl = ref<HTMLElement | null>(null)
const activeCustomerId = ref<string | null>(null)

const onlineCustomers = computed(() =>
  onlineUsers.value.filter((u) => !u.role || u.role === 'customer'),
)

const messages = computed(() => chat.activeMessages.value)

const activeCustomerLabel = computed(() => {
  if (!activeCustomerId.value) {
    return 'Khách hàng'
  }
  return customerLabel(activeCustomerId.value)
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
  return 'Khách đang nhập...'
})

function shortId(id: string) {
  return `Khách #${id.slice(0, 8)}`
}

function customerLabel(customerId: string) {
  const online = onlineUsers.value.find((u) => u.user_id === customerId)
  if (online?.email) {
    return online.email
  }
  const room = supportRooms.value.find((r) => r.customer_id === customerId)
  if (room) {
    return shortId(customerId)
  }
  return shortId(customerId)
}

function formatRoomDate(value: string) {
  return new Date(value).toLocaleString('vi-VN', {
    day: '2-digit',
    month: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesEl.value) {
      messagesEl.value.scrollTop = messagesEl.value.scrollHeight
    }
  })
}

async function loadOnline() {
  loadingOnline.value = true
  try {
    const res = await analytics.fetchOnlineUsers()
    onlineUsers.value = res.users
  } catch {
    onlineUsers.value = []
  } finally {
    loadingOnline.value = false
  }
}

async function loadRooms() {
  loadingRooms.value = true
  try {
    supportRooms.value = await chat.fetchStaffSupportRooms()
  } catch {
    supportRooms.value = []
  } finally {
    loadingRooms.value = false
  }
}

async function openChatWithCustomer(customerId: string) {
  activeCustomerId.value = customerId
  const room = await chat.openCustomerSupportRoom(customerId)
  activeCustomerId.value = room.customer_id || customerId
  await loadRooms()
  scrollToBottom()
}

async function selectRoom(room: SupportChatRoom) {
  activeCustomerId.value = room.customer_id
  if (!chat.connected.value) {
    chat.connect()
  }
  chat.joinRoom(room.id)
  await chat.fetchMessages(room.id)
  scrollToBottom()
}

const send = () => {
  const text = input.value.trim()
  const roomId = chat.activeRoomId.value
  if (!text || !roomId) {
    return
  }
  input.value = ''
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
  const roomId = chat.activeRoomId.value
  if (roomId) {
    chat.sendTyping(roomId)
  }
}

watch(
  () => messages.value.length,
  () => scrollToBottom(),
)

let onlineTimer: ReturnType<typeof setInterval>

onMounted(async () => {
  chat.connect()
  await Promise.all([loadOnline(), loadRooms()])
  onlineTimer = setInterval(loadOnline, 30_000)
})

onUnmounted(() => {
  clearInterval(onlineTimer)
})
</script>

<style scoped>
.admin-chat__layout {
  display: grid;
  grid-template-columns: minmax(280px, 340px) 1fr;
  gap: 16px;
  min-height: calc(100vh - 220px);
}

.admin-chat__sidebar,
.admin-chat__panel {
  background: var(--color-surface, #fff);
  border: 1px solid var(--color-border, rgba(0, 0, 0, 0.08));
  border-radius: var(--admin-radius-sm);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.admin-chat__sidebar-body {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
}

.admin-chat__section-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  font-size: 0.8125rem;
  font-weight: 600;
  margin-bottom: 12px;
  color: var(--color-text-muted, #666);
}

.admin-chat__empty {
  text-align: center;
  padding: 24px 12px;
  font-size: 0.875rem;
  color: var(--color-text-muted, #888);
}

.admin-chat__panel {
  min-height: 520px;
}

.admin-chat__placeholder {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 24px;
  text-align: center;
}

.admin-chat__panel-header {
  padding: 14px 16px;
  border-bottom: 1px solid var(--color-border, rgba(0, 0, 0, 0.08));
}

.admin-chat__messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  min-height: 360px;
  background: var(--color-bg, #fafafa);
}

.admin-chat__composer {
  padding: 12px 16px 16px;
  border-top: 1px solid var(--color-border, rgba(0, 0, 0, 0.08));
}

@media (max-width: 960px) {
  .admin-chat__layout {
    grid-template-columns: 1fr;
    min-height: auto;
  }
}
</style>
