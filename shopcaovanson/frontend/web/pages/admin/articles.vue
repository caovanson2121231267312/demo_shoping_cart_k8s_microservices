<template>
  <v-container class="page-container py-6">
    <div class="d-flex align-center justify-space-between mb-6 flex-wrap ga-3">
      <h1 class="text-h4 font-weight-bold">Quản lý bài viết</h1>
      <v-btn color="primary" prepend-icon="mdi-plus" @click="openCreate">Thêm bài viết</v-btn>
    </div>

    <div class="d-flex flex-wrap ga-3 mb-4">
      <v-select
        v-model="filterCategory"
        :items="categoryItems"
        label="Danh mục"
        density="compact"
        hide-details
        variant="outlined"
        style="max-width:200px"
        clearable
        @update:model-value="load"
      />
      <v-text-field
        v-model="filterSearch"
        density="compact"
        hide-details
        label="Tìm kiếm"
        prepend-inner-icon="mdi-magnify"
        variant="outlined"
        style="max-width:260px"
        @keyup.enter="load"
      />
      <v-btn variant="outlined" @click="load">Lọc</v-btn>
    </div>

    <LoadingSpinner v-if="loading" />

    <v-card v-else>
      <v-data-table :headers="headers" :items="articles" :items-per-page="10">
        <template #item.title="{ item }">
          <div class="font-weight-medium">{{ item.title }}</div>
          <div class="text-caption text-grey">/{{ item.slug }}</div>
        </template>
        <template #item.category="{ item }">
          {{ ARTICLE_CATEGORIES[item.category] || item.category }}
        </template>
        <template #item.is_published="{ item }">
          <v-chip :color="item.is_published ? 'success' : 'grey'" size="small">
            {{ item.is_published ? 'Đã xuất bản' : 'Nháp' }}
          </v-chip>
        </template>
        <template #item.is_featured="{ item }">
          <v-icon v-if="item.is_featured" color="warning" size="small">mdi-star</v-icon>
        </template>
        <template #item.view_count="{ item }">
          {{ item.view_count }}
        </template>
        <template #item.created_at="{ item }">
          {{ formatDate(item.created_at) }}
        </template>
        <template #item.actions="{ item }">
          <v-btn size="small" variant="text" :to="`/blog/${item.slug}`" target="_blank">Xem</v-btn>
          <v-btn size="small" variant="text" @click="openEdit(item)">Sửa</v-btn>
          <v-btn size="small" variant="text" color="error" @click="remove(item.id)">Xóa</v-btn>
        </template>
      </v-data-table>
    </v-card>

    <v-dialog v-model="dialog" max-width="900" scrollable persistent>
      <v-card>
        <v-card-title>{{ editing ? 'Sửa bài viết' : 'Thêm bài viết' }}</v-card-title>
        <v-card-text>
          <v-row dense>
            <v-col cols="12" md="8">
              <v-text-field v-model="form.title" label="Tiêu đề" variant="outlined" class="mb-2" />
            </v-col>
            <v-col cols="12" md="4">
              <v-text-field v-model="form.slug" label="Slug (URL)" variant="outlined" class="mb-2" hint="Để trống sẽ tự tạo" persistent-hint />
            </v-col>
            <v-col cols="12" md="6">
              <v-select
                v-model="form.category"
                :items="categoryItems"
                label="Danh mục"
                variant="outlined"
                class="mb-2"
              />
            </v-col>
            <v-col cols="12" md="6">
              <v-text-field v-model="form.author_name" label="Tác giả" variant="outlined" class="mb-2" />
            </v-col>
            <v-col cols="12">
              <v-text-field v-model="form.cover_image" label="Ảnh bìa (URL)" variant="outlined" class="mb-2" />
            </v-col>
            <v-col cols="12">
              <v-textarea v-model="form.excerpt" label="Mô tả ngắn" rows="2" variant="outlined" class="mb-2" />
            </v-col>
            <v-col cols="12">
              <div class="text-caption text-grey mb-1">Nội dung</div>
              <SummernoteEditor v-model="form.content" :height="360" />
            </v-col>
            <v-col cols="12">
              <v-combobox
                v-model="form.tags"
                label="Thẻ (tags)"
                chips
                multiple
                closable-chips
                variant="outlined"
                hint="Nhập và Enter để thêm tag"
                persistent-hint
              />
            </v-col>
            <v-col cols="6">
              <v-switch v-model="form.is_published" label="Xuất bản" color="primary" />
            </v-col>
            <v-col cols="6">
              <v-switch v-model="form.is_featured" label="Nổi bật trang chủ" color="warning" />
            </v-col>
          </v-row>
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
import type { Article } from '~/types'
import { ARTICLE_CATEGORIES } from '~/types'

definePageMeta({ layout: 'admin' })

const { adminFetchArticles, createArticle, updateArticle, deleteArticle } = useArticles()
const snackbar = useSnackbar()

const loading = ref(true)
const saving = ref(false)
const dialog = ref(false)
const editing = ref(false)
const editId = ref('')
const articles = ref<Article[]>([])
const filterCategory = ref<string | null>(null)
const filterSearch = ref('')

const form = reactive({
  title: '',
  slug: '',
  excerpt: '',
  content: '',
  cover_image: '',
  category: 'tin-tuc',
  tags: [] as string[],
  author_name: 'Shop Cao Văn Sơn',
  is_published: true,
  is_featured: false,
})

const categoryItems = Object.entries(ARTICLE_CATEGORIES).map(([value, title]) => ({ title, value }))

const headers = [
  { title: 'Tiêu đề', key: 'title', sortable: false },
  { title: 'Danh mục', key: 'category' },
  { title: 'Lượt xem', key: 'view_count' },
  { title: 'Xuất bản', key: 'is_published' },
  { title: 'Nổi bật', key: 'is_featured' },
  { title: 'Ngày tạo', key: 'created_at' },
  { title: '', key: 'actions', sortable: false },
]

const formatDate = (iso: string) => new Date(iso).toLocaleDateString('vi-VN')

const resetForm = () => {
  form.title = ''
  form.slug = ''
  form.excerpt = ''
  form.content = ''
  form.cover_image = ''
  form.category = 'tin-tuc'
  form.tags = []
  form.author_name = 'Shop Cao Văn Sơn'
  form.is_published = true
  form.is_featured = false
}

const load = async () => {
  loading.value = true
  try {
    const query: Record<string, string | number> = { limit: 100 }
    if (filterCategory.value) query.category = filterCategory.value
    if (filterSearch.value.trim()) query.search = filterSearch.value.trim()
    const res = await adminFetchArticles(query)
    articles.value = res.items
  } finally {
    loading.value = false
  }
}

const openCreate = () => {
  editing.value = false
  editId.value = ''
  resetForm()
  dialog.value = true
}

const openEdit = (item: Article) => {
  editing.value = true
  editId.value = item.id
  form.title = item.title
  form.slug = item.slug
  form.excerpt = item.excerpt || ''
  form.content = item.content
  form.cover_image = item.cover_image || ''
  form.category = item.category
  form.tags = [...(item.tags || [])]
  form.author_name = item.author_name
  form.is_published = item.is_published
  form.is_featured = item.is_featured
  dialog.value = true
}

const save = async () => {
  if (!form.title.trim()) {
    snackbar.show('Vui lòng nhập tiêu đề', 'error')
    return
  }
  saving.value = true
  try {
    const body = { ...form, slug: form.slug.trim() || undefined }
    if (editing.value) {
      await updateArticle(editId.value, body)
      snackbar.show('Đã cập nhật bài viết', 'success')
    } else {
      await createArticle(body)
      snackbar.show('Đã tạo bài viết', 'success')
    }
    dialog.value = false
    await load()
  } catch (e: unknown) {
    snackbar.show(e instanceof Error ? e.message : 'Lỗi lưu bài viết', 'error')
  } finally {
    saving.value = false
  }
}

const remove = async (id: string) => {
  if (!confirm('Xóa bài viết này?')) return
  try {
    await deleteArticle(id)
    snackbar.show('Đã xóa bài viết', 'success')
    await load()
  } catch {
    snackbar.show('Không thể xóa bài viết', 'error')
  }
}

onMounted(load)
</script>
