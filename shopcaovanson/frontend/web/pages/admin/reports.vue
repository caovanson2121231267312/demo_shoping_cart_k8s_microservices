<template>
  <v-container class="page-container py-6">
    <div class="d-flex align-center justify-space-between flex-wrap ga-3 mb-6">
      <div>
        <h1 class="text-h4 font-weight-bold">Báo cáo & Thống kê</h1>
        <p class="text-body-2 text-muted mb-0">Doanh thu, sản phẩm, review, user online — Kafka + Analytics Service</p>
      </div>
      <div class="d-flex flex-wrap ga-2 align-center">
        <v-btn-toggle v-model="period" mandatory color="primary" density="compact" rounded="lg">
          <v-btn value="day" class="text-none">Ngày</v-btn>
          <v-btn value="week" class="text-none">Tuần</v-btn>
          <v-btn value="month" class="text-none">Tháng</v-btn>
          <v-btn value="quarter" class="text-none">Quý</v-btn>
        </v-btn-toggle>
        <v-btn color="success" variant="flat" prepend-icon="mdi-microsoft-excel" class="text-none" :loading="exporting" @click="onExport">
          Xuất Excel
        </v-btn>
        <v-btn variant="outlined" prepend-icon="mdi-refresh" class="text-none" :loading="loading" @click="loadAll">
          Làm mới
        </v-btn>
      </div>
    </div>

    <LoadingSpinner v-if="loading && !overview" />

    <template v-else>
      <!-- KPI -->
      <v-row class="mb-4">
        <v-col cols="6" md="3">
          <v-card class="kpi-card" color="primary" variant="tonal">
            <v-card-text>
              <div class="text-overline">Doanh thu</div>
              <div class="text-h6 font-weight-bold">{{ formatVND(overview?.orders.revenue ?? 0) }}</div>
            </v-card-text>
          </v-card>
        </v-col>
        <v-col cols="6" md="3">
          <v-card class="kpi-card" color="info" variant="tonal">
            <v-card-text>
              <div class="text-overline">Đơn hàng</div>
              <div class="text-h5 font-weight-bold">{{ overview?.orders.total ?? 0 }}</div>
            </v-card-text>
          </v-card>
        </v-col>
        <v-col cols="6" md="3">
          <v-card class="kpi-card" color="success" variant="tonal">
            <v-card-text>
              <div class="text-overline">Review mới</div>
              <div class="text-h5 font-weight-bold">{{ overview?.products.new_reviews ?? 0 }}</div>
              <div class="text-caption">TB {{ (overview?.products.avg_rating ?? 0).toFixed(1) }} ★</div>
            </v-card-text>
          </v-card>
        </v-col>
        <v-col cols="6" md="3">
          <v-card class="kpi-card" color="secondary" variant="tonal">
            <v-card-text>
              <div class="text-overline d-flex align-center ga-1">
                Online
                <v-chip size="x-small" color="success" variant="flat">{{ onlineCount }}</v-chip>
              </div>
              <div class="text-h5 font-weight-bold">{{ overview?.users.new_customers ?? 0 }}</div>
              <div class="text-caption">khách mới trong kỳ</div>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <!-- Charts row 1 -->
      <v-row>
        <v-col cols="12" lg="8">
          <v-card rounded="lg" class="pa-4">
            <div class="text-subtitle-1 font-weight-bold mb-3">Doanh thu & Đơn hàng</div>
            <AdminBarChart :series="revenueSeries" value-key="revenue" color="#1565c0" />
          </v-card>
        </v-col>
        <v-col cols="12" lg="4">
          <v-card rounded="lg" class="pa-4 h-100">
            <div class="text-subtitle-1 font-weight-bold mb-3">Trạng thái đơn</div>
            <div v-for="(count, status) in overview?.orders.by_status" :key="status" class="d-flex justify-space-between py-2 border-b">
              <span class="text-capitalize">{{ statusLabel(status) }}</span>
              <strong>{{ count }}</strong>
            </div>
          </v-card>
        </v-col>
      </v-row>

      <!-- Charts row 2 -->
      <v-row class="mt-2">
        <v-col cols="12" md="6">
          <v-card rounded="lg" class="pa-4">
            <div class="text-subtitle-1 font-weight-bold mb-3">Khách đăng ký mới</div>
            <AdminBarChart :series="usersSeries" value-key="count" color="#2e7d32" />
          </v-card>
        </v-col>
        <v-col cols="12" md="6">
          <v-card rounded="lg" class="pa-4">
            <div class="text-subtitle-1 font-weight-bold mb-3">Review & Đánh giá</div>
            <AdminBarChart :series="reviewsSeries" value-key="count" color="#f57c00" />
          </v-card>
        </v-col>
      </v-row>

      <!-- Top products + Online users -->
      <v-row class="mt-2">
        <v-col cols="12" lg="7">
          <v-card rounded="lg" class="pa-4">
            <div class="text-subtitle-1 font-weight-bold mb-3">Top sản phẩm bán chạy</div>
            <v-table density="compact">
              <thead>
                <tr>
                  <th>#</th>
                  <th>Sản phẩm</th>
                  <th class="text-right">SL</th>
                  <th class="text-right">Doanh thu</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(item, i) in topProducts" :key="item.name">
                  <td>{{ i + 1 }}</td>
                  <td class="text-truncate" style="max-width:280px">{{ item.name }}</td>
                  <td class="text-right">{{ item.quantity_sold }}</td>
                  <td class="text-right">{{ formatVND(item.revenue) }}</td>
                </tr>
              </tbody>
            </v-table>
          </v-card>
        </v-col>
        <v-col cols="12" lg="5">
          <v-card rounded="lg" class="pa-4">
            <div class="d-flex align-center justify-space-between mb-3">
              <div class="text-subtitle-1 font-weight-bold">User đang online</div>
              <v-chip color="success" size="small" variant="flat">{{ onlineCount }} người</v-chip>
            </div>
            <div v-if="!onlineUsers.length" class="text-caption text-muted py-4 text-center">
              Không có user online (heartbeat 90s)
            </div>
            <v-list v-else density="compact" class="pa-0">
              <v-list-item v-for="u in onlineUsers" :key="u.user_id" class="px-0">
                <template #prepend>
                  <v-avatar color="primary" size="32">
                    <v-icon size="18" color="white">mdi-account</v-icon>
                  </v-avatar>
                </template>
                <v-list-item-title class="text-body-2">{{ u.email || u.user_id.slice(0, 8) }}</v-list-item-title>
                <v-list-item-subtitle class="text-caption">{{ u.page }}</v-list-item-subtitle>
              </v-list-item>
            </v-list>
          </v-card>
        </v-col>
      </v-row>
    </template>
  </v-container>
</template>

<script setup lang="ts">
import type { AnalyticsOverview, AnalyticsPeriod, AnalyticsSeriesPoint, OnlineUser, TopProductRow } from '~/types'

definePageMeta({ layout: 'admin' })

const analytics = useAnalytics()
const { formatVND } = useFormat()
const snackbar = useSnackbar()

const period = ref<AnalyticsPeriod>('month')
const loading = ref(true)
const exporting = ref(false)
const overview = ref<AnalyticsOverview | null>(null)
const revenueSeries = ref<AnalyticsSeriesPoint[]>([])
const usersSeries = ref<AnalyticsSeriesPoint[]>([])
const reviewsSeries = ref<AnalyticsSeriesPoint[]>([])
const topProducts = ref<TopProductRow[]>([])
const onlineUsers = ref<OnlineUser[]>([])
const onlineCount = ref(0)

let onlineTimer: ReturnType<typeof setInterval>

const statusLabel = (s: string) => ({
  pending: 'Chờ xử lý',
  confirmed: 'Đã xác nhận',
  shipping: 'Đang giao',
  delivered: 'Đã giao',
  cancelled: 'Đã hủy',
}[s] ?? s)

const loadAll = async () => {
  loading.value = true
  try {
    const [ov, rev, usr, revw, top, online] = await Promise.all([
      analytics.fetchOverview(period.value),
      analytics.fetchRevenue(period.value),
      analytics.fetchUsers(period.value),
      analytics.fetchReviews(period.value),
      analytics.fetchTopProducts(period.value, 10),
      analytics.fetchOnlineUsers(),
    ])
    overview.value = ov
    revenueSeries.value = rev.series
    usersSeries.value = usr.series
    reviewsSeries.value = revw.series
    topProducts.value = top.items
    onlineUsers.value = online.users
    onlineCount.value = online.count
  } catch {
    snackbar.show('Không tải được báo cáo. Kiểm tra analytics-service (port 8086).', 'error')
  } finally {
    loading.value = false
  }
}

const onExport = async () => {
  exporting.value = true
  try {
    await analytics.exportExcel(period.value)
    snackbar.show('Đã xuất file Excel', 'success')
  } catch {
    snackbar.show('Xuất Excel thất bại', 'error')
  } finally {
    exporting.value = false
  }
}

watch(period, loadAll)

onMounted(() => {
  loadAll()
  onlineTimer = setInterval(() => {
    analytics.fetchOnlineUsers().then((r) => {
      onlineUsers.value = r.users
      onlineCount.value = r.count
    }).catch(() => {})
  }, 30000)
})

onUnmounted(() => clearInterval(onlineTimer))
</script>

<style scoped>
.kpi-card {
  height: 100%;
}

.border-b {
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
}
</style>
