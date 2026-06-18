<template>
  <div class="category-nav border-b">
    <v-container class="page-container py-0 d-flex align-center" style="min-height: 44px">
      <v-menu open-on-hover location="bottom">
        <template #activator="{ props: menuProps }">
          <v-btn
            v-bind="menuProps"
            variant="text"
            size="small"
            class="category-nav__all text-none font-weight-bold flex-shrink-0"
            prepend-icon="mdi-menu"
          >
            Tất cả danh mục
          </v-btn>
        </template>
        <v-card min-width="300" max-height="420" class="overflow-y-auto">
          <v-list density="compact">
            <template v-for="root in rootCategories" :key="root.id">
              <v-list-item
                :to="`/products?category=${root.slug}`"
                :prepend-icon="categoryIcon(root.slug)"
                :title="root.name"
                class="font-weight-bold"
              />
              <v-list-item
                v-for="child in root.children || []"
                :key="child.id"
                :to="`/products?category=${child.slug}`"
                :prepend-icon="categoryIcon(child.slug)"
                :title="child.name"
                class="pl-8"
              />
              <v-divider class="my-1" />
            </template>
          </v-list>
        </v-card>
      </v-menu>

      <v-divider vertical class="mx-2 flex-shrink-0" />

      <div class="category-nav__scroll d-flex align-center flex-grow-1">
        <v-btn
          v-for="cat in rootCategories"
          :key="cat.id"
          :to="`/products?category=${cat.slug}`"
          variant="text"
          size="small"
          class="text-none flex-shrink-0"
        >
          <v-icon start size="16">{{ categoryIcon(cat.slug) }}</v-icon>
          {{ cat.name }}
        </v-btn>
        <v-chip to="/products?sort=price_desc" color="error" size="small" variant="flat" class="mx-1 flex-shrink-0">
          <v-icon start size="14">mdi-fire</v-icon> Sale
        </v-chip>
        <v-chip to="/products?sort=newest" size="small" variant="outlined" class="mx-1 flex-shrink-0">Mới</v-chip>
        <v-chip to="/wishlist" size="small" variant="outlined" color="pink" class="mx-1 flex-shrink-0">
          <v-icon start size="14">mdi-heart</v-icon>Yêu thích
        </v-chip>
        <v-chip to="/orders/track" size="small" variant="outlined" class="mx-1 flex-shrink-0">
          Tra cứu đơn
        </v-chip>
      </div>
    </v-container>
  </div>
</template>

<script setup lang="ts">
import type { Category } from '~/types'
import { categoryIcon } from '~/utils/categoryIcons'

const props = defineProps<{ categories: Category[] }>()
const rootCategories = computed(() => props.categories.filter((c) => !c.parent_id))
</script>

<style scoped>
.category-nav {
  background: #fff;
  box-shadow: 0 1px 0 rgba(0, 0, 0, 0.06);
}

.category-nav__all {
  color: rgb(var(--v-theme-primary));
}

.category-nav__scroll {
  overflow-x: auto;
  scrollbar-width: none;
}

.category-nav__scroll::-webkit-scrollbar {
  display: none;
}

.border-b {
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
}
</style>
