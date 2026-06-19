<template>
  <div v-if="article" class="article-detail">
    <PageBanner
      :title="article.title"
      :subtitle="article.excerpt || ''"
      compact
      :breadcrumbs="[
        { label: 'Bài viết', to: '/blog' },
        { label: categoryLabel },
      ]"
    />

    <v-container class="page-container py-8">
      <v-row>
        <v-col cols="12" lg="8">
          <v-img
            :src="article.cover_image || `https://picsum.photos/seed/${article.slug}/1200/630`"
            :alt="article.title"
            rounded="lg"
            class="mb-6"
            cover
            max-height="420"
          />
          <div class="d-flex flex-wrap align-center ga-3 mb-4 text-caption text-muted">
            <v-chip size="small" color="primary" variant="tonal">{{ categoryLabel }}</v-chip>
            <span><v-icon size="14" class="mr-1">mdi-account</v-icon>{{ article.author_name }}</span>
            <span><v-icon size="14" class="mr-1">mdi-calendar</v-icon>{{ formatDate(article.created_at) }}</span>
            <span><v-icon size="14" class="mr-1">mdi-eye</v-icon>{{ article.view_count }} lượt xem</span>
          </div>
          <div v-if="article.tags?.length" class="mb-4">
            <v-chip v-for="tag in article.tags" :key="tag" size="x-small" class="mr-1 mb-1" variant="outlined">
              #{{ tag }}
            </v-chip>
          </div>
          <div class="article-content" v-html="article.content" />
        </v-col>
        <v-col cols="12" lg="4">
          <v-card class="pa-4 mb-4" rounded="lg" variant="outlined">
            <div class="text-subtitle-2 font-weight-bold mb-3">Bài viết liên quan</div>
            <v-list density="compact" class="pa-0">
              <v-list-item
                v-for="rel in related"
                :key="rel.id"
                :to="`/blog/${rel.slug}`"
                :title="rel.title"
                :subtitle="formatDate(rel.created_at)"
                class="px-0"
              />
            </v-list>
          </v-card>
          <v-card class="pa-4" rounded="lg" color="primary" variant="tonal">
            <div class="font-weight-bold mb-2">Mua sắm ngay</div>
            <p class="text-body-2 mb-3">Khám phá hàng ngàn sản phẩm chính hãng với giá tốt.</p>
            <v-btn color="primary" to="/products" block class="text-none">Xem sản phẩm</v-btn>
          </v-card>
        </v-col>
      </v-row>
    </v-container>
  </div>
  <LoadingSpinner v-else-if="loading" />
  <EmptyState v-else icon="mdi-post-outline" title="Không tìm thấy bài viết" description="Bài viết có thể đã bị gỡ hoặc đường dẫn không đúng.">
    <v-btn color="primary" to="/blog" class="text-none mt-4">Về danh sách bài viết</v-btn>
  </EmptyState>
</template>

<script setup lang="ts">
import type { Article } from '~/types'
import { ARTICLE_CATEGORIES } from '~/types'

definePageMeta({ layout: 'default' })

const route = useRoute()
const { fetchArticles, fetchArticleBySlug } = useArticles()

const article = ref<Article | null>(null)
const related = ref<Article[]>([])
const loading = ref(true)

const categoryLabel = computed(() =>
  article.value ? (ARTICLE_CATEGORIES[article.value.category] || article.value.category) : '',
)

const formatDate = (iso: string) => new Date(iso).toLocaleDateString('vi-VN')

onMounted(async () => {
  const slug = route.params.slug as string
  try {
    article.value = await fetchArticleBySlug(slug)
    const res = await fetchArticles({ category: article.value.category, limit: 5 })
    related.value = res.items.filter((a) => a.slug !== slug).slice(0, 4)
  } catch {
    article.value = null
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.article-content :deep(h2),
.article-content :deep(h3) {
  margin: 1.2em 0 0.6em;
  font-weight: 700;
}

.article-content :deep(p) {
  margin-bottom: 1em;
  line-height: 1.7;
  color: var(--color-text-secondary, #616161);
}

.article-content :deep(ul) {
  padding-left: 1.25rem;
  margin-bottom: 1em;
}

.article-content :deep(blockquote) {
  border-left: 4px solid var(--color-primary, #1565c0);
  padding-left: 1rem;
  margin: 1rem 0;
  color: var(--color-text-secondary);
  font-style: italic;
}

.article-content :deep(code) {
  background: #f5f5f5;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 0.9em;
}
</style>
