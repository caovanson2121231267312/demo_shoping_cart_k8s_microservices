<template>
  <section class="category-showcase">
    <div class="section-header">
      <div>
        <h2 class="section-title">Mua sắm theo danh mục</h2>
        <p class="section-subtitle">{{ displayCategories.length }} hạng mục — chọn nhanh sản phẩm bạn cần</p>
      </div>
      <v-btn variant="text" color="primary" to="/products" class="text-none" size="small">
        Tất cả danh mục
        <v-icon end size="18">mdi-arrow-right</v-icon>
      </v-btn>
    </div>

    <div class="category-showcase__grid">
      <NuxtLink
        v-for="(cat, idx) in displayCategories"
        :key="cat.id"
        :to="`/products?category=${cat.slug}`"
        class="category-tile animate-fade-in"
        :style="{ animationDelay: `${idx * 40}ms`, background: categoryBg(idx) }"
      >
        <v-avatar size="52" class="category-tile__avatar" :style="{ background: categoryAvatarBg(idx) }">
          <v-icon :icon="categoryIcon(cat.slug)" size="26" color="primary" />
        </v-avatar>
        <div class="category-tile__name">{{ cat.name }}</div>
        <div v-if="cat.children?.length" class="category-tile__meta">
          {{ cat.children.length }} nhóm
        </div>
      </NuxtLink>
    </div>

    <div v-if="subcategories.length" class="category-showcase__subs mt-5">
      <div class="category-showcase__subs-label text-caption font-weight-bold text-muted mb-2">
        Danh mục phổ biến
      </div>
      <div class="category-showcase__subs-scroll">
        <v-chip
          v-for="sub in subcategories"
          :key="sub.id"
          :to="`/products?category=${sub.slug}`"
          size="small"
          variant="outlined"
          color="primary"
          class="text-none"
        >
          <v-icon start size="16">{{ categoryIcon(sub.slug) }}</v-icon>
          {{ sub.name }}
        </v-chip>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import type { Category } from '~/types'
import { categoryAvatarBg, categoryBg, categoryIcon } from '~/utils/categoryIcons'

const props = defineProps<{ categories: Category[] }>()

const displayCategories = computed(() => props.categories.filter((c) => !c.parent_id))

const subcategories = computed(() => {
  const subs: Category[] = []
  for (const cat of props.categories) {
    if (cat.children?.length) {
      subs.push(...cat.children.slice(0, 2))
    }
  }
  return subs.slice(0, 16)
})
</script>

<style scoped>
.section-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.category-showcase__grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 10px;
}

@media (min-width: 600px) {
  .category-showcase__grid {
    grid-template-columns: repeat(6, 1fr);
  }
}

@media (min-width: 960px) {
  .category-showcase__grid {
    grid-template-columns: repeat(8, 1fr);
  }
}

.category-tile {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 14px 8px;
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border-light);
  text-decoration: none;
  transition: transform 0.25s ease, box-shadow 0.25s ease;
}

.category-tile:hover {
  transform: translateY(-5px);
  box-shadow: var(--shadow-lg);
}

.category-tile__avatar {
  border: 2px solid var(--color-surface);
  margin-bottom: 8px;
}

.category-tile__name {
  font-size: 12px;
  font-weight: 700;
  color: var(--color-text);
  line-height: 1.2;
}

.category-tile__meta {
  font-size: 10px;
  color: var(--color-text-muted);
  margin-top: 3px;
}

.category-showcase__subs-scroll {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
</style>
