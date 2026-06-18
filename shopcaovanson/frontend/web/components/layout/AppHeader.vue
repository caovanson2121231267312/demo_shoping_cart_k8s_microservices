<template>
  <v-app-bar color="primary" elevation="0" density="comfortable" class="app-header">
    <v-app-bar-nav-icon class="d-lg-none" color="white" @click="$emit('toggle-sidebar')" />

    <NuxtLink to="/" class="app-header__brand text-white d-flex align-center ml-1">
      <v-icon class="mr-2" size="28">mdi-storefront</v-icon>
      <span class="text-h6 font-weight-bold d-none d-sm-inline">Shop Cao Van Son</span>
    </NuxtLink>

    <div class="d-none d-md-flex flex-grow-1 mx-4" style="max-width: 520px">
      <ProductSearch />
    </div>

    <v-spacer class="d-md-none" />

    <v-btn icon variant="text" color="white" class="d-md-none" to="/products">
      <v-icon>mdi-magnify</v-icon>
    </v-btn>

    <v-btn icon variant="text" color="white" to="/wishlist">
      <v-badge :content="wishlist.count.value" :model-value="wishlist.count.value > 0" color="error">
        <v-icon>mdi-heart-outline</v-icon>
      </v-badge>
    </v-btn>

    <v-btn icon variant="text" color="white" @click="cartDrawerOpen = true">
      <v-badge :content="cart.count.value" :model-value="cart.count.value > 0" color="secondary">
        <v-icon>mdi-cart-outline</v-icon>
      </v-badge>
    </v-btn>

    <template v-if="auth.isLoggedIn.value">
      <v-menu>
        <template #activator="{ props: menuProps }">
          <v-btn v-bind="menuProps" icon variant="text" color="white" class="ml-1">
            <v-avatar color="secondary" size="36">
              <span class="text-white text-body-2">{{ initials }}</span>
            </v-avatar>
          </v-btn>
        </template>
        <v-list density="compact" min-width="220">
          <v-list-item>
            <v-list-item-title>{{ auth.user.value?.full_name }}</v-list-item-title>
            <v-list-item-subtitle>{{ auth.user.value?.email }}</v-list-item-subtitle>
          </v-list-item>
          <v-divider />
          <v-list-item to="/profile" prepend-icon="mdi-account-outline" title="Hồ sơ" />
          <v-list-item to="/orders" prepend-icon="mdi-package-variant" title="Đơn hàng" />
          <v-list-item to="/orders/track" prepend-icon="mdi-truck-fast" title="Tra cứu đơn" />
          <v-list-item to="/wishlist" prepend-icon="mdi-heart-outline" title="Yêu thích" />
          <v-list-item
            v-if="auth.isStaff.value"
            to="/admin"
            prepend-icon="mdi-view-dashboard"
            title="Quản trị"
          />
          <v-list-item prepend-icon="mdi-logout" title="Đăng xuất" @click="handleLogout" />
        </v-list>
      </v-menu>
    </template>
    <template v-else>
      <v-btn variant="text" color="white" to="/auth/login" class="d-none d-sm-inline-flex">Đăng nhập</v-btn>
      <v-btn variant="outlined" color="white" to="/auth/register" class="mr-2 text-none">Đăng ký</v-btn>
    </template>
  </v-app-bar>
</template>

<script setup lang="ts">
defineEmits<{ 'toggle-sidebar': [] }>()

const auth = useAuth()
const cart = useCart()
const wishlist = useWishlist()
const cartDrawerOpen = useState('cartDrawerOpen', () => false)

onMounted(() => useWishlistStore().hydrate())

const initials = computed(() => {
  const name = auth.user.value?.full_name || ''
  return name.split(' ').map((p) => p[0]).join('').slice(0, 2).toUpperCase()
})

const handleLogout = async () => {
  await auth.logout()
  await navigateTo('/')
}
</script>

<style scoped>
.app-header {
  background: linear-gradient(135deg, #1565c0 0%, #0d47a1 100%) !important;
}

.app-header__brand {
  text-decoration: none;
}
</style>
