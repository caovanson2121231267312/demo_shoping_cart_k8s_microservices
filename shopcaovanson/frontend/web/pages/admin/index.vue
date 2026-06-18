<template>
  <v-container class="page-container py-6">
    <h1 class="text-h4 font-weight-bold mb-6">Bảng điều khiển</h1>

    <LoadingSpinner v-if="loading" />

    <v-row v-else>
      <v-col cols="12" sm="6" md="3">
        <v-card color="primary" variant="tonal">
          <v-card-text>
            <div class="text-overline">Người dùng</div>
            <div class="text-h4 font-weight-bold">{{ stats.totalUsers }}</div>
          </v-card-text>
        </v-card>
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <v-card color="info" variant="tonal">
          <v-card-text>
            <div class="text-overline">Đơn hàng</div>
            <div class="text-h4 font-weight-bold">{{ stats.totalOrders }}</div>
          </v-card-text>
        </v-card>
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <v-card color="success" variant="tonal">
          <v-card-text>
            <div class="text-overline">Doanh thu</div>
            <div class="text-h6 font-weight-bold">{{ formatVND(stats.revenue) }}</div>
          </v-card-text>
        </v-card>
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <v-card color="secondary" variant="tonal">
          <v-card-text>
            <div class="text-overline">Sản phẩm</div>
            <div class="text-h4 font-weight-bold">{{ stats.totalProducts }}</div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-row class="mt-4">
      <v-col v-if="admin.canManageProducts.value" cols="12" sm="6" md="4">
        <v-card :to="'/admin/products'" hover>
          <v-card-title><v-icon class="mr-2">mdi-package-variant</v-icon>Sản phẩm</v-card-title>
        </v-card>
      </v-col>
      <v-col v-if="admin.canManageProducts.value" cols="12" sm="6" md="4">
        <v-card :to="'/admin/categories'" hover>
          <v-card-title><v-icon class="mr-2">mdi-shape</v-icon>Danh mục</v-card-title>
        </v-card>
      </v-col>
      <v-col v-if="admin.canManageOrders.value" cols="12" sm="6" md="4">
        <v-card :to="'/admin/orders'" hover>
          <v-card-title><v-icon class="mr-2">mdi-clipboard-list</v-icon>Đơn hàng</v-card-title>
        </v-card>
      </v-col>
      <v-col v-if="admin.canManageOrders.value" cols="12" sm="6" md="4">
        <v-card :to="'/admin/coupons'" hover>
          <v-card-title><v-icon class="mr-2">mdi-ticket-percent</v-icon>Mã giảm giá</v-card-title>
        </v-card>
      </v-col>
      <v-col v-if="admin.canViewUsers.value" cols="12" sm="6" md="4">
        <v-card :to="'/admin/users'" hover>
          <v-card-title><v-icon class="mr-2">mdi-account-group</v-icon>Người dùng</v-card-title>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin' })

const admin = useAdmin()
const authStore = useAuthStore()
const { fetchProducts } = useProducts()
const { formatVND } = useFormat()

const loading = ref(true)
const stats = reactive({
  totalUsers: 0,
  totalOrders: 0,
  revenue: 0,
  totalProducts: 0,
})

onMounted(async () => {
  try {
    const tasks: Promise<void>[] = [
      fetchProducts({ limit: 1 }).then((r) => { stats.totalProducts = r.total }),
    ]
    if (admin.canViewUsers.value) {
      tasks.push(admin.fetchUserStats().then((r) => { stats.totalUsers = r.total_users }))
    }
    if (authStore.can('manager')) {
      tasks.push(admin.fetchOrderStats().then((r) => {
        stats.totalOrders = r.total_orders
        stats.revenue = r.revenue
      }))
    }
    await Promise.all(tasks)
  } finally {
    loading.value = false
  }
})
</script>
