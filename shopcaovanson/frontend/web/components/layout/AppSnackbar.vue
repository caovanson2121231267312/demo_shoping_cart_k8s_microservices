<template>
  <v-snackbar v-model="visible" :color="color" location="bottom right" :timeout="2800">
    {{ message }}
    <template #actions>
      <v-btn variant="text" @click="visible = false">Đóng</v-btn>
    </template>
  </v-snackbar>
</template>

<script setup lang="ts">
const { message, visible, color } = useSnackbar()
</script>
