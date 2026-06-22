<template>
  <div>
    <AdminPageHeader title="Báo cáo & Thống kê" subtitle="Doanh thu, sản phẩm, review, user online — Kafka + Analytics Service">
      <template #actions>
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
      </template>
    </AdminPageHeader>

    <LoadingSpinner v-if="loading && !overview" />

    <template v-else>
      <v-row class="mb-2">
        <v-col cols="6" md="3">
          <AdminStatCard
            label="Doanh thu"
            :value="formatVND(overview?.orders.revenue ?? 0)"
            icon="mdi-cash-multiple"
            color="var(--admin-primary)"
            :subtitle="`${overview?.orders.total ?? 0} đơn hàng`"
          />
        </v-col>
        <v-col cols="6" md="3">
          <AdminStatCard
            label="Đơn hàng"
            :value="overview?.orders.total ?? 0"
            icon="mdi-clipboard-check-outline"
            color="var(--admin-accent)"
            :subtitle="`${overview?.orders.delivered ?? 0} đã giao`"
          />
        </v-col>
        <v-col cols="6" md="3">
          <AdminStatCard
            label="Review mới"
            :value="overview?.products.new_reviews ?? 0"
            icon="mdi-star-outline"
            color="#059669"
            :subtitle="`TB ${(overview?.products.avg_rating ?? 0).toFixed(1)} ★`"
          />
        </v-col>
        <v-col cols="6" md="3">
          <AdminStatCard
            label="Khách mới"
            :value="overview?.users.new_customers ?? 0"
            icon="mdi-account-plus-outline"
            color="#f57c00"
            :subtitle="`${onlineCount} online ngay`"
          />
        </v-col>
      </v-row>

      <v-row>
        <v-col cols="12" lg="8">
          <v-card rounded="lg" class="admin-chart-card pa-4">
            <div class="admin-chart-card__head">
              <div>
                <div class="admin-chart-card__title">Doanh thu & Đơn hàng</div>
                <div class="admin-chart-card__sub">Biểu đồ kết hợp theo {{ periodLabel }}</div>
              </div>
              <v-chip size="small" color="primary" variant="tonal">Chart.js</v-chip>
            </div>
            <AdminComboChart :series="revenueSeries" :format-revenue="formatVND" />
          </v-card>
        </v-col>
        <v-col cols="12" lg="4">
          <v-card rounded="lg" class="admin-chart-card pa-4 h-100">
            <div class="admin-chart-card__head mb-2">
              <div>
                <div class="admin-chart-card__title">Trạng thái đơn</div>
                <div class="admin-chart-card__sub">Phân bổ theo trạng thái</div>
              </div>
            </div>
            <AdminDoughnutChart :items="overview?.orders.by_status ?? {}" />
          </v-card>
        </v-col>
      </v-row>

      <v-row class="mt-2">
        <v-col cols="12" md="6">
          <v-card rounded="lg" class="admin-chart-card pa-4">
            <div class="admin-chart-card__head mb-2">
              <div>
                <div class="admin-chart-card__title">Khách đăng ký mới</div>
                <div class="admin-chart-card__sub">Xu hướng tăng trưởng user</div>
              </div>
            </div>
            <AdminLineChart :series="usersSeries" value-key="count" color="#059669" label="Khách mới" />
          </v-card>
        </v-col>
        <v-col cols="12" md="6">
          <v-card rounded="lg" class="admin-chart-card pa-4">
            <div class="admin-chart-card__head mb-2">
              <div>
                <div class="admin-chart-card__title">Review & Đánh giá</div>
                <div class="admin-chart-card__sub">Số review theo thời gian</div>
              </div>
            </div>
            <AdminBarChart :series="reviewsSeries" value-key="count" color="#f57c00" label="Review" />
          </v-card>
        </v-col>
      </v-row>

      <v-row class="mt-2">
        <v-col cols="12" lg="7">
          <v-card rounded="lg" class="admin-chart-card pa-4">
            <div class="admin-chart-card__head mb-2">
              <div>
                <div class="admin-chart-card__title">Top sản phẩm bán chạy</div>
                <div class="admin-chart-card__sub">Theo doanh thu trong kỳ</div>
              </div>
            </div>
            <AdminHorizontalBarChart
              :items="topProductBars"
              :format-value="formatVND"
            />
          </v-card>
        </v-col>
        <v-col cols="12" lg="5">
          <v-card rounded="lg" class="admin-chart-card pa-4 h-100">
            <div class="d-flex align-center justify-space-between mb-3">
              <div>
                <div class="admin-chart-card__title">User đang online</div>
                <div class="admin-chart-card__sub">Heartbeat 90 giây</div>
              </div>
              <v-chip color="success" size="small" variant="flat">{{ onlineCount }} người</v-chip>
            </div>
            <div v-if="!onlineUsers.length" class="admin-chart-wrap__empty-inline">
              <v-icon size="36" color="grey-lighten-1">mdi-account-off-outline</v-icon>
              <span>Không có user online</span>
            </div>
            <v-list v-else density="compact" class="pa-0">
              <v-list-item v-for="u in onlineUsers" :key="u.user_id" class="px-0 online-user-item">
                <template #prepend>
                  <v-avatar color="primary" size="36">
                    <v-icon size="18" color="white">mdi-account</v-icon>
                  </v-avatar>
                </template>
                <v-list-item-title class="text-body-2 font-weight-medium">
                  {{ u.email || u.user_id.slice(0, 8) }}
                </v-list-item-title>
                <v-list-item-subtitle class="text-caption">{{ u.page }}</v-list-item-subtitle>
                <template #append>
                  <v-icon size="8" color="success">mdi-circle</v-icon>
                </template>
              </v-list-item>
            </v-list>
          </v-card>
        </v-col>
      </v-row>
    </template>
  </div>
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

const periodLabel = computed(() => ({
  day: 'ngày',
  week: 'tuần',
  month: 'tháng',
  quarter: 'quý',
}[period.value]))

const topProductBars = computed(() =>
  topProducts.value.slice(0, 8).map((p) => ({
    label: p.name.length > 28 ? `${p.name.slice(0, 27)}…` : p.name,
    value: p.revenue,
    subLabel: `${p.quantity_sold} sp`,
  })),
)

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
.admin-chart-card {
  border: 1px solid var(--admin-border, rgba(0, 0, 0, 0.06));
}

.admin-chart-card__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 8px;
}

.admin-chart-card__title {
  font-size: 15px;
  font-weight: 700;
  color: var(--admin-text, #0f172a);
}

.admin-chart-card__sub {
  font-size: 12px;
  color: var(--admin-text-muted, #64748b);
  margin-top: 2px;
}

.admin-chart-wrap__empty-inline {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 32px 16px;
  color: var(--admin-text-muted, #94a3b8);
  font-size: 13px;
}

.online-user-item {
  border-radius: var(--admin-radius-xs);
  margin-bottom: 4px;
}

.online-user-item:hover {
  background: rgba(148, 163, 184, 0.08);
}
</style>
