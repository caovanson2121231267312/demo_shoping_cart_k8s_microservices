<template>
  <nav class="category-nav">
    <v-container class="category-nav__inner">
      <div class="category-nav__roots">
        <div
          v-for="root in rootCategories"
          :key="root.id"
          class="category-nav__root"
          :class="{ 'category-nav__root--has-children': root.children?.length }"
        >
          <NuxtLink
            :to="`/products?category=${root.slug}`"
            class="category-nav__root-link"
          >
            <v-icon size="17">{{ categoryIcon(root.slug) }}</v-icon>
            <span>{{ root.name }}</span>
            <v-icon
              v-if="root.children?.length"
              size="14"
              class="category-nav__root-chevron"
            >
              mdi-chevron-down
            </v-icon>
          </NuxtLink>

          <div v-if="root.children?.length" class="category-nav__panel">
            <div class="category-nav__panel-head">
              <span class="category-nav__panel-title">{{ root.name }}</span>
              <NuxtLink :to="`/products?category=${root.slug}`" class="category-nav__panel-all">
                Xem tất cả
                <v-icon size="14">mdi-arrow-right</v-icon>
              </NuxtLink>
            </div>
            <div class="category-nav__panel-grid">
              <NuxtLink
                v-for="child in root.children"
                :key="child.id"
                :to="`/products?category=${child.slug}`"
                class="category-nav__child-link"
              >
                <v-icon size="18">{{ categoryIcon(child.slug) }}</v-icon>
                <span>{{ child.name }}</span>
              </NuxtLink>
            </div>
          </div>
        </div>
      </div>

      <div class="category-nav__divider" />

      <div class="category-nav__links">
        <NuxtLink
          v-for="link in quickLinks"
          :key="link.to"
          :to="link.to"
          class="category-nav__link"
          :class="{ 'category-nav__link--hot': link.hot }"
        >
          <v-icon v-if="link.icon" size="16">{{ link.icon }}</v-icon>
          {{ link.label }}
        </NuxtLink>
      </div>
    </v-container>
  </nav>
</template>

<script setup lang="ts">
import type { Category } from '~/types'
import { categoryIcon } from '~/utils/categoryIcons'

const props = defineProps<{ categories: Category[] }>()
const rootCategories = computed(() => props.categories.filter((c) => !c.parent_id))

const quickLinks = [
  { to: '/products?sort=price_desc', label: 'Flash Sale', icon: 'mdi-flash', hot: true },
  { to: '/products?sort=newest', label: 'Hàng mới', icon: 'mdi-new-box', hot: false },
  { to: '/blog', label: 'Tin tức', icon: 'mdi-post-outline', hot: false },
  { to: '/products', label: 'Tất cả SP', icon: 'mdi-shopping', hot: false },
  { to: '/wishlist', label: 'Yêu thích', icon: 'mdi-heart', hot: false },
  { to: '/orders/track', label: 'Tra cứu đơn', icon: 'mdi-truck-fast', hot: false },
]
</script>

<style scoped>
.category-nav {
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  display: none;
  position: relative;
  z-index: 90;
}

@media (min-width: 960px) {
  .category-nav {
    display: block;
  }
}

.category-nav__inner {
  max-width: 1280px;
  display: flex;
  align-items: stretch;
  min-height: 46px;
  padding-top: 0;
  padding-bottom: 0;
  gap: 0;
}

.category-nav__roots {
  display: flex;
  align-items: stretch;
  flex: 1;
  min-width: 0;
  overflow-x: auto;
  scrollbar-width: none;
}

.category-nav__roots::-webkit-scrollbar {
  display: none;
}

.category-nav__root {
  position: relative;
  flex-shrink: 0;
}

.category-nav__root-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 46px;
  padding: 0 14px;
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text-secondary);
  text-decoration: none;
  white-space: nowrap;
  border-bottom: 2px solid transparent;
  transition: color 0.2s, background 0.2s, border-color 0.2s;
}

.category-nav__root-link:hover,
.category-nav__root:hover .category-nav__root-link,
.category-nav__root--has-children:focus-within .category-nav__root-link {
  color: var(--color-primary);
  background: var(--color-surface-hover);
  border-bottom-color: var(--color-primary);
}

.category-nav__root-chevron {
  opacity: 0.55;
  transition: transform 0.2s;
}

.category-nav__root:hover .category-nav__root-chevron,
.category-nav__root--has-children:focus-within .category-nav__root-chevron {
  transform: rotate(180deg);
  opacity: 1;
}

.category-nav__panel {
  position: absolute;
  top: calc(100% + 1px);
  left: 0;
  min-width: 280px;
  max-width: 360px;
  padding: 12px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-lg);
  opacity: 0;
  visibility: hidden;
  transform: translateY(6px);
  transition: opacity 0.18s ease, transform 0.18s ease, visibility 0.18s;
  pointer-events: none;
  z-index: 120;
}

.category-nav__root:hover .category-nav__panel,
.category-nav__root--has-children:focus-within .category-nav__panel {
  opacity: 1;
  visibility: visible;
  transform: translateY(0);
  pointer-events: auto;
}

.category-nav__panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 4px 6px 10px;
  margin-bottom: 4px;
  border-bottom: 1px solid var(--color-border-light);
}

.category-nav__panel-title {
  font-size: 13px;
  font-weight: 700;
  color: var(--color-text);
}

.category-nav__panel-all {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  font-size: 12px;
  font-weight: 600;
  color: var(--color-primary);
  text-decoration: none;
}

.category-nav__panel-all:hover {
  text-decoration: underline;
}

.category-nav__panel-grid {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.category-nav__child-link {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 10px;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text-secondary);
  text-decoration: none;
  transition: background 0.15s, color 0.15s;
}

.category-nav__child-link:hover {
  background: var(--color-primary-50);
  color: var(--color-primary);
}

.category-nav__divider {
  width: 1px;
  align-self: center;
  height: 24px;
  background: var(--color-border);
  margin: 0 8px 0 4px;
  flex-shrink: 0;
}

.category-nav__links {
  display: flex;
  align-items: center;
  gap: 2px;
  overflow-x: auto;
  scrollbar-width: none;
  flex-shrink: 0;
  max-width: 42%;
}

.category-nav__links::-webkit-scrollbar {
  display: none;
}

.category-nav__link {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 8px 10px;
  font-size: 12px;
  font-weight: 600;
  color: var(--color-text-muted);
  text-decoration: none;
  border-radius: var(--radius-sm);
  white-space: nowrap;
  transition: background 0.2s, color 0.2s;
  flex-shrink: 0;
}

.category-nav__link:hover {
  background: var(--color-surface-hover);
  color: var(--color-primary);
}

.category-nav__link--hot {
  color: var(--color-sale);
}

.category-nav__link--hot:hover {
  background: var(--color-sale-bg);
  color: var(--color-sale-dark);
}
</style>
