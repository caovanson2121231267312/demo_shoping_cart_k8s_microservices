<template>
  <v-container class="page-container py-6" style="max-width: 560px">
    <h1 class="text-h4 font-weight-bold mb-6">Hồ sơ cá nhân</h1>

    <v-card>
      <v-card-text>
        <v-alert v-if="success" type="success" variant="tonal" class="mb-4">
          Cập nhật hồ sơ thành công!
        </v-alert>

        <v-form @submit.prevent="handleSave">
          <v-text-field
            v-model="email"
            label="Email"
            readonly
            variant="outlined"
            class="mb-2"
          />
          <v-text-field
            v-model="fullName"
            label="Họ và tên"
            :rules="[rules.required]"
            variant="outlined"
            class="mb-2"
          />
          <v-text-field
            :model-value="roleLabel"
            label="Vai trò"
            readonly
            variant="outlined"
            class="mb-4"
          />
          <v-btn type="submit" color="primary" :loading="loading">Lưu thay đổi</v-btn>
        </v-form>
      </v-card-text>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const auth = useAuth()

const fullName = ref('')
const email = ref('')
const loading = ref(false)
const success = ref(false)

const rules = {
  required: (v: string) => !!v || 'Trường này là bắt buộc',
}

const roleLabel = computed(() =>
  auth.user.value?.role === 'admin' ? 'Quản trị viên' : 'Khách hàng',
)

onMounted(() => {
  if (auth.user.value) {
    fullName.value = auth.user.value.full_name
    email.value = auth.user.value.email
  }
})

const handleSave = async () => {
  loading.value = true
  success.value = false
  try {
    await auth.updateProfile(fullName.value)
    success.value = true
  } finally {
    loading.value = false
  }
}
</script>
