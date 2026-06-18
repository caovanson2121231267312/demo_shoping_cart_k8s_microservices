<template>
  <v-container class="page-container py-6">
    <h1 class="text-h4 font-weight-bold mb-6">Tra cứu đơn hàng</h1>

    <v-card max-width="560" class="pa-6 mb-6">
      <v-form @submit.prevent="search">
        <v-text-field
          v-model="form.order_number"
          label="Mã đơn hàng"
          placeholder="ORD-0000000001"
          prepend-inner-icon="mdi-barcode"
          variant="outlined"
          required
        />
        <v-text-field
          v-model="form.shipping_phone"
          label="Số điện thoại nhận hàng"
          prepend-inner-icon="mdi-phone"
          variant="outlined"
          required
        />
        <v-btn type="submit" color="primary" :loading="loading" block>
          Tra cứu
        </v-btn>
      </v-form>
      <v-alert v-if="error" type="error" class="mt-4" variant="tonal">{{ error }}</v-alert>
    </v-card>

    <v-card v-if="order" class="pa-6">
      <v-card-title class="px-0">Đơn {{ order.order_number || order.id }}</v-card-title>
      <v-chip :color="statusColor(order.status)" class="mb-4">{{ statusLabel(order.status) }}</v-chip>
      <v-list density="compact">
        <v-list-item title="Người nhận" :subtitle="order.shipping_name" />
        <v-list-item title="SĐT" :subtitle="order.shipping_phone" />
        <v-list-item title="Địa chỉ" :subtitle="order.shipping_address" />
        <v-list-item title="Tổng tiền" :subtitle="formatVND(order.total_amount)" />
        <v-list-item title="Ngày đặt" :subtitle="formatDate(order.created_at)" />
      </v-list>
      <v-divider class="my-4" />
      <div v-for="item in order.items || []" :key="item.id" class="d-flex justify-space-between py-2">
        <span>{{ item.product_name_snapshot }} × {{ item.quantity }}</span>
        <span>{{ formatVND(item.unit_price * item.quantity) }}</span>
      </div>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import type { Order } from '~/types'

definePageMeta({ layout: 'default' })

const { trackOrder } = useOrders()
const { formatVND } = useFormat()

const form = reactive({ order_number: '', shipping_phone: '' })
const loading = ref(false)
const error = ref('')
const order = ref<Order | null>(null)

const statusLabel = (s: string) =>
  ({ pending: 'Chờ xử lý', confirmed: 'Đã xác nhận', shipping: 'Đang giao', delivered: 'Đã giao', cancelled: 'Đã hủy' }[s] ?? s)

const statusColor = (s: string) =>
  ({ pending: 'warning', confirmed: 'info', shipping: 'primary', delivered: 'success', cancelled: 'error' }[s] ?? 'default')

const formatDate = (d: string) => new Date(d).toLocaleString('vi-VN')

async function search() {
  loading.value = true
  error.value = ''
  order.value = null
  try {
    order.value = await trackOrder({ ...form })
  } catch {
    error.value = 'Không tìm thấy đơn hàng. Kiểm tra mã đơn và số điện thoại.'
  } finally {
    loading.value = false
  }
}
</script>
