<template>
  <div>
    <AdminPageHeader
      title="Lịch sử đăng nhập"
      subtitle="Theo dõi chi tiết mỗi lần login và tải báo cáo Excel theo ngày"
    >
      <template #actions>
        <v-btn
          v-if="admin.canManageUsers.value"
          color="primary"
          variant="tonal"
          prepend-icon="mdi-file-excel"
          :loading="generating"
          @click="onGenerate"
        >
          Tạo báo cáo hôm qua
        </v-btn>
      </template>
    </AdminPageHeader>

    <v-tabs v-model="tab" color="primary" class="mb-4">
      <v-tab value="history">Lịch sử chi tiết</v-tab>
      <v-tab value="reports">Báo cáo Excel</v-tab>
    </v-tabs>

    <v-tabs-window v-model="tab">
      <v-tabs-window-item value="history">
        <AdminFilterBar>
          <v-text-field
            v-model="emailFilter"
            label="Email"
            density="compact"
            variant="outlined"
            hide-details
            clearable
            prepend-inner-icon="mdi-magnify"
            @keyup.enter="onSearchHistory"
          />
          <v-select
            v-model="successFilter"
            :items="successOptions"
            label="Kết quả"
            clearable
            density="compact"
            variant="outlined"
            hide-details
          />
          <AdminDateRangeFilter v-model:from="dateFilter.createdFrom" v-model:to="dateFilter.createdTo" />
          <v-btn color="primary" prepend-icon="mdi-magnify" @click="onSearchHistory">Tìm kiếm</v-btn>
          <v-btn v-if="hasHistoryFilters" variant="text" @click="clearHistoryFilters">Xóa lọc</v-btn>
        </AdminFilterBar>

        <AdminDataTable
          server
          cursor-mode
          v-model:items-per-page="limit"
          :headers="historyHeaders"
          :items="history"
          :total-items="total"
          :count="total"
          :loading="loadingHistory"
          :has-more="hasMore"
          :has-prev="hasPrev"
          :cursor-page-number="pageIndex + 1"
          :range-offset="rangeOffset"
          title="Nhật ký đăng nhập"
          @update:options="onHistoryTableOptions"
          @cursor-next="onNextPage"
          @cursor-prev="onPrevPage"
        >
          <template #item.email="{ item }">
            <div>
              <div class="admin-table__cell-title">{{ item.email }}</div>
              <div v-if="item.full_name" class="admin-table__cell-sub">{{ item.full_name }}</div>
            </div>
          </template>
          <template #item.success="{ item }">
            <v-chip :color="item.success ? 'success' : 'error'" size="small" variant="tonal">
              {{ item.success ? 'Thành công' : 'Thất bại' }}
            </v-chip>
          </template>
          <template #item.failure_reason="{ item }">
            {{ failureLabel(item.failure_reason) }}
          </template>
          <template #item.device="{ item }">
            <div class="admin-table__cell-sub">
              {{ [item.device, item.browser, item.os].filter(Boolean).join(' · ') || '—' }}
            </div>
          </template>
          <template #item.created_at="{ item }">
            {{ formatDateTime(item.created_at) }}
          </template>
        </AdminDataTable>
      </v-tabs-window-item>

      <v-tabs-window-item value="reports">
        <AdminFilterBar>
          <AdminDateRangeFilter v-model:from="reportFrom" v-model:to="reportTo" />
          <v-btn color="primary" prepend-icon="mdi-filter-outline" @click="loadReports">Lọc</v-btn>
          <v-btn v-if="reportFrom || reportTo" variant="text" @click="clearReportFilters">Xóa lọc</v-btn>
        </AdminFilterBar>

        <AdminDataTable
          server
          v-model:page="reportPage"
          v-model:items-per-page="reportLimit"
          :headers="reportHeaders"
          :items="reports"
          :total-items="reportTotal"
          :count="reportTotal"
          :loading="loadingReports"
          title="Báo cáo đăng nhập theo ngày"
          @update:options="loadReports"
        >
          <template #item.report_date="{ item }">
            {{ formatReportDate(item.report_date) }}
          </template>
          <template #item.file_size="{ item }">
            {{ formatBytes(item.file_size) }}
          </template>
          <template #item.status="{ item }">
            <v-chip :color="item.status === 'ready' ? 'success' : 'error'" size="small" variant="tonal">
              {{ item.status === 'ready' ? 'Sẵn sàng' : item.status }}
            </v-chip>
          </template>
          <template #item.actions="{ item }">
            <v-btn
              size="small"
              color="success"
              variant="tonal"
              prepend-icon="mdi-download"
              :disabled="item.status !== 'ready'"
              :loading="downloadingId === item.id"
              @click="onDownload(item)"
            >
              Tải Excel
            </v-btn>
          </template>
        </AdminDataTable>
      </v-tabs-window-item>
    </v-tabs-window>
  </div>
</template>

<script setup lang="ts">
import type { LoginHistoryItem, LoginReport } from '~/types'

definePageMeta({ layout: 'admin' })

const admin = useAdmin()
const dateFilter = useAdminDateFilter()
const snackbar = useSnackbar()
const {
  limit,
  total,
  hasMore,
  hasPrev,
  pageIndex,
  rangeOffset,
  applyMeta,
  reset,
  goNext,
  goPrev,
  queryParams,
} = useAdminCursorTable(20)

const tab = ref('history')
const history = ref<LoginHistoryItem[]>([])
const loadingHistory = ref(false)
const emailFilter = ref('')
const successFilter = ref<string | null>(null)

const reports = ref<LoginReport[]>([])
const loadingReports = ref(false)
const reportPage = ref(1)
const reportLimit = ref(20)
const reportTotal = ref(0)
const reportFrom = ref('')
const reportTo = ref('')
const downloadingId = ref('')
const generating = ref(false)

const successOptions = [
  { title: 'Thành công', value: 'true' },
  { title: 'Thất bại', value: 'false' },
]

const historyHeaders = [
  { title: 'Thời gian', key: 'created_at' },
  { title: 'Tài khoản', key: 'email', sortable: false },
  { title: 'Kết quả', key: 'success', sortable: false },
  { title: 'Lý do', key: 'failure_reason', sortable: false },
  { title: 'IP', key: 'ip_address', sortable: false },
  { title: 'Thiết bị', key: 'device', sortable: false },
]

const reportHeaders = [
  { title: 'Ngày', key: 'report_date' },
  { title: 'Tổng login', key: 'total_logins', align: 'end' as const },
  { title: 'Thành công', key: 'success_count', align: 'end' as const },
  { title: 'Thất bại', key: 'failure_count', align: 'end' as const },
  { title: 'User unique', key: 'unique_users', align: 'end' as const },
  { title: 'Dung lượng', key: 'file_size', align: 'end' as const },
  { title: 'Trạng thái', key: 'status', sortable: false },
  { title: 'Thao tác', key: 'actions', sortable: false, align: 'end' as const, width: 140 },
]

const hasHistoryFilters = computed(() =>
  Boolean(emailFilter.value || successFilter.value || dateFilter.hasDateFilter),
)

function failureLabel(reason?: string | null) {
  if (!reason) return '—'
  const map: Record<string, string> = {
    invalid_credentials: 'Sai email/mật khẩu',
    account_disabled: 'Tài khoản bị khóa',
    email_not_verified: 'Chưa xác minh email',
  }
  return map[reason] || reason
}

function formatDateTime(iso: string) {
  return new Date(iso).toLocaleString('vi-VN')
}

function formatReportDate(value: string) {
  const d = value.slice(0, 10)
  const [y, m, day] = d.split('-')
  return `${day}/${m}/${y}`
}

function formatBytes(n: number) {
  if (!n) return '0 B'
  if (n < 1024) return `${n} B`
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`
  return `${(n / (1024 * 1024)).toFixed(1)} MB`
}

async function loadHistory() {
  loadingHistory.value = true
  try {
    const query = dateFilter.withDateQuery(queryParams({
      ...(emailFilter.value ? { email: emailFilter.value } : {}),
      ...(successFilter.value ? { success: successFilter.value } : {}),
    }))
    const result = await admin.fetchLoginHistory(query)
    history.value = result.items
    applyMeta(result)
  } catch (e: unknown) {
    snackbar.show(e instanceof Error ? e.message : 'Không tải được lịch sử đăng nhập', 'error')
  } finally {
    loadingHistory.value = false
  }
}

function onHistoryTableOptions() {
  reset()
  loadHistory()
}

function onNextPage() {
  if (goNext()) loadHistory()
}

function onPrevPage() {
  if (goPrev()) loadHistory()
}

function onSearchHistory() {
  reset()
  loadHistory()
}

function clearHistoryFilters() {
  emailFilter.value = ''
  successFilter.value = null
  dateFilter.resetDates()
  reset()
  loadHistory()
}

async function loadReports() {
  loadingReports.value = true
  try {
    const query: Record<string, string | number> = {
      page: reportPage.value,
      limit: reportLimit.value,
    }
    if (reportFrom.value) query.from = reportFrom.value
    if (reportTo.value) query.to = reportTo.value
    const result = await admin.fetchLoginReports(query)
    reports.value = result.items
    reportTotal.value = result.total
  } catch (e: unknown) {
    snackbar.show(e instanceof Error ? e.message : 'Không tải được báo cáo', 'error')
  } finally {
    loadingReports.value = false
  }
}

function clearReportFilters() {
  reportFrom.value = ''
  reportTo.value = ''
  reportPage.value = 1
  loadReports()
}

async function onDownload(item: LoginReport) {
  downloadingId.value = item.id
  try {
    await admin.downloadLoginReport(item.id, item.file_name)
    snackbar.show('Đã tải báo cáo Excel', 'success')
  } catch (e: unknown) {
    snackbar.show(e instanceof Error ? e.message : 'Tải báo cáo thất bại', 'error')
  } finally {
    downloadingId.value = ''
  }
}

async function onGenerate() {
  generating.value = true
  try {
    const report = await admin.generateLoginReport()
    snackbar.show(`Đã tạo báo cáo ${formatReportDate(String(report.report_date))}`, 'success')
    tab.value = 'reports'
    await loadReports()
  } catch (e: unknown) {
    snackbar.show(e instanceof Error ? e.message : 'Tạo báo cáo thất bại (cần MinIO)', 'error')
  } finally {
    generating.value = false
  }
}

watch(tab, (v) => {
  if (v === 'reports' && !reports.value.length) {
    loadReports()
  }
})
</script>
