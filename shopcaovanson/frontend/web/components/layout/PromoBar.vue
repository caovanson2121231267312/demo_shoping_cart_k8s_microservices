<template>

  <div class="promo-bar">

    <v-container class="promo-bar__inner">

      <div class="promo-bar__left d-none d-sm-flex align-center ga-4">

        <NuxtLink to="/" class="promo-bar__utility">

          <v-icon size="14">mdi-store</v-icon>

          Kênh người bán

        </NuxtLink>

        <span class="promo-bar__divider" />

        <NuxtLink to="/auth/register" class="promo-bar__utility">

          <v-icon size="14">mdi-account-plus</v-icon>

          Đăng ký

        </NuxtLink>

      </div>



      <div class="promo-bar__scroll flex-grow-1">

        <div class="promo-bar__track">

          <span v-for="n in 2" :key="n" class="promo-bar__content">

            <span v-for="(msg, i) in messages" :key="`${n}-${i}`" class="promo-bar__item">

              <v-icon size="14" class="mr-1">{{ msg.icon }}</v-icon>{{ msg.text }}

            </span>

          </span>

        </div>

      </div>



      <div class="promo-bar__right d-flex align-center ga-3 flex-shrink-0">

        <NuxtLink to="/orders/track" class="promo-bar__utility d-none d-md-inline-flex">

          <v-icon size="14">mdi-truck-fast</v-icon>

          Tra cứu đơn

        </NuxtLink>

        <a href="tel:19001234" class="promo-bar__hotline">

          <v-icon size="14">mdi-phone-in-talk</v-icon>

          <span class="d-none d-sm-inline">Hotline</span>

          <strong>1900 1234</strong>

        </a>

      </div>

    </v-container>

  </div>

</template>



<script setup lang="ts">

const messages = [

  { icon: 'mdi-truck-delivery', text: 'Freeship đơn từ 500K — mã FREESHIP' },

  { icon: 'mdi-tag', text: 'Giảm 10% thành viên mới — WELCOME10' },

  { icon: 'mdi-sync', text: 'Đổi trả miễn phí trong 7 ngày' },

  { icon: 'mdi-flash', text: 'Giao nhanh 2h nội thành HCM' },

]

</script>



<style scoped>

.promo-bar {

  background: var(--gradient-primary-horizontal);

  color: var(--color-text-inverse);

  font-size: 12px;

  height: 38px;

}



.promo-bar__inner {

  max-width: 1280px;

  height: 38px;

  display: flex;

  align-items: center;

  gap: 16px;

  padding-top: 0;

  padding-bottom: 0;

}



.promo-bar__utility {

  color: rgba(255, 255, 255, 0.9);

  text-decoration: none;

  font-size: 12px;

  display: inline-flex;

  align-items: center;

  gap: 4px;

  white-space: nowrap;

  transition: color 0.2s;

}



.promo-bar__utility:hover {

  color: var(--color-text-inverse);

}



.promo-bar__divider {

  width: 1px;

  height: 14px;

  background: rgba(255, 255, 255, 0.3);

}



.promo-bar__scroll {

  overflow: hidden;

  min-width: 0;

}



.promo-bar__track {

  display: flex;

  width: max-content;

  animation: promo-scroll 36s linear infinite;

}



.promo-bar__content {

  display: flex;

  white-space: nowrap;

  padding-right: 48px;

}



.promo-bar__item {

  display: inline-flex;

  align-items: center;

  margin-right: 36px;

  font-weight: 500;

}



.promo-bar__hotline {

  color: var(--color-text-inverse);

  text-decoration: none;

  font-size: 12px;

  display: inline-flex;

  align-items: center;

  gap: 6px;

  white-space: nowrap;

  background: rgba(255, 255, 255, 0.12);

  padding: 4px 10px;

  border-radius: var(--radius-pill);

  transition: background 0.2s;

}



.promo-bar__hotline:hover {

  background: rgba(255, 255, 255, 0.22);

}



.promo-bar__hotline strong {

  font-weight: 700;

  letter-spacing: 0.3px;

}



@keyframes promo-scroll {

  0% { transform: translateX(0); }

  100% { transform: translateX(-50%); }

}

</style>

