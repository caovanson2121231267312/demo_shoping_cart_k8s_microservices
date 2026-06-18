<template>
  <div class="d-flex justify-center align-center py-8">
    <v-progress-circular indeterminate color="primary" :size="size" />
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{ size?: number }>(), { size: 48 })
</script>
