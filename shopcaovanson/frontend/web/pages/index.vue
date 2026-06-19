<template>
  <div class="home-page">
    <!-- Hero -->
    <section class="home-hero">
      <HeroCarousel />
    </section>

    <TrustBar />
    <HomeQuickLinks />
    <TrendingTags />

    <PromoBannerGrid />
    <CouponStrip />

    <!-- Danh mục -->
    <v-container class="page-container home-section">
      <CategoryShowcase v-if="categories.length" :categories="categories" />
    </v-container>

    <!-- Flash Sale -->
    <section v-if="saleProducts.length" class="home-flash">
      <v-container class="page-container">
        <FlashSaleSection :products="saleProducts" :summaries="summaries" />
      </v-container>
    </section>

    <v-container class="page-container home-section">
      <MegaDealBanner />
    </v-container>

    <!-- Sản phẩm nổi bật -->
    <v-container class="page-container home-section">
      <section class="home-block">
        <div class="section-header">
          <div class="d-flex align-center ga-2">
            <v-icon color="primary" size="28">mdi-star-circle</v-icon>
            <div>
              <h2 class="section-title">Sản phẩm nổi bật</h2>
              <p class="section-subtitle">Được yêu thích nhất tuần này</p>
            </div>
          </div>
          <v-btn variant="outlined" color="primary" to="/products" rounded="lg" class="text-none">
            Xem tất cả
            <v-icon end>mdi-arrow-right</v-icon>
          </v-btn>
        </div>
        <ProductGrid
          :products="featuredProducts"
          :loading="loadingProducts"
          :summaries="summaries"
          :fetch-summaries="false"
        />
      </section>

      <!-- Giá tốt -->
      <section v-if="budgetProducts.length" class="home-block">
        <div class="section-header">
          <div class="d-flex align-center ga-2">
            <v-icon color="success" size="28">mdi-tag-heart</v-icon>
            <div>
              <h2 class="section-title">Giá tốt mỗi ngày</h2>
              <p class="section-subtitle">Săn deal giá rẻ — chất lượng đảm bảo</p>
            </div>
          </div>
          <v-btn variant="text" color="primary" to="/products?sort=price_asc" class="text-none">
            Xem thêm
          </v-btn>
        </div>
        <v-row>
          <v-col v-for="(product, idx) in budgetProducts" :key="product.id" cols="6" sm="4" md="3" lg="2">
            <div class="animate-fade-in" :style="{ animationDelay: `${idx * 50}ms` }">
              <ProductCard :product="product" :summary="summaries[product.id]" />
            </div>
          </v-col>
        </v-row>
      </section>

      <!-- Hàng mới -->
      <section v-if="newProducts.length" class="home-block">
        <div class="section-header">
          <div class="d-flex align-center ga-2">
            <v-icon color="secondary" size="28">mdi-new-box</v-icon>
            <div>
              <h2 class="section-title">Hàng mới về</h2>
              <p class="section-subtitle">Cập nhật mỗi ngày — xu hướng mới nhất</p>
            </div>
          </div>
          <v-btn variant="text" color="primary" to="/products?sort=newest" class="text-none">
            Xem thêm
          </v-btn>
        </div>
        <v-row>
          <v-col v-for="(product, idx) in newProducts" :key="product.id" cols="6" sm="4" md="3">
            <div class="animate-fade-in" :style="{ animationDelay: `${idx * 50}ms` }">
              <ProductCard :product="product" :summary="summaries[product.id]" />
            </div>
          </v-col>
        </v-row>
      </section>
    </v-container>

    <!-- Kệ hàng theo danh mục -->
    <section v-if="categoryShelves.length" class="home-shelves">
      <v-container class="page-container">
        <CategoryShelf
          v-for="(shelf, idx) in categoryShelves"
          :key="shelf.category.id"
          :category="shelf.category"
          :products="shelf.products"
          :color-index="idx"
          :summaries="summaries"
        />
      </v-container>
    </section>

    <BrandShowcase />
    <BlogSection />
    <ServiceHighlights />

    <!-- Cao cấp -->
    <v-container v-if="premiumProducts.length" class="page-container home-section">
      <section class="home-block home-block--premium">
        <div class="section-header">
          <div class="d-flex align-center ga-2">
            <v-icon color="warning" size="28">mdi-crown</v-icon>
            <div>
              <h2 class="section-title">Sản phẩm cao cấp</h2>
              <p class="section-subtitle">Hàng premium — trải nghiệm đỉnh cao</p>
            </div>
          </div>
          <v-btn variant="text" color="primary" to="/products?sort=price_desc" class="text-none">
            Xem thêm
          </v-btn>
        </div>
        <v-row>
          <v-col v-for="(product, idx) in premiumProducts" :key="product.id" cols="6" sm="4" md="3">
            <div class="animate-fade-in" :style="{ animationDelay: `${idx * 50}ms` }">
              <ProductCard :product="product" :summary="summaries[product.id]" />
            </div>
          </v-col>
        </v-row>
      </section>
    </v-container>

    <TestimonialsSection />
    <PaymentStrip />
    <NewsletterSection />
  </div>
</template>

<script setup lang="ts">
import type { Category, Product } from '~/types'
import type { ReviewSummary } from '~/composables/useReviewSummary'

definePageMeta({ layout: 'default' })

const { fetchProducts, fetchCategories } = useProducts()
const { getSummariesBatch } = useReviewSummary()

interface CategoryShelf {
  category: Category
  products: Product[]
}

const categories = ref<Category[]>([])
const featuredProducts = ref<Product[]>([])
const saleProducts = ref<Product[]>([])
const newProducts = ref<Product[]>([])
const budgetProducts = ref<Product[]>([])
const premiumProducts = ref<Product[]>([])
const categoryShelves = ref<CategoryShelf[]>([])
const loadingProducts = ref(true)
const summaries = ref<Record<string, ReviewSummary>>({})

const loadSummaries = async (products: Product[]) => {
  const ids = products.map((p) => p.id)
  if (!ids.length) {
    return
  }
  const batch = await getSummariesBatch(ids)
  summaries.value = { ...summaries.value, ...batch }
}

onMounted(async () => {
  try {
    const cats = await fetchCategories()
    categories.value = cats

    const topCats = cats.filter((c) => !c.parent_id).slice(0, 4)
    const shelfPromises = topCats.map((cat) =>
      fetchProducts({ category: cat.slug, limit: 4 }).catch(() => ({ items: [] as Product[] })),
    )

    const [featured, sale, budget, premium, ...shelfResults] = await Promise.all([
      fetchProducts({ limit: 8, sort: 'newest' }),
      fetchProducts({ limit: 8, sort: 'price_desc' }),
      fetchProducts({ limit: 6, sort: 'price_asc' }),
      fetchProducts({ limit: 4, sort: 'price_desc' }),
      ...shelfPromises,
    ])

    featuredProducts.value = featured.items
    saleProducts.value = sale.items.filter((p) => p.sale_price != null)
    if (!saleProducts.value.length) saleProducts.value = sale.items.slice(0, 8)
    newProducts.value = featured.items.slice(0, 8)
    budgetProducts.value = budget.items
    premiumProducts.value = premium.items.slice(0, 4)

    categoryShelves.value = topCats
      .map((cat, i) => ({
        category: cat,
        products: shelfResults[i]?.items ?? [],
      }))
      .filter((s) => s.products.length > 0)

    const all = [
      ...featured.items,
      ...saleProducts.value,
      ...newProducts.value,
      ...budget.items,
      ...premium.items,
      ...categoryShelves.value.flatMap((s) => s.products),
    ]
    const unique = [...new Map(all.map((p) => [p.id, p])).values()]
    await loadSummaries(unique)
  } finally {
    loadingProducts.value = false
  }
})
</script>

<style scoped>
.home-hero {
  background: var(--color-surface);
}

.home-section {
  padding-top: 28px;
  padding-bottom: 8px;
}

.home-block {
  margin-bottom: 44px;
}

.home-block--premium {
  background: var(--color-surface);
  border-radius: var(--radius-lg);
  padding: 24px;
  border: 1px solid var(--color-border-light);
  box-shadow: var(--shadow-sm);
}

.home-flash {
  background: var(--gradient-sale-section);
  padding: 28px 0 12px;
  border-top: 1px solid var(--color-sale-border);
  border-bottom: 1px solid var(--color-sale-border);
}

.home-shelves {
  background: var(--color-surface-muted);
  padding: 32px 0 8px;
  border-top: 1px solid var(--color-border);
}

.section-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}
</style>
