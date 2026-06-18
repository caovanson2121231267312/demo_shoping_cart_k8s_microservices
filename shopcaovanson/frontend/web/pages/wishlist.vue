<template>
  <v-container class="page-container py-6">
    <h1 class="text-h4 font-weight-bold mb-2">
      <v-icon color="error" class="mr-2">mdi-heart</v-icon>
      Sản phẩm yêu thích
    </h1>
    <p class="text-body-2 text-grey mb-6">{{ wishlist.count.value }} sản phẩm đã lưu</p>

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
</template>

<script setup lang="ts">
import type { ReviewSummary } from '~/composables/useReviewSummary'

definePageMeta({ layout: 'default' })

const wishlist = useWishlist()
const { getSummary } = useReviewSummary()
const summaries = ref<Record<string, ReviewSummary>>({})

onMounted(async () => {
  useWishlistStore().hydrate()
  for (const p of wishlist.items.value) {
    summaries.value[p.id] = await getSummary(p.id)
  }
})

const clearAll = () => {
  if (confirm('Xóa toàn bộ danh sách yêu thích?')) {
    wishlist.clear()
    summaries.value = {}
  }
}
</script>
