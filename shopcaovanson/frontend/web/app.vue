<template>
  <v-app>
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
  await auth.ensureAuth()
  if (auth.isLoggedIn.value) {
    await cart.fetchCart()
  }
})
</script>
