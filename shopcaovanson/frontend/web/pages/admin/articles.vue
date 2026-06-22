<template>
  <div>
    <AdminPageHeader title="Quản lý bài viết" subtitle="Blog, tin tức và nội dung SEO">
      <template #actions>
        <v-btn color="primary" prepend-icon="mdi-plus" @click="openCreate">Thêm bài viết</v-btn>
      </template>
    </AdminPageHeader>

    <AdminFilterBar>
      <v-select
        v-model="filterCategory"
        :items="categoryItems"
        label="Danh mục"
        density="compact"
        hide-details
        variant="outlined"
        clearable
        @update:model-value="onFilterCategory"
      />
      <v-text-field
        v-model="filterSearch"
        density="compact"
        hide-details
        label="Tìm kiếm"
        prepend-inner-icon="mdi-magnify"
        variant="outlined"
        @keyup.enter="onFilter"
      />
      <AdminDateRangeFilter v-model:from="dateFilter.createdFrom" v-model:to="dateFilter.createdTo" />
      <v-btn variant="tonal" color="primary" prepend-icon="mdi-filter-outline" @click="onFilter">Lọc</v-btn>
      <v-btn v-if="dateFilter.hasDateFilter" variant="text" @click="clearFilters">Xóa lọc</v-btn>
    </AdminFilterBar>

    <AdminDataTable
      server
      v-model:page="page"
      v-model:items-per-page="limit"
      :headers="headers"
      :items="articles"
      :total-items="total"
      :count="total"
      :loading="loading"
      title="Danh sách bài viết"
      @update:options="load"
    >
      <template #item.title="{ item }">
        <div>
          <div class="admin-table__cell-title">{{ item.title }}</div>
          <div class="admin-table__cell-sub">/{{ item.slug }}</div>
        </div>
      </template>
      <template #item.category="{ item }">
        <v-chip size="small" variant="tonal" color="info">
          {{ ARTICLE_CATEGORIES[item.category] || item.category }}
        </v-chip>
      </template>
      <template #item.is_published="{ item }">
        <v-chip :color="item.is_published ? 'success' : 'grey'" size="small" variant="tonal">
          {{ item.is_published ? 'Đã xuất bản' : 'Nháp' }}
        </v-chip>
      </template>
      <template #item.is_featured="{ item }">
        <v-icon v-if="item.is_featured" color="warning" size="small">mdi-star</v-icon>
        <span v-else class="admin-table__cell-sub">—</span>
      </template>
      <template #item.view_count="{ item }">
        <span class="admin-table__money">{{ item.view_count.toLocaleString('vi-VN') }}</span>
      </template>
      <template #item.created_at="{ item }">
        {{ formatDate(item.created_at) }}
      </template>
      <template #item.actions="{ item }">
        <div class="admin-table-actions">
          <v-btn size="small" variant="text" icon="mdi-open-in-new" :to="`/blog/${item.slug}`" target="_blank" />
          <v-btn size="small" variant="tonal" color="primary" @click="openEdit(item)">Sửa</v-btn>
          <v-btn size="small" variant="tonal" color="error" @click="remove(item.id)">Xóa</v-btn>
        </div>
      </template>
    </AdminDataTable>

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
  </div>
</template>

<script setup lang="ts">
import type { Article } from '~/types'
import { ARTICLE_CATEGORIES } from '~/types'

definePageMeta({ layout: 'admin' })

const { adminFetchArticles, createArticle, updateArticle, deleteArticle } = useArticles()
const snackbar = useSnackbar()
const dateFilter = useAdminDateFilter()
const { page, limit, total, applyMeta, resetPage } = useAdminServerTable(20)

const loading = ref(false)
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
  { title: 'Danh mục', key: 'category', sortable: false },
  { title: 'Lượt xem', key: 'view_count', align: 'end' as const },
  { title: 'Xuất bản', key: 'is_published', sortable: false },
  { title: 'Nổi bật', key: 'is_featured', sortable: false, align: 'center' as const },
  { title: 'Ngày tạo', key: 'created_at' },
  { title: 'Thao tác', key: 'actions', sortable: false, align: 'end' as const, width: 160 },
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
    const query = dateFilter.withDateQuery({
      page: page.value,
      limit: limit.value,
    })
    if (filterCategory.value) query.category = filterCategory.value
    if (filterSearch.value.trim()) query.search = filterSearch.value.trim()
    const res = await adminFetchArticles(query)
    articles.value = res.items
    applyMeta(res)
  } finally {
    loading.value = false
  }
}

function onFilter() {
  resetPage()
  load()
}

function onFilterCategory() {
  resetPage()
  load()
}

const clearFilters = () => {
  filterCategory.value = null
  filterSearch.value = ''
  dateFilter.resetDates()
  resetPage()
  load()
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
</script>
