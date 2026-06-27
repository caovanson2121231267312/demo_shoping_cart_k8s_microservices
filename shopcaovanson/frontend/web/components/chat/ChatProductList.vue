<template>
  <div class="chat-product-list">
    <NuxtLink
      v-for="item in products"
      :key="item.id"
      :to="`/products/${item.slug}`"
      class="chat-product-list__item"
    >
      <div class="chat-product-list__thumb">
        <img
          :src="imageFor(item)"
          :alt="item.name"
          loading="lazy"
          class="chat-product-list__img"
        >
        <v-chip
          v-if="isOnSale(item)"
          size="x-small"
          color="error"
          variant="flat"
          class="chat-product-list__badge"
        >
          Sale
        </v-chip>
      </div>
      <div class="chat-product-list__info">
        <div class="chat-product-list__name text-body-2">{{ item.name }}</div>
        <div class="d-flex align-center flex-wrap ga-2 mt-1">
          <span class="chat-product-list__price text-body-2 font-weight-bold text-primary">
            {{ formatVND(displayPrice(item)) }}
          </span>
          <span
            v-if="isOnSale(item)"
            class="chat-product-list__price-old text-caption text-medium-emphasis"
          >
            {{ formatVND(item.price) }}
          </span>
        </div>
        <div v-if="item.stock !== undefined && item.stock <= 0" class="text-caption text-error mt-1">
          Hết hàng
        </div>
      </div>
      <v-icon size="18" class="chat-product-list__arrow text-medium-emphasis">
        mdi-chevron-right
      </v-icon>
    </NuxtLink>
  </div>
</template>

<script setup lang="ts">
import type { ChatProductSuggestion } from '~/types'
import { formatVND, getProductImage } from '~/composables/useFormat'

defineProps<{
  products: ChatProductSuggestion[]
}>()

const imageFor = (item: ChatProductSuggestion) =>
  getProductImage({ slug: item.slug, images: item.image_url ? [item.image_url] : [] })

const displayPrice = (item: ChatProductSuggestion) =>
  item.sale_price != null && item.sale_price < item.price ? item.sale_price : item.price

const isOnSale = (item: ChatProductSuggestion) =>
  item.sale_price != null && item.sale_price < item.price
</script>

<style scoped>
.chat-product-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 10px;
}

.chat-product-list__item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px;
  border-radius: 10px;
  border: 1px solid var(--color-border, rgba(0, 0, 0, 0.08));
  background: var(--color-surface, #fff);
  text-decoration: none;
  color: inherit;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

.chat-product-list__item:hover {
  border-color: rgba(var(--v-theme-primary), 0.35);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.chat-product-list__thumb {
  position: relative;
  flex-shrink: 0;
  width: 56px;
  height: 56px;
  border-radius: 8px;
  overflow: hidden;
  background: var(--color-bg-muted, #f5f5f5);
}

.chat-product-list__img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.chat-product-list__badge {
  position: absolute;
  top: 4px;
  left: 4px;
  font-size: 0.625rem;
  height: 16px;
  padding: 0 4px;
}

.chat-product-list__info {
  flex: 1;
  min-width: 0;
}

.chat-product-list__name {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  line-height: 1.35;
}

.chat-product-list__price-old {
  text-decoration: line-through;
}

.chat-product-list__arrow {
  flex-shrink: 0;
}
</style>
