<template>
  <NuxtLink :to="`/blog/${article.slug}`" class="article-card">
    <v-card class="h-100" elevation="0" rounded="lg">
      <v-img :src="cover" :alt="article.title" aspect-ratio="16/9" cover class="article-card__img" />
      <v-card-text class="pa-4">
        <v-chip size="x-small" color="primary" variant="tonal" class="mb-2">{{ categoryLabel }}</v-chip>
        <div v-if="article.is_featured" class="mb-1">
          <v-chip size="x-small" color="secondary" variant="flat">Nổi bật</v-chip>
        </div>
        <h3 class="article-card__title text-body-1 font-weight-bold mb-2">{{ article.title }}</h3>
        <p v-if="article.excerpt" class="article-card__excerpt text-caption text-muted mb-3">{{ article.excerpt }}</p>
        <div class="d-flex align-center justify-space-between text-caption text-muted">
          <span>{{ article.author_name }}</span>
          <span>{{ formatDate(article.created_at) }}</span>
        </div>
      </v-card-text>
    </v-card>
  </NuxtLink>
</template>

<script setup lang="ts">
import type { Article } from '~/types'
import { ARTICLE_CATEGORIES } from '~/types'

const props = defineProps<{ article: Article }>()

const cover = computed(() => props.article.cover_image || `https://picsum.photos/seed/${props.article.slug}/800/450`)
const categoryLabel = computed(() => ARTICLE_CATEGORIES[props.article.category] || props.article.category)

const formatDate = (iso: string) => {
  try {
    return new Date(iso).toLocaleDateString('vi-VN')
  } catch {
    return ''
  }
}
</script>

<style scoped>
.article-card {
  text-decoration: none;
  color: inherit;
  display: block;
  height: 100%;
}

.article-card .v-card {
  border: 1px solid var(--color-border-light, rgba(0, 0, 0, 0.06));
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.article-card:hover .v-card {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lg, 0 8px 24px rgba(0, 0, 0, 0.1));
}

.article-card__title {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  line-height: 1.35;
  color: var(--color-text, #212121);
}

.article-card__excerpt {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
