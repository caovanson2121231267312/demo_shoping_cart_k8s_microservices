<template>
  <div class="admin-page-header">
    <div class="d-flex align-center justify-space-between flex-wrap ga-3">
      <div>
        <h1 class="admin-page-header__title">{{ title }}</h1>
        <p v-if="subtitle" class="admin-page-header__subtitle">{{ subtitle }}</p>
      </div>
      <div v-if="$slots.actions" class="d-flex flex-wrap ga-2 align-center">
        <slot name="actions" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  subtitle?: string
}>()
</script>
