<template>
  <div class="admin-table-pagination admin-table-pagination--cursor">
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
    <div class="admin-table-pagination__cursor-nav">
      <v-btn
        variant="outlined"
        size="small"
        :disabled="!hasPrev || loading"
        prepend-icon="mdi-chevron-left"
        @click="emit('prev')"
      >
        Trước
      </v-btn>
      <span class="admin-table-pagination__cursor-page">Trang {{ pageNumber }}</span>
      <v-btn
        variant="outlined"
        size="small"
        :disabled="!hasMore || loading"
        append-icon="mdi-chevron-right"
        @click="emit('next')"
      >
        Sau
      </v-btn>
    </div>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  rangeText: string
  itemsPerPageOptions: { value: number; title: string }[]
  hasPrev: boolean
  hasMore: boolean
  pageNumber: number
  loading?: boolean
}>(), {
  loading: false,
})

const emit = defineEmits<{
  prev: []
  next: []
  change: []
}>()

const itemsPerPage = defineModel<number>('itemsPerPage', { required: true })

function onPerPageChange(value: number) {
  if (itemsPerPage.value === value) return
  itemsPerPage.value = value
  emit('change')
}
</script>

<style scoped>
.admin-table-pagination--cursor {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 16px;
}

.admin-table-pagination__cursor-nav {
  display: flex;
  align-items: center;
  gap: 12px;
}

.admin-table-pagination__cursor-page {
  font-size: 13px;
  color: var(--color-text-muted, #64748b);
  min-width: 72px;
  text-align: center;
}
</style>
