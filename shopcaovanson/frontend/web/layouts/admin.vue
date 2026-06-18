<template>
  <v-app>
    <v-navigation-drawer v-model="drawer" permanent :rail="rail" color="grey-darken-4">
      <div class="pa-4 d-flex align-center">
        <v-icon color="primary" size="28" class="mr-2">mdi-store-cog</v-icon>
        <span v-if="!rail" class="text-white font-weight-bold">Admin Panel</span>
      </div>
      <v-divider class="border-opacity-25" />
      <v-list density="compact" nav class="text-grey-lighten-1">
        <v-list-item to="/admin" prepend-icon="mdi-view-dashboard" title="Tổng quan" />
        <v-list-item
          v-if="admin.canManageProducts.value"
          to="/admin/products"
          prepend-icon="mdi-package-variant"
          title="Sản phẩm"
        />
        <v-list-item
          v-if="admin.canManageProducts.value"
          to="/admin/categories"
          prepend-icon="mdi-shape"
          title="Danh mục"
        />
        <v-list-item
          v-if="admin.canManageOrders.value"
          to="/admin/orders"
          prepend-icon="mdi-clipboard-list"
          title="Đơn hàng"
        />
        <v-list-item
          v-if="admin.canManageOrders.value"
          to="/admin/coupons"
          prepend-icon="mdi-ticket-percent"
          title="Mã giảm giá"
        />
        <v-list-item
          v-if="admin.canViewUsers.value"
          to="/admin/users"
          prepend-icon="mdi-account-group"
          title="Người dùng"
        />
      </v-list>
      <template #append>
        <v-list density="compact" nav>
          <v-list-item prepend-icon="mdi-storefront" title="Về cửa hàng" to="/" />
          <v-list-item prepend-icon="mdi-chevron-left" :title="rail ? 'Mở rộng' : 'Thu gọn'" @click="rail = !rail" />
        </v-list>
      </template>
    </v-navigation-drawer>

    <v-app-bar color="white" elevation="1" density="comfortable">
      <v-app-bar-title class="text-h6 font-weight-bold text-primary">
        Quản trị Shop Cao Văn Sơn
      </v-app-bar-title>
      <v-spacer />
      <span class="text-body-2 text-grey mr-4 d-none d-sm-inline">{{ auth.user.value?.email }}</span>
      <v-btn variant="text" color="primary" to="/">Cửa hàng</v-btn>
    </v-app-bar>

    <v-main class="bg-grey-lighten-4">
      <slot />
    </v-main>
    <AppSnackbar />
  </v-app>
</template>

<script setup lang="ts">
const auth = useAuth()
const admin = useAdmin()
const drawer = ref(true)
const rail = ref(false)

onMounted(() => auth.ensureAuth())
</script>
