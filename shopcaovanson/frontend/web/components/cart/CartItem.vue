<template>
  <article :class="['cart-item', variant === 'drawer' ? 'cart-item--drawer' : 'cart-item--page']">
    <div v-if="variant === 'drawer'" class="cart-item__thumb">
      <v-img :src="imageUrl" cover class="cart-item__image" />
    </div>
    <v-img
      v-else
      :src="imageUrl"
      width="72"
      height="72"
      cover
      class="rounded flex-shrink-0"
    />

    <div class="cart-item__content">
      <div class="cart-item__top">
        <div v-if="variant === 'drawer'" class="cart-item__name">
          {{ item.product_name }}
        </div>
        <div v-else class="text-body-1 font-weight-medium text-truncate">{{ item.product_name }}</div>

        <button
          v-if="variant === 'drawer'"
          type="button"
          class="cart-item__remove"
          :disabled="removing"
          aria-label="Xóa sản phẩm"
          @click="handleRemove"
        >
          <v-progress-circular v-if="removing" indeterminate size="16" width="2" color="error" />
          <v-icon v-else size="18">mdi-close</v-icon>
        </button>
      </div>

      <div :class="variant === 'drawer' ? 'cart-item__price' : 'text-primary font-weight-bold'">
        {{ formatVND(item.unit_price) }}
      </div>

      <div :class="variant === 'drawer' ? 'cart-item__footer' : 'd-flex align-center ga-2 mt-2'">
        <div :class="variant === 'drawer' ? 'cart-item__qty' : 'd-flex align-center ga-2'">
          <button
            type="button"
            class="cart-item__qty-btn"
            :class="{ 'cart-item__qty-btn--page': variant === 'page' }"
            :disabled="item.quantity <= 1 || updating"
            aria-label="Giảm số lượng"
            @click="changeQty(item.quantity - 1)"
          >
            <v-icon size="16">mdi-minus</v-icon>
          </button>
          <span :class="variant === 'drawer' ? 'cart-item__qty-value' : 'text-body-2'">{{ item.quantity }}</span>
          <button
            type="button"
            class="cart-item__qty-btn"
            :class="{ 'cart-item__qty-btn--page': variant === 'page' }"
            :disabled="updating"
            aria-label="Tăng số lượng"
            @click="changeQty(item.quantity + 1)"
          >
            <v-icon size="16">mdi-plus</v-icon>
          </button>
        </div>

        <span v-if="variant === 'drawer'" class="cart-item__line-total">
          {{ formatVND(lineTotal) }}
        </span>
      </div>
    </div>

    <template v-if="variant === 'page'">
      <div class="text-right">
        <div class="font-weight-bold mb-2">{{ formatVND(lineTotal) }}</div>
        <v-btn
          icon
          size="small"
          variant="text"
          color="error"
          :loading="removing"
          @click="handleRemove"
        >
          <v-icon>mdi-delete-outline</v-icon>
        </v-btn>
      </div>
    </template>
  </article>
</template>

<script setup lang="ts">
import type { CartItem } from '~/types'

const props = withDefaults(
  defineProps<{
    item: CartItem
    variant?: 'page' | 'drawer'
  }>(),
  { variant: 'page' },
)

const { formatVND } = useFormat()
const cart = useCart()

const updating = ref(false)
const removing = ref(false)

const lineTotal = computed(() => props.item.unit_price * props.item.quantity)
const imageUrl = computed(() => {
  if (props.item.product_image) {
    return props.item.product_image
  }
  return productImageUrl(props.item.product_id)
})
const changeQty = async (qty: number) => {
  if (qty < 1) {
    return
  }
  updating.value = true
  try {
    await cart.updateItem(props.item.product_id, qty)
  } finally {
    updating.value = false
  }
}

const handleRemove = async () => {
  removing.value = true
  try {
    await cart.removeItem(props.item.product_id)
  } finally {
    removing.value = false
  }
}
</script>

<style scoped>
.cart-item--page {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
  padding: 16px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-surface);
}

.cart-item--drawer {
  display: flex;
  gap: 14px;
  padding: 14px;
  border-radius: var(--radius-md);
  background: var(--color-surface);
  border: 1px solid var(--color-border-light);
  box-shadow: var(--shadow-sm);
  transition: border-color 0.2s, box-shadow 0.2s;
}

.cart-item--drawer:hover {
  border-color: var(--color-primary-alpha-35);
  box-shadow: var(--shadow-md);
}

.cart-item__thumb {
  flex-shrink: 0;
  width: 76px;
  height: 76px;
  border-radius: var(--radius-sm);
  overflow: hidden;
  border: 1px solid var(--color-border-light);
  background: var(--color-surface-muted);
  text-decoration: none;
}

.cart-item__image {
  width: 100%;
  height: 100%;
}

.cart-item__content {
  flex: 1 1 auto;
  min-width: 0;
}

.cart-item__top {
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.cart-item__name {
  flex: 1;
  min-width: 0;
  font-size: 0.875rem;
  font-weight: 600;
  line-height: 1.4;
  color: var(--color-text);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.cart-item__remove {
  width: 28px;
  height: 28px;
  border: none;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-text-muted);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: background 0.2s, color 0.2s;
}

.cart-item__remove:hover:not(:disabled) {
  background: var(--color-sale-bg);
  color: var(--color-sale);
}

.cart-item__remove:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.cart-item__price {
  margin-top: 4px;
  font-size: 0.875rem;
  font-weight: 700;
  color: var(--color-primary);
}

.cart-item__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-top: 10px;
}

.cart-item__qty {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  padding: 2px;
  border-radius: var(--radius-pill);
  background: var(--color-surface-muted);
  border: 1px solid var(--color-border-light);
}

.cart-item__qty-btn {
  width: 28px;
  height: 28px;
  border: none;
  border-radius: 50%;
  background: var(--color-surface);
  color: var(--color-text-secondary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s, color 0.2s, box-shadow 0.2s;
}

.cart-item__qty-btn--page {
  width: 32px;
  height: 32px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
}

.cart-item__qty-btn:hover:not(:disabled) {
  background: var(--color-primary-50);
  color: var(--color-primary);
}

.cart-item__qty-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.cart-item__qty-value {
  min-width: 28px;
  text-align: center;
  font-size: 0.875rem;
  font-weight: 700;
  color: var(--color-text);
}

.cart-item__line-total {
  font-size: 0.875rem;
  font-weight: 800;
  color: var(--color-text);
  white-space: nowrap;
}
</style>
