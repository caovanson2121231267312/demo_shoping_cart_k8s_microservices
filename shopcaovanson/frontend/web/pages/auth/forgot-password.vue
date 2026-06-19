<template>
  <v-card>
    <v-card-title class="text-h5">Quên mật khẩu</v-card-title>
    <v-card-text>
      <v-alert v-if="success" type="success" variant="tonal" class="mb-4">
        {{ success }}
      </v-alert>
      <v-alert v-if="error" type="error" variant="tonal" class="mb-4" closable @click:close="error = ''">
        {{ error }}
      </v-alert>

      <p v-if="!success" class="text-body-2 text-medium-emphasis mb-4">
        Nhập email đăng ký. Chúng tôi sẽ gửi mã OTP qua email để đặt lại mật khẩu.
      </p>

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
        <v-btn type="submit" color="primary" size="large" block :loading="loading" class="mt-2">
          Gửi mã OTP
        </v-btn>
      </v-form>

      <v-btn
        v-else
        color="primary"
        block
        class="mt-2"
        :to="resetLink"
      >
        Nhập mã OTP
      </v-btn>
    </v-card-text>
    <v-card-actions class="justify-center pb-4">
      <v-btn variant="text" color="primary" to="/auth/login">Quay lại đăng nhập</v-btn>
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'auth' })

const auth = useAuth()
const route = useRoute()

const email = ref((route.query.email as string) || '')
const loading = ref(false)
const error = ref('')
const success = ref('')

const rules = {
  required: (v: string) => !!v || 'Trường này là bắt buộc',
  email: (v: string) => /.+@.+\..+/.test(v) || 'Email không hợp lệ',
}

const resetLink = computed(() => ({
  path: '/auth/reset-password',
  query: { email: email.value },
}))

const handleSubmit = async () => {
  if (!email.value) return
  loading.value = true
  error.value = ''
  try {
    const resp = await auth.forgotPassword(email.value)
    success.value = resp.message
    useSnackbar().show('Đã gửi mã OTP (nếu email tồn tại)', 'success')
  } catch (e: unknown) {
    error.value = (e as { data?: { error?: string } })?.data?.error || 'Không thể gửi yêu cầu. Vui lòng thử lại.'
  } finally {
    loading.value = false
  }
}
</script>
