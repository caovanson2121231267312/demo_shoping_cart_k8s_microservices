<template>
  <section class="coupon-strip">
    <v-container class="coupon-strip__inner">
      <div class="coupon-strip__header">
        <v-icon color="error" size="22">mdi-ticket-percent</v-icon>
        <span class="coupon-strip__title">Mã ưu đãi hôm nay</span>
      </div>
      <div class="coupon-strip__list">
        <div v-for="coupon in coupons" :key="coupon.code" class="coupon-card">
          <div class="coupon-card__value">{{ coupon.value }}</div>
          <div class="coupon-card__info">
            <div class="coupon-card__code">{{ coupon.code }}</div>
            <div class="coupon-card__desc">{{ coupon.desc }}</div>
          </div>
          <v-btn
            size="small"
            variant="outlined"
            color="error"
            rounded="lg"
            class="text-none coupon-card__btn"
            @click="copyCode(coupon.code)"
          >
            {{ copied === coupon.code ? 'Đã copy' : 'Lấy mã' }}
          </v-btn>
        </div>
      </div>
    </v-container>
  </section>
</template>

<script setup lang="ts">
const coupons = [
  { code: 'WELCOME10', value: 'Giảm 10%', desc: 'Đơn đầu tiên từ 200K' },
  { code: 'FREESHIP', value: 'Freeship', desc: 'Đơn từ 500K toàn quốc' },
  { code: 'SALE50K', value: 'Giảm 50K', desc: 'Đơn từ 1 triệu' },
  { code: 'VIP15', value: 'Giảm 15%', desc: 'Thành viên thân thiết' },
]

const copied = ref<string | null>(null)

const copyCode = async (code: string) => {
  try {
    await navigator.clipboard.writeText(code)
    copied.value = code
    useSnackbar().show(`Đã copy mã ${code}`, 'success')
    setTimeout(() => { copied.value = null }, 2000)
  } catch {
    useSnackbar().show('Không thể copy mã', 'error')
  }
}
</script>

<style scoped>
.coupon-strip {
  background: var(--color-sale-bg);
  border-top: 1px solid var(--color-sale-border);
  border-bottom: 1px solid var(--color-sale-border);
}

.coupon-strip__inner {
  max-width: 1280px;
  padding-top: 20px;
  padding-bottom: 20px;
}

.coupon-strip__header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 14px;
}

.coupon-strip__title {
  font-size: 15px;
  font-weight: 800;
  color: var(--color-sale-dark);
}

.coupon-strip__list {
  display: grid;
  grid-template-columns: 1fr;
  gap: 10px;
}

@media (min-width: 600px) {
  .coupon-strip__list {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (min-width: 960px) {
  .coupon-strip__list {
    grid-template-columns: repeat(4, 1fr);
  }
}

.coupon-card {
  display: flex;
  align-items: center;
  gap: 12px;
  background: var(--color-surface);
  border: 1px dashed var(--color-sale);
  border-radius: var(--radius-md);
  padding: 12px 14px;
}

.coupon-card__value {
  flex-shrink: 0;
  font-size: 13px;
  font-weight: 800;
  color: var(--color-sale);
  background: var(--color-sale-bg);
  padding: 6px 10px;
  border-radius: var(--radius-sm);
  white-space: nowrap;
}

.coupon-card__info {
  flex: 1;
  min-width: 0;
}

.coupon-card__code {
  font-size: 13px;
  font-weight: 700;
  color: var(--color-text);
}

.coupon-card__desc {
  font-size: 11px;
  color: var(--color-text-secondary);
  margin-top: 2px;
}

.coupon-card__btn {
  flex-shrink: 0;
  font-size: 11px !important;
}
</style>
