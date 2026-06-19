<template>
  <section class="quick-links">
    <v-container class="quick-links__inner">
      <NuxtLink
        v-for="link in links"
        :key="link.label"
        :to="link.to"
        class="quick-links__item"
      >
        <div class="quick-links__icon" :style="{ background: link.bg }">
          <v-icon :color="link.color" size="24">{{ link.icon }}</v-icon>
        </div>
        <span class="quick-links__label">{{ link.label }}</span>
        <v-chip v-if="link.badge" size="x-small" color="error" variant="flat" class="quick-links__badge">
          {{ link.badge }}
        </v-chip>
      </NuxtLink>
    </v-container>
  </section>
</template>

<script setup lang="ts">
const links = [
  { label: 'Flash Sale', icon: 'mdi-flash', to: '/products?sort=price_desc', color: 'error', bg: 'var(--color-sale-bg)', badge: 'HOT' },
  { label: 'Mã giảm giá', icon: 'mdi-ticket-percent', to: '/products', color: 'secondary', bg: 'var(--color-secondary-50)' },
  { label: 'Hàng mới', icon: 'mdi-new-box', to: '/products?sort=newest', color: 'primary', bg: 'var(--color-primary-50)' },
  { label: 'Yêu thích', icon: 'mdi-heart', to: '/wishlist', color: 'error', bg: 'var(--color-sale-bg)' },
  { label: 'Theo dõi đơn', icon: 'mdi-package-variant', to: '/orders/track', color: 'info', bg: 'var(--color-info-bg)' },
  { label: 'Giỏ hàng', icon: 'mdi-cart', to: '/cart', color: 'primary', bg: 'var(--color-primary-50)' },
  { label: 'Đơn hàng', icon: 'mdi-receipt-text', to: '/orders', color: 'success', bg: 'var(--color-success-bg)' },
  { label: 'Hỗ trợ', icon: 'mdi-headset', to: '/profile', color: 'primary', bg: 'var(--color-primary-50)' },
]
</script>

<style scoped>
.quick-links {
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
}

.quick-links__inner {
  max-width: 1280px;
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
  padding-top: 16px;
  padding-bottom: 16px;
}

@media (min-width: 600px) {
  .quick-links__inner {
    grid-template-columns: repeat(8, 1fr);
  }
}

.quick-links__item {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 8px 4px;
  text-decoration: none;
  border-radius: var(--radius-md);
  transition: background 0.2s ease, transform 0.2s ease;
}

.quick-links__item:hover {
  background: var(--color-surface-hover);
  transform: translateY(-2px);
}

.quick-links__icon {
  width: 44px;
  height: 44px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
}

.quick-links__label {
  font-size: 11px;
  font-weight: 600;
  color: var(--color-text);
  text-align: center;
  line-height: 1.2;
}

@media (min-width: 600px) {
  .quick-links__label {
    font-size: 12px;
  }
}

.quick-links__badge {
  position: absolute;
  top: 2px;
  right: 2px;
  font-size: 9px !important;
  height: 16px !important;
}
</style>
