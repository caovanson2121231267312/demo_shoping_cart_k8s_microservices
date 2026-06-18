<template>
  <v-card>
    <v-card-title class="text-h5">Đăng nhập</v-card-title>
    <v-card-text>
      <v-alert v-if="error" type="error" variant="tonal" class="mb-4" closable @click:close="error = ''">
        {{ error }}
      </v-alert>

      <v-form @submit.prevent="handleLogin">
        <v-text-field
          v-model="email"
          label="Email"
          type="email"
          prepend-inner-icon="mdi-email-outline"
          :rules="[rules.required, rules.email]"
          variant="outlined"
          class="mb-2"
        />
        <v-text-field
          v-model="password"
          label="Mật khẩu"
          :type="showPassword ? 'text' : 'password'"
          prepend-inner-icon="mdi-lock-outline"
          :append-inner-icon="showPassword ? 'mdi-eye-off' : 'mdi-eye'"
          :rules="[rules.required]"
          variant="outlined"
          @click:append-inner="showPassword = !showPassword"
        />
        <v-btn type="submit" color="primary" size="large" block :loading="loading" class="mt-4">
          Đăng nhập
        </v-btn>
      </v-form>
    </v-card-text>
    <v-card-actions class="justify-center pb-4">
      <span class="text-body-2">Chưa có tài khoản?</span>
      <v-btn variant="text" color="primary" to="/auth/register">Đăng ký</v-btn>
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'auth' })

const auth = useAuth()
const cart = useCart()
const router = useRouter()

const email = ref('')
const password = ref('')
const showPassword = ref(false)
const loading = ref(false)
const error = ref('')

const rules = {
  required: (v: string) => !!v || 'Trường này là bắt buộc',
  email: (v: string) => /.+@.+\..+/.test(v) || 'Email không hợp lệ',
}

const handleLogin = async () => {
  if (!email.value || !password.value) {
    return
  }
  loading.value = true
  error.value = ''
  try {
    await auth.login(email.value, password.value)
    await cart.fetchCart()
    await router.push('/')
  } catch (e: unknown) {
    const msg = (e as { data?: { message?: string } })?.data?.message
    error.value = msg || 'Đăng nhập thất bại. Vui lòng kiểm tra lại thông tin.'
  } finally {
    loading.value = false
  }
}
</script>
