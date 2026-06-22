<template>
  <header class="store-header">
    <v-container class="store-header__main">
      <v-app-bar-nav-icon
        class="store-header__menu d-lg-none flex-shrink-0"
        color="primary"
        @click="$emit('toggle-sidebar')"
      />

      <NuxtLink to="/" class="store-header__brand flex-shrink-0">
        <div class="store-header__logo">
          <v-icon color="white" size="28">mdi-storefront-outline</v-icon>
        </div>
        <div class="store-header__brand-text">
          <span class="store-header__name">Shop Cao Văn Sơn</span>
          <span class="store-header__tagline">Thương mại điện tử uy tín</span>
        </div>
      </NuxtLink>

      <div class="store-header__search d-none d-md-flex">
        <ProductSearch v-model="searchQuery" show-button @search="onSearch" />
      </div>

      <v-spacer class="d-md-none" />

      <div class="store-header__actions">
        <a href="tel:19001234" class="store-header__hotline d-none d-xl-flex">
          <div class="store-header__hotline-icon">
            <v-icon color="primary" size="22">mdi-headset</v-icon>
          </div>
          <div>
            <span class="store-header__hotline-label">Hỗ trợ / Mua hàng</span>
            <span class="store-header__hotline-number">1900 1234</span>
          </div>
        </a>

        <NuxtLink to="/wishlist" class="store-header__action d-none d-sm-flex">
          <v-badge
            :content="wishlistCount"
            :model-value="wishlistCount > 0"
            color="error"
            location="top end"
            offset-x="2"
            offset-y="2"
          >
            <v-icon size="26">mdi-heart-outline</v-icon>
          </v-badge>
          <span class="store-header__action-label">Yêu thích</span>
        </NuxtLink>

        <button type="button" class="store-header__action store-header__action--cart" @click="cartDrawerOpen = true">
          <v-badge
            :content="cartCount"
            :model-value="cartCount > 0"
            color="error"
            location="top end"
            offset-x="2"
            offset-y="2"
          >
            <v-icon size="26">mdi-cart-outline</v-icon>
          </v-badge>
          <span class="store-header__action-label d-none d-sm-inline">Giỏ hàng</span>
        </button>

        <template v-if="isLoggedIn">
          <v-menu location="bottom end" offset="8">
            <template #activator="{ props: menuProps }">
              <button type="button" class="store-header__action store-header__action--account" v-bind="menuProps">
                <v-avatar color="primary" size="38">
                  <span class="text-white text-caption font-weight-bold">{{ initials }}</span>
                </v-avatar>
                <span class="store-header__action-label d-none d-lg-inline">Tài khoản</span>
              </button>
            </template>
            <v-list density="compact" min-width="260" rounded="lg" elevation="4">
              <v-list-item class="py-3">
                <template #prepend>
                  <v-avatar color="primary" size="40">
                    <span class="text-white text-body-2 font-weight-bold">{{ initials }}</span>
                  </v-avatar>
                </template>
                <v-list-item-title class="font-weight-bold">{{ user?.full_name }}</v-list-item-title>
                <v-list-item-subtitle class="text-truncate">{{ user?.email }}</v-list-item-subtitle>
              </v-list-item>
              <v-divider class="my-1" />
              <v-list-item to="/profile" prepend-icon="mdi-account-outline" title="Hồ sơ của tôi" />
              <v-list-item to="/orders" prepend-icon="mdi-package-variant-closed" title="Đơn hàng" />
              <v-list-item to="/orders/track" prepend-icon="mdi-truck-fast-outline" title="Tra cứu đơn hàng" />
              <v-list-item to="/wishlist" prepend-icon="mdi-heart-outline" title="Sản phẩm yêu thích" />
              <v-list-item
                v-if="isStaff"
                to="/admin"
                prepend-icon="mdi-view-dashboard-outline"
                title="Quản trị hệ thống"
              />
              <v-divider class="my-1" />
              <v-list-item prepend-icon="mdi-logout" title="Đăng xuất" base-color="error" @click="handleLogout" />
            </v-list>
          </v-menu>
        </template>
        <template v-else>
          <div class="store-header__auth d-none d-sm-flex align-center ga-1">
            <v-btn variant="text" color="grey-darken-3" to="/auth/login" class="text-none font-weight-medium" size="small">
              Đăng nhập
            </v-btn>
            <v-btn color="primary" to="/auth/register" class="text-none font-weight-bold" size="small" rounded="lg" elevation="0">
              Đăng ký
            </v-btn>
          </div>
          <v-btn icon variant="text" color="grey-darken-3" to="/auth/login" class="d-sm-none">
            <v-icon>mdi-account-outline</v-icon>
          </v-btn>
        </template>
      </div>
    </v-container>

    <div class="store-header__mobile-search d-md-none">
      <v-container class="py-2">
        <ProductSearch v-model="searchQuery" show-button @search="onSearch" />
      </v-container>
    </div>
  </header>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia'

defineEmits<{ 'toggle-sidebar': [] }>()

const auth = useAuth()
const cartStore = useCartStore()
const wishlistStore = useWishlistStore()
const { items: cartItems } = storeToRefs(cartStore)
const { items: wishlistItems } = storeToRefs(wishlistStore)
const router = useRouter()
const { isLoggedIn, user, isStaff } = auth
const cartDrawerOpen = useState('cartDrawerOpen', () => false)
const searchQuery = ref('')

const cartCount = computed(() =>
  cartItems.value.reduce((sum, item) => sum + item.quantity, 0),
)
const wishlistCount = computed(() => wishlistItems.value.length)

onMounted(() => {
  wishlistStore.hydrate()
  cartStore.hydrate()
})

const initials = computed(() => {
  const name = user.value?.full_name || ''
  return name.split(' ').map((p) => p[0]).join('').slice(0, 2).toUpperCase() || '?'
})

const onSearch = (query: string) => {
  router.push({
    path: '/products',
    query: query.trim() ? { search: query.trim() } : {},
  })
}

const handleLogout = async () => {
  await auth.logout()
  await navigateTo('/')
  await auth.ensureAuth()
}
</script>

<style scoped>
.store-header {
  background: var(--color-surface);
  box-shadow: var(--shadow-header);
  position: relative;
  z-index: 10;
}

.store-header__main {
  max-width: 1280px;
  min-height: 76px;
  display: flex;
  align-items: center;
  gap: 12px;
  padding-top: 10px;
  padding-bottom: 10px;
}

.store-header__menu {
  margin-left: -8px;
}

.store-header__brand {
  display: flex;
  align-items: center;
  gap: 12px;
  text-decoration: none;
  color: inherit;
  flex-shrink: 0;
}

.store-header__logo {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-md);
  background: var(--gradient-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: var(--shadow-primary);
  transition: transform 0.2s;
}

.store-header__brand:hover .store-header__logo {
  transform: scale(1.04);
}

.store-header__brand-text {
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.store-header__name {
  font-size: 20px;
  font-weight: 800;
  color: var(--color-primary);
  line-height: 1.15;
  letter-spacing: -0.3px;
}

.store-header__tagline {
  font-size: 11px;
  color: var(--color-text-muted);
  font-weight: 500;
}

@media (max-width: 599px) {
  .store-header__brand-text {
    display: none;
  }

  .store-header__logo {
    width: 42px;
    height: 42px;
  }
}

.store-header__search {
  flex: 1;
  max-width: 580px;
  margin: 0 20px;
  display: flex;
  align-items: center;
  min-width: 0;
}

.store-header__actions {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
  overflow: visible;
}

.store-header__hotline {
  display: flex;
  align-items: center;
  gap: 10px;
  text-decoration: none;
  color: inherit;
  padding: 6px 12px;
  border-radius: var(--radius-sm);
  margin-right: 8px;
  transition: background 0.2s;
}

.store-header__hotline:hover {
  background: var(--color-surface-hover);
}

.store-header__hotline-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: var(--color-primary-50);
  display: flex;
  align-items: center;
  justify-content: center;
}

.store-header__hotline-label {
  display: block;
  font-size: 11px;
  color: var(--color-text-muted);
  line-height: 1.2;
}

.store-header__hotline-number {
  display: block;
  font-size: 15px;
  font-weight: 800;
  color: var(--color-primary);
  line-height: 1.2;
}

.store-header__action {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  padding: 8px 12px;
  border: none;
  background: transparent;
  cursor: pointer;
  color: var(--color-text-secondary);
  text-decoration: none;
  border-radius: var(--radius-sm);
  transition: background 0.2s, color 0.2s;
  position: relative;
  overflow: visible;
}

.store-header__action:hover {
  background: var(--color-bg);
  color: var(--color-primary);
}

.store-header__action-label {
  font-size: 11px;
  font-weight: 600;
  white-space: nowrap;
}

.store-header__action--cart {
  position: relative;
}

.store-header__mobile-search {
  background: var(--color-surface-muted);
  border-top: 1px solid var(--color-border-light);
}

.store-header__mobile-search .product-search {
  width: 100%;
}
</style>
