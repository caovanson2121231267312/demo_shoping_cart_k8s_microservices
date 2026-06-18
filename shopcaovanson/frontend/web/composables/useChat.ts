import { useDebounceFn } from '@vueuse/core'
import type { ChatMessage, ChatRoom, MessageListResult, WSClientMessage, WSServerMessage } from '~/types'

const RECONNECT_DELAYS = [1000, 2000, 4000, 8000, 16000, 30000]

export const useChat = () => {
  const chatStore = useChatStore()
  const authStore = useAuthStore()
  const config = useRuntimeConfig()
  const { apiFetch, isLoggedIn } = useAuth()

  let socket: WebSocket | null = null
  let reconnectAttempt = 0
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let typingTimeout: ReturnType<typeof setTimeout> | null = null

  const wsBaseUrl = computed(() => config.public.wsUrl as string)

  const connect = () => {
    if (!import.meta.client || !isLoggedIn.value || !authStore.accessToken) {
      return
    }
    if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) {
      return
    }

    const token = encodeURIComponent(authStore.accessToken)
    const url = `${wsBaseUrl.value}/ws?token=${token}`
    socket = new WebSocket(url)

    socket.onopen = () => {
      chatStore.setConnected(true)
      reconnectAttempt = 0
    }

    socket.onclose = () => {
      chatStore.setConnected(false)
      scheduleReconnect()
    }

    socket.onerror = () => {
      socket?.close()
    }

    socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data) as WSServerMessage
        handleServerMessage(data)
      } catch {
        // ignore malformed messages
      }
    }
  }

  const disconnect = () => {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    reconnectAttempt = 0
    if (socket) {
      socket.onclose = null
      socket.close()
      socket = null
    }
    chatStore.setConnected(false)
  }

  const scheduleReconnect = () => {
    if (!isLoggedIn.value) {
      return
    }
    const delay = RECONNECT_DELAYS[Math.min(reconnectAttempt, RECONNECT_DELAYS.length - 1)]
    reconnectAttempt += 1
    reconnectTimer = setTimeout(() => {
      connect()
    }, delay)
  }

  const sendWs = (message: WSClientMessage) => {
    if (socket?.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify(message))
    }
  }

  const handleServerMessage = (data: WSServerMessage) => {
    if (data.type === 'message' && data.room_id && data.sender_id && data.content) {
      const message: ChatMessage = {
        id: `${data.room_id}-${data.created_at || Date.now()}`,
        room_id: data.room_id,
        sender_id: data.sender_id,
        content: data.content,
        type: 'text',
        created_at: data.created_at || new Date().toISOString(),
      }
      chatStore.addMessage(data.room_id, message)
      if (!chatStore.panelOpen || chatStore.activeRoomId !== data.room_id) {
        chatStore.incrementUnread()
      }
    }

    if (data.type === 'typing' && data.room_id && data.user_id) {
      chatStore.setTyping(data.room_id, data.user_id, true)
      setTimeout(() => {
        chatStore.setTyping(data.room_id!, data.user_id!, false)
      }, 3000)
    }
  }

  const fetchRooms = async () => {
    const rooms = await apiFetch<ChatRoom[]>('/api/chat/rooms')
    chatStore.setRooms(rooms)
    return rooms
  }

  const fetchMessages = async (roomId: string, before?: string, limit = 50) => {
    const query: Record<string, string | number> = { limit }
    if (before) {
      query.before = before
    }
    const result = await apiFetch<MessageListResult>(`/api/chat/rooms/${roomId}/messages`, { query })
    if (before) {
      chatStore.prependMessages(roomId, result.items)
    } else {
      chatStore.setMessages(roomId, result.items)
    }
    return result
  }

  const createRoom = async (participantId: string) => {
    const room = await apiFetch<ChatRoom>('/api/chat/rooms', {
      method: 'POST',
      body: { participant_id: participantId },
    })
    chatStore.setRooms([room, ...chatStore.rooms.filter((r) => r.id !== room.id)])
    return room
  }

  const joinRoom = (roomId: string) => {
    chatStore.setActiveRoom(roomId)
    sendWs({ type: 'join', room_id: roomId })
  }

  const sendMessage = (roomId: string, content: string) => {
    if (!content.trim()) {
      return
    }
    sendWs({ type: 'message', room_id: roomId, content: content.trim() })
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
    const supportRoom = rooms.find((r) => r.room_type === 'support') || rooms[0]
    if (supportRoom) {
      joinRoom(supportRoom.id)
      if (!chatStore.messages[supportRoom.id]) {
        await fetchMessages(supportRoom.id)
      }
      return supportRoom
    }
    return null
  }

  const openChat = async () => {
    chatStore.setPanelOpen(true)
    connect()
    await ensureSupportRoom()
  }

  const closeChat = () => {
    chatStore.setPanelOpen(false)
  }

  watch(isLoggedIn, (loggedIn) => {
    if (loggedIn) {
      connect()
      fetchRooms().catch(() => {})
    } else {
      disconnect()
      chatStore.setRooms([])
      chatStore.setActiveRoom(null)
    }
  })

  onUnmounted(() => {
    disconnect()
  })

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
    joinRoom,
    sendMessage,
    sendTyping,
    ensureSupportRoom,
    openChat,
    closeChat,
  }
}
