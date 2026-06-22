<template>
  <div>
    <AdminPageHeader title="Mã giảm giá" subtitle="Quản lý voucher và khuyến mãi">
      <template #actions>
        <v-btn color="primary" prepend-icon="mdi-plus" @click="openCreate">Thêm mã</v-btn>
      </template>
    </AdminPageHeader>

    <AdminFilterBar>
      <AdminDateRangeFilter v-model:from="dateFilter.createdFrom" v-model:to="dateFilter.createdTo" />
      <v-btn color="primary" prepend-icon="mdi-filter-outline" @click="onFilter">Lọc</v-btn>
      <v-btn v-if="dateFilter.hasDateFilter" variant="text" @click="clearFilters">Xóa lọc</v-btn>
    </AdminFilterBar>

    <AdminDataTable
      server
      v-model:page="page"
      v-model:items-per-page="limit"
      :headers="headers"
      :items="coupons"
      :total-items="total"
      :count="total"
      :loading="loading"
      title="Danh sách mã giảm giá"
      @update:options="load"
    >
      <template #item.code="{ item }">
        <span class="admin-table__mono">{{ item.code }}</span>
      </template>
      <template #item.type="{ item }">
        <v-chip size="small" variant="tonal" :color="item.type === 'percent' ? 'info' : 'primary'">
          {{ item.type === 'percent' ? 'Phần trăm' : 'Cố định' }}
        </v-chip>
      </template>
      <template #item.value="{ item }">
        <span class="admin-table__money">
          {{ item.type === 'percent' ? `${item.value}%` : formatVND(item.value) }}
        </span>
      </template>
      <template #item.used_count="{ item }">
        {{ item.used_count }}
      </template>
      <template #item.created_at="{ item }">
        {{ formatDate(item.created_at) }}
      </template>
      <template #item.is_active="{ item }">
        <v-chip :color="item.is_active ? 'success' : 'grey'" size="small" variant="tonal">
          {{ item.is_active ? 'Bật' : 'Tắt' }}
        </v-chip>
      </template>
      <template #item.actions="{ item }">
        <div class="admin-table-actions">
          <v-btn size="small" variant="tonal" color="primary" @click="openEdit(item)">Sửa</v-btn>
          <v-btn size="small" variant="tonal" color="error" @click="remove(item.id)">Xóa</v-btn>
        </div>
      </template>
    </AdminDataTable>

    <v-dialog v-model="dialog" max-width="520">
      <v-card>
        <v-card-title>{{ editing ? 'Sửa mã' : 'Thêm mã giảm giá' }}</v-card-title>
        <v-card-text>
          <v-text-field v-model="form.code" label="Mã" :disabled="editing" variant="outlined" class="mb-2" />
          <v-select
            v-model="form.type"
            :items="[{ title: 'Phần trăm', value: 'percent' }, { title: 'Cố định', value: 'fixed' }]"
            label="Loại"
            variant="outlined"
            class="mb-2"
          />
          <v-text-field v-model.number="form.value" label="Giá trị" type="number" variant="outlined" class="mb-2" />
          <v-text-field v-model.number="form.min_order" label="Đơn tối thiểu" type="number" variant="outlined" class="mb-2" />
          <v-switch v-model="form.is_active" label="Kích hoạt" color="primary" />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="dialog = false">Hủy</v-btn>
          <v-btn color="primary" :loading="saving" @click="save">Lưu</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import type { Coupon } from '~/types'

definePageMeta({ layout: 'admin' })

const admin = useAdmin()
const { formatVND } = useFormat()
const dateFilter = useAdminDateFilter()
const snackbar = useSnackbar()
const { page, limit, total, applyMeta, resetPage } = useAdminServerTable(20)

const loading = ref(false)
const saving = ref(false)
const dialog = ref(false)
const editing = ref(false)
const editId = ref('')
const coupons = ref<Coupon[]>([])

const form = reactive({
  code: '',
  type: 'percent' as 'percent' | 'fixed',
  value: 10,
  min_order: 0,
  is_active: true,
})

const headers = [
  { title: 'Mã', key: 'code' },
  { title: 'Loại', key: 'type', sortable: false },
  { title: 'Giá trị', key: 'value', align: 'end' as const },
  { title: 'Đã dùng', key: 'used_count', align: 'center' as const },
  { title: 'Ngày tạo', key: 'created_at' },
  { title: 'Trạng thái', key: 'is_active', sortable: false },
  { title: 'Thao tác', key: 'actions', sortable: false, align: 'end' as const, width: 140 },
]

const load = async () => {
  loading.value = true
  try {
    const res = await admin.fetchCoupons(dateFilter.withDateQuery({
      page: page.value,
      limit: limit.value,
    }))
    coupons.value = res.items
    applyMeta(res)
  } finally {
    loading.value = false
  }
}

function onFilter() {
  resetPage()
  load()
}

const clearFilters = () => {
  dateFilter.resetDates()
  resetPage()
  load()
}

const formatDate = (iso: string) => new Date(iso).toLocaleDateString('vi-VN')

const openCreate = () => {
  editing.value = false
  editId.value = ''
  form.code = ''
  form.type = 'percent'
  form.value = 10
  form.min_order = 200000
  form.is_active = true
  dialog.value = true
}

const openEdit = (item: Coupon) => {
  editing.value = true
  editId.value = item.id
  form.code = item.code
  form.type = item.type
  form.value = item.value
  form.min_order = item.min_order
  form.is_active = item.is_active
  dialog.value = true
}

const save = async () => {
  saving.value = true
  try {
    if (editing.value) {
      await admin.updateCoupon(editId.value, {
        type: form.type,
        value: form.value,
        min_order: form.min_order,
        is_active: form.is_active,
      })
    } else {
      await admin.createCoupon({ ...form })
    }
    dialog.value = false
    await load()
    snackbar.show('Đã lưu mã giảm giá', 'success')
  } catch (e: unknown) {
    snackbar.show(e instanceof Error ? e.message : 'Không thể lưu mã giảm giá', 'error')
  } finally {
    saving.value = false
  }
}

const remove = async (id: string) => {
  if (!confirm('Xóa mã giảm giá này?')) return
  try {
    await admin.deleteCoupon(id)
    await load()
    snackbar.show('Đã xóa mã giảm giá', 'success')
  } catch (e: unknown) {
    snackbar.show(e instanceof Error ? e.message : 'Không thể xóa mã giảm giá', 'error')
  }
}
</script>
