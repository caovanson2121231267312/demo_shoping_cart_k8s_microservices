<template>
  <v-app class="store-app">
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
    <ChatWidget />
  </v-app>
</template>

<script setup lang="ts">
const auth = useAuth()
const cart = useCart()

onMounted(async () => {
  useCartStore().hydrate()
  await auth.ensureAuth()
  if (auth.isLoggedIn.value) {
    await cart.mergeLocalToServer()
  }
})
</script>

<style>
.store-app {
  min-height: 100vh;
}

.store-app .v-application__wrap {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}
</style>
