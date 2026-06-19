<template>
  <section class="service-highlights">
    <v-container class="page-container">
      <div class="section-header text-center mb-6">
        <h2 class="section-title">Dịch vụ mua sắm</h2>
        <p class="section-subtitle">Trải nghiệm mua sắm an tâm từ A đến Z</p>
      </div>
      <v-row>
        <v-col v-for="item in services" :key="item.title" cols="6" md="4" lg="2">
          <div class="service-card">
            <v-icon :color="item.color" size="32" class="mb-3">{{ item.icon }}</v-icon>
            <div class="service-card__title">{{ item.title }}</div>
            <div class="service-card__desc">{{ item.desc }}</div>
          </div>
        </v-col>
      </v-row>
    </v-container>
  </section>
</template>

<script setup lang="ts">
const services = [
  { icon: 'mdi-truck-fast-outline', title: 'Giao 2H', desc: 'Nội thành HCM & HN', color: 'primary' },
  { icon: 'mdi-credit-card-outline', title: 'Trả góp 0%', desc: 'Thẻ tín dụng', color: 'secondary' },
  { icon: 'mdi-shield-check-outline', title: 'Bảo hành', desc: 'Chính hãng 12–24 tháng', color: 'success' },
  { icon: 'mdi-sync', title: 'Đổi trả', desc: 'Miễn phí 7 ngày', color: 'info' },
  { icon: 'mdi-gift-outline', title: 'Quà tặng', desc: 'Kèm phụ kiện', color: 'error' },
  { icon: 'mdi-star-circle-outline', title: 'Tích điểm', desc: 'Đổi voucher', color: 'warning' },
]
</script>

<style scoped>
.service-highlights {
  padding: 32px 0;
}

.service-card {
  text-align: center;
  padding: 20px 12px;
  background: var(--color-surface);
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border-light);
  height: 100%;
  transition: box-shadow 0.2s ease, transform 0.2s ease;
}

.service-card:hover {
  box-shadow: var(--shadow-md);
  transform: translateY(-2px);
}

.service-card__title {
  font-size: 14px;
  font-weight: 700;
  color: var(--color-text);
}

.service-card__desc {
  font-size: 11px;
  color: var(--color-text-secondary);
  margin-top: 4px;
  line-height: 1.3;
}
</style>
