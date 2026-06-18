<template>
  <v-carousel
    cycle
    height="360"
    hide-delimiter-background
    show-arrows="hover"
    class="hero-carousel rounded-xl overflow-hidden mb-6"
  >
    <v-carousel-item v-for="(slide, i) in slides" :key="i" :src="slide.image" cover>
      <div class="hero-carousel__overlay d-flex align-center">
        <v-container>
          <v-row>
            <v-col cols="12" md="7" class="text-white">
              <v-chip v-if="slide.tag" color="secondary" size="small" class="mb-3">{{ slide.tag }}</v-chip>
              <h2 class="text-h3 text-md-h2 font-weight-bold mb-3 hero-animate">{{ slide.title }}</h2>
              <p class="text-h6 text-md-h5 font-weight-regular mb-6 opacity-90 hero-animate hero-animate--delay">
                {{ slide.subtitle }}
              </p>
              <v-btn :to="slide.to" size="large" color="secondary" class="hero-animate hero-animate--delay2">
                {{ slide.cta }}
              </v-btn>
            </v-col>
          </v-row>
        </v-container>
      </div>
    </v-carousel-item>
  </v-carousel>
</template>

<script setup lang="ts">
const slides = [
  {
    title: 'Siêu sale cuối tuần',
    subtitle: 'Giảm đến 50% hàng ngàn sản phẩm điện tử & thời trang',
    cta: 'Săn deal ngay',
    to: '/products?sort=price_desc',
    tag: 'HOT',
    image: 'https://picsum.photos/seed/hero-electronics/1600/600',
  },
  {
    title: 'Bộ sưu tập mùa hè',
    subtitle: 'Thời trang & làm đẹp — phong cách tươi mới mỗi ngày',
    cta: 'Khám phá',
    to: '/products?category=fashion',
    tag: 'NEW',
    image: 'https://picsum.photos/seed/hero-fashion/1600/600',
  },
  {
    title: 'Gia dụng thông minh',
    subtitle: 'Nhà cửa tiện nghi — giao nhanh 2h nội thành',
    cta: 'Mua sắm ngay',
    to: '/products?category=home',
    tag: 'SALE',
    image: 'https://picsum.photos/seed/hero-home/1600/600',
  },
]
</script>

<style scoped>
.hero-carousel__overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(90deg, rgba(0, 0, 0, 0.65) 0%, rgba(0, 0, 0, 0.15) 70%);
}

.hero-animate {
  animation: hero-fade-up 0.7s ease both;
}

.hero-animate--delay {
  animation-delay: 0.15s;
}

.hero-animate--delay2 {
  animation-delay: 0.3s;
}

@keyframes hero-fade-up {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
