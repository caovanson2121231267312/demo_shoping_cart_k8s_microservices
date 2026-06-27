import { useDebounceFn } from '@vueuse/core'
import type { ChatMessage, ChatRoom, SupportChatRoom, WSClientMessage, WSServerMessage } from '~/types'
import { chatLog, chatWarn, wsStateLabel } from '~/utils/chatDebug'

const RECONNECT_DELAYS = [1000, 2000, 4000, 8000, 16000, 30000]

// Shared singleton — one WebSocket for the whole app (avoid duplicate connections per component).
let socket: WebSocket | null = null
let reconnectAttempt = 0
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
let typingTimeout: ReturnType<typeof setTimeout> | null = null
let loginWatchRegistered = false
const pendingOutbound: WSClientMessage[] = []

function dedupePendingJoins() {
  const latestJoinByRoom = new Map<string, WSClientMessage>()
  const rest: WSClientMessage[] = []
  for (const msg of pendingOutbound) {
    if (msg.type === 'join') {
      latestJoinByRoom.set(msg.room_id, msg)
    } else {
      rest.push(msg)
    }
  }
  pendingOutbound.length = 0
  pendingOutbound.push(...latestJoinByRoom.values(), ...rest)
}

function flushPendingOutbound() {
  dedupePendingJoins()
  while (pendingOutbound.length > 0 && socket?.readyState === WebSocket.OPEN) {
    const message = pendingOutbound.shift()!
    socket.send(JSON.stringify(message))
    chatLog('ws>>flush', message.type, message.room_id, message.content?.slice(0, 60) || '')
  }
}

export const useChat = () => {
  const chatStore = useChatStore()
  const authStore = useAuthStore()
  const config = useRuntimeConfig()
  const { apiFetch, isLoggedIn } = useAuth()

  const wsBaseUrl = computed(() => config.public.wsUrl as string)

  const sendWs = (message: WSClientMessage) => {
    if (socket?.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify(message))
      chatLog('ws>>', message.type, message.room_id, message.content?.slice(0, 60) || '')
      return
    }

    if (message.type === 'join' || message.type === 'message') {
      pendingOutbound.push(message)
      dedupePendingJoins()
      chatWarn('ws>>queued', message.type, message.room_id, 'socket=', wsStateLabel(socket))
      return
    }

    chatWarn('ws>>DROP', message.type, message.room_id, 'socket=', wsStateLabel(socket))
  }

  const connect = () => {
    if (!import.meta.client || !isLoggedIn.value || !authStore.accessToken) {
      chatLog('connect skipped', {
        client: import.meta.client,
        loggedIn: isLoggedIn.value,
        hasToken: Boolean(authStore.accessToken),
      })
      return
    }
    if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) {
      chatLog('connect skipped — socket already', wsStateLabel(socket))
      return
    }

    const token = encodeURIComponent(authStore.accessToken)
    const url = `${wsBaseUrl.value}/ws?token=${token.slice(0, 12)}...`
    chatLog('connecting', url, 'user=', authStore.user?.id, 'role=', authStore.user?.role)
    socket = new WebSocket(`${wsBaseUrl.value}/ws?token=${token}`)

    socket.onopen = () => {
      chatLog('ws OPEN', 'user=', authStore.user?.id, 'activeRoom=', chatStore.activeRoomId)
      chatStore.setConnected(true)
      reconnectAttempt = 0
      if (chatStore.activeRoomId) {
        sendWs({ type: 'join', room_id: chatStore.activeRoomId })
      }
      flushPendingOutbound()
    }

    socket.onclose = (event) => {
      chatWarn('ws CLOSED', 'code=', event.code, 'reason=', event.reason || '(none)', 'wasClean=', event.wasClean)
      chatStore.setConnected(false)
      socket = null
      scheduleReconnect()
    }

    socket.onerror = () => {
      chatWarn('ws ERROR', 'socket=', wsStateLabel(socket))
      socket?.close()
    }

    socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data) as WSServerMessage | { error?: string }
        if ('error' in data && data.error) {
          chatWarn('ws<<error', data.error)
          return
        }
        chatLog('ws<<', (data as WSServerMessage).type, (data as WSServerMessage).room_id, (data as WSServerMessage).content?.slice(0, 60) || '')
        handleServerMessage(data as WSServerMessage)
      } catch (err) {
        chatWarn('ws<<parse failed', event.data, err)
      }
    }
  }

  const disconnect = () => {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    reconnectAttempt = 0
    pendingOutbound.length = 0
    if (socket) {
      socket.onclose = null
      socket.close()
      socket = null
    }
    chatLog('disconnect')
    chatStore.setConnected(false)
  }

  const scheduleReconnect = () => {
    if (!isLoggedIn.value) {
      return
    }
    const delay = RECONNECT_DELAYS[Math.min(reconnectAttempt, RECONNECT_DELAYS.length - 1)]
    reconnectAttempt += 1
    chatLog('reconnect in', delay, 'ms', 'attempt=', reconnectAttempt)
    reconnectTimer = setTimeout(() => {
      connect()
    }, delay)
  }

  const handleServerMessage = (data: WSServerMessage) => {
    if (data.type === 'message' && data.room_id && data.sender_id && data.content) {
      const message: ChatMessage = {
        id: data.message_id || `${data.room_id}-${data.created_at || Date.now()}`,
        room_id: data.room_id,
        sender_id: data.sender_id,
        content: data.content,
        type: data.msg_type || 'text',
        reactions: data.reactions,
        created_at: data.created_at || new Date().toISOString(),
      }
      chatLog('message received', message.id, 'room=', data.room_id, 'from=', data.sender_id)
      chatStore.addMessage(data.room_id, message)
      if (!chatStore.panelOpen || chatStore.activeRoomId !== data.room_id) {
        chatStore.incrementUnread()
      }
    }

    if (data.type === 'reaction' && data.room_id && data.message_id && data.reactions) {
      chatStore.updateMessageReactions(data.room_id, data.message_id, data.reactions)
    }

    if (data.type === 'typing' && data.room_id && data.user_id) {
      chatStore.setTyping(data.room_id, data.user_id, true)
      setTimeout(() => {
        chatStore.setTyping(data.room_id!, data.user_id!, false)
      }, 3000)
    }

    if (data.type === 'joined' && data.room_id) {
      chatLog('joined ack', data.room_id, 'user=', data.user_id)
    }

    if (data.type === 'user_joined' && data.room_id && data.user_id) {
      chatLog('user joined room', data.room_id, 'user=', data.user_id)
    }
  }

  const fetchRooms = async () => {
    const res = await apiFetch<{ data: ChatRoom[] }>('/api/chat/rooms')
    const rooms = res.data || []
    chatLog('fetchRooms', rooms.length, rooms.map((r) => `${r.id.slice(0, 8)}:${r.room_type}`))
    chatStore.setRooms(rooms)
    return rooms
  }

  const fetchMessages = async (roomId: string, before?: string, limit = 50) => {
    const query: Record<string, string | number> = { limit }
    if (before) {
      query.before = before
    }
    chatLog('fetchMessages', roomId)
    const res = await apiFetch<{ data: ChatMessage[] }>(`/api/chat/rooms/${roomId}/messages`, { query })
    const items = (res.data || []).slice().reverse()
    chatLog('fetchMessages done', roomId, 'count=', items.length)
    if (before) {
      chatStore.prependMessages(roomId, items)
    } else {
      chatStore.setMessages(roomId, items)
    }
    return { items, total: items.length }
  }

  const createRoom = async (participantId: string) => {
    const res = await apiFetch<{ data: ChatRoom }>('/api/chat/rooms', {
      method: 'POST',
      body: { participant_id: participantId },
    })
    const room = res.data
    chatStore.setRooms([room, ...chatStore.rooms.filter((r) => r.id !== room.id)])
    return room
  }

  const createSupportRoom = async () => {
    const res = await apiFetch<{ data: ChatRoom }>('/api/chat/rooms/support', {
      method: 'POST',
    })
    const room = res.data
    chatLog('createSupportRoom', room.id)
    chatStore.setRooms([room, ...chatStore.rooms.filter((r) => r.id !== room.id)])
    return room
  }

  const joinRoom = (roomId: string) => {
    chatLog('joinRoom', roomId, 'socket=', wsStateLabel(socket))
    chatStore.setActiveRoom(roomId)
    sendWs({ type: 'join', room_id: roomId })
  }

  const sendMessage = (roomId: string, content: string, msgType: 'text' | 'image' = 'text') => {
    if (!content.trim()) {
      return
    }
    const trimmed = content.trim()
    const senderId = authStore.user?.id
    chatLog('sendMessage', roomId, trimmed.slice(0, 60), 'socket=', wsStateLabel(socket), 'activeRoom=', chatStore.activeRoomId)
    if (senderId) {
      chatStore.addMessage(roomId, {
        id: `local-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
        room_id: roomId,
        sender_id: senderId,
        content: trimmed,
        type: msgType,
        created_at: new Date().toISOString(),
      })
    }
    sendWs({ type: 'message', room_id: roomId, content: trimmed, msg_type: msgType })
  }

  const uploadChatImage = async (file: File): Promise<string> => {
    const base = (config.public.apiUrl as string || '').replace(/\/$/, '')
    const form = new FormData()
    form.append('file', file)
    const res = await fetch(`${base}/api/chat/upload`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${authStore.accessToken}`,
      },
      body: form,
    })
    if (!res.ok) {
      const err = await res.json().catch(() => ({}))
      throw new Error((err as { error?: string }).error || 'Upload ảnh thất bại')
    }
    const json = await res.json() as { data: { url: string } }
    return json.data.url
  }

  const sendImageMessage = async (roomId: string, file: File) => {
    const url = await uploadChatImage(file)
    sendMessage(roomId, url, 'image')
  }

  const toggleReaction = (roomId: string, messageId: string, emoji: string) => {
    if (!messageId || messageId.startsWith('local-')) {
      return
    }
    const userId = authStore.user?.id
    if (userId) {
      const list = chatStore.messages[roomId] || []
      const msg = list.find((m) => m.id === messageId)
      if (msg) {
        const reactions = { ...(msg.reactions || {}) }
        const users = [...(reactions[emoji] || [])]
        const idx = users.indexOf(userId)
        if (idx >= 0) {
          users.splice(idx, 1)
        } else {
          users.push(userId)
        }
        if (users.length) {
          reactions[emoji] = users
        } else {
          delete reactions[emoji]
        }
        chatStore.updateMessageReactions(roomId, messageId, reactions)
      }
    }
    sendWs({
      type: 'reaction',
      room_id: roomId,
      message_id: messageId,
      emoji,
    })
  }

  const sendTyping = useDebounceFn((roomId: string) => {
    sendWs({ type: 'typing', room_id: roomId })
    if (typingTimeout) {
      clearTimeout(typingTimeout)
    }
    typingTimeout = setTimeout(() => {
      // typing indicator auto-clears on server after 3s
    }, 3000)
  }, 500)

  const ensureSupportRoom = async () => {
    let rooms = chatStore.rooms
    if (rooms.length === 0) {
      rooms = await fetchRooms()
    }
    let supportRoom = rooms.find((r) => r.room_type === 'support')
    if (!supportRoom) {
      supportRoom = await createSupportRoom()
    }
    if (supportRoom) {
      if (!chatStore.connected) {
        connect()
      }
      joinRoom(supportRoom.id)
      if (!chatStore.messages[supportRoom.id]) {
        await fetchMessages(supportRoom.id)
      }
      return supportRoom
    }
    return null
  }

  const fetchStaffSupportRooms = async () => {
    const res = await apiFetch<{ data: SupportChatRoom[] }>('/api/chat/admin/support-rooms')
    const rooms = res.data || []
    chatLog('fetchStaffSupportRooms', rooms.length)
    chatStore.setRooms(rooms)
    return rooms
  }

  const openCustomerSupportRoom = async (customerId: string) => {
    chatLog('openCustomerSupportRoom', customerId)
    const res = await apiFetch<{ data: SupportChatRoom }>('/api/chat/admin/support-rooms', {
      method: 'POST',
      body: { customer_id: customerId },
    })
    const room = res.data
    chatLog('openCustomerSupportRoom ok', room.id, 'customer=', room.customer_id)
    chatStore.setRooms([room, ...chatStore.rooms.filter((r) => r.id !== room.id)])
    if (!chatStore.connected) {
      connect()
    }
    joinRoom(room.id)
    await fetchMessages(room.id)
    return room
  }

  const openChat = async () => {
    chatStore.setPanelOpen(true)
  }

  const closeChat = () => {
    chatStore.setPanelOpen(false)
  }

  if (!loginWatchRegistered) {
    loginWatchRegistered = true
    const auth = useAuth()
    watch(
      auth.isLoggedIn,
      (loggedIn) => {
        chatLog('auth changed', loggedIn ? 'login' : 'logout')
        if (loggedIn) {
          connect()
          fetchRooms().catch((err) => chatWarn('fetchRooms failed', err))
        } else {
          disconnect()
          chatStore.setRooms([])
          chatStore.setActiveRoom(null)
        }
      },
      { immediate: true },
    )
  }

  return {
    connected: computed(() => chatStore.connected),
    rooms: computed(() => chatStore.rooms),
    activeRoomId: computed(() => chatStore.activeRoomId),
    activeMessages: computed(() => chatStore.activeMessages),
    typingUsers: computed(() => chatStore.typingUsers),
    unreadCount: computed(() => chatStore.unreadCount),
    panelOpen: computed(() => chatStore.panelOpen),
    connect,
    disconnect,
    fetchRooms,
    fetchMessages,
    createRoom,
    createSupportRoom,
    joinRoom,
    sendMessage,
    sendImageMessage,
    toggleReaction,
    uploadChatImage,
    sendTyping,
    ensureSupportRoom,
    fetchStaffSupportRooms,
    openCustomerSupportRoom,
    openChat,
    closeChat,
  }
}
