<template>
  <nav class="category-nav">
    <v-container class="category-nav__inner">
      <v-menu open-on-hover location="bottom start" :close-on-content-click="true">
        <template #activator="{ props: menuProps }">
          <button type="button" class="category-nav__all" v-bind="menuProps">
            <v-icon size="20">mdi-menu</v-icon>
            <span>Danh mục</span>
            <v-icon size="16" class="category-nav__chevron">mdi-chevron-down</v-icon>
          </button>
        </template>
        <v-card min-width="280" max-width="320" max-height="460" class="category-nav__dropdown" elevation="8" rounded="lg">
          <v-list density="compact" class="py-2">
            <template v-for="root in rootCategories" :key="root.id">
              <v-list-item
                :to="`/products?category=${root.slug}`"
                :prepend-icon="categoryIcon(root.slug)"
                :title="root.name"
                class="category-nav__dropdown-item font-weight-bold"
                rounded="lg"
              />
              <v-list-item
                v-for="child in root.children || []"
                :key="child.id"
                :to="`/products?category=${child.slug}`"
                :prepend-icon="categoryIcon(child.slug)"
                :title="child.name"
                class="category-nav__dropdown-item pl-10"
                rounded="lg"
              />
              <v-divider v-if="root !== rootCategories[rootCategories.length - 1]" class="my-1 mx-3" />
            </template>
          </v-list>
        </v-card>
      </v-menu>

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
        <NuxtLink
          v-for="cat in rootCategories.slice(0, 6)"
          :key="cat.id"
          :to="`/products?category=${cat.slug}`"
          class="category-nav__link d-none d-lg-inline-flex"
        >
          <v-icon size="16">{{ categoryIcon(cat.slug) }}</v-icon>
          {{ cat.name }}
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
}

@media (min-width: 960px) {
  .category-nav {
    display: block;
  }
}

.category-nav__inner {
  max-width: 1280px;
  display: flex;
  align-items: center;
  min-height: 46px;
  padding-top: 0;
  padding-bottom: 0;
  gap: 0;
}

.category-nav__all {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border: none;
  background: var(--gradient-primary);
  color: var(--color-text-inverse);
  font-size: 14px;
  font-weight: 700;
  border-radius: var(--radius-sm);
  cursor: pointer;
  white-space: nowrap;
  transition: opacity 0.2s, transform 0.2s;
  flex-shrink: 0;
}

.category-nav__all:hover {
  opacity: 0.92;
  transform: translateY(-1px);
}

.category-nav__chevron {
  opacity: 0.8;
}

.category-nav__divider {
  width: 1px;
  height: 24px;
  background: var(--color-border);
  margin: 0 12px;
  flex-shrink: 0;
}

.category-nav__links {
  display: flex;
  align-items: center;
  gap: 4px;
  overflow-x: auto;
  scrollbar-width: none;
  flex: 1;
  min-width: 0;
}

.category-nav__links::-webkit-scrollbar {
  display: none;
}

.category-nav__link {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 8px 12px;
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text-secondary);
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

.category-nav__dropdown-item {
  margin: 0 4px;
}
</style>
