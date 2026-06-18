<template>
  <v-container class="page-container py-6">
    <h1 class="text-h4 font-weight-bold mb-6">Giỏ hàng</h1>

    <LoadingSpinner v-if="cart.loading.value && !cart.items.value.length" />
    <EmptyState
      v-else-if="!cart.items.value.length"
      icon="mdi-cart-outline"
      title="Giỏ hàng trống"
      description="Bạn chưa thêm sản phẩm nào vào giỏ."
    >
      <v-btn color="primary" to="/products" class="mt-4">Tiếp tục mua sắm</v-btn>
    </EmptyState>

    <v-row v-else>
      <v-col cols="12" md="8">
        <CartItem
          v-for="item in cart.items.value"
          :key="item.product_id"
          :item="item"
        />
        <v-btn variant="text" color="error" @click="showClearDialog = true">Xóa toàn bộ giỏ</v-btn>
      </v-col>

      <v-col cols="12" md="4">
        <v-card>
          <v-card-title>Tóm tắt đơn hàng</v-card-title>
          <v-card-text>
            <div class="d-flex justify-space-between mb-2">
              <span>Số lượng</span>
              <span>{{ cart.count.value }} sản phẩm</span>
            </div>
            <div class="d-flex justify-space-between text-h6">
              <span>Tổng cộng</span>
              <span class="text-primary font-weight-bold">{{ cart.formattedTotal.value }}</span>
            </div>
          </v-card-text>
          <v-card-actions>
            <v-btn block color="primary" size="large" to="/checkout">Thanh toán</v-btn>
          </v-card-actions>
        </v-card>
      </v-col>
    </v-row>

    <ConfirmDialog
      v-model="showClearDialog"
      title="Xóa giỏ hàng"
      message="Bạn có chắc muốn xóa tất cả sản phẩm trong giỏ?"
      confirm-text="Xóa"
      :loading="clearing"
      @confirm="handleClear"
    />
  </v-container>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const cart = useCart()
const showClearDialog = ref(false)
const clearing = ref(false)

onMounted(() => cart.fetchCart())

const handleClear = async () => {
  clearing.value = true
  try {
    await cart.clearCart()
    showClearDialog.value = false
  } finally {
    clearing.value = false
  }
}
</script>
