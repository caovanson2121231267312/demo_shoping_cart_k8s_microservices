<template>
  <section class="newsletter">
    <v-container class="newsletter__inner">
      <div class="newsletter__content">
        <v-icon color="white" size="40" class="mb-3">mdi-email-newsletter</v-icon>
        <h3 class="text-h5 font-weight-bold text-white mb-2">Đăng ký nhận ưu đãi</h3>
        <p class="text-body-2 text-white mb-0 opacity-90">
          Nhận mã giảm 10% cho đơn đầu tiên và cập nhật deal mới mỗi tuần
        </p>
      </div>
      <div class="newsletter__form">
        <v-text-field
          v-model="email"
          placeholder="Nhập email của bạn"
          variant="solo"
          density="comfortable"
          hide-details
          rounded="lg"
          bg-color="white"
          class="newsletter__input"
        />
        <v-btn color="secondary" size="large" rounded="lg" class="text-none" @click="subscribe">
          Đăng ký
        </v-btn>
      </div>
    </v-container>
  </section>
</template>

<script setup lang="ts">
const email = ref('')

const subscribe = () => {
  if (!email.value) return
  useSnackbar().show('Cảm ơn bạn đã đăng ký nhận tin!', 'success')
  email.value = ''
}
</script>

<style scoped>
.newsletter {
  background: linear-gradient(135deg, #0d47a1 0%, #1565c0 50%, #1976d2 100%);
  margin: 32px 0 0;
}

.newsletter__inner {
  max-width: 1280px;
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding-top: 40px;
  padding-bottom: 40px;
}

@media (min-width: 768px) {
  .newsletter__inner {
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
  }
}

.newsletter__content {
  flex: 1;
}

.newsletter__form {
  display: flex;
  flex-direction: column;
  gap: 10px;
  width: 100%;
  max-width: 420px;
}

@media (min-width: 600px) {
  .newsletter__form {
    flex-direction: row;
    align-items: center;
  }

  .newsletter__input {
    flex: 1;
  }
}
</style>
