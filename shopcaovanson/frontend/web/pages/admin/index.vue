<template>
  <div>
    <div class="admin-welcome">
      <h2 class="admin-welcome__title">Xin chào, {{ user?.full_name || 'Admin' }} 👋</h2>
      <p class="admin-welcome__text">
        Tổng quan hoạt động cửa hàng hôm nay. Theo dõi doanh thu, đơn hàng và quản lý nội dung từ một nơi.
      </p>
    </div>

    <LoadingSpinner v-if="loading" />

    <template v-else>
      <v-row class="mb-6">
        <v-col cols="12" sm="6" lg="3">
          <AdminStatCard
            label="Người dùng"
            :value="stats.totalUsers"
            icon="mdi-account-group-outline"
            color="var(--admin-primary)"
            subtitle="Tài khoản đã đăng ký"
          />
        </v-col>
        <v-col cols="12" sm="6" lg="3">
          <AdminStatCard
            label="Đơn hàng"
            :value="stats.totalOrders"
            icon="mdi-clipboard-check-outline"
            color="var(--admin-accent)"
            subtitle="Tổng đơn trong hệ thống"
          />
        </v-col>
        <v-col cols="12" sm="6" lg="3">
          <AdminStatCard
            label="Doanh thu"
            :value="formatVND(stats.revenue)"
            icon="mdi-cash-multiple"
            color="#059669"
            subtitle="Tích lũy"
          />
        </v-col>
        <v-col cols="12" sm="6" lg="3">
          <AdminStatCard
            label="Sản phẩm"
            :value="stats.totalProducts"
            icon="mdi-package-variant-closed"
            color="#f57c00"
            subtitle="Đang bán"
          />
        </v-col>
      </v-row>

      <AdminPageHeader title="Truy cập nhanh" subtitle="Đi tới các module quản trị thường dùng" />

      <v-row>
        <v-col v-if="admin.canManageProducts.value" cols="12" md="6" lg="4">
          <AdminQuickLink to="/admin/products" icon="mdi-package-variant-closed" title="Sản phẩm" description="Thêm, sửa, quản lý kho" />
        </v-col>
        <v-col v-if="admin.canManageProducts.value" cols="12" md="6" lg="4">
          <AdminQuickLink to="/admin/categories" icon="mdi-shape-outline" title="Danh mục" description="Cấu trúc ngành hàng" />
        </v-col>
        <v-col v-if="admin.canManageOrders.value" cols="12" md="6" lg="4">
          <AdminQuickLink to="/admin/orders" icon="mdi-clipboard-list-outline" title="Đơn hàng" description="Xử lý & theo dõi giao hàng" />
        </v-col>
        <v-col v-if="admin.canManageOrders.value" cols="12" md="6" lg="4">
          <AdminQuickLink to="/admin/coupons" icon="mdi-ticket-percent-outline" title="Mã giảm giá" description="Khuyến mãi & voucher" />
        </v-col>
        <v-col v-if="authStore.isStaff" cols="12" md="6" lg="4">
          <AdminQuickLink to="/admin/chat" icon="mdi-headset" title="Chat hỗ trợ" description="Khách online & nhắn tin trực tiếp" />
        </v-col>
        <v-col v-if="authStore.can('manager')" cols="12" md="6" lg="4">
          <AdminQuickLink to="/admin/reports" icon="mdi-chart-areaspline" title="Báo cáo chi tiết" description="Biểu đồ, Excel, online" />
        </v-col>
        <v-col v-if="authStore.can('manager')" cols="12" md="6" lg="4">
          <AdminQuickLink to="/admin/articles" icon="mdi-post-outline" title="Bài viết" description="Blog & nội dung SEO" />
        </v-col>
        <v-col v-if="admin.canViewUsers.value" cols="12" md="6" lg="4">
          <AdminQuickLink to="/admin/users" icon="mdi-account-cog-outline" title="Người dùng" description="Phân quyền & trạng thái" />
        </v-col>
      </v-row>

      <v-row class="mt-4">
        <v-col cols="12" lg="8">
          <v-card rounded="lg" class="pa-5">
            <div class="text-subtitle-1 font-weight-bold mb-1">Mẹo sử dụng</div>
            <p class="text-body-2 text-medium-emphasis mb-4">
              Nhấn biểu tượng palette trên header — panel có 4 tab: <strong>Theme</strong>, <strong>Navbar</strong> (sidebar/top + kiểu menu), <strong>Header</strong> (glass/màu/viền), <strong>Bố cục</strong> (khoảng cách gọn/cân bằng/thoáng).
            </p>
            <div class="d-flex flex-wrap ga-2">
              <v-chip color="primary" variant="tonal" size="small">Navbar sidebar / top</v-chip>
              <v-chip color="primary" variant="tonal" size="small">Menu pill / gạch / soft</v-chip>
              <v-chip color="primary" variant="tonal" size="small">Header glass / màu / viền</v-chip>
              <v-chip color="primary" variant="tonal" size="small">Spacing gọn / cân bằng / thoáng</v-chip>
            </div>
          </v-card>
        </v-col>
        <v-col cols="12" lg="4">
          <v-card rounded="lg" class="pa-5 h-100">
            <div class="text-subtitle-1 font-weight-bold mb-3">Trạng thái hệ thống</div>
            <div class="d-flex align-center justify-space-between py-2 border-b">
              <span class="text-body-2">API Gateway</span>
              <v-chip size="x-small" color="success" variant="flat">Online</v-chip>
            </div>
            <div class="d-flex align-center justify-space-between py-2 border-b">
              <span class="text-body-2">Vai trò của bạn</span>
              <span class="text-body-2 font-weight-medium text-capitalize">{{ user?.role?.replace('_', ' ') }}</span>
            </div>
            <div class="d-flex align-center justify-space-between py-2">
              <span class="text-body-2">Theme hiện tại</span>
              <span class="text-body-2 font-weight-medium">{{ layout.theme.value.label }}</span>
            </div>
          </v-card>
        </v-col>
      </v-row>
    </template>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin' })

const admin = useAdmin()
const authStore = useAuthStore()
const user = computed(() => authStore.user)
const layout = useAdminLayout()
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

<style scoped>
.border-b {
  border-bottom: 1px solid var(--admin-border, rgba(0, 0, 0, 0.06));
}
</style>
