<template>
  <section class="trending-tags">
    <v-container class="trending-tags__inner">
      <div class="trending-tags__label">
        <v-icon size="18" color="error">mdi-fire</v-icon>
        <span>Tìm kiếm phổ biến</span>
      </div>
      <div class="trending-tags__list">
        <NuxtLink
          v-for="tag in tags"
          :key="tag"
          :to="`/products?search=${encodeURIComponent(tag)}`"
          class="trending-tags__chip"
        >
          {{ tag }}
        </NuxtLink>
      </div>
    </v-container>
  </section>
</template>

<script setup lang="ts">
const tags = [
  'iPhone 16', 'Samsung Galaxy', 'Tai nghe Bluetooth', 'Laptop gaming',
  'Áo thun nam', 'Giày sneaker', 'Nồi cơm điện', 'Máy lọc không khí',
  'Son môi', 'Vitamin C', 'Balo du lịch', 'Đồng hồ thông minh',
]
</script>

<style scoped>
.trending-tags {
  background: var(--color-surface-muted);
  border-bottom: 1px solid var(--color-border);
}

.trending-tags__inner {
  max-width: 1280px;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
  padding-top: 14px;
  padding-bottom: 14px;
}

.trending-tags__label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 700;
  color: var(--color-text);
  flex-shrink: 0;
}

.trending-tags__list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.trending-tags__chip {
  font-size: 12px;
  font-weight: 500;
  color: var(--color-text-secondary);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 999px;
  padding: 5px 14px;
  text-decoration: none;
  transition: color 0.2s ease, border-color 0.2s ease, background 0.2s ease;
}

.trending-tags__chip:hover {
  color: var(--color-primary);
  border-color: var(--color-primary-alpha-35);
  background: var(--color-primary-subtle);
}
</style>
