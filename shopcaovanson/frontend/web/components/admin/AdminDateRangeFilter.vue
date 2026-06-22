<template>
  <v-text-field
    :model-value="from"
    type="date"
    label="Từ ngày"
    density="compact"
    variant="outlined"
    hide-details
    clearable
    class="admin-date-field"
    @update:model-value="emit('update:from', $event || '')"
  />
  <v-text-field
    :model-value="to"
    type="date"
    label="Đến ngày"
    density="compact"
    variant="outlined"
    hide-details
    clearable
    class="admin-date-field"
    :min="from || undefined"
    @update:model-value="emit('update:to', $event || '')"
  />
</template>

<script setup lang="ts">
defineProps<{
  from: string
  to: string
}>()

const emit = defineEmits<{
  'update:from': [value: string]
  'update:to': [value: string]
}>()
</script>

<style scoped>
.admin-date-field {
  flex: 0 1 160px;
  min-width: 140px;
  max-width: 180px;
}
</style>
