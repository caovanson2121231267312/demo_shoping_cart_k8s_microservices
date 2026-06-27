export const BOT_USER_ID = '00000000-0000-0000-0000-000000000001'

export const BOT_DISPLAY_NAME = 'Trợ lý Shop'

export const CHAT_INPUT_EMOJIS = ['😀', '😊', '😂', '❤️', '👍', '🎉', '🔥', '🙏', '👏', '😍', '😢', '🤔']

export const CHAT_REACTION_EMOJIS = ['👍', '❤️', '😂', '😮', '😢', '🙏']

export function getChatTypingLabel(othersCount: number, isStaffView: boolean): string {
  if (othersCount <= 0) {
    return ''
  }
  if (isStaffView) {
    return othersCount === 1
      ? 'Khách hàng đang nhập tin nhắn...'
      : `${othersCount} người đang nhập tin nhắn...`
  }
  return 'Nhân viên đang nhập tin nhắn...'
}

export function chatMediaUrl(content: string): string {
  if (!content) return ''
  if (content.startsWith('http://') || content.startsWith('https://') || content.startsWith('data:')) {
    return content
  }
  const config = useRuntimeConfig()
  const base = (config.public.apiUrl as string || '').replace(/\/$/, '')
  return `${base}${content.startsWith('/') ? content : `/${content}`}`
}
