<template>
  <div>
    <PromoBar />
    <AppHeader @toggle-sidebar="drawer = !drawer" />
    <CategoryNavBar v-if="categories.length" :categories="categories" />
    <AppSidebar v-model="drawer" />
    <v-main class="bg-grey-lighten-5">
      <slot />
    </v-main>
    <AppFooter />
    <CartDrawer />
    <AppSnackbar />
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
