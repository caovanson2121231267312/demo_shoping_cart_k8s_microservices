const DEBUG_KEY = 'chat:debug'

export function isChatDebugEnabled(): boolean {
  if (!import.meta.client) {
    return false
  }
  if (import.meta.dev) {
    return true
  }
  try {
    return localStorage.getItem(DEBUG_KEY) === '1'
  } catch {
    return false
  }
}

export function wsStateLabel(socket: WebSocket | null): string {
  if (!socket) {
    return 'null'
  }
  switch (socket.readyState) {
    case WebSocket.CONNECTING:
      return 'CONNECTING'
    case WebSocket.OPEN:
      return 'OPEN'
    case WebSocket.CLOSING:
      return 'CLOSING'
    case WebSocket.CLOSED:
      return 'CLOSED'
    default:
      return String(socket.readyState)
  }
}

export function chatLog(...args: unknown[]) {
  if (!isChatDebugEnabled()) {
    return
  }
  const ts = new Date().toISOString().slice(11, 23)
  console.log(`[chat ${ts}]`, ...args)
}

export function chatWarn(...args: unknown[]) {
  if (!isChatDebugEnabled()) {
    return
  }
  const ts = new Date().toISOString().slice(11, 23)
  console.warn(`[chat ${ts}]`, ...args)
}
