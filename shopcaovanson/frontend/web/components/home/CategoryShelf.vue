<template>
  <section class="category-shelf">
    <div class="section-header">
      <div class="d-flex align-center ga-3">
        <v-avatar size="40" :style="{ background: categoryAvatarBg(colorIndex) }">
          <v-icon :icon="categoryIcon(category.slug)" color="primary" size="22" />
        </v-avatar>
        <div>
          <h2 class="section-title">{{ category.name }}</h2>
          <p v-if="category.children?.length" class="section-subtitle">
            {{ category.children.length }} nhóm sản phẩm
          </p>
        </div>
      </div>
      <v-btn
        variant="outlined"
        color="primary"
        :to="`/products?category=${category.slug}`"
        rounded="lg"
        class="text-none"
        size="small"
      >
        Xem tất cả
        <v-icon end size="18">mdi-arrow-right</v-icon>
      </v-btn>
    </div>

    <div v-if="category.children?.length" class="category-shelf__chips mb-4">
      <v-chip
        v-for="child in category.children.slice(0, 6)"
        :key="child.id"
        :to="`/products?category=${child.slug}`"
        size="small"
        variant="outlined"
        color="primary"
        class="text-none"
      >
        {{ child.name }}
      </v-chip>
    </div>

    <v-row v-if="products.length">
      <v-col v-for="(product, idx) in products" :key="product.id" cols="6" sm="4" md="3">
        <div class="animate-fade-in" :style="{ animationDelay: `${idx * 50}ms` }">
          <ProductCard :product="product" :summary="summaries?.[product.id]" />
        </div>
      </v-col>
    </v-row>
    <div v-else class="category-shelf__empty text-center py-6 text-muted">
      Đang cập nhật sản phẩm...
    </div>
  </section>
</template>

<script setup lang="ts">
import type { Category, Product } from '~/types'
import type { ReviewSummary } from '~/composables/useReviewSummary'
import { categoryAvatarBg, categoryIcon } from '~/utils/categoryIcons'

defineProps<{
  category: Category
  products: Product[]
  colorIndex?: number
  summaries?: Record<string, ReviewSummary>
}>()
</script>

<style scoped>
.category-shelf {
  margin-bottom: 40px;
}

.category-shelf__chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.section-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.category-shelf__empty {
  background: var(--color-surface-muted);
  border-radius: var(--radius-lg);
  border: 1px dashed var(--color-border);
}
</style>
