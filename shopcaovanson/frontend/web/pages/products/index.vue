<template>
  <div>
    <PageBanner
      title="Sản phẩm"
      subtitle="Khám phá hàng ngàn sản phẩm chất lượng với giá tốt nhất"
      :breadcrumbs="[{ label: 'Sản phẩm' }]"
      compact
    />

    <v-container class="page-container py-6">
    <v-row>
      <v-col cols="12" md="3">
        <ProductFilter
          :categories="categories"
          :category="filters.category"
          :min-price="filters.min_price"
          :max-price="filters.max_price"
          :sort="filters.sort"
          @filter="onFilter"
        />
      </v-col>

      <v-col cols="12" md="9">
        <div class="mb-4">
          <ProductSearch v-model="search" @search="onSearch" />
        </div>

        <ProductGrid :products="products" :loading="loading" />

        <div v-if="totalPages > 1" class="d-flex justify-center mt-6">
          <v-pagination
            v-model="page"
            :length="totalPages"
            :total-visible="7"
            @update:model-value="loadProducts"
          />
        </div>
      </v-col>
    </v-row>
    </v-container>
  </div>
</template>

<script setup lang="ts">
import type { Category, Product, ProductFilters } from '~/types'

definePageMeta({ layout: 'default' })

const route = useRoute()
const router = useRouter()
const { fetchProducts, fetchCategories } = useProducts()

const categories = ref<Category[]>([])
const products = ref<Product[]>([])
const loading = ref(true)
const page = ref(1)
const totalPages = ref(1)
const search = ref('')

const filters = ref<ProductFilters>({
  sort: 'newest',
  limit: 12,
})

const loadProducts = async () => {
  loading.value = true
  try {
    const result = await fetchProducts({
      ...filters.value,
      page: page.value,
      search: search.value || undefined,
    })
    products.value = result.items
    totalPages.value = result.total_pages
  } finally {
    loading.value = false
  }
}

const onFilter = (payload: ProductFilters) => {
  filters.value = { ...filters.value, ...payload }
  page.value = 1
  loadProducts()
}

const onSearch = () => {
  page.value = 1
  loadProducts()
}

onMounted(async () => {
  if (route.query.category) {
    filters.value.category = String(route.query.category)
  }
  if (route.query.search) {
    search.value = String(route.query.search)
  }
  categories.value = await fetchCategories()
  await loadProducts()
})

watch(page, () => {
  router.replace({ query: { ...route.query, page: String(page.value) } })
})
</script>
