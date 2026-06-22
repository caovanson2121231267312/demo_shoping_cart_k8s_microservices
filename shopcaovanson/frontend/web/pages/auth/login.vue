<template>

  <v-card>

    <v-card-title class="text-h5">Đăng nhập</v-card-title>

    <v-card-text>

      <v-alert v-if="needsVerify" type="warning" variant="tonal" class="mb-4">

        Email chưa được xác nhận.

        <v-btn size="small" variant="text" :loading="resending" @click="resendVerify">Gửi lại email xác nhận</v-btn>

      </v-alert>

      <v-alert v-if="error" type="error" variant="tonal" class="mb-4" closable @click:close="error = ''">

        {{ error }}

      </v-alert>



      <v-form autocomplete="on" @submit.prevent="handleLogin">

        <v-combobox

          v-model="email"

          :items="savedEmails"

          label="Email"

          type="email"

          name="username"

          autocomplete="username"

          prepend-inner-icon="mdi-email-outline"

          :rules="[rules.required, rules.email]"

          variant="outlined"

          class="mb-2"

          clearable

          hide-no-data

          @update:model-value="onEmailChange"

        >

          <template #item="{ item, props: itemProps }">

            <v-list-item v-bind="itemProps" :title="item.title">

              <template #append>

                <v-btn

                  icon="mdi-close"

                  size="x-small"

                  variant="text"

                  aria-label="Xóa tài khoản đã lưu"

                  @click.stop="forgetSaved(item.title)"

                />

              </template>

            </v-list-item>

          </template>

        </v-combobox>



        <v-text-field

          v-model="password"

          label="Mật khẩu"

          name="password"

          autocomplete="current-password"

          :type="showPassword ? 'text' : 'password'"

          prepend-inner-icon="mdi-lock-outline"

          :append-inner-icon="showPassword ? 'mdi-eye-off' : 'mdi-eye'"

          :rules="[rules.required]"

          variant="outlined"

          @click:append-inner="showPassword = !showPassword"

        />

        <div class="text-right mb-2">
          <v-btn variant="text" size="small" color="primary" class="px-0 text-none" to="/auth/forgot-password">
            Quên mật khẩu?
          </v-btn>
        </div>



        <v-checkbox

          v-model="rememberMe"

          label="Ghi nhớ đăng nhập"

          color="primary"

          density="compact"

          hide-details

          class="mt-1"

        />



        <p v-if="savedEmails.length" class="text-caption text-medium-emphasis mt-2 mb-0">

          Chọn email đã lưu để điền nhanh. Trình duyệt cũng có thể gợi ý mật khẩu đã lưu.

        </p>



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

const route = useRoute()

const remembered = useRememberedLogin()



const email = ref('')

const password = ref('')

const showPassword = ref(false)

const loading = ref(false)

const resending = ref(false)

const error = ref('')

const needsVerify = ref(false)



const rememberMe = remembered.rememberMe

const savedEmails = remembered.savedEmails



const rules = {

  required: (v: string) => !!v || 'Trường này là bắt buộc',

  email: (v: string) => /.+@.+\..+/.test(v) || 'Email không hợp lệ',

}



onMounted(async () => {

  const saved = await remembered.init()

  if (saved) {

    email.value = saved.email

    password.value = saved.password

  }

})



const onEmailChange = async (value: string | null) => {

  if (!value || !rememberMe.value) {

    return

  }

  const savedPassword = await remembered.fillPasswordForEmail(value)

  if (savedPassword) {

    password.value = savedPassword

  }

}



const forgetSaved = (accountEmail: string) => {

  remembered.forgetAccount(accountEmail)

  if (email.value.trim().toLowerCase() === accountEmail.trim().toLowerCase()) {

    password.value = ''

  }

}



const handleLogin = async () => {

  if (!email.value || !password.value) {

    return

  }

  loading.value = true

  error.value = ''

  needsVerify.value = false

  try {

    await auth.login(email.value, password.value)

    await remembered.persistAfterLogin(email.value, password.value)

    await cart.mergeLocalToServer()

    const target = resolveLoginRedirect(route.query.redirect, '/')

    await navigateTo(target, { replace: true })

  } catch (e: unknown) {

    const data = (e as { data?: { error?: string; code?: string } })?.data

    if (data?.code === 'EMAIL_NOT_VERIFIED') {

      needsVerify.value = true

      error.value = 'Vui lòng xác nhận email trước khi đăng nhập.'

    } else {

      error.value = data?.error || 'Đăng nhập thất bại. Vui lòng kiểm tra lại thông tin.'

    }

  } finally {

    loading.value = false

  }

}



const resendVerify = async () => {

  if (!email.value) return

  resending.value = true

  try {

    await auth.resendVerification(email.value)

    error.value = ''

    needsVerify.value = false

    useSnackbar().show('Đã gửi lại email xác nhận', 'success')

  } finally {

    resending.value = false

  }

}

</script>

