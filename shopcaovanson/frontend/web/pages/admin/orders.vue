<template>
  <v-container class="page-container py-6">
    <h1 class="text-h4 font-weight-bold mb-6">Quản lý đơn hàng</h1>

    <v-card class="mb-4 pa-4">
      <v-row dense>
        <v-col cols="12" md="3">
          <v-text-field
            v-model="searchQuery"
            label="Mã đơn / SĐT / tên"
            density="compact"
            variant="outlined"
            hide-details
            clearable
            @keyup.enter="loadOrders"
          />
        </v-col>
        <v-col cols="12" md="3">
          <v-select
            v-model="statusFilter"
            :items="statusOptions"
            label="Trạng thái"
            clearable
            density="compact"
            variant="outlined"
            hide-details
          />
        </v-col>
        <v-col cols="12" md="2">
          <v-btn color="primary" @click="loadOrders">Tìm</v-btn>
        </v-col>
      </v-row>
    </v-card>

    <LoadingSpinner v-if="loading" />

    <v-card v-else>
      <v-data-table
        :headers="headers"
        :items="orders"
        :items-per-page="10"
        class="elevation-0"
      >
        <template #item.order_number="{ item }">
          <span class="text-caption font-weight-medium">{{ item.order_number || item.id.slice(0, 8) }}</span>
        </template>
        <template #item.id="{ item }">
          <span class="text-caption">{{ item.id.slice(0, 8) }}...</span>
        </template>
        <template #item.created_at="{ item }">
          {{ formatDate(item.created_at) }}
        </template>
        <template #item.status="{ item }">
          <v-chip :color="statusColor(item.status)" size="small" variant="flat">
            {{ statusLabel(item.status) }}
          </v-chip>
        </template>
        <template #item.total_amount="{ item }">
          {{ formatVND(item.total_amount) }}
        </template>
        <template #item.actions="{ item }">
          <v-btn size="small" variant="text" :to="`/admin/orders/${item.id}`">Chi tiết</v-btn>
          <v-menu v-if="admin.canManageOrders.value">
            <template #activator="{ props }">
              <v-btn v-bind="props" size="small" variant="outlined" color="primary">
                Cập nhật
              </v-btn>
            </template>
            <v-list density="compact">
              <v-list-item
                v-for="status in availableStatuses(item.status)"
                :key="status"
                :title="statusLabel(status)"
                @click="updateStatus(item.id, status)"
              />
            </v-list>
          </v-menu>
        </template>
      </v-data-table>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import type { Order } from '~/types'

definePageMeta({ layout: 'admin' })

const { searchAdminOrders, updateOrderStatus } = useOrders()
const admin = useAdmin()

const orders = ref<Order[]>([])
const loading = ref(true)
const statusFilter = ref<string | null>(null)
const searchQuery = ref('')
const page = ref(1)

const statusOptions = [
  { title: 'Chờ xử lý', value: 'pending' },
  { title: 'Đã xác nhận', value: 'confirmed' },
  { title: 'Đang giao', value: 'shipping' },
  { title: 'Đã giao', value: 'delivered' },
  { title: 'Đã hủy', value: 'cancelled' },
]

const headers = [
  { title: 'Mã đơn', key: 'order_number' },
  { title: 'Khách hàng', key: 'shipping_name' },
  { title: 'Ngày đặt', key: 'created_at' },
  { title: 'Trạng thái', key: 'status' },
  { title: 'Tổng tiền', key: 'total_amount' },
  { title: '', key: 'actions', sortable: false },
]

const statusLabel = (status: string) => {
  const map: Record<string, string> = {
    pending: 'Chờ xử lý',
    confirmed: 'Đã xác nhận',
    shipping: 'Đang giao',
    delivered: 'Đã giao',
    cancelled: 'Đã hủy',
  }
  return map[status] || status
}

const statusColor = (status: string) => {
  const map: Record<string, string> = {
    pending: 'warning',
    confirmed: 'info',
    shipping: 'primary',
    delivered: 'success',
    cancelled: 'error',
  }
  return map[status] || 'grey'
}

const formatDate = (date: string) => new Date(date).toLocaleDateString('vi-VN')

const availableStatuses = (current: string) => {
  const flow: Record<string, string[]> = {
    pending: ['confirmed', 'cancelled'],
    confirmed: ['shipping', 'cancelled'],
    shipping: ['delivered'],
    delivered: [],
    cancelled: [],
  }
  return flow[current] || []
}

const loadOrders = async () => {
  loading.value = true
  try {
    const query: Record<string, string | number> = { limit: 50, page: page.value }
    if (statusFilter.value) query.status = statusFilter.value
    if (searchQuery.value) query.search = searchQuery.value
    const result = await searchAdminOrders(query)
    orders.value = result.items
  } finally {
    loading.value = false
  }
}

const updateStatus = async (orderId: string, status: string) => {
  await updateOrderStatus(orderId, status)
  await loadOrders()
}

onMounted(loadOrders)
</script>
