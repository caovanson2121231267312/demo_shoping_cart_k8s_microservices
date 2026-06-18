export const useSnackbar = () => {
  const message = useState<string>('snackbar-msg', () => '')
  const visible = useState<boolean>('snackbar-visible', () => false)
  const color = useState<string>('snackbar-color', () => 'primary')
  let timer: ReturnType<typeof setTimeout> | null = null

  const show = (msg: string, tone: 'success' | 'error' | 'info' | 'primary' = 'primary') => {
    message.value = msg
    color.value = tone === 'info' ? 'grey-darken-3' : tone
    visible.value = true
    if (timer) clearTimeout(timer)
    timer = setTimeout(() => {
      visible.value = false
    }, 2800)
  }

  return { message, visible, color, show }
}
