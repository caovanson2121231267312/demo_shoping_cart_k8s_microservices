<template>
  <section class="brand-showcase">
    <v-container class="page-container">
      <div class="section-header text-center mb-6">
        <h2 class="section-title">Thương hiệu nổi bật</h2>
        <p class="section-subtitle">Hàng chính hãng — bảo hành đầy đủ</p>
      </div>
      <div class="brand-showcase__grid">
        <NuxtLink
          v-for="brand in brands"
          :key="brand.name"
          :to="brand.to"
          class="brand-showcase__item"
        >
          <v-avatar :color="brand.color" size="48" class="mb-2">
            <span class="text-white font-weight-bold text-caption">{{ brand.initials }}</span>
          </v-avatar>
          <span class="brand-showcase__name">{{ brand.name }}</span>
        </NuxtLink>
      </div>
    </v-container>
  </section>
</template>

<script setup lang="ts">
const brands = [
  { name: 'Samsung', initials: 'SS', color: 'blue-darken-2', to: '/products?search=samsung' },
  { name: 'Apple', initials: 'AP', color: 'grey-darken-3', to: '/products?search=apple' },
  { name: 'Xiaomi', initials: 'XM', color: 'orange-darken-2', to: '/products?search=xiaomi' },
  { name: 'Sony', initials: 'SN', color: 'indigo-darken-1', to: '/products?search=sony' },
  { name: 'LG', initials: 'LG', color: 'red-darken-2', to: '/products?search=lg' },
  { name: 'Nike', initials: 'NK', color: 'grey-darken-4', to: '/products?search=nike' },
  { name: 'Adidas', initials: 'AD', color: 'blue-grey-darken-2', to: '/products?search=adidas' },
  { name: 'Unilever', initials: 'UL', color: 'cyan-darken-2', to: '/products?search=unilever' },
  { name: 'Philips', initials: 'PH', color: 'blue-darken-3', to: '/products?search=philips' },
  { name: 'Dell', initials: 'DL', color: 'blue-darken-4', to: '/products?search=dell' },
]
</script>

<style scoped>
.brand-showcase {
  background: var(--color-surface);
  padding: 32px 0;
  border-top: 1px solid var(--color-border);
  border-bottom: 1px solid var(--color-border);
}

.brand-showcase__grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 12px;
}

@media (max-width: 599px) {
  .brand-showcase__grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

.brand-showcase__item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 16px 8px;
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border-light);
  text-decoration: none;
  transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease;
}

.brand-showcase__item:hover {
  transform: translateY(-3px);
  box-shadow: var(--shadow-md);
  border-color: var(--color-primary-alpha-35);
}

.brand-showcase__name {
  font-size: 12px;
  font-weight: 600;
  color: var(--color-text);
  text-align: center;
}
</style>
