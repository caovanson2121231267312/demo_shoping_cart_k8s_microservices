<template>
  <div class="admin-stat-card">
    <div class="admin-stat-card__glow" :style="{ background: color }" />
    <div class="admin-stat-card__icon" :style="{ background: `${color}18`, color }">
      <v-icon :icon="icon" size="22" />
    </div>
    <div class="admin-stat-card__label">{{ label }}</div>
    <div class="admin-stat-card__value">{{ value }}</div>
    <div v-if="subtitle" class="admin-stat-card__sub">{{ subtitle }}</div>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  label: string
  value: string | number
  icon: string
  color?: string
  subtitle?: string
}>(), {
  color: 'var(--admin-primary)',
})
</script>
