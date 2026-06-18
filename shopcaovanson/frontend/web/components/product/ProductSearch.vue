<template>
  <v-text-field
    v-model="query"
    :placeholder="placeholder"
    prepend-inner-icon="mdi-magnify"
    density="compact"
    variant="solo-filled"
    flat
    hide-details
    clearable
    @keyup.enter="emitSearch"
    @click:clear="emitSearch"
  />
</template>

<script setup lang="ts">
import { useDebounceFn } from '@vueuse/core'

const props = withDefaults(
  defineProps<{
    modelValue?: string
    placeholder?: string
    debounce?: number
  }>(),
  {
    modelValue: '',
    placeholder: 'Tìm kiếm sản phẩm...',
    debounce: 400,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
  search: [value: string]
}>()

const query = ref(props.modelValue)

watch(
  () => props.modelValue,
  (val) => {
    query.value = val
  },
)

const emitSearch = useDebounceFn(() => {
  emit('update:modelValue', query.value)
  emit('search', query.value)
}, props.debounce)

watch(query, () => {
  emitSearch()
})
</script>
