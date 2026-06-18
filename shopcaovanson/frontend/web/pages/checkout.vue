<template>
  <v-container class="page-container py-6">
    <h1 class="text-h4 font-weight-bold mb-6">Thanh toán</h1>

    <EmptyState
      v-if="!cart.items.value.length && !cart.loading.value"
      icon="mdi-cart-outline"
      title="Giỏ hàng trống"
      description="Thêm sản phẩm trước khi thanh toán."
    >
      <v-btn color="primary" to="/products" class="mt-4">Mua sắm ngay</v-btn>
    </EmptyState>

    <v-row v-else>
      <v-col cols="12" md="7">
        <v-card>
          <v-card-title>Thông tin giao hàng</v-card-title>
          <v-card-text>
            <v-form ref="formRef" @submit.prevent="placeOrder">
              <v-text-field
                v-model="form.shipping_name"
                label="Họ và tên"
                :rules="[rules.required]"
                variant="outlined"
                class="mb-2"
              />
              <v-text-field
                v-model="form.shipping_phone"
                label="Số điện thoại"
                :rules="[rules.required, rules.phone]"
                variant="outlined"
                class="mb-2"
              />
              <v-textarea
                v-model="form.shipping_address"
                label="Địa chỉ giao hàng"
                :rules="[rules.required]"
                rows="3"
                variant="outlined"
              />
              <v-btn type="submit" color="primary" size="large" :loading="submitting" block class="mt-4">
                Đặt hàng (COD)
              </v-btn>
            </v-form>
          </v-card-text>
        </v-card>
      </v-col>

      <v-col cols="12" md="5">
        <v-card>
          <v-card-title>Đơn hàng</v-card-title>
          <v-card-text>
            <div
              v-for="item in cart.items.value"
              :key="item.product_id"
              class="d-flex justify-space-between mb-2 text-body-2"
            >
              <span>{{ item.product_name }} × {{ item.quantity }}</span>
              <span>{{ formatVND(item.unit_price * item.quantity) }}</span>
            </div>
            <v-divider class="my-3" />
            <div class="d-flex ga-2 mb-2">
              <v-text-field
                v-model="couponCode"
                label="Mã giảm giá"
                variant="outlined"
                density="compact"
                hide-details
                placeholder="WELCOME10"
              />
              <v-btn color="secondary" :loading="validatingCoupon" @click="applyCoupon">Áp dụng</v-btn>
            </div>
            <v-alert
              v-if="couponMessage"
              :type="couponValid ? 'success' : 'warning'"
              density="compact"
              variant="tonal"
              class="mb-2"
            >
              {{ couponMessage }}
            </v-alert>
            <div class="d-flex justify-space-between text-body-2 mb-1">
              <span>Tạm tính</span>
              <span>{{ formatVND(subtotal) }}</span>
            </div>
            <div v-if="discount > 0" class="d-flex justify-space-between text-body-2 mb-1 text-success">
              <span>Giảm giá</span>
              <span>-{{ formatVND(discount) }}</span>
            </div>
            <v-divider class="my-2" />
            <div class="d-flex justify-space-between text-h6">
              <span>Tổng cộng</span>
              <span class="text-primary font-weight-bold">{{ formatVND(total) }}</span>
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-snackbar v-model="snackbar" color="success">
      Đặt hàng thành công! Mã đơn: {{ orderNumber || orderId }}
      <template #actions>
        <v-btn variant="text" :to="`/orders/${orderId}`">Xem đơn</v-btn>
      </template>
    </v-snackbar>
  </v-container>
</template>

<script setup lang="ts">
import type { CheckoutInput, Order } from '~/types'

definePageMeta({ layout: 'default' })

const cart = useCart()
const auth = useAuth()
const { validateCoupon } = useCoupons()
const { formatVND } = useFormat()

const formRef = ref()
const submitting = ref(false)
const snackbar = ref(false)
const orderId = ref('')
const orderNumber = ref('')
const couponCode = ref('')
const discount = ref(0)
const couponMessage = ref('')
const couponValid = ref(false)
const validatingCoupon = ref(false)

const form = reactive<CheckoutInput>({
  shipping_name: '',
  shipping_phone: '',
  shipping_address: '',
  coupon_code: '',
})

const rules = {
  required: (v: string) => !!v || 'Trường này là bắt buộc',
  phone: (v: string) => /^[0-9]{9,11}$/.test(v) || 'Số điện thoại không hợp lệ',
}

const subtotal = computed(() =>
  cart.items.value.reduce((sum, item) => sum + item.unit_price * item.quantity, 0),
)
const total = computed(() => Math.max(0, subtotal.value - discount.value))

onMounted(async () => {
  await cart.fetchCart()
  if (auth.user.value) {
    form.shipping_name = auth.user.value.full_name
  }
})

const applyCoupon = async () => {
  if (!couponCode.value.trim()) {
    return
  }
  validatingCoupon.value = true
  try {
    const result = await validateCoupon(couponCode.value.trim(), subtotal.value)
    couponValid.value = result.valid
    if (result.valid) {
      discount.value = result.discount_amount
      form.coupon_code = result.code || couponCode.value.trim()
      couponMessage.value = `Đã áp dụng mã ${form.coupon_code}`
    } else {
      discount.value = 0
      form.coupon_code = ''
      couponMessage.value = result.message || 'Mã không hợp lệ'
    }
  } catch {
    couponMessage.value = 'Không thể kiểm tra mã giảm giá'
    couponValid.value = false
  } finally {
    validatingCoupon.value = false
  }
}

const placeOrder = async () => {
  const { valid } = await formRef.value.validate()
  if (!valid) {
    return
  }
  submitting.value = true
  try {
    const order = await auth.apiFetch<Order>('/api/orders', {
      method: 'POST',
      body: form,
    })
    orderId.value = order.id
    orderNumber.value = order.order_number || ''
    snackbar.value = true
    await cart.fetchCart()
  } finally {
    submitting.value = false
  }
}
</script>
