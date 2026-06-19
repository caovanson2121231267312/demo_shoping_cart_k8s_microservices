<template>
  <div class="storefront-layout">
    <div class="storefront-layout__header">
      <PromoBar />
      <AppHeader @toggle-sidebar="drawer = !drawer" />
      <CategoryNavBar v-if="categories.length" :categories="categories" />
    </div>

    <AppSidebar v-model="drawer" />

    <v-main class="storefront-layout__main">
      <slot />
    </v-main>

    <AppFooter />
    <CartDrawer />
    <AppSnackbar />
    <BackToTop />
  </div>
</template>

<script setup lang="ts">
import type { Category } from '~/types'

const drawer = ref(false)
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
.storefront-layout {
  display: flex;
  flex-direction: column;
  min-height: 100%;
  width: 100%;
}

.storefront-layout__header {
  position: sticky;
  top: 0;
  z-index: 100;
  flex-shrink: 0;
}

.storefront-layout__main {
  flex: 1 1 auto;
  background: var(--color-bg);
  min-height: 0;
  padding: 0 !important;
}
</style>
