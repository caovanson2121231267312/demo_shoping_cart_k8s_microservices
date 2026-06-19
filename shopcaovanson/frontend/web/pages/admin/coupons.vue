<template>
  <v-container class="page-container py-6">
    <div class="d-flex align-center justify-space-between mb-6">
      <h1 class="text-h4 font-weight-bold">Mã giảm giá</h1>
      <v-btn color="primary" prepend-icon="mdi-plus" @click="openCreate">Thêm mã</v-btn>
    </div>

    <LoadingSpinner v-if="loading" />

    <v-card v-else>
      <v-data-table :headers="headers" :items="coupons" :items-per-page="10">
        <template #item.type="{ item }">
          {{ item.type === 'percent' ? 'Phần trăm' : 'Cố định' }}
        </template>
        <template #item.value="{ item }">
          {{ item.type === 'percent' ? `${item.value}%` : formatVND(item.value) }}
        </template>
        <template #item.is_active="{ item }">
          <v-chip :color="item.is_active ? 'success' : 'grey'" size="small">{{ item.is_active ? 'Bật' : 'Tắt' }}</v-chip>
        </template>
        <template #item.actions="{ item }">
          <v-btn size="small" variant="text" @click="openEdit(item)">Sửa</v-btn>
          <v-btn size="small" variant="text" color="error" @click="remove(item.id)">Xóa</v-btn>
        </template>
      </v-data-table>
    </v-card>

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
  </v-container>
</template>

<script setup lang="ts">
import type { Coupon } from '~/types'

definePageMeta({ layout: 'admin' })

const admin = useAdmin()
const { formatVND } = useFormat()

const loading = ref(true)
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
  { title: 'Loại', key: 'type' },
  { title: 'Giá trị', key: 'value' },
  { title: 'Đã dùng', key: 'used_count' },
  { title: 'TT', key: 'is_active' },
  { title: '', key: 'actions', sortable: false },
]

const load = async () => {
  loading.value = true
  try {
    const res = await admin.fetchCoupons({ limit: 100 })
    coupons.value = res.items
  } finally {
    loading.value = false
  }
}

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
    useSnackbar().show('Đã lưu mã giảm giá', 'success')
  } finally {
    saving.value = false
  }
}

const remove = async (id: string) => {
  if (!confirm('Xóa mã giảm giá này?')) return
  await admin.deleteCoupon(id)
  await load()
  useSnackbar().show('Đã xóa mã giảm giá', 'success')
}

onMounted(load)
</script>
