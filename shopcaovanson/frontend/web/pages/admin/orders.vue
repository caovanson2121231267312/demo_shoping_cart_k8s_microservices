<template>
  <div>
    <AdminPageHeader title="Quản lý đơn hàng" subtitle="Theo dõi và cập nhật trạng thái đơn hàng">
      <template #actions>
        <v-btn
          color="success"
          variant="flat"
          prepend-icon="mdi-microsoft-excel"
          class="text-none"
          :loading="exporting"
          @click="onExportExcel"
        >
          Xuất Excel
        </v-btn>
      </template>
    </AdminPageHeader>

    <AdminFilterBar>
      <v-text-field
        v-model="searchQuery"
        label="Mã đơn / SĐT / tên"
        density="compact"
        variant="outlined"
        hide-details
        clearable
        prepend-inner-icon="mdi-magnify"
        @keyup.enter="loadOrders"
      />
      <v-select
        v-model="statusFilter"
        :items="statusOptions"
        label="Trạng thái"
        clearable
        density="compact"
        variant="outlined"
        hide-details
      />
      <AdminDateRangeFilter v-model:from="dateFilter.createdFrom" v-model:to="dateFilter.createdTo" />
      <v-btn color="primary" prepend-icon="mdi-magnify" @click="onSearch">Tìm kiếm</v-btn>
      <v-btn v-if="dateFilter.hasDateFilter" variant="text" @click="clearFilters">Xóa lọc</v-btn>
    </AdminFilterBar>

    <AdminDataTable
      server
      v-model:page="page"
      v-model:items-per-page="limit"
      :headers="headers"
      :items="orders"
      :total-items="total"
      :count="total"
      :loading="loading"
      title="Danh sách đơn hàng"
      @update:options="loadOrders"
    >
      <template #item.order_number="{ item }">
        <span class="admin-table__mono">{{ item.order_number || item.id.slice(0, 8) }}</span>
      </template>
      <template #item.shipping_name="{ item }">
        <div>
          <div class="admin-table__cell-title">{{ item.shipping_name }}</div>
          <div v-if="item.shipping_phone" class="admin-table__cell-sub">{{ item.shipping_phone }}</div>
        </div>
      </template>
      <template #item.created_at="{ item }">
        {{ formatDate(item.created_at) }}
      </template>
      <template #item.status="{ item }">
        <v-chip :color="statusColor(item.status)" size="small" variant="tonal">
          {{ statusLabel(item.status) }}
        </v-chip>
      </template>
      <template #item.total_amount="{ item }">
        <span class="admin-table__money">{{ formatVND(item.total_amount) }}</span>
      </template>
      <template #item.actions="{ item }">
        <div class="admin-table-actions">
          <v-btn
            size="small"
            variant="tonal"
            color="primary"
            @click.stop="goToDetail(item)"
          >
            Chi tiết
          </v-btn>
          <v-menu v-if="admin.canManageOrders.value">
            <template #activator="{ props }">
              <v-btn v-bind="props" size="small" variant="outlined" color="primary">
                Cập nhật
              </v-btn>
            </template>
            <v-list density="compact">
              <v-list-item
                v-for="status in availableStatuses(item.status)"
                :key="status"
                :title="statusLabel(status)"
                @click="updateStatus(item.id, status)"
              />
            </v-list>
          </v-menu>
        </div>
      </template>
    </AdminDataTable>

    <v-dialog v-model="exportDialog" persistent max-width="440">
      <v-card rounded="lg">
        <v-card-title class="d-flex align-center ga-2">
          <v-icon color="success">mdi-microsoft-excel</v-icon>
          Đang xuất Excel
        </v-card-title>
        <v-card-text>
          <p class="text-body-2 mb-3">{{ exportProgressMessage }}</p>
          <v-progress-linear
            :model-value="exportProgress"
            color="success"
            height="10"
            rounded
            striped
          />
          <p class="text-caption text-medium-emphasis mt-2 text-end">{{ exportProgress }}%</p>
        </v-card-text>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import type { Order } from '~/types'

definePageMeta({ layout: 'admin' })

const { searchAdminOrders, updateOrderStatus, exportOrdersExcel } = useOrders()
const admin = useAdmin()
const dateFilter = useAdminDateFilter()
const snackbar = useSnackbar()
const { page, limit, total, applyMeta, resetPage } = useAdminServerTable(20)

const orders = ref<Order[]>([])
const loading = ref(false)
const exporting = ref(false)
const exportDialog = ref(false)
const exportProgress = ref(0)
const exportProgressMessage = ref('Đang khởi tạo...')
const statusFilter = ref<string | null>(null)
const searchQuery = ref('')

const statusOptions = [
  { title: 'Chờ xử lý', value: 'pending' },
  { title: 'Đã xác nhận', value: 'confirmed' },
  { title: 'Đang giao', value: 'shipping' },
  { title: 'Đã giao', value: 'delivered' },
  { title: 'Đã hủy', value: 'cancelled' },
]

const headers = [
  { title: 'Mã đơn', key: 'order_number' },
  { title: 'Khách hàng', key: 'shipping_name', sortable: false },
  { title: 'Ngày đặt', key: 'created_at' },
  { title: 'Trạng thái', key: 'status', sortable: false },
  { title: 'Tổng tiền', key: 'total_amount', align: 'end' as const },
  { title: 'Thao tác', key: 'actions', sortable: false, align: 'end' as const, width: 180 },
]

const statusLabel = (status: string) => {
  const map: Record<string, string> = {
    pending: 'Chờ xử lý',
    confirmed: 'Đã xác nhận',
    shipping: 'Đang giao',
    delivered: 'Đã giao',
    cancelled: 'Đã hủy',
  }
  return map[status] || status
}

const statusColor = (status: string) => {
  const map: Record<string, string> = {
    pending: 'warning',
    confirmed: 'info',
    shipping: 'primary',
    delivered: 'success',
    cancelled: 'error',
  }
  return map[status] || 'grey'
}

const formatDate = (date: string) => new Date(date).toLocaleDateString('vi-VN')

function resolveOrderId(item: Order | { raw?: Order }) {
  return item.id || item.raw?.id
}

function goToDetail(item: Order | { raw?: Order }) {
  const id = resolveOrderId(item)
  if (!id) {
    snackbar.show('Không xác định được mã đơn hàng', 'error')
    return
  }
  navigateTo(`/admin/orders/${id}`)
}

const availableStatuses = (current: string) => {
  const flow: Record<string, string[]> = {
    pending: ['confirmed', 'cancelled'],
    confirmed: ['shipping', 'cancelled'],
    shipping: ['delivered'],
    delivered: [],
    cancelled: [],
  }
  return flow[current] || []
}

const loadOrders = async () => {
  loading.value = true
  try {
    const query = dateFilter.withDateQuery({
      limit: limit.value,
      page: page.value,
    })
    if (statusFilter.value) query.status = statusFilter.value
    if (searchQuery.value) query.search = searchQuery.value
    const result = await searchAdminOrders(query)
    orders.value = result.items
    applyMeta(result)
  } finally {
    loading.value = false
  }
}

function onSearch() {
  resetPage()
  loadOrders()
}

const clearFilters = () => {
  searchQuery.value = ''
  statusFilter.value = null
  dateFilter.resetDates()
  resetPage()
  loadOrders()
}

const updateStatus = async (orderId: string, status: string) => {
  try {
    await updateOrderStatus(orderId, status)
    await loadOrders()
    snackbar.show('Đã cập nhật trạng thái đơn hàng', 'success')
  } catch (e: unknown) {
    snackbar.show(e instanceof Error ? e.message : 'Không thể cập nhật đơn hàng', 'error')
  }
}

const onExportExcel = async () => {
  exporting.value = true
  exportDialog.value = true
  exportProgress.value = 0
  exportProgressMessage.value = 'Đang khởi tạo...'
  try {
    const filters: Record<string, string> = {}
    if (statusFilter.value) filters.status = statusFilter.value
    if (searchQuery.value?.trim()) filters.search = searchQuery.value.trim()
    if (dateFilter.createdFrom) filters.created_from = dateFilter.createdFrom
    if (dateFilter.createdTo) filters.created_to = dateFilter.createdTo
    await exportOrdersExcel(
      filters,
      ({ progress, message }) => {
        exportProgress.value = progress
        exportProgressMessage.value = message
      },
      async () => {
        exportDialog.value = false
        await nextTick()
      },
    )
    snackbar.show('Đã xuất file Excel đơn hàng', 'success')
  } catch (e: unknown) {
    snackbar.show(e instanceof Error ? e.message : 'Xuất Excel thất bại', 'error')
  } finally {
    exporting.value = false
    exportDialog.value = false
  }
}
</script>
