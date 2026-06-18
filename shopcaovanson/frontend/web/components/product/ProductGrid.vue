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
          <ProductCard :product="product" :summary="summaries?.[product.id]" />
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
}>()

const { getSummary } = useReviewSummary()
const localSummaries = ref<Record<string, ReviewSummary>>({})

const summaries = computed(() => ({ ...localSummaries.value, ...props.summaries }))

watch(
  () => props.products,
  async (list) => {
    if (!list?.length) return
    await Promise.all(
      list.slice(0, 24).map(async (p) => {
        if (!summaries.value[p.id]) {
          localSummaries.value[p.id] = await getSummary(p.id)
        }
      }),
    )
  },
  { immediate: true },
)
</script>
