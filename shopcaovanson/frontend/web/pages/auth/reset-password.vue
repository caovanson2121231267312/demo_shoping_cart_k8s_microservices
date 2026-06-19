<template>
  <v-card>
    <v-card-title class="text-h5">Đặt lại mật khẩu</v-card-title>
    <v-card-text>
      <v-alert v-if="success" type="success" variant="tonal" class="mb-4">
        {{ success }}
      </v-alert>
      <v-alert v-if="error" type="error" variant="tonal" class="mb-4" closable @click:close="error = ''">
        {{ error }}
      </v-alert>

      <v-form v-if="!success" @submit.prevent="handleSubmit">
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
          v-model="otp"
          label="Mã OTP (6 số)"
          prepend-inner-icon="mdi-shield-key-outline"
          maxlength="6"
          inputmode="numeric"
          :rules="[rules.required, rules.otp]"
          variant="outlined"
          class="mb-2"
        />
        <v-text-field
          v-model="password"
          label="Mật khẩu mới"
          :type="showPassword ? 'text' : 'password'"
          prepend-inner-icon="mdi-lock-outline"
          :append-inner-icon="showPassword ? 'mdi-eye-off' : 'mdi-eye'"
          :rules="[rules.required, rules.minLength]"
          variant="outlined"
          class="mb-2"
          @click:append-inner="showPassword = !showPassword"
        />
        <v-text-field
          v-model="confirmPassword"
          label="Xác nhận mật khẩu mới"
          :type="showPassword ? 'text' : 'password'"
          prepend-inner-icon="mdi-lock-check-outline"
          :rules="[rules.required, rules.match]"
          variant="outlined"
        />
        <v-btn type="submit" color="primary" size="large" block :loading="loading" class="mt-4">
          Đặt lại mật khẩu
        </v-btn>
      </v-form>

      <v-btn v-else color="primary" block to="/auth/login" class="mt-2">
        Đăng nhập ngay
      </v-btn>
    </v-card-text>
    <v-card-actions class="justify-center pb-4 flex-wrap ga-2">
      <v-btn variant="text" color="primary" to="/auth/forgot-password">Gửi lại OTP</v-btn>
      <v-btn variant="text" color="primary" to="/auth/login">Quay lại đăng nhập</v-btn>
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'auth' })

const auth = useAuth()
const route = useRoute()

const email = ref((route.query.email as string) || '')
const otp = ref('')
const password = ref('')
const confirmPassword = ref('')
const showPassword = ref(false)
const loading = ref(false)
const error = ref('')
const success = ref('')

const rules = {
  required: (v: string) => !!v || 'Trường này là bắt buộc',
  email: (v: string) => /.+@.+\..+/.test(v) || 'Email không hợp lệ',
  otp: (v: string) => /^\d{6}$/.test(v) || 'Mã OTP gồm 6 chữ số',
  minLength: (v: string) => v.length >= 8 || 'Mật khẩu tối thiểu 8 ký tự',
  match: (v: string) => v === password.value || 'Mật khẩu xác nhận không khớp',
}

const handleSubmit = async () => {
  if (!email.value || !otp.value || !password.value) return
  if (password.value !== confirmPassword.value) {
    error.value = 'Mật khẩu xác nhận không khớp'
    return
  }
  loading.value = true
  error.value = ''
  try {
    const resp = await auth.resetPassword(email.value, otp.value, password.value)
    success.value = resp.message
    useSnackbar().show('Đặt lại mật khẩu thành công', 'success')
  } catch (e: unknown) {
    const data = (e as { data?: { error?: string; code?: string } })?.data
    error.value = data?.error || 'Không thể đặt lại mật khẩu. Vui lòng kiểm tra mã OTP.'
  } finally {
    loading.value = false
  }
}
</script>
