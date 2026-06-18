<template>
  <v-container class="page-container py-6">
    <div class="d-flex align-center justify-space-between mb-6">
      <h1 class="text-h4 font-weight-bold">Quản lý danh mục</h1>
      <v-btn color="primary" prepend-icon="mdi-plus" @click="openCreate">Thêm danh mục</v-btn>
    </div>

    <LoadingSpinner v-if="loading" />

    <v-row v-else>
      <v-col v-for="cat in flatCategories" :key="cat.id" cols="12" sm="6" md="4">
        <v-card>
          <v-img v-if="cat.image_url" :src="cat.image_url" height="140" cover />
          <v-card-title>{{ cat.name }}</v-card-title>
          <v-card-subtitle>{{ cat.slug }}</v-card-subtitle>
          <v-card-actions>
            <v-btn size="small" variant="text" @click="openEdit(cat)">Sửa</v-btn>
            <v-btn size="small" color="error" variant="text" @click="remove(cat.id)">Xóa</v-btn>
          </v-card-actions>
        </v-card>
      </v-col>
    </v-row>

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
  </v-container>
</template>

<script setup lang="ts">
import type { Category } from '~/types'

definePageMeta({ layout: 'admin' })

const { fetchCategories, createCategory, updateCategory, deleteCategory } = useProducts()

const categories = ref<Category[]>([])
const loading = ref(true)
const dialog = ref(false)
const saving = ref(false)
const editing = ref<Category | null>(null)
const form = reactive({ name: '', slug: '', image_url: '', parent_id: null as string | null })

const flatCategories = computed(() => {
  const out: Category[] = []
  const walk = (items: Category[], depth = 0) => {
    for (const c of items) {
      out.push({ ...c, name: `${'— '.repeat(depth)}${c.name}` })
      if (c.children?.length) walk(c.children, depth + 1)
    }
  }
  walk(categories.value)
  return out
})

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
  } finally {
    saving.value = false
  }
}

async function remove(id: string) {
  if (!confirm('Xóa danh mục này?')) return
  await deleteCategory(id)
  await load()
}

onMounted(load)
</script>
