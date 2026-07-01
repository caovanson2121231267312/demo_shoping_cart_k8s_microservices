<template>
  <v-navigation-drawer
    :model-value="open"
    class="cart-drawer"
    location="right"
    temporary
    width="420"
    @update:model-value="setOpen"
  >
    <div class="cart-drawer__shell">
      <header class="cart-drawer__header">
        <div class="cart-drawer__header-main">
          <div class="cart-drawer__header-icon">
            <v-icon color="white" size="22">mdi-cart-outline</v-icon>
          </div>
          <div>
            <h2 class="cart-drawer__title">Giỏ hàng</h2>
            <p v-if="cart.items.value.length" class="cart-drawer__subtitle">
              {{ cart.count.value }} sản phẩm
            </p>
          </div>
        </div>
        <button type="button" class="cart-drawer__close" aria-label="Đóng giỏ hàng" @click="setOpen(false)">
          <v-icon size="20">mdi-close</v-icon>
        </button>
      </header>

      <div class="cart-drawer__body">
        <div v-if="cart.loading.value" class="cart-drawer__loading">
          <LoadingSpinner />
        </div>

        <div v-else-if="!cart.items.value.length" class="cart-drawer__empty">
          <div class="cart-drawer__empty-icon">
            <v-icon size="40" color="primary">mdi-cart-off-outline</v-icon>
          </div>
          <h3 class="cart-drawer__empty-title">Giỏ hàng trống</h3>
          <p class="cart-drawer__empty-desc">Thêm sản phẩm yêu thích để bắt đầu mua sắm.</p>
          <v-btn
            color="primary"
            variant="flat"
            rounded="lg"
            class="text-none font-weight-bold mt-2"
            to="/products"
            @click="setOpen(false)"
          >
            Khám phá sản phẩm
          </v-btn>
        </div>

        <div v-else class="cart-drawer__items">
          <CartItem
            v-for="item in cart.items.value"
            :key="item.product_id"
            :item="item"
            variant="drawer"
          />
        </div>
      </div>

      <footer v-if="cart.items.value.length" class="cart-drawer__footer">
        <div class="cart-drawer__summary">
          <div class="cart-drawer__summary-row">
            <span>Tạm tính</span>
            <span>{{ cart.formattedTotal.value }}</span>
          </div>
          <div class="cart-drawer__summary-row cart-drawer__summary-row--total">
            <span>Tổng cộng</span>
            <strong>{{ cart.formattedTotal.value }}</strong>
          </div>
        </div>

        <v-btn
          block
          color="primary"
          size="large"
          rounded="lg"
          elevation="0"
          class="cart-drawer__btn text-none font-weight-bold"
          to="/checkout"
          prepend-icon="mdi-lock-check-outline"
          @click="setOpen(false)"
        >
          Thanh toán
        </v-btn>
        <v-btn
          block
          variant="outlined"
          color="primary"
          size="large"
          rounded="lg"
          class="cart-drawer__btn cart-drawer__btn--secondary text-none font-weight-medium"
          to="/cart"
          @click="setOpen(false)"
        >
          Xem giỏ hàng
        </v-btn>
        <button type="button" class="cart-drawer__continue" @click="setOpen(false)">
          Tiếp tục mua sắm
        </button>
      </footer>
    </div>
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

<style scoped>
.cart-drawer :deep(.v-navigation-drawer__content) {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.cart-drawer__shell {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--color-surface);
}

.cart-drawer__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 20px 20px 18px;
  background: var(--gradient-primary);
  flex-shrink: 0;
}

.cart-drawer__header-main {
  display: flex;
  align-items: center;
  gap: 14px;
  min-width: 0;
}

.cart-drawer__header-icon {
  width: 44px;
  height: 44px;
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.16);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.cart-drawer__title {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--color-text-inverse);
  line-height: 1.25;
  letter-spacing: -0.02em;
}

.cart-drawer__subtitle {
  margin: 2px 0 0;
  font-size: 0.8125rem;
  color: rgba(255, 255, 255, 0.82);
  font-weight: 500;
}

.cart-drawer__close {
  width: 36px;
  height: 36px;
  border: none;
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.14);
  color: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
  flex-shrink: 0;
}

.cart-drawer__close:hover {
  background: rgba(255, 255, 255, 0.24);
}

.cart-drawer__body {
  flex: 1 1 auto;
  overflow-y: auto;
  min-height: 0;
}

.cart-drawer__loading {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 240px;
}

.cart-drawer__empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 48px 28px;
}

.cart-drawer__empty-icon {
  width: 88px;
  height: 88px;
  border-radius: 50%;
  background: var(--color-primary-50);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 20px;
}

.cart-drawer__empty-title {
  margin: 0;
  font-size: 1.0625rem;
  font-weight: 700;
  color: var(--color-text);
}

.cart-drawer__empty-desc {
  margin: 8px 0 0;
  max-width: 260px;
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  line-height: 1.5;
}

.cart-drawer__items {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.cart-drawer__footer {
  flex-shrink: 0;
  padding: 16px 20px 20px;
  border-top: 1px solid var(--color-border);
  background: var(--color-surface-muted);
  box-shadow: 0 -4px 20px rgba(0, 0, 0, 0.04);
}

.cart-drawer__summary {
  margin-bottom: 16px;
  padding: 14px 16px;
  border-radius: var(--radius-md);
  background: var(--color-surface);
  border: 1px solid var(--color-border-light);
}

.cart-drawer__summary-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  font-size: 0.875rem;
  color: var(--color-text-secondary);
}

.cart-drawer__summary-row + .cart-drawer__summary-row {
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px dashed var(--color-border);
}

.cart-drawer__summary-row--total {
  font-size: 1rem;
  color: var(--color-text);
}

.cart-drawer__summary-row--total strong {
  font-size: 1.125rem;
  font-weight: 800;
  color: var(--color-primary);
  letter-spacing: -0.02em;
}

.cart-drawer__btn {
  letter-spacing: 0.02em;
}

.cart-drawer__btn--secondary {
  margin-top: 10px;
  background: var(--color-surface) !important;
}

.cart-drawer__continue {
  display: block;
  width: 100%;
  margin-top: 12px;
  padding: 8px;
  border: none;
  background: transparent;
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--color-text-muted);
  cursor: pointer;
  transition: color 0.2s;
}

.cart-drawer__continue:hover {
  color: var(--color-primary);
}
</style>
