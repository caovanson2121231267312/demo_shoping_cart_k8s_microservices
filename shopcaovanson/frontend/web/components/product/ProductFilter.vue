<template>
  <v-card>
    <v-card-title class="text-subtitle-1">Bộ lọc</v-card-title>
    <v-card-text>
      <v-select
        v-model="localCategory"
        :items="categoryItems"
        item-title="name"
        item-value="slug"
        label="Danh mục"
        clearable
        density="compact"
        variant="outlined"
        hide-details
        class="mb-4"
      />

      <div class="text-body-2 mb-2">Khoảng giá (VND)</div>
      <v-range-slider
        v-model="priceRange"
        :min="0"
        :max="maxPrice"
        :step="100000"
        color="primary"
        hide-details
        class="mb-2"
      />
      <div class="text-caption text-grey mb-4">
        {{ formatVND(priceRange[0]) }} – {{ formatVND(priceRange[1]) }}
      </div>

      <v-select
        v-model="localSort"
        :items="sortOptions"
        label="Sắp xếp"
        density="compact"
        variant="outlined"
        hide-details
        class="mb-4"
      />

      <v-select
        v-model="localRating"
        :items="ratingOptions"
        label="Đánh giá tối thiểu"
        clearable
        density="compact"
        variant="outlined"
        hide-details
        class="mb-4"
      />

      <v-btn block color="primary" @click="applyFilters">Áp dụng</v-btn>
      <v-btn block variant="text" class="mt-2" @click="resetFilters">Xóa bộ lọc</v-btn>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import type { Category } from '~/types'

const props = defineProps<{
  categories: Category[]
  category?: string
  minPrice?: number
  maxPrice?: number
  sort?: string
}>()

const emit = defineEmits<{
  filter: [payload: { category?: string; min_price?: number; max_price?: number; sort?: string }]
}>()

const { formatVND } = useFormat()

const maxPrice = 50_000_000
const localCategory = ref(props.category || null)
const localSort = ref(props.sort || 'newest')
const localRating = ref<number | null>(null)
const priceRange = ref<[number, number]>([
  props.minPrice ?? 0,
  props.maxPrice ?? maxPrice,
])

const sortOptions = [
  { title: 'Mới nhất', value: 'newest' },
  { title: 'Giá tăng dần', value: 'price_asc' },
  { title: 'Giá giảm dần', value: 'price_desc' },
]

const ratingOptions = [
  { title: '4 sao trở lên', value: 4 },
  { title: '3 sao trở lên', value: 3 },
  { title: '2 sao trở lên', value: 2 },
]

const categoryItems = computed(() => {
  const flat: Category[] = []
  for (const cat of props.categories) {
    flat.push(cat)
    if (cat.children) {
      flat.push(...cat.children)
    }
  }
  return flat
})

const applyFilters = () => {
  emit('filter', {
    category: localCategory.value || undefined,
    min_price: priceRange.value[0] > 0 ? priceRange.value[0] : undefined,
    max_price: priceRange.value[1] < maxPrice ? priceRange.value[1] : undefined,
    sort: localSort.value as 'newest' | 'price_asc' | 'price_desc',
  })
}

const resetFilters = () => {
  localCategory.value = null
  localSort.value = 'newest'
  localRating.value = null
  priceRange.value = [0, maxPrice]
  emit('filter', { sort: 'newest' })
}

watch(
  () => [props.category, props.minPrice, props.maxPrice, props.sort],
  () => {
    localCategory.value = props.category || null
    localSort.value = props.sort || 'newest'
    priceRange.value = [props.minPrice ?? 0, props.maxPrice ?? maxPrice]
  },
)
</script>
