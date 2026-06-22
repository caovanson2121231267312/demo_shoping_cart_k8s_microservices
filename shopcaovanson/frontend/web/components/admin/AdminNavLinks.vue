<template>

  <NuxtLink

    v-for="item in items"

    :key="item.to"

    :to="item.to"

    class="admin-nav-item"

    :class="{ 'admin-nav-item--active': isActive(item.to) }"

    @click="$emit('navigate')"

  >

    <span v-if="isActive(item.to)" class="admin-nav-item__indicator" aria-hidden="true" />

    <span class="admin-nav-item__icon">

      <v-icon :icon="item.icon" size="20" />

    </span>

    <span class="admin-nav-item__label">{{ item.title }}</span>

  </NuxtLink>

</template>



<script setup lang="ts">

export interface AdminNavLinkItem {

  to: string

  title: string

  icon: string

}



defineProps<{

  items: AdminNavLinkItem[]

  isActive: (path: string) => boolean

}>()



defineEmits<{ navigate: [] }>()

</script>

