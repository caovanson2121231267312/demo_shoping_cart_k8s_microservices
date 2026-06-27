import { defineStore } from 'pinia'
import type { ChatMessage, ChatRoom } from '~/types'

export const useChatStore = defineStore('chat', () => {
  const rooms = ref<ChatRoom[]>([])
  const messages = ref<Record<string, ChatMessage[]>>({})
  const activeRoomId = ref<string | null>(null)
  const connected = ref(false)
  const typingUsers = ref<Record<string, string[]>>({})
  const unreadCount = ref(0)
  const panelOpen = ref(false)

  const activeMessages = computed(() => {
    if (!activeRoomId.value) {
      return []
    }
    return messages.value[activeRoomId.value] || []
  })

  function setRooms(value: ChatRoom[]) {
    rooms.value = value
  }

  function setActiveRoom(roomId: string | null) {
    activeRoomId.value = roomId
  }

  function setConnected(value: boolean) {
    connected.value = value
  }

  function addMessage(roomId: string, message: ChatMessage) {
    const list = messages.value[roomId] || []
    const exists = list.some((m) => m.id === message.id)
    if (exists) {
      return
    }
    const localIdx = list.findIndex(
      (m) =>
        m.id.startsWith('local-') &&
        m.sender_id === message.sender_id &&
        m.content === message.content &&
        m.type === message.type,
    )
    if (localIdx >= 0) {
      const next = [...list]
      next[localIdx] = message
      messages.value[roomId] = next
      return
    }
    messages.value[roomId] = [...list, message]
  }

  function updateMessageReactions(roomId: string, messageId: string, reactions: Record<string, string[]>) {
    const list = messages.value[roomId]
    if (!list) {
      return
    }
    const idx = list.findIndex((m) => m.id === messageId)
    if (idx < 0) {
      return
    }
    const next = [...list]
    next[idx] = { ...next[idx], reactions: { ...reactions } }
    messages.value[roomId] = next
  }

  function setMessages(roomId: string, list: ChatMessage[]) {
    messages.value[roomId] = list
  }

  function prependMessages(roomId: string, list: ChatMessage[]) {
    const existing = messages.value[roomId] || []
    const ids = new Set(existing.map((m) => m.id))
    const merged = [...list.filter((m) => !ids.has(m.id)), ...existing]
    messages.value[roomId] = merged
  }

  function setTyping(roomId: string, userId: string, isTyping: boolean) {
    const current = typingUsers.value[roomId] || []
    if (isTyping) {
      if (current.includes(userId)) {
        return
      }
      typingUsers.value = { ...typingUsers.value, [roomId]: [...current, userId] }
    } else {
      if (!current.includes(userId)) {
        return
      }
      typingUsers.value = { ...typingUsers.value, [roomId]: current.filter((id) => id !== userId) }
    }
  }

  function incrementUnread() {
    unreadCount.value += 1
  }

  function resetUnread() {
    unreadCount.value = 0
  }

  function setPanelOpen(value: boolean) {
    panelOpen.value = value
    if (value) {
      resetUnread()
    }
  }

  return {
    rooms,
    messages,
    activeRoomId,
    connected,
    typingUsers,
    unreadCount,
    panelOpen,
    activeMessages,
    setRooms,
    setActiveRoom,
    setConnected,
    addMessage,
    updateMessageReactions,
    setMessages,
    prependMessages,
    setTyping,
    incrementUnread,
    resetUnread,
    setPanelOpen,
  }
})
