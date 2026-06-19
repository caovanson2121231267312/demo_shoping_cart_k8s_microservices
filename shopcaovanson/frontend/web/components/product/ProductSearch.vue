<template>
  <div class="product-search" :class="{ 'product-search--with-btn': showButton }">
    <v-text-field
      v-model="query"
      :placeholder="placeholder"
      prepend-inner-icon="mdi-magnify"
      density="compact"
      :variant="variant"
      :rounded="showButton ? '0' : rounded"
      hide-details
      clearable
      :bg-color="bgColor"
      class="product-search__field"
      @keyup.enter="emitSearchNow"
      @click:clear="emitSearchNow"
    />
    <v-btn
      v-if="showButton"
      color="primary"
      class="product-search__btn text-none font-weight-bold"
      rounded="0"
      elevation="0"
      @click="emitSearchNow"
    >
      <v-icon start>mdi-magnify</v-icon>
      <span class="d-none d-sm-inline">Tìm kiếm</span>
    </v-btn>
  </div>
</template>



<script setup lang="ts">

import { useDebounceFn } from '@vueuse/core'



const props = withDefaults(

  defineProps<{

    modelValue?: string

    placeholder?: string

    debounce?: number

    showButton?: boolean

    variant?: 'solo' | 'outlined' | 'plain' | 'filled' | 'underlined' | 'solo-inverted' | 'solo-filled'

    rounded?: string | boolean

    bgColor?: string

  }>(),

  {

    modelValue: '',

    placeholder: 'Tìm kiếm sản phẩm, thương hiệu, danh mục...',

    debounce: 400,

    showButton: false,

    variant: 'solo',

    rounded: 'lg',

    bgColor: 'white',

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



const emitSearchNow = () => {

  emit('update:modelValue', query.value)

  emit('search', query.value)

}



watch(query, () => {

  if (!props.showButton) {

    emitSearch()

  }

})

</script>



<style scoped>
.product-search {
  width: 100%;
}

.product-search--with-btn {
  --search-height: 44px;
  display: flex;
  align-items: center;
  height: var(--search-height);
  border: 2px solid var(--color-primary);
  border-radius: var(--radius-sm);
  overflow: hidden;
  background: var(--color-surface);
  box-shadow: var(--shadow-search);
}

.product-search--with-btn .product-search__field {
  flex: 1;
  min-width: 0;
  height: 100%;
}

.product-search--with-btn :deep(.v-input) {
  height: 100%;
}

.product-search--with-btn :deep(.v-input__control),
.product-search--with-btn :deep(.v-field) {
  height: 100% !important;
  min-height: var(--search-height) !important;
  border-radius: 0 !important;
  box-shadow: none !important;
}

.product-search--with-btn :deep(.v-field__outline) {
  display: none;
}

.product-search--with-btn :deep(.v-field__overlay) {
  opacity: 0;
}

.product-search--with-btn :deep(.v-field__input) {
  min-height: var(--search-height);
  padding-top: 0;
  padding-bottom: 0;
}

.product-search__btn {
  flex-shrink: 0;
  height: 100% !important;
  min-height: var(--search-height) !important;
  border-radius: 0 !important;
  min-width: 52px;
  padding-left: 16px;
  padding-right: 16px;
  margin: 0;
}

.product-search:not(.product-search--with-btn) :deep(.v-field) {
  box-shadow: inset 0 1px 3px var(--color-border-light);
}
</style>

