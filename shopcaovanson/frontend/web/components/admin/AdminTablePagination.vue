<template>
  <div class="admin-table-pagination">
    <div class="admin-table-pagination__left">
      <span class="admin-table-pagination__label">Hiển thị</span>
      <v-select
        :model-value="itemsPerPage"
        :items="itemsPerPageOptions"
        item-title="title"
        item-value="value"
        density="compact"
        hide-details
        variant="outlined"
        class="admin-table-pagination__select"
        @update:model-value="onPerPageChange"
      />
      <span class="admin-table-pagination__range">{{ rangeText }}</span>
    </div>
    <v-pagination
      v-if="totalPages > 1"
      :model-value="page"
      :length="totalPages"
      :total-visible="paginationVisible"
      density="comfortable"
      active-color="primary"
      class="admin-table-pagination__pages"
      @update:model-value="onPageChange"
    />
    <span v-else class="admin-table-pagination__single-page">Trang 1</span>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  totalPages: number
  rangeText: string
  itemsPerPageOptions: { value: number; title: string }[]
  paginationVisible?: number
}>(), {
  paginationVisible: 7,
})

const emit = defineEmits<{
  change: []
}>()

const page = defineModel<number>('page', { required: true })
const itemsPerPage = defineModel<number>('itemsPerPage', { required: true })

function onPageChange(value: number) {
  if (page.value === value) return
  page.value = value
  emit('change')
}

function onPerPageChange(value: number) {
  if (itemsPerPage.value === value && page.value === 1) return
  itemsPerPage.value = value
  page.value = 1
  emit('change')
}
</script>
