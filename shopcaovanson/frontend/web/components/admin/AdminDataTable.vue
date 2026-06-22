<template>
  <div class="admin-table-card">
    <div v-if="title || $slots.toolbar" class="admin-table-card__head">
      <div v-if="title" class="admin-table-card__title-wrap">
        <h3 class="admin-table-card__title">{{ title }}</h3>
        <span v-if="displayCount != null" class="admin-table-card__badge">{{ displayCount }}</span>
      </div>
      <div v-if="$slots.toolbar" class="admin-table-card__toolbar">
        <slot name="toolbar" />
      </div>
    </div>

    <v-data-table-server
      v-if="server"
      v-bind="tableAttrs"
      v-model:page="pageModel"
      v-model:items-per-page="itemsPerPageModel"
      :items-length="totalItems ?? 0"
      class="admin-data-table elevation-0"
      :class="{
        'admin-data-table--striped': striped,
        'admin-data-table--loading': loading,
      }"
      :hover="hover"
      :loading="loading"
      :density="tableDensity"
      items-per-page-text="Hiển thị"
      @update:options="onServerOptions"
    >
      <template v-for="(_, name) in $slots" #[name]="slotData">
        <slot v-if="name !== 'toolbar' && name !== 'footer' && name !== 'bottom'" :name="name" v-bind="slotData ?? {}" />
      </template>

      <template #no-data>
        <div class="admin-table-empty">
          <div class="admin-table-empty__icon">
            <v-icon icon="mdi-database-off-outline" size="28" />
          </div>
          <p class="admin-table-empty__title">{{ emptyText }}</p>
          <p class="admin-table-empty__hint">Thử điều chỉnh bộ lọc hoặc thêm dữ liệu mới</p>
        </div>
      </template>

      <template #loading>
        <div class="admin-table-loading">
          <v-progress-circular indeterminate color="primary" size="32" width="3" />
          <span>Đang tải dữ liệu...</span>
        </div>
      </template>

      <template v-if="!$slots.bottom" #bottom>
        <AdminTablePagination
          v-model:page="pageModel"
          v-model:items-per-page="itemsPerPageModel"
          :total-pages="totalPages"
          :range-text="rangeText"
          :items-per-page-options="itemsPerPageOptions"
          :pagination-visible="paginationVisible"
          @change="requestLoad"
        />
      </template>
    </v-data-table-server>

    <v-data-table
      v-else
      v-bind="tableAttrs"
      v-model:page="pageModel"
      v-model:items-per-page="itemsPerPageModel"
      class="admin-data-table elevation-0"
      :class="{
        'admin-data-table--striped': striped,
        'admin-data-table--loading': loading,
      }"
      :hover="hover"
      :loading="loading"
      :density="tableDensity"
      items-per-page-text="Hiển thị"
    >
      <template v-for="(_, name) in $slots" #[name]="slotData">
        <slot v-if="name !== 'toolbar' && name !== 'footer' && name !== 'bottom'" :name="name" v-bind="slotData ?? {}" />
      </template>

      <template #no-data>
        <div class="admin-table-empty">
          <div class="admin-table-empty__icon">
            <v-icon icon="mdi-database-off-outline" size="28" />
          </div>
          <p class="admin-table-empty__title">{{ emptyText }}</p>
          <p class="admin-table-empty__hint">Thử điều chỉnh bộ lọc hoặc thêm dữ liệu mới</p>
        </div>
      </template>

      <template #loading>
        <div class="admin-table-loading">
          <v-progress-circular indeterminate color="primary" size="32" width="3" />
          <span>Đang tải dữ liệu...</span>
        </div>
      </template>

      <template v-if="!$slots.bottom" #bottom>
        <AdminTablePagination
          v-model:page="pageModel"
          v-model:items-per-page="itemsPerPageModel"
          :total-pages="totalPages"
          :range-text="rangeText"
          :items-per-page-options="itemsPerPageOptions"
          :pagination-visible="paginationVisible"
        />
      </template>
    </v-data-table>

    <div v-if="$slots.footer" class="admin-table-card__footer">
      <slot name="footer" />
    </div>
  </div>
</template>

<script setup lang="ts">
defineOptions({ inheritAttrs: false })

const props = withDefaults(defineProps<{
  title?: string
  count?: number
  totalItems?: number
  server?: boolean
  striped?: boolean
  hover?: boolean
  loading?: boolean
  emptyText?: string
  density?: 'default' | 'comfortable' | 'compact'
  paginationVisible?: number
}>(), {
  server: false,
  striped: false,
  hover: true,
  loading: false,
  emptyText: 'Không có dữ liệu',
  paginationVisible: 7,
})

const emit = defineEmits<{
  'update:options': []
}>()

const pageModel = defineModel<number>('page', { default: 1 })
const itemsPerPageModel = defineModel<number>('itemsPerPage', { default: 20 })

const attrs = useAttrs()

const tableAttrs = computed(() => {
  const { class: _c, ...rest } = attrs
  return rest
})

const clientItemsCount = computed(() => {
  const items = attrs.items
  return Array.isArray(items) ? items.length : 0
})

const itemsPerPageOptions = [
  { value: 10, title: '10' },
  { value: 20, title: '20' },
  { value: 50, title: '50' },
  { value: 100, title: '100' },
]

const layout = useAdminLayout()

const tableDensity = computed(() => {
  if (props.density) return props.density
  return layout.prefs.value.density === 'compact' ? 'compact' : 'comfortable'
})

const displayCount = computed(() => {
  if (props.server) return props.totalItems ?? props.count
  return props.count ?? clientItemsCount.value
})

const recordTotal = computed(() => {
  if (props.server) return props.totalItems ?? 0
  return clientItemsCount.value
})

const totalPages = computed(() => {
  const total = recordTotal.value
  const perPage = itemsPerPageModel.value || 20
  if (total === 0) return 1
  return Math.ceil(total / perPage)
})

const rangeText = computed(() => {
  const total = recordTotal.value
  if (total === 0) return '0 bản ghi'
  const start = (pageModel.value - 1) * itemsPerPageModel.value + 1
  const end = Math.min(pageModel.value * itemsPerPageModel.value, total)
  return `${start}–${end} / ${total} bản ghi`
})

const bootstrapped = ref(false)
const lastRequest = ref({ page: 0, itemsPerPage: 0 })

function requestLoad() {
  lastRequest.value = {
    page: pageModel.value,
    itemsPerPage: itemsPerPageModel.value,
  }
  bootstrapped.value = true
  emit('update:options')
}

function onServerOptions(options: { page: number; itemsPerPage: number }) {
  const nextPage = options.page || 1
  const nextPerPage = options.itemsPerPage || 20
  const sameAsLast = lastRequest.value.page === nextPage
    && lastRequest.value.itemsPerPage === nextPerPage

  if (bootstrapped.value && sameAsLast) {
    return
  }

  bootstrapped.value = true
  lastRequest.value = { page: nextPage, itemsPerPage: nextPerPage }

  if (pageModel.value !== nextPage) pageModel.value = nextPage
  if (itemsPerPageModel.value !== nextPerPage) itemsPerPageModel.value = nextPerPage

  emit('update:options')
}
</script>
