<template>

  <section class="trust-bar">

    <v-container class="trust-bar__inner">

      <div v-for="item in items" :key="item.title" class="trust-bar__item">

        <v-icon :color="item.color" size="32" class="mb-2">{{ item.icon }}</v-icon>

        <div class="trust-bar__title">{{ item.title }}</div>

        <div class="trust-bar__desc">{{ item.desc }}</div>

      </div>

    </v-container>

  </section>

</template>



<script setup lang="ts">

const items = [

  { icon: 'mdi-truck-fast', title: 'Giao hàng nhanh', desc: 'Freeship đơn 500K', color: 'primary' },

  { icon: 'mdi-shield-check', title: 'Chính hãng', desc: 'Hoàn tiền 200%', color: 'success' },

  { icon: 'mdi-sync', title: 'Đổi trả 7 ngày', desc: 'Miễn phí đổi trả', color: 'info' },

  { icon: 'mdi-headset', title: 'Hỗ trợ 24/7', desc: 'Chat & hotline', color: 'secondary' },

]

</script>



<style scoped>

.trust-bar {

  background: var(--color-surface);

  border-bottom: 1px solid var(--color-border);

}



.trust-bar__inner {

  max-width: 1280px;

  display: grid;

  grid-template-columns: repeat(2, 1fr);

  gap: 16px;

  padding-top: 20px;

  padding-bottom: 20px;

}



@media (min-width: 768px) {

  .trust-bar__inner {

    grid-template-columns: repeat(4, 1fr);

  }

}



.trust-bar__item {

  text-align: center;

  padding: 8px;

}



.trust-bar__title {

  font-weight: 700;

  font-size: 14px;

  color: var(--color-text);

}



.trust-bar__desc {

  font-size: 12px;

  color: var(--color-text-secondary);

  margin-top: 2px;

}

</style>

