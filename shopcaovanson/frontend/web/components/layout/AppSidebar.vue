<template>
  <v-navigation-drawer
    :model-value="modelValue"
    temporary
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <v-list nav density="comfortable">
      <v-list-item prepend-icon="mdi-home" title="Trang chủ" to="/" />
      <v-list-item prepend-icon="mdi-shopping" title="Sản phẩm" to="/products" />
      <v-list-item prepend-icon="mdi-cart" title="Giỏ hàng" to="/cart" />
      <v-list-item prepend-icon="mdi-heart-outline" title="Yêu thích" to="/wishlist" />
      <v-list-item prepend-icon="mdi-package-variant" title="Đơn hàng" to="/orders" />
      <v-list-item prepend-icon="mdi-truck-fast" title="Tra cứu đơn" to="/orders/track" />
      <v-list-item prepend-icon="mdi-truck-fast" title="Tra cứu đơn" to="/orders/track" />
      <v-list-item prepend-icon="mdi-account" title="Hồ sơ" to="/profile" />
      <v-list-item
        v-if="auth.isStaff.value"
        prepend-icon="mdi-view-dashboard"
        title="Quản trị"
        to="/admin"
      />
    </v-list>
  </v-navigation-drawer>
</template>

<script setup lang="ts">
defineProps<{ modelValue: boolean }>()
defineEmits<{ 'update:modelValue': [value: boolean] }>()

const auth = useAuth()
</script>
