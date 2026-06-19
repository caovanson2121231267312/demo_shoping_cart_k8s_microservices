<template>
  <section class="blog-section py-8">
    <v-container class="page-container">
      <div class="section-header d-flex align-end justify-space-between mb-5 flex-wrap ga-3">
        <div>
          <h2 class="section-title">Tin tức & Bài viết</h2>
          <p class="section-subtitle">Cẩm nang mua sắm, review và khuyến mãi mới nhất</p>
        </div>
        <v-btn variant="outlined" color="primary" to="/blog" class="text-none" rounded="lg">
          Xem tất cả
          <v-icon end>mdi-arrow-right</v-icon>
        </v-btn>
      </div>
      <v-row v-if="articles.length">
        <v-col v-for="article in articles" :key="article.id" cols="12" sm="6" md="3">
          <ArticleCard :article="article" />
        </v-col>
      </v-row>
      <div v-else-if="!loading" class="text-center text-muted py-8">Chưa có bài viết</div>
    </v-container>
  </section>
</template>

<script setup lang="ts">
import type { Article } from '~/types'

const { fetchArticles } = useArticles()
const articles = ref<Article[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const res = await fetchArticles({ limit: 4, featured: 'true' })
    articles.value = res.items.length ? res.items : (await fetchArticles({ limit: 4 })).items
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.blog-section {
  background: var(--color-surface);
  border-top: 1px solid var(--color-border-light);
}
</style>
