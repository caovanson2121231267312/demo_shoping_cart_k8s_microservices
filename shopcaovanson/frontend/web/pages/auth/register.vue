<template>
  <v-card>
    <v-card-title class="text-h5">Đăng ký</v-card-title>
    <v-card-text>
      <v-alert v-if="success" type="success" variant="tonal" class="mb-4">
        {{ success }}
        <div class="mt-2">
          <v-btn size="small" variant="text" :loading="resending" @click="resend">Gửi lại email xác nhận</v-btn>
        </div>
      </v-alert>
      <v-alert v-if="error" type="error" variant="tonal" class="mb-4" closable @click:close="error = ''">
        {{ error }}
      </v-alert>

      <v-form v-if="!success" @submit.prevent="handleRegister">
        <v-text-field
          v-model="fullName"
          label="Họ và tên"
          prepend-inner-icon="mdi-account-outline"
          :rules="[rules.required]"
          variant="outlined"
          class="mb-2"
        />
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
          :rules="[rules.required, rules.minLength]"
          variant="outlined"
          @click:append-inner="showPassword = !showPassword"
        />
        <v-btn type="submit" color="primary" size="large" block :loading="loading" class="mt-4">
          Đăng ký
        </v-btn>
      </v-form>
    </v-card-text>
    <v-card-actions class="justify-center pb-4">
      <span class="text-body-2">Đã có tài khoản?</span>
      <v-btn variant="text" color="primary" to="/auth/login">Đăng nhập</v-btn>
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'auth' })

const auth = useAuth()

const fullName = ref('')
const email = ref('')
const password = ref('')
const showPassword = ref(false)
const loading = ref(false)
const resending = ref(false)
const error = ref('')
const success = ref('')
const registeredEmail = ref('')

const rules = {
  required: (v: string) => !!v || 'Trường này là bắt buộc',
  email: (v: string) => /.+@.+\..+/.test(v) || 'Email không hợp lệ',
  minLength: (v: string) => v.length >= 8 || 'Mật khẩu tối thiểu 8 ký tự',
}

const handleRegister = async () => {
  if (!fullName.value || !email.value || !password.value) {
    return
  }
  loading.value = true
  error.value = ''
  try {
    const res = await auth.register(fullName.value, email.value, password.value)
    registeredEmail.value = res.email
    success.value = res.message || 'Vui lòng kiểm tra email để xác nhận tài khoản.'
  } catch (e: unknown) {
    const msg = (e as { data?: { error?: string } })?.data?.error
    error.value = msg || 'Đăng ký thất bại. Email có thể đã được sử dụng.'
  } finally {
    loading.value = false
  }
}

const resend = async () => {
  if (!registeredEmail.value) {
    return
  }
  resending.value = true
  try {
    await auth.resendVerification(registeredEmail.value)
    success.value = 'Email xác nhận đã được gửi lại.'
  } finally {
    resending.value = false
  }
}
</script>
