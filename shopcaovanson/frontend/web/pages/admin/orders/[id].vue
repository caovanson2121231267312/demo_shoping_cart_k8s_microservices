<template>
  <v-container class="page-container py-6">
    <div class="d-flex align-center mb-6">
      <v-btn icon variant="text" to="/admin/orders" class="mr-2"><v-icon>mdi-arrow-left</v-icon></v-btn>
      <h1 class="text-h5 font-weight-bold">Chi tiết đơn #{{ order?.order_number || orderId }}</h1>
    </div>

    <LoadingSpinner v-if="loading" />
    <EmptyState v-else-if="!order" title="Không tìm thấy đơn hàng" />

    <v-row v-else>
      <v-col cols="12" md="8">
        <v-card class="mb-4">
          <v-card-title>Sản phẩm</v-card-title>
          <v-card-text>
            <div v-for="item in order.items" :key="item.id" class="d-flex justify-space-between mb-2">
              <span>{{ item.product_name_snapshot }} × {{ item.quantity }}</span>
              <span>{{ formatVND(item.unit_price * item.quantity) }}</span>
            </div>
          </v-card-text>
        </v-card>
        <v-card>
          <v-card-title>Giao hàng</v-card-title>
          <v-card-text>
            <p class="mb-1"><strong>{{ order.shipping_name }}</strong> — {{ order.shipping_phone }}</p>
            <p class="mb-0">{{ order.shipping_address }}</p>
          </v-card-text>
        </v-card>
      </v-col>
      <v-col cols="12" md="4">
        <v-card>
          <v-card-text>
            <v-chip :color="statusColor(order.status)" class="mb-4">{{ statusLabel(order.status) }}</v-chip>
            <div v-if="order.subtotal_amount" class="d-flex justify-space-between mb-1">
              <span>Tạm tính</span><span>{{ formatVND(order.subtotal_amount) }}</span>
            </div>
            <div v-if="order.discount_amount" class="d-flex justify-space-between mb-1 text-success">
              <span>Giảm giá</span><span>-{{ formatVND(order.discount_amount) }}</span>
            </div>
            <div class="d-flex justify-space-between text-h6 mb-4">
              <span>Tổng</span><span class="text-primary font-weight-bold">{{ formatVND(order.total_amount) }}</span>
            </div>
            <v-select
              v-if="admin.canManageOrders.value"
              v-model="newStatus"
              :items="statusOptions"
              label="Cập nhật trạng thái"
              variant="outlined"
              density="compact"
              class="mb-2"
            />
            <v-btn
              v-if="admin.canManageOrders.value"
              color="primary"
              block
              :loading="saving"
              @click="saveStatus"
            >
              Lưu trạng thái
            </v-btn>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup lang="ts">
import type { Order } from '~/types'

definePageMeta({ layout: 'admin' })

const route = useRoute()
const admin = useAdmin()
const snackbar = useSnackbar()
const { fetchAdminOrder, updateOrderStatus } = useOrders()
const { formatVND } = useFormat()

const orderId = route.params.id as string
const order = ref<Order | null>(null)
const loading = ref(true)
const saving = ref(false)
const newStatus = ref('')

const statusOptions = ['pending', 'confirmed', 'shipping', 'delivered', 'cancelled']

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

const load = async () => {
  loading.value = true
  try {
    order.value = await fetchAdminOrder(orderId)
    newStatus.value = order.value?.status || 'pending'
  } catch (e: unknown) {
    order.value = null
    snackbar.show(e instanceof Error ? e.message : 'Không thể tải chi tiết đơn hàng', 'error')
  } finally {
    loading.value = false
  }
}

const saveStatus = async () => {
  if (!order.value || !newStatus.value) return
  saving.value = true
  try {
    order.value = await updateOrderStatus(orderId, newStatus.value)
    snackbar.show('Đã cập nhật trạng thái đơn hàng', 'success')
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>
