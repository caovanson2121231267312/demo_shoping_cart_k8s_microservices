<template>
  <div class="star-rating d-inline-flex align-center" :class="{ 'star-rating--sm': size === 'sm' }">
    <v-icon
      v-for="i in 5"
      :key="i"
      :icon="starIcon(i)"
      :color="i <= Math.round(value) ? color : 'grey-lighten-1'"
      :size="size === 'sm' ? 14 : 18"
      class="star-rating__icon"
    />
    <span v-if="showValue" class="star-rating__value text-caption ml-1">{{ value.toFixed(1) }}</span>
    <span v-if="count != null" class="star-rating__count text-caption text-grey ml-1">({{ count }})</span>
  </div>
</template>

<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    value: number
    count?: number
    size?: 'sm' | 'md'
    color?: string
    showValue?: boolean
  }>(),
  { size: 'md', color: 'amber-darken-2', showValue: false },
)

const starIcon = (i: number) => {
  const rounded = Math.round(props.value)
  if (i <= rounded) return 'mdi-star'
  if (i - 0.5 <= props.value) return 'mdi-star-half-full'
  return 'mdi-star-outline'
}
</script>

<style scoped>
.star-rating__icon {
  margin-right: -2px;
}
.star-rating--sm .star-rating__icon {
  margin-right: -3px;
}
</style>
