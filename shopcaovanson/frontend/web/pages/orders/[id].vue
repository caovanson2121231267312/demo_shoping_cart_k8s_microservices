<template>
  <div>
    <PageBanner
      :title="order ? `Đơn hàng #${order.order_number || orderId.slice(0, 8)}` : 'Chi tiết đơn hàng'"
      subtitle="Chi tiết và trạng thái giao hàng"
      :breadcrumbs="[{ label: 'Đơn hàng', to: '/orders' }, { label: 'Chi tiết' }]"
      compact
    />
    <v-container class="page-container py-6">
    <LoadingSpinner v-if="loading" />
    <EmptyState
      v-else-if="!order"
      icon="mdi-alert-circle-outline"
      title="Không tìm thấy đơn hàng"
    >
      <v-btn color="primary" to="/orders" class="mt-4">Quay lại</v-btn>
    </EmptyState>

    <template v-else-if="order">
      <v-row>
        <v-col cols="12" md="8">
          <v-card class="mb-4">
            <v-card-title>Sản phẩm</v-card-title>
            <v-card-text>
              <div
                v-for="item in order.items"
                :key="item.id"
                class="d-flex align-center ga-3 mb-3"
              >
                <v-img
                  :src="item.product_image_snapshot || productImageUrl(item.product_id)"
                  width="64"
                  height="64"
                  cover
                  class="rounded"
                />
                <div class="flex-grow-1">
                  <div class="font-weight-medium">{{ item.product_name_snapshot }}</div>
                  <div class="text-body-2 text-grey">
                    {{ formatVND(item.unit_price) }} × {{ item.quantity }}
                  </div>
                </div>
                <div class="font-weight-bold">{{ formatVND(item.unit_price * item.quantity) }}</div>
              </div>
            </v-card-text>
          </v-card>

          <v-card>
            <v-card-title>Địa chỉ giao hàng</v-card-title>
            <v-card-text>
              <p class="mb-1"><strong>{{ order.shipping_name }}</strong></p>
              <p class="mb-1">{{ order.shipping_phone }}</p>
              <p class="mb-0">{{ order.shipping_address }}</p>
            </v-card-text>
          </v-card>
        </v-col>

        <v-col cols="12" md="4">
          <v-card class="mb-4">
            <v-card-title>Trạng thái</v-card-title>
            <v-card-text>
              <v-timeline side="end" density="compact" truncate-line="both">
                <v-timeline-item
                  v-for="step in timeline"
                  :key="step.status"
                  :dot-color="step.active ? step.color : 'grey-lighten-1'"
                  size="small"
                >
                  <div :class="step.active ? 'font-weight-bold' : 'text-grey'">
                    {{ step.label }}
                  </div>
                </v-timeline-item>
              </v-timeline>
            </v-card-text>
          </v-card>

          <v-card>
            <v-card-text>
              <div class="d-flex justify-space-between text-body-2 mb-1" v-if="order.subtotal_amount">
                <span>Tạm tính</span>
                <span>{{ formatVND(order.subtotal_amount) }}</span>
              </div>
              <div
                v-if="order.discount_amount && order.discount_amount > 0"
                class="d-flex justify-space-between text-body-2 mb-1 text-success"
              >
                <span>Giảm giá {{ order.coupon_code ? `(${order.coupon_code})` : '' }}</span>
                <span>-{{ formatVND(order.discount_amount) }}</span>
              </div>
              <div class="d-flex justify-space-between text-h6 mb-4">
                <span>Tổng cộng</span>
                <span class="text-primary font-weight-bold">{{ formatVND(order.total_amount) }}</span>
              </div>
              <v-btn
                block
                color="primary"
                variant="outlined"
                prepend-icon="mdi-file-pdf-box"
                class="mb-2"
                :loading="downloadingInvoice"
                @click="downloadInvoice"
              >
                Tải hóa đơn PDF
              </v-btn>
              <v-btn
                v-if="order.status === 'pending'"
                block
                color="error"
                variant="outlined"
                :loading="cancelling"
                @click="showCancelDialog = true"
              >
                Hủy đơn hàng
              </v-btn>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </template>

    <ConfirmDialog
      v-model="showCancelDialog"
      title="Hủy đơn hàng"
      message="Bạn có chắc muốn hủy đơn hàng này?"
      confirm-text="Hủy đơn"
      :loading="cancelling"
      @confirm="cancelOrder"
    />
    </v-container>
  </div>
</template>

<script setup lang="ts">
import type { Order } from '~/types'

definePageMeta({ layout: 'default' })

const route = useRoute()
const auth = useAuth()
const { formatVND } = useFormat()

const orderId = route.params.id as string

const downloadInvoice = async () => {
  downloadingInvoice.value = true
  try {
    const blob = await auth.apiFetch<Blob>(`/api/invoices/${orderId}.pdf`, {
      responseType: 'blob',
    })
    const url = URL.createObjectURL(blob)
    window.open(url, '_blank')
    setTimeout(() => URL.revokeObjectURL(url), 60000)
  } catch {
    useSnackbar().show('Hóa đơn chưa sẵn sàng. Vui lòng thử lại sau vài giây.', 'info')
  } finally {
    downloadingInvoice.value = false
  }
}
const order = ref<Order | null>(null)
const loading = ref(true)
const cancelling = ref(false)
const downloadingInvoice = ref(false)
const showCancelDialog = ref(false)

const statusSteps = [
  { status: 'pending', label: 'Chờ xử lý', color: 'warning' },
  { status: 'confirmed', label: 'Đã xác nhận', color: 'info' },
  { status: 'shipping', label: 'Đang giao hàng', color: 'primary' },
  { status: 'delivered', label: 'Đã giao hàng', color: 'success' },
]

const timeline = computed(() => {
  if (!order.value) {
    return []
  }
  if (order.value.status === 'cancelled') {
    return [{ status: 'cancelled', label: 'Đã hủy', color: 'error', active: true }]
  }
  const currentIdx = statusSteps.findIndex((s) => s.status === order.value!.status)
  return statusSteps.map((step, idx) => ({
    ...step,
    active: idx <= currentIdx,
  }))
})

const loadOrder = async () => {
  loading.value = true
  try {
    order.value = await auth.apiFetch<Order>(`/api/orders/${orderId}`)
  } catch {
    order.value = null
  } finally {
    loading.value = false
  }
}

const cancelOrder = async () => {
  cancelling.value = true
  try {
    await auth.apiFetch(`/api/orders/${orderId}/cancel`, { method: 'PUT' })
    showCancelDialog.value = false
    await loadOrder()
  } finally {
    cancelling.value = false
  }
}

onMounted(loadOrder)
</script>
