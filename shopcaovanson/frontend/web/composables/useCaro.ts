import { io, type Socket } from 'socket.io-client'
import type { CaroChatMessage, CaroPublicRoom, CaroRoom, CaroSymbol } from '~/types'

type Board = CaroSymbol[][]

let socket: Socket | null = null

export const useCaro = () => {
  const auth = useAuth()
  const authStore = useAuthStore()
  const config = useRuntimeConfig()

  const connected = ref(false)
  const room = ref<CaroRoom | null>(null)
  const board = ref<Board>([])
  const boardSize = ref(15)
  const status = ref<string>('lobby')
  const currentTurn = ref<CaroSymbol>(null)
  const chatMessages = ref<CaroChatMessage[]>([])
  const publicRooms = ref<CaroPublicRoom[]>([])
  const winnerId = ref<string | null>(null)
  const rpsPhase = ref(false)
  const myRpsSent = ref(false)
  const voiceEnabled = ref(false)
  const remoteAudio = ref<HTMLAudioElement | null>(null)

  let peer: RTCPeerConnection | null = null
  let localStream: MediaStream | null = null

  const wsBase = computed(() => (config.public.wsUrl as string) || (config.public.apiUrl as string) || '')

  function ensureBoard(size: number) {
    boardSize.value = size
    board.value = Array.from({ length: size }, () => Array(size).fill(null))
  }

  function connect() {
    if (socket?.connected || !authStore.accessToken) return
    socket = io(wsBase.value, {
      path: '/ws/caro/socket.io',
      transports: ['websocket', 'polling'],
      auth: { token: authStore.accessToken },
    })

    socket.on('connect', () => { connected.value = true })
    socket.on('disconnect', () => { connected.value = false })

    socket.on('rps_start', () => {
      rpsPhase.value = true
      myRpsSent.value = false
      status.value = 'rps'
    })
    socket.on('rps_tie', () => {
      myRpsSent.value = false
    })
    socket.on('game_start', (payload: { current_turn: CaroSymbol }) => {
      rpsPhase.value = false
      status.value = 'playing'
      currentTurn.value = payload.current_turn
      if (room.value) room.value.status = 'playing'
    })
    socket.on('move', (payload: { row: number; col: number; symbol: CaroSymbol; current_turn: CaroSymbol; board: Board }) => {
      board.value = payload.board
      currentTurn.value = payload.current_turn
    })
    socket.on('game_over', (payload: { winner_id: string | null; board: Board }) => {
      board.value = payload.board
      winnerId.value = payload.winner_id
      status.value = 'finished'
      if (room.value) room.value.status = 'finished'
    })
    socket.on('chat', (msg: CaroChatMessage) => {
      chatMessages.value.push(msg)
    })
    socket.on('room_state', (payload: { board: Board | null; current_turn: CaroSymbol; status: string }) => {
      if (payload.board) board.value = payload.board
      currentTurn.value = payload.current_turn
      status.value = payload.status
    })
    socket.on('webrtc_signal', async (payload: { from: string; sdp?: RTCSessionDescriptionInit; candidate?: RTCIceCandidateInit }) => {
      await handleSignal(payload)
    })
  }

  function disconnect() {
    stopVoice()
    socket?.disconnect()
    socket = null
    connected.value = false
  }

  async function fetchPublicRooms() {
    const res = await auth.apiFetch<{ items: CaroPublicRoom[] }>('/api/caro/rooms/public')
    publicRooms.value = res.items
    return res.items
  }

  async function createRoom(visibility: 'public' | 'private', size: 15 | 20) {
    const user = auth.user.value
    const created = await auth.apiFetch<CaroRoom>('/api/caro/rooms', {
      method: 'POST',
      body: {
        visibility,
        board_size: size,
        host_name: user?.full_name || user?.email,
      },
    })
    room.value = created
    ensureBoard(size)
    await joinRoomSocket(created.code)
    return created
  }

  async function joinRoom(code: string) {
    const info = await auth.apiFetch<CaroRoom>(`/api/caro/rooms/${code}`)
    room.value = info
    ensureBoard(info.board_size)
    await joinRoomSocket(code)
  }

  function joinRoomSocket(code: string) {
    return new Promise<void>((resolve, reject) => {
      if (!socket) connect()
      socket?.emit('join_room', { code }, (res: { ok: boolean; error?: string; room?: CaroRoom; chat?: CaroChatMessage[] }) => {
        if (!res?.ok) {
          reject(new Error(res?.error || 'join failed'))
          return
        }
        if (res.room) {
          room.value = res.room
          boardSize.value = res.room.board_size
          status.value = res.room.status
          currentTurn.value = (res.room.current_turn as CaroSymbol) || null
        }
        if (res.chat) chatMessages.value = res.chat
        resolve()
      })
    })
  }

  function sendRps(choice: 'rock' | 'paper' | 'scissors') {
    myRpsSent.value = true
    socket?.emit('rps', { choice })
  }

  function sendMove(row: number, col: number) {
    socket?.emit('move', { row, col })
  }

  function sendChat(content: string) {
    if (!content.trim()) return
    socket?.emit('chat', { content })
  }

  async function getIceServers() {
    const cfg = await auth.apiFetch<{
      turn_url?: string
      turn_username?: string
      turn_credential?: string
      stun_urls?: string[]
    }>('/api/caro/webrtc/config')
    const servers: RTCIceServer[] = (cfg.stun_urls || ['stun:stun.l.google.com:19302']).map((u) => ({ urls: u }))
    if (cfg.turn_url && cfg.turn_username && cfg.turn_credential) {
      servers.push({
        urls: cfg.turn_url,
        username: cfg.turn_username,
        credential: cfg.turn_credential,
      })
    }
    return servers
  }

  async function startVoice() {
    if (voiceEnabled.value || !socket) return
    const iceServers = await getIceServers()
    peer = new RTCPeerConnection({ iceServers })
    localStream = await navigator.mediaDevices.getUserMedia({ audio: true, video: false })
    localStream.getTracks().forEach((t) => peer!.addTrack(t, localStream!))

    peer.ontrack = (ev) => {
      if (!remoteAudio.value) remoteAudio.value = new Audio()
      remoteAudio.value.srcObject = ev.streams[0]
      remoteAudio.value.play().catch(() => {})
    }
    peer.onicecandidate = (ev) => {
      if (ev.candidate) socket?.emit('webrtc_signal', { candidate: ev.candidate.toJSON() })
    }

    const offer = await peer.createOffer()
    await peer.setLocalDescription(offer)
    socket.emit('webrtc_signal', { sdp: offer })
    voiceEnabled.value = true
  }

  async function handleSignal(payload: { sdp?: RTCSessionDescriptionInit; candidate?: RTCIceCandidateInit }) {
    if (!peer) {
      const iceServers = await getIceServers()
      peer = new RTCPeerConnection({ iceServers })
      peer.ontrack = (ev) => {
        if (!remoteAudio.value) remoteAudio.value = new Audio()
        remoteAudio.value.srcObject = ev.streams[0]
        remoteAudio.value.play().catch(() => {})
      }
      peer.onicecandidate = (ev) => {
        if (ev.candidate) socket?.emit('webrtc_signal', { candidate: ev.candidate.toJSON() })
      }
      if (!localStream && voiceEnabled.value) {
        localStream = await navigator.mediaDevices.getUserMedia({ audio: true, video: false })
        localStream.getTracks().forEach((t) => peer!.addTrack(t, localStream!))
      }
    }

    if (payload.sdp) {
      await peer.setRemoteDescription(payload.sdp)
      if (payload.sdp.type === 'offer') {
        const answer = await peer.createAnswer()
        await peer.setLocalDescription(answer)
        socket?.emit('webrtc_signal', { sdp: answer })
      }
    } else if (payload.candidate) {
      await peer.addIceCandidate(payload.candidate)
    }
  }

  function stopVoice() {
    voiceEnabled.value = false
    localStream?.getTracks().forEach((t) => t.stop())
    localStream = null
    peer?.close()
    peer = null
    if (remoteAudio.value) {
      remoteAudio.value.pause()
      remoteAudio.value.srcObject = null
    }
  }

  function resetLocal() {
    room.value = null
    winnerId.value = null
    chatMessages.value = []
    rpsPhase.value = false
    myRpsSent.value = false
    status.value = 'lobby'
    currentTurn.value = null
    ensureBoard(15)
  }

  const isMyTurn = computed(() => {
    if (!room.value?.you_are || !currentTurn.value) return false
    return room.value.you_are === currentTurn.value
  })

  onMounted(() => {
    if (auth.isLoggedIn.value) connect()
  })

  onUnmounted(() => {
    disconnect()
  })

  watch(() => auth.isLoggedIn.value, (loggedIn) => {
    if (loggedIn) connect()
    else disconnect()
  })

  return {
    connected,
    room,
    board,
    boardSize,
    status,
    currentTurn,
    chatMessages,
    publicRooms,
    winnerId,
    rpsPhase,
    myRpsSent,
    voiceEnabled,
    isMyTurn,
    connect,
    disconnect,
    fetchPublicRooms,
    createRoom,
    joinRoom,
    sendRps,
    sendMove,
    sendChat,
    startVoice,
    stopVoice,
    resetLocal,
  }
}
