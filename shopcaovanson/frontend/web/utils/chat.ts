export const BOT_USER_ID = '00000000-0000-0000-0000-000000000001'

export const BOT_DISPLAY_NAME = 'Trợ lý Shop'

export const CHAT_INPUT_EMOJIS = ['😀', '😊', '😂', '❤️', '👍', '🎉', '🔥', '🙏', '👏', '😍', '😢', '🤔']

export const CHAT_REACTION_EMOJIS = ['👍', '❤️', '😂', '😮', '😢', '🙏']

export function chatMediaUrl(content: string): string {
  if (!content) return ''
  if (content.startsWith('http://') || content.startsWith('https://') || content.startsWith('data:')) {
    return content
  }
  const config = useRuntimeConfig()
  const base = (config.public.apiUrl as string || '').replace(/\/$/, '')
  return `${base}${content.startsWith('/') ? content : `/${content}`}`
}
