<template>
  <v-container class="page-container py-6">
    <h1 class="text-h4 font-weight-bold mb-6">Đơn hàng của tôi</h1>

    <LoadingSpinner v-if="loading" />
    <EmptyState
      v-else-if="!orders.length"
      icon="mdi-package-variant-closed"
      title="Chưa có đơn hàng"
      description="Bạn chưa đặt đơn hàng nào."
    >
      <v-btn color="primary" to="/products" class="mt-4">Mua sắm ngay</v-btn>
    </EmptyState>

    <v-card v-else>
      <v-table>
        <thead>
          <tr>
            <th>Mã đơn</th>
            <th>Ngày đặt</th>
            <th>Trạng thái</th>
            <th class="text-right">Tổng tiền</th>
            <th />
          </tr>
        </thead>
        <tbody>
          <tr v-for="order in orders" :key="order.id">
            <td class="text-caption">{{ order.id.slice(0, 8) }}...</td>
            <td>{{ formatDate(order.created_at) }}</td>
            <td>
              <v-chip :color="statusColor(order.status)" size="small" variant="flat">
                {{ statusLabel(order.status) }}
              </v-chip>
            </td>
            <td class="text-right font-weight-bold">{{ formatVND(order.total_amount) }}</td>
            <td class="text-right">
              <v-btn size="small" variant="text" color="primary" :to="`/orders/${order.id}`">
                Chi tiết
              </v-btn>
            </td>
          </tr>
        </tbody>
      </v-table>
    </v-card>

    <div v-if="totalPages > 1" class="d-flex justify-center mt-6">
      <v-pagination v-model="page" :length="totalPages" @update:model-value="loadOrders" />
    </div>
  </v-container>
</template>

<script setup lang="ts">
import type { Order } from '~/types'

definePageMeta({ layout: 'default' })

const auth = useAuth()
const { formatVND } = useFormat()

const orders = ref<Order[]>([])
const loading = ref(true)
const page = ref(1)
const totalPages = ref(1)

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

const loadOrders = async () => {
  loading.value = true
  try {
    const result = await auth.apiFetch<{ items: Order[]; total_pages: number }>('/api/orders', {
      query: { page: page.value, limit: 10 },
    })
    orders.value = result.items
    totalPages.value = result.total_pages
  } finally {
    loading.value = false
  }
}

onMounted(loadOrders)
</script>
