<template>
  <v-carousel
    cycle
    :height="carouselHeight"
    hide-delimiter-background
    show-arrows="hover"
    class="hero-carousel"
    :interval="5000"
  >
    <v-carousel-item v-for="(slide, i) in slides" :key="i" :src="slide.image" cover>
      <div class="hero-carousel__overlay d-flex align-center">
        <v-container class="hero-carousel__container">
          <v-row>
            <v-col cols="12" md="7" lg="6" class="text-white">
              <v-chip v-if="slide.tag" color="error" size="small" class="mb-3 font-weight-bold">
                {{ slide.tag }}
              </v-chip>
              <h2 class="text-h4 text-md-h3 text-lg-h2 font-weight-bold mb-3 hero-animate">{{ slide.title }}</h2>
              <p class="text-body-1 text-md-h6 font-weight-regular mb-6 opacity-90 hero-animate hero-animate--delay">
                {{ slide.subtitle }}
              </p>
              <div class="d-flex flex-wrap ga-3 hero-animate hero-animate--delay2">
                <v-btn :to="slide.to" size="large" color="error" rounded="lg" class="text-none font-weight-bold">
                  {{ slide.cta }}
                </v-btn>
                <v-btn :to="slide.to2" size="large" variant="outlined" color="white" rounded="lg" class="text-none">
                  Khám phá thêm
                </v-btn>
              </div>
            </v-col>
          </v-row>
        </v-container>
      </div>
    </v-carousel-item>
  </v-carousel>
</template>

<script setup lang="ts">
const carouselHeight = ref(420)

const updateHeight = () => {
  if (window.innerWidth < 600) carouselHeight.value = 280
  else if (window.innerWidth < 960) carouselHeight.value = 340
  else carouselHeight.value = 420
}

onMounted(() => {
  updateHeight()
  window.addEventListener('resize', updateHeight)
})

onUnmounted(() => {
  window.removeEventListener('resize', updateHeight)
})

const slides = [
  {
    title: 'Siêu sale cuối tuần',
    subtitle: 'Giảm đến 50% — hàng ngàn sản phẩm điện tử & thời trang. Mã WELCOME10 cho đơn đầu.',
    cta: 'Săn deal ngay',
    to: '/products?sort=price_desc',
    to2: '/products',
    tag: 'FLASH SALE',
    image: 'https://picsum.photos/seed/hero-electronics/1920/640',
  },
  {
    title: 'Bộ sưu tập mùa hè 2026',
    subtitle: 'Thời trang & làm đẹp — phong cách tươi mới, giao nhanh 2h nội thành',
    cta: 'Mua sắm ngay',
    to: '/products?sort=newest',
    to2: '/wishlist',
    tag: 'NEW ARRIVAL',
    image: 'https://picsum.photos/seed/hero-fashion/1920/640',
  },
  {
    title: 'Gia dụng thông minh',
    subtitle: 'Nhà cửa tiện nghi — Freeship toàn quốc đơn từ 500K',
    cta: 'Xem sản phẩm',
    to: '/products',
    to2: '/orders/track',
    tag: 'FREESHIP',
    image: 'https://picsum.photos/seed/hero-home/1920/640',
  },
]
</script>

<style scoped>
.hero-carousel {
  border-radius: 0;
}

.hero-carousel__container {
  max-width: 1280px;
}

.hero-carousel__overlay {
  position: absolute;
  inset: 0;
  background: var(--gradient-hero-overlay);
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
