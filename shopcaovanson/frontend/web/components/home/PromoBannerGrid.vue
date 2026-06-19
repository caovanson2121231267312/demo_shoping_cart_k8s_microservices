<template>
  <section class="promo-grid">
    <v-container class="promo-grid__inner">
      <NuxtLink
        v-for="banner in banners"
        :key="banner.title"
        :to="banner.to"
        class="promo-grid__card"
        :class="banner.variant"
      >
        <div class="promo-grid__content">
          <v-chip v-if="banner.tag" size="x-small" color="white" variant="flat" class="mb-2 font-weight-bold">
            {{ banner.tag }}
          </v-chip>
          <div class="promo-grid__title">{{ banner.title }}</div>
          <div class="promo-grid__subtitle">{{ banner.subtitle }}</div>
          <span class="promo-grid__cta">Mua ngay →</span>
        </div>
        <v-icon class="promo-grid__icon" size="72" color="white">{{ banner.icon }}</v-icon>
      </NuxtLink>
    </v-container>
  </section>
</template>

<script setup lang="ts">
const banners = [
  {
    title: 'Điện tử hot',
    subtitle: 'Giảm đến 40%',
    to: '/products?sort=price_desc',
    tag: 'SALE',
    icon: 'mdi-laptop',
    variant: 'promo-grid__card--primary',
  },
  {
    title: 'Thời trang mới',
    subtitle: 'Bộ sưu tập hè 2026',
    to: '/products?sort=newest',
    tag: 'NEW',
    icon: 'mdi-tshirt-crew',
    variant: 'promo-grid__card--secondary',
  },
  {
    title: 'Gia dụng tiện ích',
    subtitle: 'Freeship toàn quốc',
    to: '/products',
    tag: 'HOT',
    icon: 'mdi-sofa-outline',
    variant: 'promo-grid__card--primary-dark',
  },
  {
    title: 'Làm đẹp & Sức khỏe',
    subtitle: 'Combo tiết kiệm 25%',
    to: '/products?sort=newest',
    tag: 'DEAL',
    icon: 'mdi-spa-outline',
    variant: 'promo-grid__card--accent',
  },
]
</script>

<style scoped>
.promo-grid__inner {
  max-width: 1280px;
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  padding-top: 20px;
  padding-bottom: 8px;
}

@media (min-width: 768px) {
  .promo-grid__inner {
    grid-template-columns: repeat(4, 1fr);
  }
}

.promo-grid__card {
  position: relative;
  border-radius: var(--radius-lg);
  padding: 18px 20px;
  min-height: 110px;
  overflow: hidden;
  text-decoration: none;
  color: var(--color-text-inverse);
  display: flex;
  align-items: center;
  justify-content: space-between;
  transition: transform 0.25s ease, box-shadow 0.25s ease;
  box-shadow: var(--shadow-md);
}

.promo-grid__card--primary {
  background: var(--gradient-primary);
}

.promo-grid__card--secondary {
  background: var(--gradient-secondary);
}

.promo-grid__card--primary-dark {
  background: linear-gradient(135deg, var(--color-primary-dark) 0%, var(--color-primary) 100%);
}

.promo-grid__card--accent {
  background: linear-gradient(135deg, #c62828 0%, #ef5350 100%);
}

.promo-grid__card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lg);
}

.promo-grid__title {
  font-size: 15px;
  font-weight: 800;
  line-height: 1.2;
}

@media (min-width: 960px) {
  .promo-grid__title {
    font-size: 17px;
  }
}

.promo-grid__subtitle {
  font-size: 12px;
  opacity: 0.9;
  margin-top: 4px;
}

.promo-grid__cta {
  display: inline-block;
  margin-top: 8px;
  font-size: 11px;
  font-weight: 700;
  opacity: 0.95;
}

.promo-grid__icon {
  opacity: 0.25;
  flex-shrink: 0;
}
</style>
