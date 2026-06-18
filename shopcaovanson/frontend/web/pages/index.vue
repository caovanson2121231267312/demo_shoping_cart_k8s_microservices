<template>
  <div>
    <HeroCarousel />

    <v-container class="page-container">
      <CategoryShowcase v-if="categories.length" :categories="categories" />

      <FlashSaleSection
        v-if="saleProducts.length"
        :products="saleProducts"
        :summaries="summaries"
      />

      <section class="mb-10">
        <div class="d-flex align-center justify-space-between mb-4">
          <h2 class="text-h5 font-weight-bold">Sản phẩm nổi bật</h2>
          <v-btn variant="text" color="primary" to="/products">Xem tất cả</v-btn>
        </div>
        <ProductGrid :products="featuredProducts" :loading="loadingProducts" :summaries="summaries" />
      </section>

      <section v-if="newProducts.length" class="mb-8">
        <div class="d-flex align-center justify-space-between mb-4">
          <h2 class="text-h5 font-weight-bold">Hàng mới về</h2>
          <v-btn variant="text" color="primary" to="/products?sort=newest">Xem thêm</v-btn>
        </div>
        <v-row>
          <v-col v-for="product in newProducts" :key="product.id" cols="6" sm="4" md="3">
            <ProductCard :product="product" :summary="summaries[product.id]" />
          </v-col>
        </v-row>
      </section>

      <v-row class="mb-8">
        <v-col cols="12" md="4">
          <v-card class="pa-6 text-center h-100" variant="tonal" color="primary" rounded="lg">
            <v-icon size="48" class="mb-3">mdi-truck-fast</v-icon>
            <div class="text-h6 font-weight-bold">Giao hàng nhanh</div>
            <p class="text-body-2 mt-2 mb-0">Freeship đơn từ 500K, giao 2h nội thành</p>
          </v-card>
        </v-col>
        <v-col cols="12" md="4">
          <v-card class="pa-6 text-center h-100" variant="tonal" color="success" rounded="lg">
            <v-icon size="48" class="mb-3">mdi-shield-check</v-icon>
            <div class="text-h6 font-weight-bold">Chính hãng 100%</div>
            <p class="text-body-2 mt-2 mb-0">Hoàn tiền nếu phát hiện hàng giả</p>
          </v-card>
        </v-col>
        <v-col cols="12" md="4">
          <v-card class="pa-6 text-center h-100" variant="tonal" color="secondary" rounded="lg">
            <v-icon size="48" class="mb-3">mdi-headset</v-icon>
            <div class="text-h6 font-weight-bold">Hỗ trợ 24/7</div>
            <p class="text-body-2 mt-2 mb-0">Chat trực tuyến & hotline tận tâm</p>
          </v-card>
        </v-col>
      </v-row>
    </v-container>
  </div>
</template>

<script setup lang="ts">
import type { Category, Product } from '~/types'
import type { ReviewSummary } from '~/composables/useReviewSummary'

definePageMeta({ layout: 'default' })

const { fetchProducts, fetchCategories } = useProducts()
const { getSummary } = useReviewSummary()

const categories = ref<Category[]>([])
const featuredProducts = ref<Product[]>([])
const saleProducts = ref<Product[]>([])
const newProducts = ref<Product[]>([])
const loadingProducts = ref(true)
const summaries = ref<Record<string, ReviewSummary>>({})

const loadSummaries = async (products: Product[]) => {
  await Promise.all(
    products.map(async (p) => {
      if (!summaries.value[p.id]) {
        summaries.value[p.id] = await getSummary(p.id)
      }
    }),
  )
}

onMounted(async () => {
  try {
    const [cats, featured, sale, newest] = await Promise.all([
      fetchCategories(),
      fetchProducts({ limit: 8, sort: 'newest' }),
      fetchProducts({ limit: 6, sort: 'price_desc' }),
      fetchProducts({ limit: 4, sort: 'newest' }),
    ])
    categories.value = cats
    featuredProducts.value = featured.items
    saleProducts.value = sale.items.filter((p) => p.sale_price != null)
    if (!saleProducts.value.length) saleProducts.value = sale.items.slice(0, 6)
    newProducts.value = newest.items

    const all = [...featured.items, ...saleProducts.value, ...newProducts.value]
    const unique = [...new Map(all.map((p) => [p.id, p])).values()]
    await loadSummaries(unique)
  } finally {
    loadingProducts.value = false
  }
})
</script>
