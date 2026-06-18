<template>
  <section class="mb-10">
    <h2 class="text-h5 font-weight-bold mb-4">Mua sắm theo danh mục</h2>
    <v-row>
      <v-col
        v-for="(cat, idx) in displayCategories"
        :key="cat.id"
        cols="6"
        sm="4"
        md="3"
        lg="2"
      >
        <v-card
          :to="`/products?category=${cat.slug}`"
          class="category-tile text-center pa-4 animate-fade-in"
          :style="{ animationDelay: `${idx * 50}ms`, background: categoryColor(idx) }"
          elevation="0"
          rounded="lg"
        >
          <v-avatar :color="categoryColor(idx + 3)" size="56" class="mb-3">
            <v-icon :icon="categoryIcon(cat.slug)" size="28" color="primary" />
          </v-avatar>
          <div class="text-body-2 font-weight-bold">{{ cat.name }}</div>
          <div v-if="cat.children?.length" class="text-caption text-grey mt-1">
            {{ cat.children.length }} nhóm
          </div>
        </v-card>
      </v-col>
    </v-row>
  </section>
</template>

<script setup lang="ts">
import type { Category } from '~/types'
import { categoryColor, categoryIcon } from '~/utils/categoryIcons'

const props = defineProps<{ categories: Category[] }>()

const displayCategories = computed(() => props.categories.filter((c) => !c.parent_id).slice(0, 12))
</script>

<style scoped>
.category-tile {
  transition: transform 0.25s ease, box-shadow 0.25s ease;
  border: 1px solid rgba(0, 0, 0, 0.04);
}

.category-tile:hover {
  transform: translateY(-6px) scale(1.02);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
}
</style>
