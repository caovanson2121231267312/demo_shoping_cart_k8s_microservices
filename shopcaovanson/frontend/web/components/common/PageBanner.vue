<template>
  <div class="page-banner" :class="{ 'page-banner--compact': compact }">
    <v-container class="page-banner__inner">
      <nav v-if="breadcrumbs.length" class="page-banner__breadcrumb d-flex align-center flex-wrap ga-1 mb-2">
        <NuxtLink to="/" class="page-banner__crumb">Trang chủ</NuxtLink>
        <template v-for="(crumb, i) in breadcrumbs" :key="i">
          <v-icon size="14" color="grey">mdi-chevron-right</v-icon>
          <NuxtLink v-if="crumb.to" :to="crumb.to" class="page-banner__crumb">{{ crumb.label }}</NuxtLink>
          <span v-else class="page-banner__crumb page-banner__crumb--current">{{ crumb.label }}</span>
        </template>
      </nav>
      <h1 v-if="title" class="page-banner__title">{{ title }}</h1>
      <p v-if="subtitle" class="page-banner__subtitle mb-0">{{ subtitle }}</p>
    </v-container>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title?: string
  subtitle?: string
  compact?: boolean
  breadcrumbs?: { label: string; to?: string }[]
}>()
</script>

<style scoped>
.page-banner {
  background: linear-gradient(135deg, #e3f2fd 0%, #f5f5f5 100%);
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
  padding: 28px 0;
}

.page-banner--compact {
  padding: 16px 0;
}

.page-banner__inner {
  max-width: 1280px;
}

.page-banner__breadcrumb {
  font-size: 13px;
}

.page-banner__crumb {
  color: #616161;
  text-decoration: none;
}

.page-banner__crumb:hover {
  color: #1565c0;
}

.page-banner__crumb--current {
  color: #212121;
  font-weight: 600;
}

.page-banner__title {
  font-size: 1.75rem;
  font-weight: 800;
  color: #1565c0;
  line-height: 1.2;
}

.page-banner__subtitle {
  font-size: 14px;
  color: #757575;
  margin-top: 6px;
}
</style>
