<template>
  <v-container class="page-container py-6">
    <div class="d-flex align-center justify-space-between mb-6">
      <h1 class="text-h4 font-weight-bold">Quản lý sản phẩm</h1>
      <v-btn color="primary" prepend-icon="mdi-plus" @click="openCreate">Thêm sản phẩm</v-btn>
    </div>

    <LoadingSpinner v-if="loading" />

    <v-card v-else>
      <v-data-table
        :headers="headers"
        :items="products"
        :items-per-page="10"
        class="elevation-0"
      >
        <template #item.name="{ item }">
          <div class="d-flex align-center ga-2">
            <v-avatar size="40" rounded>
              <v-img :src="getProductImage(item)" />
            </v-avatar>
            <span>{{ item.name }}</span>
          </div>
        </template>
        <template #item.price="{ item }">
          {{ formatVND(item.price) }}
        </template>
        <template #item.sale_price="{ item }">
          {{ item.sale_price ? formatVND(item.sale_price) : '—' }}
        </template>
        <template #item.is_active="{ item }">
          <v-chip :color="item.is_active ? 'success' : 'grey'" size="small">
            {{ item.is_active ? 'Hiển thị' : 'Ẩn' }}
          </v-chip>
        </template>
        <template #item.actions="{ item }">
          <v-btn icon size="small" variant="text" @click="openEdit(item)">
            <v-icon>mdi-pencil</v-icon>
          </v-btn>
          <v-btn icon size="small" variant="text" color="error" @click="confirmDelete(item)">
            <v-icon>mdi-delete</v-icon>
          </v-btn>
        </template>
      </v-data-table>
    </v-card>

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
  </v-container>
</template>

<script setup lang="ts">
import type { Category, Product } from '~/types'

definePageMeta({ layout: 'admin' })

const { fetchProducts, fetchCategories, createProduct, updateProduct, deleteProduct } = useProducts()
const { formatVND } = useFormat()

const products = ref<Product[]>([])
const categories = ref<Category[]>([])
const loading = ref(true)
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
  { title: 'Sản phẩm', key: 'name' },
  { title: 'Giá', key: 'price' },
  { title: 'KM', key: 'sale_price' },
  { title: 'Tồn kho', key: 'stock' },
  { title: 'Trạng thái', key: 'is_active' },
  { title: '', key: 'actions', sortable: false },
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
      fetchProducts({ limit: 100 }),
      fetchCategories(),
    ])
    products.value = prodResult.items
    categories.value = cats
  } finally {
    loading.value = false
  }
}

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
    dialog.value = false
    await loadData()
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
  } finally {
    deleting.value = false
  }
}

onMounted(loadData)
</script>
