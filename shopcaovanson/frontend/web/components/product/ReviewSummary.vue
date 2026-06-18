<template>
  <v-card variant="outlined" class="pa-4 mb-6">
    <v-row align="center">
      <v-col cols="12" md="4" class="text-center">
        <div class="text-h2 font-weight-bold text-primary">{{ summary.average.toFixed(1) }}</div>
        <StarRating :value="summary.average" size="md" class="justify-center" />
        <div class="text-caption text-grey mt-1">{{ summary.total }} đánh giá</div>
      </v-col>
      <v-col cols="12" md="8">
        <div v-for="star in [5, 4, 3, 2, 1]" :key="star" class="d-flex align-center ga-2 mb-1">
          <span class="text-caption" style="width: 48px">{{ star }} sao</span>
          <v-progress-linear
            :model-value="barPercent(star)"
            color="amber-darken-2"
            height="8"
            rounded
            class="flex-grow-1"
          />
          <span class="text-caption text-grey" style="width: 32px">{{ summary.distribution[star] || 0 }}</span>
        </div>
      </v-col>
    </v-row>
  </v-card>
</template>

<script setup lang="ts">
import type { ReviewSummary } from '~/composables/useReviewSummary'

const props = defineProps<{ summary: ReviewSummary }>()

const barPercent = (star: number) => {
  if (!props.summary.total) return 0
  return ((props.summary.distribution[star] || 0) / props.summary.total) * 100
}
</script>
