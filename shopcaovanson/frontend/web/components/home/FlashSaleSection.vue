<template>

  <section class="flash-sale">

    <div class="flash-sale__header d-flex align-center justify-space-between mb-4 flex-wrap ga-2">

      <div class="d-flex align-center ga-2">

        <v-icon color="error" size="28">mdi-flash</v-icon>

        <h2 class="section-title flash-sale__title">Flash Sale</h2>

        <div class="flash-countdown d-flex align-center ga-1 ml-2">

          <span class="flash-countdown__box">{{ pad(hours) }}</span>

          <span class="flash-countdown__sep">:</span>

          <span class="flash-countdown__box">{{ pad(minutes) }}</span>

          <span class="flash-countdown__sep">:</span>

          <span class="flash-countdown__box">{{ pad(seconds) }}</span>

        </div>

      </div>

      <v-btn variant="text" color="error" to="/products?sort=price_desc" class="text-none font-weight-bold">

        Xem tất cả

      </v-btn>

    </div>

    <v-row>

      <v-col v-for="(product, idx) in products" :key="product.id" cols="6" sm="4" md="3" lg="2">

        <div class="animate-fade-in" :style="{ animationDelay: `${idx * 60}ms` }">

          <ProductCard :product="product" :summary="summaries[product.id]" />

        </div>

      </v-col>

    </v-row>

  </section>

</template>



<script setup lang="ts">

import type { Product } from '~/types'

import type { ReviewSummary } from '~/composables/useReviewSummary'



defineProps<{

  products: Product[]

  summaries?: Record<string, ReviewSummary>

}>()



const endTime = Date.now() + 4 * 60 * 60 * 1000

const remaining = ref(endTime - Date.now())



let timer: ReturnType<typeof setInterval>

onMounted(() => {

  timer = setInterval(() => {

    remaining.value = Math.max(0, endTime - Date.now())

  }, 1000)

})

onUnmounted(() => clearInterval(timer))



const hours = computed(() => Math.floor(remaining.value / 3600000))

const minutes = computed(() => Math.floor((remaining.value % 3600000) / 60000))

const seconds = computed(() => Math.floor((remaining.value % 60000) / 1000))

const pad = (n: number) => String(n).padStart(2, '0')

</script>



<style scoped>

.flash-sale {

  background: var(--color-surface);

  border-radius: var(--radius-lg);

  padding: 20px;

  box-shadow: var(--shadow-sale);

  border: 1px solid var(--color-sale-border);

}



.flash-sale__title {

  color: var(--color-sale-dark);

}



.flash-countdown__box {

  background: var(--color-sale);

  color: var(--color-text-inverse);

  padding: 4px 8px;

  border-radius: var(--radius-sm);

  font-size: 14px;

  font-weight: 700;

  min-width: 32px;

  text-align: center;

}



.flash-countdown__sep {

  color: var(--color-sale);

  font-weight: 700;

}

</style>

