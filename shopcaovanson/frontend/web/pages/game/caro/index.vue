<template>
  <v-container class="page-container py-6 caro-page">
    <div class="d-flex align-center justify-space-between mb-6 flex-wrap ga-3">
      <div>
        <h1 class="text-h4 font-weight-bold">Cờ Caro Online</h1>
        <p class="text-body-2 text-medium-emphasis mb-0">Tạo phòng, mời bạn hoặc vào phòng công khai — oẳn tù xì chọn X/O</p>
      </div>
      <v-chip :color="caro.connected.value ? 'success' : 'warning'" variant="tonal">
        {{ caro.connected.value ? 'Đã kết nối' : 'Đang kết nối...' }}
      </v-chip>
    </div>

    <!-- Lobby -->
    <v-row v-if="!caro.room.value">
      <v-col cols="12" md="5">
        <v-card class="pa-4 mb-4">
          <v-card-title class="px-0">Tạo phòng mới</v-card-title>
          <v-select
            v-model="boardSize"
            :items="boardOptions"
            label="Kích thước bàn"
            variant="outlined"
            class="mb-3"
          />
          <v-btn-toggle v-model="visibility" mandatory color="primary" class="mb-4">
            <v-btn value="public">Công khai</v-btn>
            <v-btn value="private">Riêng tư</v-btn>
          </v-btn-toggle>
          <v-btn color="primary" block :loading="creating" @click="onCreate">Tạo phòng</v-btn>
        </v-card>

        <v-card class="pa-4">
          <v-card-title class="px-0">Vào phòng</v-card-title>
          <v-text-field
            v-model="joinCode"
            label="Mã phòng"
            variant="outlined"
            class="mb-3"
            @keyup.enter="onJoin"
          />
          <v-btn color="secondary" block :loading="joining" @click="onJoin">Tham gia</v-btn>
        </v-card>
      </v-col>

      <v-col cols="12" md="7">
        <v-card>
          <v-card-title class="d-flex align-center justify-space-between">
            Phòng công khai
            <v-btn icon variant="text" :loading="loadingRooms" @click="loadPublic">
              <v-icon>mdi-refresh</v-icon>
            </v-btn>
          </v-card-title>
          <v-list>
            <v-list-item
              v-for="r in caro.publicRooms.value"
              :key="r.id"
              :title="`Phòng ${r.code}`"
              :subtitle="`${r.host_name} · ${r.board_size}×${r.board_size} · ${statusLabel(r.status)}`"
            >
              <template #append>
                <v-btn size="small" color="primary" variant="tonal" @click="joinPublic(r.code)">Vào</v-btn>
              </template>
            </v-list-item>
            <v-list-item v-if="!caro.publicRooms.value.length" title="Chưa có phòng công khai" />
          </v-list>
        </v-card>
      </v-col>
    </v-row>

    <!-- In room -->
    <v-row v-else>
      <v-col cols="12" lg="8">
        <v-card class="pa-4">
          <div class="d-flex align-center justify-space-between mb-4 flex-wrap ga-2">
            <div>
              <div class="text-h6 font-weight-bold">Phòng {{ caro.room.value.code }}</div>
              <div class="text-caption text-medium-emphasis">
                {{ caro.room.value.player_x_name || '—' }} (X) vs {{ caro.room.value.player_o_name || 'Chờ người chơi' }} (O)
              </div>
            </div>
            <v-chip variant="tonal" color="primary">{{ statusLabel(caro.status.value) }}</v-chip>
          </div>

          <!-- RPS -->
          <div v-if="caro.rpsPhase.value" class="text-center py-6">
            <p class="text-body-1 mb-4">Oẳn tù xì để chọn ai đi X (trước)</p>
            <div class="d-flex justify-center ga-2 flex-wrap">
              <v-btn
                v-for="c in rpsChoices"
                :key="c.value"
                size="large"
                variant="tonal"
                :disabled="caro.myRpsSent.value"
                @click="caro.sendRps(c.value)"
              >
                <v-icon start>{{ c.icon }}</v-icon>
                {{ c.label }}
              </v-btn>
            </div>
            <p v-if="caro.myRpsSent.value" class="text-caption mt-3">Đã chọn — chờ đối thủ...</p>
          </div>

          <!-- Board -->
          <div v-else class="caro-board-wrap">
            <div
              class="caro-board"
              :style="{ gridTemplateColumns: `repeat(${caro.boardSize.value}, 1fr)` }"
            >
              <button
                v-for="(cell, idx) in flatBoard"
                :key="idx"
                type="button"
                class="caro-cell"
                :class="{
                  'caro-cell--x': cell === 'X',
                  'caro-cell--o': cell === 'O',
                  'caro-cell--disabled': !canPlay,
                }"
                :disabled="!canPlay || cell !== null"
                @click="onCellClick(idx)"
              >
                {{ cell || '' }}
              </button>
            </div>
          </div>

          <v-alert v-if="caro.status.value === 'finished'" type="success" variant="tonal" class="mt-4">
            {{ gameResultText }}
          </v-alert>

          <div class="d-flex ga-2 mt-4 flex-wrap">
            <v-btn
              :color="caro.voiceEnabled.value ? 'error' : 'primary'"
              variant="tonal"
              prepend-icon="mdi-microphone"
              @click="toggleVoice"
            >
              {{ caro.voiceEnabled.value ? 'Tắt mic' : 'Bật mic' }}
            </v-btn>
            <v-btn variant="outlined" @click="leaveRoom">Rời phòng</v-btn>
          </div>
        </v-card>
      </v-col>

      <v-col cols="12" lg="4">
        <v-card class="d-flex flex-column" style="height: 420px">
          <v-card-title>Chat phòng</v-card-title>
          <v-card-text class="flex-grow-1 overflow-y-auto py-0">
            <div v-for="(m, i) in caro.chatMessages.value" :key="m.id || i" class="mb-2">
              <span class="font-weight-medium text-primary">{{ m.user_name }}:</span>
              {{ m.content }}
            </div>
          </v-card-text>
          <v-card-actions class="pa-3">
            <v-text-field
              v-model="chatInput"
              density="compact"
              variant="outlined"
              hide-details
              placeholder="Nhắn tin..."
              @keyup.enter="sendChat"
            />
            <v-btn icon color="primary" @click="sendChat"><v-icon>mdi-send</v-icon></v-btn>
          </v-card-actions>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const caro = useCaro()
const snackbar = useSnackbar()
const auth = useAuth()

const boardSize = ref<15 | 20>(15)
const visibility = ref<'public' | 'private'>('public')
const joinCode = ref('')
const creating = ref(false)
const joining = ref(false)
const loadingRooms = ref(false)
const chatInput = ref('')

const boardOptions = [
  { title: '15 × 15', value: 15 },
  { title: '20 × 20', value: 20 },
]

const rpsChoices = [
  { value: 'rock' as const, label: 'Búa', icon: 'mdi-gesture' },
  { value: 'scissors' as const, label: 'Kéo', icon: 'mdi-content-cut' },
  { value: 'paper' as const, label: 'Bao', icon: 'mdi-hand-back-left' },
]

const flatBoard = computed(() => caro.board.value.flat())

const canPlay = computed(() =>
  caro.status.value === 'playing' &&
  caro.isMyTurn.value &&
  !caro.winnerId.value,
)

const gameResultText = computed(() => {
  if (!caro.winnerId.value) return 'Hòa — cả hai giỏi quá!'
  if (caro.winnerId.value === auth.user.value?.id) return 'Bạn thắng!'
  return 'Bạn thua — chơi ván khác nhé!'
})

function statusLabel(s: string) {
  const map: Record<string, string> = {
    waiting: 'Chờ người',
    rps: 'Oẳn tù xì',
    playing: 'Đang chơi',
    finished: 'Kết thúc',
    lobby: 'Sảnh',
  }
  return map[s] || s
}

async function loadPublic() {
  loadingRooms.value = true
  try {
    await caro.fetchPublicRooms()
  } finally {
    loadingRooms.value = false
  }
}

async function onCreate() {
  creating.value = true
  try {
    await caro.createRoom(visibility.value, boardSize.value)
    snackbar.show(`Đã tạo phòng ${caro.room.value?.code}`, 'success')
  } catch (e: unknown) {
    snackbar.show(e instanceof Error ? e.message : 'Tạo phòng thất bại', 'error')
  } finally {
    creating.value = false
  }
}

async function onJoin() {
  if (!joinCode.value.trim()) return
  joining.value = true
  try {
    await caro.joinRoom(joinCode.value.trim().toUpperCase())
  } catch (e: unknown) {
    snackbar.show(e instanceof Error ? e.message : 'Vào phòng thất bại', 'error')
  } finally {
    joining.value = false
  }
}

async function joinPublic(code: string) {
  joining.value = true
  try {
    await caro.joinRoom(code)
  } catch (e: unknown) {
    snackbar.show(e instanceof Error ? e.message : 'Vào phòng thất bại', 'error')
  } finally {
    joining.value = false
  }
}

function onCellClick(idx: number) {
  const size = caro.boardSize.value
  const row = Math.floor(idx / size)
  const col = idx % size
  caro.sendMove(row, col)
}

function sendChat() {
  caro.sendChat(chatInput.value)
  chatInput.value = ''
}

async function toggleVoice() {
  try {
    if (caro.voiceEnabled.value) caro.stopVoice()
    else await caro.startVoice()
  } catch {
    snackbar.show('Không thể bật mic — kiểm tra quyền trình duyệt', 'error')
  }
}

function leaveRoom() {
  caro.stopVoice()
  caro.resetLocal()
  loadPublic()
}

onMounted(loadPublic)
</script>

<style scoped>
.caro-board-wrap {
  overflow: auto;
  max-width: 100%;
}

.caro-board {
  display: grid;
  gap: 2px;
  background: #e2e8f0;
  padding: 4px;
  border-radius: 8px;
  margin: 0 auto;
  max-width: min(100%, 560px);
}

.caro-cell {
  aspect-ratio: 1;
  min-width: 0;
  border: none;
  background: #fff;
  font-weight: 700;
  font-size: clamp(0.65rem, 2.5vw, 1rem);
  cursor: pointer;
  border-radius: 4px;
  transition: background 0.12s;
}

.caro-cell:hover:not(:disabled) {
  background: #f1f5f9;
}

.caro-cell--disabled {
  cursor: default;
}

.caro-cell--x {
  color: #1565c0;
}

.caro-cell--o {
  color: #c62828;
}
</style>
