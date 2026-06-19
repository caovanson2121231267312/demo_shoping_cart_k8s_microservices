<template>
  <div>
    <PageBanner
      title="Sản phẩm yêu thích"
      :subtitle="`${wishlist.count.value} sản phẩm đã lưu`"
      :breadcrumbs="[{ label: 'Yêu thích' }]"
      compact
    />
    <v-container class="page-container py-6">
    <EmptyState
      v-if="!wishlist.items.value.length"
      icon="mdi-heart-outline"
      title="Danh sách trống"
      description="Nhấn biểu tượng trái tim trên sản phẩm để lưu vào đây."
    >
      <v-btn color="primary" to="/products" class="mt-4">Khám phá sản phẩm</v-btn>
    </EmptyState>

    <v-row v-else>
      <v-col v-for="product in wishlist.items.value" :key="product.id" cols="6" sm="4" md="3">
        <ProductCard :product="product" :summary="summaries[product.id]" />
      </v-col>
    </v-row>

    <div v-if="wishlist.items.value.length" class="text-center mt-6">
      <v-btn variant="outlined" color="error" @click="clearAll">Xóa tất cả</v-btn>
    </div>
    </v-container>
  </div>
</template>

<script setup lang="ts">
import type { ReviewSummary } from '~/composables/useReviewSummary'

definePageMeta({ layout: 'default' })

const wishlist = useWishlist()
const { getSummariesBatch } = useReviewSummary()
const summaries = ref<Record<string, ReviewSummary>>({})

onMounted(async () => {
  useWishlistStore().hydrate()
  const ids = wishlist.items.value.map((p) => p.id)
  if (ids.length) {
    summaries.value = await getSummariesBatch(ids)
  }
})

const clearAll = () => {
  if (confirm('Xóa toàn bộ danh sách yêu thích?')) {
    wishlist.clear()
    summaries.value = {}
  }
}
</script>
