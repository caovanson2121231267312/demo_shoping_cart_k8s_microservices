<template>
  <div class="storefront-layout">
    <div class="storefront-layout__header">
      <PromoBar />
      <AppHeader />
      <CategoryNavBar v-if="categories.length" :categories="categories" />
    </div>

    <v-main class="storefront-layout__main d-flex align-center">
      <v-container class="auth-container py-8">
        <slot />
      </v-container>
    </v-main>

    <PaymentStrip />
    <AppFooter />
    <AppSnackbar />
  </div>
</template>

<script setup lang="ts">
import type { Category } from '~/types'

const categories = ref<Category[]>([])
const { fetchCategories } = useProducts()

onMounted(async () => {
  try {
    categories.value = await fetchCategories()
  } catch {
    categories.value = []
  }
})
</script>

<style scoped>
.storefront-layout__header {
  position: sticky;
  top: 0;
  z-index: 100;
}

.storefront-layout__main {
  background: linear-gradient(180deg, #e3f2fd 0%, #f5f5f5 40%);
  min-height: calc(100vh - 200px);
}

.auth-container {
  max-width: 480px;
}
</style>
