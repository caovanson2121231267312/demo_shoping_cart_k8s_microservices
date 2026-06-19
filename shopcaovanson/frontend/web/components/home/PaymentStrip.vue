<template>
  <section class="payment-strip">
    <v-container class="payment-strip__inner">
      <span class="payment-strip__label">Thanh toán:</span>
      <div class="payment-strip__items">
        <v-chip v-for="m in methods" :key="m" size="small" variant="outlined" class="payment-strip__chip">
          {{ m }}
        </v-chip>
      </div>
      <v-divider vertical class="d-none d-md-flex mx-4" />
      <span class="payment-strip__label d-none d-md-inline">Đối tác:</span>
      <div class="payment-strip__items d-none d-md-flex">
        <v-chip v-for="b in brands" :key="b" size="small" color="grey-lighten-3" variant="flat">
          {{ b }}
        </v-chip>
      </div>
    </v-container>
  </section>
</template>

<script setup lang="ts">
const methods = ['COD', 'Chuyển khoản', 'Ví điện tử', 'Thẻ tín dụng']
const brands = ['GHN', 'GHTK', 'Viettel Post', 'J&T']
</script>

<style scoped>
.payment-strip {
  background: #fff;
  border-top: 1px solid rgba(0, 0, 0, 0.06);
}

.payment-strip__inner {
  max-width: 1280px;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
  padding-top: 16px;
  padding-bottom: 16px;
}

.payment-strip__label {
  font-size: 12px;
  font-weight: 700;
  color: #757575;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.payment-strip__items {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.payment-strip__chip {
  font-weight: 600;
}
</style>
