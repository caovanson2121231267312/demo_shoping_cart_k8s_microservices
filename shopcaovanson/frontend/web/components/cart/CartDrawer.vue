<template>
  <v-navigation-drawer
    :model-value="open"
    location="right"
    temporary
    width="400"
    @update:model-value="setOpen"
  >
    <div class="d-flex align-center pa-4 bg-primary">
      <span class="text-h6 text-white">Giỏ hàng</span>
      <v-spacer />
      <v-btn icon variant="text" color="white" @click="setOpen(false)">
        <v-icon>mdi-close</v-icon>
      </v-btn>
    </div>

    <div class="pa-4">
      <LoadingSpinner v-if="cart.loading.value" />
      <EmptyState
        v-else-if="!cart.items.value.length"
        icon="mdi-cart-outline"
        title="Giỏ hàng trống"
        description="Thêm sản phẩm để bắt đầu mua sắm."
      />
      <div v-else>
        <CartItem
          v-for="item in cart.items.value"
          :key="item.product_id"
          :item="item"
        />
      </div>
    </div>

    <template v-if="cart.items.value.length">
      <v-divider />
      <div class="pa-4">
        <div class="d-flex justify-space-between text-h6 mb-4">
          <span>Tổng cộng</span>
          <span class="text-primary font-weight-bold">{{ cart.formattedTotal.value }}</span>
        </div>
        <v-btn block color="primary" to="/cart" @click="setOpen(false)">Xem giỏ hàng</v-btn>
        <v-btn block variant="outlined" color="primary" class="mt-2" to="/checkout" @click="setOpen(false)">
          Thanh toán
        </v-btn>
      </div>
    </template>
  </v-navigation-drawer>
</template>

<script setup lang="ts">
const cart = useCart()
const open = useState('cartDrawerOpen', () => false)

const setOpen = (value: boolean) => {
  open.value = value
}

watch(open, async (isOpen) => {
  if (isOpen) {
    await cart.fetchCart()
  }
})
</script>
