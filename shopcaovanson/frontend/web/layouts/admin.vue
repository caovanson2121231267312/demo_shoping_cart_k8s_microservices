<template>
  <v-app class="admin-app">
    <div
      ref="shellRef"
      class="admin-shell"
      :class="shellClasses"
      :style="shellStyle"
    >
      <div
        v-if="mobileDrawer && prefs.navPosition === 'sidebar'"
        class="admin-sidebar-overlay d-lg-none"
        @click="mobileDrawer = false"
      />

      <!-- Sidebar (sidebar mode only) -->
      <aside
        v-if="prefs.navPosition === 'sidebar'"
        class="admin-sidebar"
        :class="{ 'admin-sidebar--open': mobileDrawer }"
        :style="[shellStyle, sidebarChromeStyle]"
      >
        <div class="admin-sidebar__brand">
          <div class="admin-sidebar__logo">
            <v-icon color="white" size="22">mdi-store-cog</v-icon>
          </div>
          <div class="admin-sidebar__brand-text">
            <span class="admin-sidebar__brand-name">Shop CVS</span>
            <span class="admin-sidebar__brand-sub">Bảng quản trị</span>
          </div>
        </div>

        <nav class="admin-sidebar__nav">
          <div class="admin-sidebar__section">Menu chính</div>
          <AdminNavLinks
            :items="visibleNav"
            :is-active="isActive"
            @navigate="mobileDrawer = false"
          />
        </nav>

        <div class="admin-sidebar__footer">
          <div class="admin-sidebar__user">
            <v-avatar size="36" class="admin-sidebar__avatar">
              <span class="text-white text-caption font-weight-bold">{{ userInitials }}</span>
            </v-avatar>
            <div class="admin-sidebar__user-info overflow-hidden">
              <div class="text-body-2 font-weight-bold text-truncate">{{ auth.user.value?.full_name }}</div>
              <div class="text-caption opacity-70 text-truncate">{{ auth.user.value?.email }}</div>
            </div>
          </div>
          <NuxtLink to="/" class="admin-nav-item" @click="mobileDrawer = false">
            <span class="admin-nav-item__icon">
              <v-icon icon="mdi-storefront-outline" size="20" />
            </span>
            <span class="admin-nav-item__label">Về cửa hàng</span>
          </NuxtLink>
          <button type="button" class="admin-nav-item w-100" @click="toggleSidebarCollapsed()">
            <span class="admin-nav-item__icon">
              <v-icon :icon="prefs.sidebarCollapsed ? 'mdi-chevron-double-right' : 'mdi-chevron-double-left'" size="20" />
            </span>
            <span class="admin-nav-item__label">{{ prefs.sidebarCollapsed ? 'Mở rộng' : 'Thu gọn' }}</span>
          </button>
        </div>
      </aside>

      <div class="admin-main">
        <header class="admin-header">
          <div class="admin-header__row">
            <v-btn
              v-if="prefs.navPosition === 'sidebar'"
              class="d-lg-none"
              icon="mdi-menu"
              variant="text"
              @click="mobileDrawer = true"
            />

            <!-- Top nav brand (top mode) -->
            <div v-if="prefs.navPosition === 'top'" class="admin-header__brand d-none d-sm-flex">
              <div class="admin-sidebar__logo admin-sidebar__logo--sm">
                <v-icon color="white" size="18">mdi-store-cog</v-icon>
              </div>
              <span class="admin-header__brand-name">Shop CVS</span>
            </div>

            <div class="admin-header__info">
              <div class="admin-header__title">{{ pageTitle }}</div>
              <div v-if="prefs.showBreadcrumb" class="admin-header__breadcrumb">
                {{ route.meta.adminSubtitle || 'Quản trị hệ thống' }}
              </div>
            </div>

            <div class="admin-header__actions">
              <v-btn
                icon="mdi-palette-outline"
                variant="text"
                :color="prefs.headerStyle === 'colored' ? 'white' : 'primary'"
                title="Tùy chỉnh giao diện"
                @click="customizeOpen = true"
              />
              <v-chip
                v-if="auth.user.value?.role"
                size="small"
                variant="tonal"
                class="d-none d-md-inline-flex text-capitalize"
                :style="{ color: theme.primary, background: `${theme.primary}18` }"
              >
                {{ auth.user.value.role.replace('_', ' ') }}
              </v-chip>
              <v-btn
                variant="tonal"
                :color="prefs.headerStyle === 'colored' ? 'white' : 'primary'"
                size="small"
                class="text-none d-none d-sm-inline-flex"
                to="/"
              >
                Cửa hàng
              </v-btn>
            </div>
          </div>

          <!-- Top horizontal navbar -->
          <nav v-if="prefs.navPosition === 'top'" class="admin-topnav">
            <div class="admin-topnav__scroll">
              <AdminNavLinks
                :items="visibleNav"
                :is-active="isActive"
              />
            </div>
          </nav>
        </header>

        <main class="admin-content admin-content--boxed">
          <slot />
        </main>
      </div>

      <v-btn
        class="admin-fab-customize d-lg-none"
        icon="mdi-palette"
        color="primary"
        size="large"
        elevation="4"
        @click="customizeOpen = true"
      />

      <AdminCustomizePanel v-model="customizeOpen" />
    </div>

    <AdminToast />
  </v-app>
</template>

<script setup lang="ts">
const authStore = useAuthStore()
const auth = useAuth()
const admin = useAdmin()
const {
  shellStyle,
  shellClasses,
  sidebarChromeStyle,
  theme,
  prefs,
  mobileDrawer,
  customizeOpen,
  init,
  applyThemeVars,
  toggleSidebarCollapsed,
} = useAdminLayout()
const route = useRoute()
const shellRef = ref<HTMLElement | null>(null)

interface NavItem {
  to: string
  title: string
  icon: string
  show: boolean
}

const navItems = computed<NavItem[]>(() => [
  { to: '/admin', title: 'Tổng quan', icon: 'mdi-view-dashboard-outline', show: true },
  { to: '/admin/chat', title: 'Chat hỗ trợ', icon: 'mdi-headset', show: authStore.isStaff },
  { to: '/admin/reports', title: 'Báo cáo', icon: 'mdi-chart-areaspline', show: authStore.can('manager') },
  { to: '/admin/products', title: 'Sản phẩm', icon: 'mdi-package-variant-closed', show: admin.canManageProducts.value },
  { to: '/admin/categories', title: 'Danh mục', icon: 'mdi-shape-outline', show: admin.canManageProducts.value },
  { to: '/admin/articles', title: 'Bài viết', icon: 'mdi-post-outline', show: authStore.can('manager') },
  { to: '/admin/orders', title: 'Đơn hàng', icon: 'mdi-clipboard-list-outline', show: admin.canManageOrders.value },
  { to: '/admin/game', title: 'Game Caro', icon: 'mdi-gamepad-variant-outline', show: authStore.isStaff },
  { to: '/admin/coupons', title: 'Mã giảm giá', icon: 'mdi-ticket-percent-outline', show: admin.canManageOrders.value },
  { to: '/admin/users', title: 'Người dùng', icon: 'mdi-account-group-outline', show: admin.canViewUsers.value },
])

const visibleNav = computed(() => navItems.value.filter((i) => i.show))

const pageTitle = computed(() => {
  const match = navItems.value.find((i) => isActive(i.to))
  return match?.title ?? 'Admin'
})

const userInitials = computed(() => {
  const name = auth.user.value?.full_name?.trim() || auth.user.value?.email || '?'
  const parts = name.split(/\s+/)
  if (parts.length >= 2) {
    return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase()
  }
  return name.slice(0, 2).toUpperCase()
})

function isActive(path: string) {
  if (path === '/admin') {
    return route.path === '/admin'
  }
  return route.path.startsWith(path)
}

onMounted(async () => {
  init()
  await nextTick()
  applyThemeVars(shellRef.value)
  await auth.ensureAuth()
})

watch([shellStyle, () => prefs.value.themeId, () => prefs.value.sidebarStyle], () => {
  nextTick(() => applyThemeVars(shellRef.value))
}, { deep: true })
</script>

<style>
@import '~/assets/css/admin.css';

.admin-app .v-application__wrap {
  min-height: 100vh;
}
</style>
