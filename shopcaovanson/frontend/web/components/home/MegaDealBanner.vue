<template>
  <section class="mega-deals">
    <v-container class="mega-deals__inner">
      <NuxtLink
        v-for="deal in deals"
        :key="deal.title"
        :to="deal.to"
        class="mega-deals__card"
        :style="{ background: deal.gradient }"
      >
        <div class="mega-deals__content">
          <v-chip v-if="deal.tag" size="x-small" color="white" variant="flat" class="mb-2 font-weight-bold">
            {{ deal.tag }}
          </v-chip>
          <div class="mega-deals__title">{{ deal.title }}</div>
          <div class="mega-deals__subtitle">{{ deal.subtitle }}</div>
          <span class="mega-deals__cta">Khám phá ngay →</span>
        </div>
        <v-icon class="mega-deals__icon" size="80" color="white">{{ deal.icon }}</v-icon>
      </NuxtLink>
    </v-container>
  </section>
</template>

<script setup lang="ts">
const deals = [
  {
    title: 'Siêu deal điện tử',
    subtitle: 'Laptop, điện thoại, tai nghe — giảm đến 45%',
    to: '/products?sort=price_desc',
    tag: 'TOP DEAL',
    icon: 'mdi-laptop',
    gradient: 'linear-gradient(135deg, #0d47a1 0%, #1976d2 100%)',
  },
  {
    title: 'Làm đẹp & Sức khỏe',
    subtitle: 'Mỹ phẩm, vitamin — combo tiết kiệm 30%',
    to: '/products?sort=newest',
    tag: 'BEST SELLER',
    icon: 'mdi-spa',
    gradient: 'linear-gradient(135deg, #e65100 0%, #f57c00 100%)',
  },
]
</script>

<style scoped>
.mega-deals__inner {
  max-width: 1280px;
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
  padding-top: 8px;
  padding-bottom: 8px;
}

@media (min-width: 768px) {
  .mega-deals__inner {
    grid-template-columns: repeat(2, 1fr);
  }
}

.mega-deals__card {
  position: relative;
  border-radius: var(--radius-lg);
  padding: 24px 28px;
  min-height: 140px;
  overflow: hidden;
  text-decoration: none;
  color: var(--color-text-inverse);
  display: flex;
  align-items: center;
  justify-content: space-between;
  transition: transform 0.25s ease, box-shadow 0.25s ease;
  box-shadow: var(--shadow-md);
}

.mega-deals__card:hover {
  transform: translateY(-3px);
  box-shadow: var(--shadow-lg);
}

.mega-deals__title {
  font-size: 20px;
  font-weight: 800;
  line-height: 1.2;
}

.mega-deals__subtitle {
  font-size: 13px;
  opacity: 0.92;
  margin-top: 6px;
  max-width: 260px;
}

.mega-deals__cta {
  display: inline-block;
  margin-top: 12px;
  font-size: 12px;
  font-weight: 700;
}

.mega-deals__icon {
  opacity: 0.2;
  flex-shrink: 0;
}
</style>
