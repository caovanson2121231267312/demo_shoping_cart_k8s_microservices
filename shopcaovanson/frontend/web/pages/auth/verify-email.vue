<template>
  <v-card max-width="480" class="mx-auto">
    <v-card-title class="text-h5">Xác nhận email</v-card-title>
    <v-card-text>
      <LoadingSpinner v-if="loading" />
      <v-alert v-else-if="error" type="error" variant="tonal">{{ error }}</v-alert>
      <v-alert v-else-if="verified" type="success" variant="tonal">
        Tài khoản đã được kích hoạt! Đang chuyển về trang chủ...
      </v-alert>
      <p v-else class="text-body-2 text-grey">Đang xử lý xác nhận...</p>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'auth' })

const route = useRoute()
const router = useRouter()
const auth = useAuth()
const cart = useCart()

const loading = ref(true)
const error = ref('')
const verified = ref(false)

onMounted(async () => {
  const token = route.query.token as string
  if (!token) {
    error.value = 'Link xác nhận không hợp lệ.'
    loading.value = false
    return
  }
  try {
    await auth.verifyEmail(token)
    verified.value = true
    await cart.fetchCart()
    setTimeout(() => router.push('/'), 2000)
  } catch (e: unknown) {
    error.value = (e as { data?: { error?: string } })?.data?.error || 'Token không hợp lệ hoặc đã hết hạn.'
  } finally {
    loading.value = false
  }
})
</script>
