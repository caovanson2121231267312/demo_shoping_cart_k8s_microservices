<template>
  <div>
    <AdminPageHeader title="Quản lý sản phẩm" subtitle="Danh sách sản phẩm, giá và tồn kho">
      <template #actions>
        <v-btn color="primary" prepend-icon="mdi-plus" @click="openCreate">Thêm sản phẩm</v-btn>
      </template>
    </AdminPageHeader>

    <AdminFilterBar>
      <AdminDateRangeFilter v-model:from="dateFilter.createdFrom" v-model:to="dateFilter.createdTo" />
      <v-btn color="primary" prepend-icon="mdi-filter-outline" @click="onFilter">Lọc</v-btn>
      <v-btn v-if="dateFilter.hasDateFilter" variant="text" @click="clearFilters">Xóa lọc</v-btn>
    </AdminFilterBar>

    <AdminDataTable
      server
      cursor-mode
      v-model:items-per-page="limit"
      :headers="headers"
      :items="products"
      :total-items="total"
      :count="total"
      :loading="loading"
      :has-more="hasMore"
      :has-prev="hasPrev"
      :cursor-page-number="pageIndex + 1"
      :range-offset="rangeOffset"
      title="Danh sách sản phẩm"
      @update:options="onTableOptions"
      @cursor-next="onNextPage"
      @cursor-prev="onPrevPage"
    >      <template #item.name="{ item }">
        <div class="admin-table__avatar-cell">
          <v-avatar size="44" rounded="lg">
            <v-img :src="getProductImage(item)" cover />
          </v-avatar>
          <div>
            <div class="admin-table__cell-title">{{ item.name }}</div>
            <div class="admin-table__cell-sub">{{ item.slug }}</div>
          </div>
        </div>
      </template>
      <template #item.price="{ item }">
        <span class="admin-table__money">{{ formatVND(item.price) }}</span>
      </template>
      <template #item.sale_price="{ item }">
        <span v-if="item.sale_price" class="admin-table__money admin-table__money--sale">
          {{ formatVND(item.sale_price) }}
        </span>
        <span v-else class="admin-table__cell-sub">—</span>
      </template>
      <template #item.stock="{ item }">
        <v-chip
          :color="item.stock > 10 ? 'success' : item.stock > 0 ? 'warning' : 'error'"
          size="small"
          variant="tonal"
        >
          {{ item.stock }}
        </v-chip>
      </template>
      <template #item.created_at="{ item }">
        {{ formatDate(item.created_at) }}
      </template>
      <template #item.is_active="{ item }">
        <v-chip :color="item.is_active ? 'success' : 'grey'" size="small" variant="tonal">
          {{ item.is_active ? 'Hiển thị' : 'Ẩn' }}
        </v-chip>
      </template>
      <template #item.actions="{ item }">
        <div class="admin-table-actions">
          <v-btn icon size="small" variant="text" color="primary" @click="openEdit(item)">
            <v-icon>mdi-pencil-outline</v-icon>
          </v-btn>
          <v-btn icon size="small" variant="text" color="error" @click="confirmDelete(item)">
            <v-icon>mdi-delete-outline</v-icon>
          </v-btn>
        </div>
      </template>
    </AdminDataTable>

    <v-dialog v-model="dialog" max-width="600" persistent>
      <v-card>
        <v-card-title>{{ editing ? 'Sửa sản phẩm' : 'Thêm sản phẩm' }}</v-card-title>
        <v-card-text>
          <v-form @submit.prevent="saveProduct">
            <v-text-field v-model="form.name" label="Tên sản phẩm" variant="outlined" class="mb-2" />
            <v-text-field v-model="form.slug" label="Slug" variant="outlined" class="mb-2" />
            <v-select
              v-model="form.category_id"
              :items="flatCategories"
              item-title="name"
              item-value="id"
              label="Danh mục"
              variant="outlined"
              class="mb-2"
            />
            <v-textarea v-model="form.description" label="Mô tả" rows="2" variant="outlined" class="mb-2" />
            <v-text-field v-model.number="form.price" label="Giá" type="number" variant="outlined" class="mb-2" />
            <v-text-field
              v-model.number="form.sale_price"
              label="Giá khuyến mãi"
              type="number"
              variant="outlined"
              class="mb-2"
            />
            <v-text-field v-model.number="form.stock" label="Tồn kho" type="number" variant="outlined" class="mb-2" />
            <v-switch v-model="form.is_active" label="Hiển thị" color="primary" />
          </v-form>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="dialog = false">Hủy</v-btn>
          <v-btn color="primary" :loading="saving" @click="saveProduct">Lưu</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <ConfirmDialog
      v-model="deleteDialog"
      title="Xóa sản phẩm"
      :message="deleteMessage"
      confirm-text="Xóa"
      :loading="deleting"
      @confirm="handleDelete"
    />
  </div>
</template>

<script setup lang="ts">
import type { Category, Product } from '~/types'

definePageMeta({ layout: 'admin' })

const { fetchProducts, fetchCategories, createProduct, updateProduct, deleteProduct } = useProducts()
const { formatVND } = useFormat()
const dateFilter = useAdminDateFilter()
const snackbar = useSnackbar()
const { limit, total, hasMore, hasPrev, pageIndex, rangeOffset, applyMeta, reset, goNext, goPrev, queryParams } = useAdminCursorTable(20)

const products = ref<Product[]>([])
const categories = ref<Category[]>([])
const loading = ref(false)
const dialog = ref(false)
const deleteDialog = ref(false)
const editing = ref(false)
const saving = ref(false)
const deleting = ref(false)
const editId = ref('')
const deleteTarget = ref<Product | null>(null)

const deleteMessage = computed(
  () => `Bạn có chắc muốn xóa "${deleteTarget.value?.name || ''}"?`,
)

const form = reactive({
  name: '',
  slug: '',
  category_id: '',
  description: '',
  price: 0,
  sale_price: null as number | null,
  stock: 0,
  is_active: true,
})

const headers = [
  { title: 'Sản phẩm', key: 'name', sortable: false },
  { title: 'Giá', key: 'price', align: 'end' as const },
  { title: 'Khuyến mãi', key: 'sale_price', align: 'end' as const },
  { title: 'Tồn kho', key: 'stock', align: 'center' as const },
  { title: 'Ngày tạo', key: 'created_at' },
  { title: 'Trạng thái', key: 'is_active', sortable: false },
  { title: 'Thao tác', key: 'actions', sortable: false, align: 'end' as const, width: 100 },
]

const flatCategories = computed(() => {
  const flat: Category[] = []
  for (const cat of categories.value) {
    flat.push(cat)
    if (cat.children) {
      flat.push(...cat.children)
    }
  }
  return flat
})

const loadData = async () => {
  loading.value = true
  try {
    const [prodResult, cats] = await Promise.all([
      fetchProducts({
        ...queryParams(),
        include_inactive: true,
        ...dateFilter.queryParams.value,
      }),
      categories.value.length ? Promise.resolve(categories.value) : fetchCategories(),
    ])
    products.value = prodResult.items
    applyMeta(prodResult)
    if (Array.isArray(cats) && !categories.value.length) {
      categories.value = cats
    }
  } finally {
    loading.value = false
  }
}

function onTableOptions() {
  reset()
  loadData()
}

function onNextPage() {
  if (goNext()) loadData()
}

function onPrevPage() {
  if (goPrev()) loadData()
}

const clearFilters = () => {
  dateFilter.resetDates()
  reset()
  loadData()
}

function onFilter() {
  reset()
  loadData()
}

const formatDate = (date: string) => new Date(date).toLocaleDateString('vi-VN')

onMounted(async () => {
  categories.value = await fetchCategories()
})

const resetForm = () => {
  form.name = ''
  form.slug = ''
  form.category_id = flatCategories.value[0]?.id || ''
  form.description = ''
  form.price = 0
  form.sale_price = null
  form.stock = 0
  form.is_active = true
}

const openCreate = () => {
  editing.value = false
  editId.value = ''
  resetForm()
  dialog.value = true
}

const openEdit = (product: Product) => {
  editing.value = true
  editId.value = product.id
  form.name = product.name
  form.slug = product.slug
  form.category_id = product.category_id
  form.description = product.description || ''
  form.price = product.price
  form.sale_price = product.sale_price ?? null
  form.stock = product.stock
  form.is_active = product.is_active
  dialog.value = true
}

const saveProduct = async () => {
  saving.value = true
  try {
    const payload = {
      ...form,
      images: [productImageUrl(form.slug || 'product')],
      sale_price: form.sale_price || undefined,
    }
    if (editing.value) {
      await updateProduct(editId.value, payload)
    } else {
      await createProduct(payload)
    }
    const wasEditing = editing.value
    dialog.value = false
    await loadData()
    snackbar.show(wasEditing ? 'Đã cập nhật sản phẩm' : 'Đã thêm sản phẩm', 'success')
  } catch (e: unknown) {
    snackbar.show(e instanceof Error ? e.message : 'Không thể lưu sản phẩm', 'error')
  } finally {
    saving.value = false
  }
}

const confirmDelete = (product: Product) => {
  deleteTarget.value = product
  deleteDialog.value = true
}

const handleDelete = async () => {
  if (!deleteTarget.value) {
    return
  }
  deleting.value = true
  try {
    await deleteProduct(deleteTarget.value.id)
    deleteDialog.value = false
    await loadData()
    snackbar.show('Đã xóa sản phẩm', 'success')
  } catch (e: unknown) {
    snackbar.show(e instanceof Error ? e.message : 'Không thể xóa sản phẩm', 'error')
  } finally {
    deleting.value = false
  }
}
</script>
