import type { ChatMessage } from '~/types'
import { BOT_USER_ID } from '~/utils/chat'

const SESSION_KEY = 'shop_chatbot_session'

export const useChatbot = () => {
  const config = useRuntimeConfig()
  const authStore = useAuthStore()
  const messages = ref<ChatMessage[]>([])
  const loading = ref(false)

  const sessionId = computed(() => {
    if (!import.meta.client) {
      return 'guest'
    }
    let id = localStorage.getItem(SESSION_KEY)
    if (!id) {
      id = `guest-${crypto.randomUUID()}`
      localStorage.setItem(SESSION_KEY, id)
    }
    return id
  })

  const sendMessage = async (text: string) => {
    const content = text.trim()
    if (!content) {
      return
    }

    const userMsg: ChatMessage = {
      id: `local-${Date.now()}`,
      room_id: 'bot',
      sender_id: sessionId.value,
      content,
      type: 'text',
      created_at: new Date().toISOString(),
    }
    messages.value.push(userMsg)
    loading.value = true

    try {
      const base = (config.public.apiUrl as string) || ''
      const metadata: Record<string, string> = {}
      if (authStore.accessToken) {
        metadata.access_token = authStore.accessToken
      }

      const res = await $fetch<{ text: string }>(`${base}/api/chatbot/message`, {
        method: 'POST',
        body: {
          sender_id: sessionId.value,
          message: content,
          metadata,
        },
      })

      messages.value.push({
        id: `bot-${Date.now()}`,
        room_id: 'bot',
        sender_id: BOT_USER_ID,
        content: res.text,
        type: 'text',
        created_at: new Date().toISOString(),
      })
    } catch {
      messages.value.push({
        id: `bot-err-${Date.now()}`,
        room_id: 'bot',
        sender_id: BOT_USER_ID,
        content: 'Xin lỗi, trợ lý đang bận. Vui lòng thử lại sau hoặc gọi 1900 1234.',
        type: 'text',
        created_at: new Date().toISOString(),
      })
    } finally {
      loading.value = false
    }
  }

  const greetIfEmpty = async () => {
    if (messages.value.length > 0) {
      return
    }
    messages.value.push({
      id: 'bot-welcome',
      room_id: 'bot',
      sender_id: BOT_USER_ID,
      content:
        'Xin chào! Mình là trợ lý ảo Shop Cao Văn Sơn. Hỏi mình về sản phẩm, đơn hàng, bảo hành, giờ làm việc, giao hàng hoặc thanh toán nhé.',
      type: 'text',
      created_at: new Date().toISOString(),
    })
  }

  return {
    messages,
    loading,
    sessionId,
    sendMessage,
    greetIfEmpty,
  }
}
