<template>
  <v-menu
    v-model="fromMenu"
    :close-on-content-click="false"
    location="bottom start"
    transition="scale-transition"
  >
    <template #activator="{ props: menuProps }">
      <v-text-field
        v-bind="menuProps"
        :model-value="displayFrom"
        label="Từ ngày"
        density="compact"
        variant="outlined"
        hide-details
        clearable
        readonly
        class="admin-date-field"
        prepend-inner-icon="mdi-calendar"
        @click:clear="clearFrom"
      />
    </template>
    <v-date-picker
      :model-value="from || undefined"
      color="primary"
      hide-header
      @update:model-value="pickFrom"
    />
  </v-menu>

  <v-menu
    v-model="toMenu"
    :close-on-content-click="false"
    location="bottom start"
    transition="scale-transition"
  >
    <template #activator="{ props: menuProps }">
      <v-text-field
        v-bind="menuProps"
        :model-value="displayTo"
        label="Đến ngày"
        density="compact"
        variant="outlined"
        hide-details
        clearable
        readonly
        class="admin-date-field"
        prepend-inner-icon="mdi-calendar"
        @click:clear="clearTo"
      />
    </template>
    <v-date-picker
      :model-value="to || undefined"
      :min="from || undefined"
      color="primary"
      hide-header
      @update:model-value="pickTo"
    />
  </v-menu>
</template>

<script setup lang="ts">
const props = defineProps<{
  from: string
  to: string
}>()

const emit = defineEmits<{
  'update:from': [value: string]
  'update:to': [value: string]
}>()

const fromMenu = ref(false)
const toMenu = ref(false)

function formatDisplay(iso: string): string {
  if (!iso?.trim()) return ''
  const [y, m, d] = iso.trim().split('-')
  if (y && m && d) return `${d}/${m}/${y}`
  return iso
}

const displayFrom = computed(() => formatDisplay(props.from))
const displayTo = computed(() => formatDisplay(props.to))

function normalizeDate(value: string | Date | null | undefined): string {
  if (!value) return ''
  if (value instanceof Date) {
    const y = value.getFullYear()
    const m = String(value.getMonth() + 1).padStart(2, '0')
    const d = String(value.getDate()).padStart(2, '0')
    return `${y}-${m}-${d}`
  }
  return String(value)
}

function pickFrom(value: string | Date | null) {
  emit('update:from', normalizeDate(value))
  fromMenu.value = false
}

function pickTo(value: string | Date | null) {
  emit('update:to', normalizeDate(value))
  toMenu.value = false
}

function clearFrom() {
  emit('update:from', '')
  fromMenu.value = false
}

function clearTo() {
  emit('update:to', '')
  toMenu.value = false
}
</script>

<style scoped>
.admin-date-field {
  flex: 0 1 160px;
  min-width: 140px;
  max-width: 180px;
}
</style>
