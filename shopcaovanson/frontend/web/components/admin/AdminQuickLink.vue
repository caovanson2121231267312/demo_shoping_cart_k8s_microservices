<template>
  <NuxtLink :to="to" class="admin-quick-card">
    <div class="admin-quick-card__icon">
      <v-icon :icon="icon" size="24" />
    </div>
    <div>
      <div class="admin-quick-card__title">{{ title }}</div>
      <div v-if="description" class="admin-quick-card__desc">{{ description }}</div>
    </div>
    <v-spacer />
    <v-icon size="18" color="grey-lighten-1">mdi-chevron-right</v-icon>
  </NuxtLink>
</template>

<script setup lang="ts">
defineProps<{
  to: string
  title: string
  icon: string
  description?: string
}>()
</script>
