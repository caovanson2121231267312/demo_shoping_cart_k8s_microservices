<template>
  <div class="text-center py-12">
    <v-icon :icon="icon" size="64" color="grey-lighten-1" class="mb-4" />
    <div class="text-h6 text-grey-darken-1">{{ title }}</div>
    <p v-if="description" class="text-body-2 text-grey mt-2">{{ description }}</p>
    <slot />
  </div>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    icon?: string
    title?: string
    description?: string
  }>(),
  {
    icon: 'mdi-inbox-outline',
    title: 'Không có dữ liệu',
    description: '',
  },
)
</script>
