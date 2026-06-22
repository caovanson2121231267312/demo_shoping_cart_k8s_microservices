export type SnackbarTone = 'success' | 'error' | 'info' | 'warning' | 'primary'

const TONE_COLORS: Record<SnackbarTone, string> = {
  success: 'success',
  error: 'error',
  info: 'info',
  warning: 'warning',
  primary: 'primary',
}

const TONE_ICONS: Record<SnackbarTone, string> = {
  success: 'mdi-check-circle-outline',
  error: 'mdi-alert-circle-outline',
  info: 'mdi-information-outline',
  warning: 'mdi-alert-outline',
  primary: 'mdi-bell-outline',
}

const TONE_TITLES: Record<SnackbarTone, string> = {
  success: 'Thành công',
  error: 'Lỗi',
  info: 'Thông báo',
  warning: 'Cảnh báo',
  primary: 'Thông báo',
}

const TONE_TIMEOUT: Record<SnackbarTone, number> = {
  success: 3200,
  error: 4500,
  info: 3500,
  warning: 4000,
  primary: 3200,
}

export const useSnackbar = () => {
  const message = useState<string>('snackbar-msg', () => '')
  const title = useState<string>('snackbar-title', () => '')
  const visible = useState<boolean>('snackbar-visible', () => false)
  const color = useState<string>('snackbar-color', () => 'primary')
  const tone = useState<SnackbarTone>('snackbar-tone', () => 'primary')
  let timer: ReturnType<typeof setTimeout> | null = null

  const icon = computed(() => TONE_ICONS[tone.value] ?? TONE_ICONS.primary)

  const hide = () => {
    visible.value = false
    if (timer) {
      clearTimeout(timer)
      timer = null
    }
  }

  const show = (
    msg: string,
    nextTone: SnackbarTone = 'primary',
    nextTitle?: string,
  ) => {
    message.value = msg
    tone.value = nextTone
    color.value = TONE_COLORS[nextTone] ?? 'primary'
    title.value = nextTitle ?? TONE_TITLES[nextTone] ?? 'Thông báo'
    visible.value = true

    if (timer) clearTimeout(timer)
    timer = setTimeout(hide, TONE_TIMEOUT[nextTone] ?? 3200)
  }

  return { message, title, visible, color, tone, icon, show, hide }
}
