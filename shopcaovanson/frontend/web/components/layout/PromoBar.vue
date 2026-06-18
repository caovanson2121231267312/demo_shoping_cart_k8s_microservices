<template>
  <div class="promo-bar">
    <div class="promo-bar__track">
      <span v-for="n in 2" :key="n" class="promo-bar__content">
        <span v-for="(msg, i) in messages" :key="`${n}-${i}`" class="promo-bar__item">
          <v-icon size="14" class="mr-1">mdi-circle-small</v-icon>{{ msg }}
        </span>
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
const messages = [
  'Freeship đơn từ 500K toàn quốc',
  'Giảm thêm 10% cho thành viên mới',
  'Đổi trả trong 7 ngày — miễn phí',
  'Hotline: 1900 1234 (8h–22h)',
]
</script>

<style scoped>
.promo-bar {
  background: linear-gradient(90deg, #1a237e, #283593);
  color: #fff;
  overflow: hidden;
  font-size: 12px;
  height: 32px;
  display: flex;
  align-items: center;
}

.promo-bar__track {
  display: flex;
  width: max-content;
  animation: promo-scroll 28s linear infinite;
}

.promo-bar__content {
  display: flex;
  white-space: nowrap;
  padding-right: 48px;
}

.promo-bar__item {
  display: inline-flex;
  align-items: center;
  margin-right: 32px;
}

@keyframes promo-scroll {
  0% { transform: translateX(0); }
  100% { transform: translateX(-50%); }
}
</style>
