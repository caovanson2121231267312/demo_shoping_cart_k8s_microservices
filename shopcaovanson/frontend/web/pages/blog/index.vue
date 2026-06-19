<template>
  <div class="blog-page">
    <PageBanner
      title="Tin tức & Bài viết"
      subtitle="Cẩm nang mua sắm, đánh giá sản phẩm, khuyến mãi và xu hướng"
      :breadcrumbs="[{ label: 'Bài viết' }]"
    />

    <v-container class="page-container py-8">
      <div class="d-flex flex-wrap ga-3 mb-6 align-center">
        <v-chip-group v-model="category" mandatory selected-class="text-primary" @update:model-value="reload">
          <v-chip value="" variant="outlined" class="text-none">Tất cả</v-chip>
          <v-chip v-for="(label, key) in ARTICLE_CATEGORIES" :key="key" :value="key" variant="outlined" class="text-none">
            {{ label }}
          </v-chip>
        </v-chip-group>
        <v-spacer />
        <v-text-field
          v-model="search"
          density="compact"
          hide-details
          placeholder="Tìm bài viết..."
          prepend-inner-icon="mdi-magnify"
          variant="outlined"
          rounded="lg"
          style="max-width:280px"
          @keyup.enter="reload"
        />
      </div>

      <LoadingSpinner v-if="loading" />
      <v-row v-else-if="articles.length">
        <v-col v-for="article in articles" :key="article.id" cols="12" sm="6" md="4" lg="3">
          <ArticleCard :article="article" />
        </v-col>
      </v-row>
      <EmptyState v-else icon="mdi-post-outline" title="Không có bài viết" description="Thử đổi bộ lọc hoặc từ khóa." />

      <div v-if="totalPages > 1" class="d-flex justify-center mt-8">
        <v-pagination v-model="page" :length="totalPages" rounded="lg" @update:model-value="load" />
      </div>
    </v-container>
  </div>
</template>

<script setup lang="ts">
import type { Article } from '~/types'
import { ARTICLE_CATEGORIES } from '~/types'

definePageMeta({ layout: 'default' })

const { fetchArticles } = useArticles()

const articles = ref<Article[]>([])
const loading = ref(true)
const page = ref(1)
const totalPages = ref(1)
const category = ref('')
const search = ref('')

const load = async () => {
  loading.value = true
  try {
    const query: Record<string, string | number> = { page: page.value, limit: 12 }
    if (category.value) query.category = category.value
    if (search.value.trim()) query.search = search.value.trim()
    const res = await fetchArticles(query)
    articles.value = res.items
    totalPages.value = res.total_pages
  } finally {
    loading.value = false
  }
}

const reload = () => {
  page.value = 1
  load()
}

onMounted(load)
</script>
