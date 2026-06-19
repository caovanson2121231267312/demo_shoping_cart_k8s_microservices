<template>
  <v-btn
    v-show="visible"
    class="back-to-top"
    icon
    color="primary"
    size="large"
    elevation="4"
    @click="scrollTop"
  >
    <v-icon>mdi-chevron-up</v-icon>
  </v-btn>
</template>

<script setup lang="ts">
const visible = ref(false)

const onScroll = () => {
  visible.value = window.scrollY > 400
}

const scrollTop = () => {
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

onMounted(() => window.addEventListener('scroll', onScroll, { passive: true }))
onUnmounted(() => window.removeEventListener('scroll', onScroll))
</script>

<style scoped>
.back-to-top {
  position: fixed;
  right: max(20px, env(safe-area-inset-right, 0px));
  bottom: calc(max(24px, env(safe-area-inset-bottom, 0px)) + 64px);
  z-index: calc(var(--chat-widget-z) - 1);
}
</style>
