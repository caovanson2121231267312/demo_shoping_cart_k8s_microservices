<template>
  <div>
    <AdminPageHeader title="Game Caro" subtitle="Giám sát phòng đang chơi realtime">
      <template #actions>
        <v-btn color="primary" variant="tonal" to="/game/caro" prepend-icon="mdi-gamepad-variant">
          Mở game
        </v-btn>
        <v-btn icon variant="text" :loading="loading" @click="load">
          <v-icon>mdi-refresh</v-icon>
        </v-btn>
      </template>
    </AdminPageHeader>

    <AdminDataTable
      :headers="headers"
      :items="rooms"
      :loading="loading"
      :count="rooms.length"
      title="Phòng đang hoạt động"
    >
      <template #item.code="{ item }">
        <span class="admin-table__mono">{{ item.code }}</span>
      </template>
      <template #item.status="{ item }">
        <v-chip size="small" variant="tonal" :color="statusColor(item.status)">
          {{ statusLabel(item.status) }}
        </v-chip>
      </template>
      <template #item.players="{ item }">
        <div class="text-body-2">
          {{ item.player_x_name || '—' }} vs {{ item.player_o_name || 'Chờ' }}
        </div>
      </template>
      <template #item.board_size="{ item }">
        {{ item.board_size }}×{{ item.board_size }}
      </template>
      <template #item.visibility="{ item }">
        <v-chip size="small" variant="tonal">{{ item.visibility === 'private' ? 'Riêng' : 'Công khai' }}</v-chip>
      </template>
      <template #item.updated_at="{ item }">
        {{ formatDate(item.updated_at) }}
      </template>
    </AdminDataTable>
  </div>
</template>

<script setup lang="ts">
import type { CaroPublicRoom } from '~/types'

definePageMeta({ layout: 'admin' })

const auth = useAuth()
const loading = ref(false)
const rooms = ref<CaroPublicRoom[]>([])
let timer: ReturnType<typeof setInterval> | null = null

const headers = [
  { title: 'Mã phòng', key: 'code' },
  { title: 'Trạng thái', key: 'status', sortable: false },
  { title: 'Người chơi', key: 'players', sortable: false },
  { title: 'Bàn', key: 'board_size', sortable: false },
  { title: 'Loại', key: 'visibility', sortable: false },
  { title: 'Cập nhật', key: 'updated_at' },
]

const statusLabel = (s: string) => {
  const map: Record<string, string> = {
    waiting: 'Chờ người',
    rps: 'Oẳn tù xì',
    playing: 'Đang chơi',
  }
  return map[s] || s
}

const statusColor = (s: string) => {
  const map: Record<string, string> = {
    waiting: 'warning',
    rps: 'info',
    playing: 'success',
  }
  return map[s] || 'grey'
}

const formatDate = (iso?: string) => (iso ? new Date(iso).toLocaleString('vi-VN') : '—')

async function load() {
  loading.value = true
  try {
    const res = await auth.apiFetch<{ items: CaroPublicRoom[] }>('/api/caro/admin/rooms')
    rooms.value = res.items
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  load()
  timer = setInterval(load, 15000)
})
onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>
