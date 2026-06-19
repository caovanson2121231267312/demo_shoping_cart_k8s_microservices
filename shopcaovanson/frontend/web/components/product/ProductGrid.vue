<template>
  <div>
    <LoadingSpinner v-if="loading" />
    <EmptyState
      v-else-if="!products.length"
      icon="mdi-package-variant-closed"
      title="Không có sản phẩm"
      description="Thử thay đổi bộ lọc hoặc từ khóa tìm kiếm."
    />
    <v-row v-else>
      <v-col
        v-for="(product, idx) in products"
        :key="product.id"
        cols="6"
        sm="4"
        md="3"
      >
        <div class="animate-fade-in" :style="{ animationDelay: `${idx * 40}ms` }">
          <ProductCard :product="product" :summary="mergedSummaries[product.id]" />
        </div>
      </v-col>
    </v-row>
  </div>
</template>

<script setup lang="ts">
import type { Product } from '~/types'
import type { ReviewSummary } from '~/composables/useReviewSummary'

const props = defineProps<{
  products: Product[]
  loading?: boolean
  summaries?: Record<string, ReviewSummary>
  fetchSummaries?: boolean
}>()

const { getSummariesBatch } = useReviewSummary()
const localSummaries = ref<Record<string, ReviewSummary>>({})

const shouldFetchSummaries = computed(() => props.fetchSummaries !== false)

const mergedSummaries = computed(() => ({ ...localSummaries.value, ...props.summaries }))

watch(
  () => props.products,
  async (list) => {
    if (!shouldFetchSummaries.value || !list?.length) {
      return
    }
    const missing = list
      .slice(0, 24)
      .map((p) => p.id)
      .filter((id) => !mergedSummaries.value[id])
    if (!missing.length) {
      return
    }
    const batch = await getSummariesBatch(missing)
    localSummaries.value = { ...localSummaries.value, ...batch }
  },
  { immediate: true },
)
</script>
