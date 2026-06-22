<template>
  <v-navigation-drawer
    :model-value="modelValue"
    temporary
    width="300"
    class="mobile-sidebar"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <div class="mobile-sidebar__header">
      <NuxtLink to="/" class="mobile-sidebar__brand" @click="close">
        <div class="mobile-sidebar__logo">
          <v-icon color="white" size="24">mdi-storefront-outline</v-icon>
        </div>
        <div>
          <div class="mobile-sidebar__name">Shop Cao Văn Sơn</div>
          <div class="mobile-sidebar__tagline">Thương mại điện tử</div>
        </div>
      </NuxtLink>
    </div>

    <div v-if="!isLoggedIn" class="mobile-sidebar__auth px-4 py-3">
      <v-btn block color="primary" to="/auth/login" class="text-none mb-2" @click="close">
        Đăng nhập
      </v-btn>
      <v-btn block variant="outlined" color="primary" to="/auth/register" class="text-none" @click="close">
        Đăng ký tài khoản
      </v-btn>
    </div>

    <v-list nav density="comfortable" class="mobile-sidebar__nav">
      <v-list-subheader class="font-weight-bold">Menu</v-list-subheader>
      <v-list-item prepend-icon="mdi-home-outline" title="Trang chủ" to="/" @click="close" />
      <v-list-item prepend-icon="mdi-shopping-outline" title="Sản phẩm" to="/products" @click="close" />
      <v-list-item prepend-icon="mdi-post-outline" title="Tin tức" to="/blog" @click="close" />
      <v-list-item prepend-icon="mdi-flash" title="Flash Sale" :to="'/products?sort=price_desc'" @click="close" />
      <v-list-item prepend-icon="mdi-cart-outline" title="Giỏ hàng" to="/cart" @click="close" />
      <v-list-item prepend-icon="mdi-heart-outline" title="Yêu thích" to="/wishlist" @click="close" />

      <v-divider class="my-2" />

      <v-list-subheader class="font-weight-bold">Tài khoản</v-list-subheader>
      <v-list-item prepend-icon="mdi-package-variant-closed" title="Đơn hàng" to="/orders" @click="close" />
      <v-list-item prepend-icon="mdi-truck-fast-outline" title="Tra cứu đơn" to="/orders/track" @click="close" />
      <v-list-item prepend-icon="mdi-account-outline" title="Hồ sơ" to="/profile" @click="close" />
      <v-list-item
        v-if="isStaff"
        prepend-icon="mdi-view-dashboard-outline"
        title="Quản trị"
        to="/admin"
        @click="close"
      />

      <template v-if="categories.length">
        <v-divider class="my-2" />
        <v-list-subheader class="font-weight-bold">Danh mục</v-list-subheader>
        <v-list-item
          v-for="cat in categories"
          :key="cat.id"
          :prepend-icon="categoryIcon(cat.slug)"
          :title="cat.name"
          :to="`/products?category=${cat.slug}`"
          @click="close"
        />
      </template>
    </v-list>

    <template v-if="isLoggedIn" #append>
      <div class="pa-4">
        <v-btn block variant="tonal" color="error" prepend-icon="mdi-logout" class="text-none" @click="handleLogout">
          Đăng xuất
        </v-btn>
      </div>
    </template>
  </v-navigation-drawer>
</template>

<script setup lang="ts">
import type { Category } from '~/types'
import { categoryIcon } from '~/utils/categoryIcons'

const props = defineProps<{ modelValue: boolean }>()
const emit = defineEmits<{ 'update:modelValue': [value: boolean] }>()

const auth = useAuth()
const { isLoggedIn, isStaff } = auth
const categories = ref<Category[]>([])
const { fetchCategories } = useProducts()

const close = () => emit('update:modelValue', false)

onMounted(async () => {
  try {
    categories.value = (await fetchCategories()).filter((c) => !c.parent_id)
  } catch {
    categories.value = []
  }
})

const handleLogout = async () => {
  close()
  await auth.logout()
  await navigateTo('/')
  await auth.ensureAuth()
}
</script>

<style scoped>
.mobile-sidebar__header {
  background: var(--gradient-primary);
  padding: 20px 16px;
}

.mobile-sidebar__brand {
  display: flex;
  align-items: center;
  gap: 12px;
  text-decoration: none;
  color: var(--color-text-inverse);
}

.mobile-sidebar__logo {
  width: 44px;
  height: 44px;
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.15);
  display: flex;
  align-items: center;
  justify-content: center;
}

.mobile-sidebar__name {
  font-size: 18px;
  font-weight: 800;
}

.mobile-sidebar__tagline {
  font-size: 12px;
  opacity: 0.85;
}

.mobile-sidebar__auth {
  background: var(--color-surface-muted);
  border-bottom: 1px solid var(--color-border);
}
</style>
