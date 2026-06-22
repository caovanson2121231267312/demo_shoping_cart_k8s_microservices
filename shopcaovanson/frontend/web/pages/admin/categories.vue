<template>
  <div>
    <AdminPageHeader title="Quản lý danh mục" subtitle="Cấu trúc ngành hàng và phân cấp">
      <template #actions>
        <v-btn color="primary" prepend-icon="mdi-plus" @click="openCreate">Thêm danh mục</v-btn>
      </template>
    </AdminPageHeader>

    <LoadingSpinner v-if="loading" />

    <AdminDataTable
      v-else
      :headers="headers"
      :items="flatCategories"
      :items-per-page="15"
      :count="flatCategories.length"
      title="Danh sách danh mục"
    >
      <template #item.name="{ item }">
        <div class="admin-table__avatar-cell">
          <v-avatar v-if="item.image_url" size="40" rounded="lg">
            <v-img :src="item.image_url" cover />
          </v-avatar>
          <v-avatar v-else size="40" rounded="lg" color="primary" variant="tonal">
            <v-icon size="20">mdi-shape-outline</v-icon>
          </v-avatar>
          <div>
            <div class="admin-table__cell-title">{{ item.displayName }}</div>
            <div class="admin-table__cell-sub">{{ item.slug }}</div>
          </div>
        </div>
      </template>
      <template #item.parent="{ item }">
        <span v-if="item.parentName" class="admin-table__cell-sub">{{ item.parentName }}</span>
        <v-chip v-else size="small" variant="tonal" color="primary">Gốc</v-chip>
      </template>
      <template #item.actions="{ item }">
        <div class="admin-table-actions">
          <v-btn size="small" variant="tonal" color="primary" @click="openEdit(item.raw)">Sửa</v-btn>
          <v-btn size="small" variant="tonal" color="error" @click="remove(item.raw.id)">Xóa</v-btn>
        </div>
      </template>
    </AdminDataTable>

    <v-dialog v-model="dialog" max-width="480">
      <v-card class="pa-4">
        <v-card-title>{{ editing ? 'Sửa danh mục' : 'Thêm danh mục' }}</v-card-title>
        <v-text-field v-model="form.name" label="Tên" variant="outlined" class="mt-2" />
        <v-text-field v-model="form.slug" label="Slug" variant="outlined" />
        <v-text-field v-model="form.image_url" label="Ảnh URL" variant="outlined" />
        <v-select
          v-model="form.parent_id"
          :items="parentOptions"
          item-title="name"
          item-value="id"
          label="Danh mục cha"
          clearable
          variant="outlined"
        />
        <v-card-actions class="px-0">
          <v-spacer />
          <v-btn variant="text" @click="dialog = false">Hủy</v-btn>
          <v-btn color="primary" :loading="saving" @click="save">Lưu</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import type { Category } from '~/types'

definePageMeta({ layout: 'admin' })

const { fetchCategories, createCategory, updateCategory, deleteCategory } = useProducts()
const snackbar = useSnackbar()

const categories = ref<Category[]>([])
const loading = ref(false)
const dialog = ref(false)
const saving = ref(false)
const editing = ref<Category | null>(null)
const form = reactive({ name: '', slug: '', image_url: '', parent_id: null as string | null })

type CategoryRow = Category & {
  displayName: string
  parentName: string | null
  raw: Category
}

const parentMap = computed(() => {
  const map = new Map<string, string>()
  const walk = (items: Category[]) => {
    for (const c of items) {
      map.set(c.id, c.name)
      if (c.children?.length) walk(c.children)
    }
  }
  walk(categories.value)
  return map
})

const flatCategories = computed<CategoryRow[]>(() => {
  const out: CategoryRow[] = []
  const walk = (items: Category[], depth = 0) => {
    for (const c of items) {
      const cleanName = c.name.replace(/^—+\s*/, '')
      const raw = { ...c, name: cleanName }
      out.push({
        ...c,
        displayName: `${'— '.repeat(depth)}${cleanName}`,
        parentName: c.parent_id ? parentMap.value.get(c.parent_id) ?? null : null,
        raw,
      })
      if (c.children?.length) walk(c.children, depth + 1)
    }
  }
  walk(categories.value)
  return out
})

const headers = [
  { title: 'Danh mục', key: 'name', sortable: false },
  { title: 'Danh mục cha', key: 'parent', sortable: false },
  { title: 'Thao tác', key: 'actions', sortable: false, align: 'end' as const, width: 160 },
]

const parentOptions = computed(() =>
  categories.value.filter((c) => !c.parent_id).map((c) => ({ id: c.id, name: c.name })),
)

async function load() {
  loading.value = true
  try {
    categories.value = await fetchCategories()
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editing.value = null
  Object.assign(form, { name: '', slug: '', image_url: '', parent_id: null })
  dialog.value = true
}

function openEdit(cat: Category) {
  editing.value = cat
  Object.assign(form, {
    name: cat.name.replace(/^—+\s*/, ''),
    slug: cat.slug,
    image_url: cat.image_url ?? '',
    parent_id: cat.parent_id ?? null,
  })
  dialog.value = true
}

async function save() {
  saving.value = true
  try {
    const body = {
      name: form.name,
      slug: form.slug,
      image_url: form.image_url || null,
      parent_id: form.parent_id || null,
    }
    if (editing.value) {
      await updateCategory(editing.value.id, body)
    } else {
      await createCategory(body)
    }
    dialog.value = false
    await load()
    snackbar.show(editing.value ? 'Đã cập nhật danh mục' : 'Đã thêm danh mục', 'success')
  } catch (e: unknown) {
    snackbar.show(e instanceof Error ? e.message : 'Không thể lưu danh mục', 'error')
  } finally {
    saving.value = false
  }
}

async function remove(id: string) {
  if (!confirm('Xóa danh mục này?')) return
  try {
    await deleteCategory(id)
    await load()
    snackbar.show('Đã xóa danh mục', 'success')
  } catch (e: unknown) {
    snackbar.show(e instanceof Error ? e.message : 'Không thể xóa danh mục', 'error')
  }
}

onMounted(load)
</script>
