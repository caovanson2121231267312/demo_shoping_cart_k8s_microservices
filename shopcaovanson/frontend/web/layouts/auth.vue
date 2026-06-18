<template>
  <v-main class="d-flex align-center justify-center bg-grey-lighten-4" style="min-height: 100vh">
    <v-container class="py-8" style="max-width: 480px">
      <div class="text-center mb-6">
        <NuxtLink to="/" class="text-h5 font-weight-bold text-primary">
          Shop Cao Van Son
        </NuxtLink>
      </div>
      <slot />
    </v-container>
  </v-main>
</template>
