import type { ChatMessage, ChatProductSuggestion } from '~/types'
import { BOT_USER_ID } from '~/utils/chat'

const SESSION_KEY = 'shop_chatbot_session'

function detectBrowserLang(): 'vi' | 'en' {
  if (!import.meta.client) {
    return 'vi'
  }
  const lang = navigator.language.toLowerCase()
  return lang.startsWith('en') ? 'en' : 'vi'
}

const WELCOME: Record<'vi' | 'en', string[]> = {
  vi: [
    'Xin chào! Mình là trợ lý ảo Shop Cao Văn Sơn. Hỏi mình về sản phẩm, đơn hàng, bảo hành, mã giảm giá, giao hàng hoặc thanh toán nhé.',
    'Chào bạn! Cần tìm sản phẩm, lọc theo giá, tra đơn hay hỏi chính sách shop? Cứ nhắn mình.',
    'Shop Cao Văn Sơn xin chào! Thử: “tìm laptop”, “dưới 5 triệu”, “giao hàng bao lâu”, “mã giảm giá”. (Chat tiếng Anh cũng được!)',
    'Hi! Mình online 24/7 — tư vấn sản phẩm, đơn hàng, bảo hành, đổi trả, thanh toán COD.',
  ],
  en: [
    "Hello! I'm the Shop Cao Van Son assistant. Ask about products, orders, warranty, coupons, shipping, or payment.",
    'Hi there! Try: find laptop, under 5 million VND, track my order, or discount code.',
    'Welcome! I can search products, filter by price, and answer shop policies 24/7. (Vietnamese works too!)',
    'Hey! Need product suggestions, order help, or policy info? Just ask.',
  ],
}

const ERROR_MSG: Record<'vi' | 'en', string[]> = {
  vi: [
    'Xin lỗi, trợ lý đang bận. Vui lòng thử lại sau hoặc gọi 1900 1234.',
    'Hệ thống đang quá tải. Bạn thử lại sau ít phút hoặc gọi hotline 1900 1234 nhé.',
  ],
  en: [
    'Sorry, the assistant is busy. Please try again later or call 1900 1234.',
    'System is overloaded. Please retry shortly or call 1900 1234.',
  ],
}

function pickRandom<T>(items: T[]): T {
  return items[Math.floor(Math.random() * items.length)]
}

export const useChatbot = () => {
  const config = useRuntimeConfig()
  const authStore = useAuthStore()
  const messages = ref<ChatMessage[]>([])
  const loading = ref(false)
  const lang = ref<'vi' | 'en'>('vi')

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
      const metadata: Record<string, string> = {
        locale: lang.value,
      }
      if (authStore.accessToken) {
        metadata.access_token = authStore.accessToken
      }

      const res = await $fetch<{
        text: string
        products?: ChatProductSuggestion[]
      }>(`${base}/api/chatbot/message`, {
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
        products: res.products?.length ? res.products : undefined,
        created_at: new Date().toISOString(),
      })
    } catch {
      messages.value.push({
        id: `bot-err-${Date.now()}`,
        room_id: 'bot',
        sender_id: BOT_USER_ID,
        content: pickRandom(ERROR_MSG[lang.value]),
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
    lang.value = detectBrowserLang()
    messages.value.push({
      id: 'bot-welcome',
      room_id: 'bot',
      sender_id: BOT_USER_ID,
      content: pickRandom(WELCOME[lang.value]),
      type: 'text',
      created_at: new Date().toISOString(),
    })
  }

  return {
    messages,
    loading,
    sessionId,
    lang,
    sendMessage,
    greetIfEmpty,
  }
}
