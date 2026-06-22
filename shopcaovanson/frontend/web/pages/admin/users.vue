<template>
  <div>
    <AdminPageHeader title="Quản lý người dùng" subtitle="Phân quyền và trạng thái tài khoản">
      <template #actions>
        <v-btn to="/admin" variant="tonal" prepend-icon="mdi-arrow-left">Dashboard</v-btn>
      </template>
    </AdminPageHeader>

    <AdminFilterBar>
      <v-text-field
        v-model="search"
        label="Tìm email / tên"
        prepend-inner-icon="mdi-magnify"
        density="compact"
        variant="outlined"
        hide-details
        @keyup.enter="load"
      />
      <v-select
        v-model="roleFilter"
        :items="roleOptions"
        label="Vai trò"
        clearable
        density="compact"
        variant="outlined"
        hide-details
      />
      <v-btn color="primary" prepend-icon="mdi-magnify" @click="onSearch">Tìm kiếm</v-btn>
      <AdminDateRangeFilter v-model:from="dateFilter.createdFrom" v-model:to="dateFilter.createdTo" />
      <v-btn v-if="dateFilter.hasDateFilter" variant="text" @click="clearFilters">Xóa lọc</v-btn>
    </AdminFilterBar>

    <AdminDataTable
      server
      v-model:page="page"
      v-model:items-per-page="limit"
      :headers="headers"
      :items="users"
      :total-items="total"
      :count="total"
      :loading="loading"
      title="Danh sách người dùng"
      @update:options="load"
    >
      <template #item.email="{ item }">
        <div>
          <div class="admin-table__cell-title">{{ item.email }}</div>
          <div v-if="item.full_name" class="admin-table__cell-sub">{{ item.full_name }}</div>
        </div>
      </template>
      <template #item.role="{ item }">
        <v-chip size="small" variant="tonal" color="primary">{{ roleLabel(item.role) }}</v-chip>
      </template>
      <template #item.is_active="{ item }">
        <v-chip :color="item.is_active ? 'success' : 'error'" size="small" variant="tonal">
          {{ item.is_active ? 'Hoạt động' : 'Khóa' }}
        </v-chip>
      </template>
      <template #item.created_at="{ item }">
        {{ formatDate(item.created_at) }}
      </template>
      <template #item.actions="{ item }">
        <div v-if="admin.canManageUsers.value" class="admin-table-actions">
          <v-menu>
            <template #activator="{ props }">
              <v-btn v-bind="props" size="small" variant="tonal" color="primary">Vai trò</v-btn>
            </template>
            <v-list density="compact">
              <v-list-item
                v-for="r in roles"
                :key="r.id"
                :title="r.label"
                @click="changeRole(item.id, r.id)"
              />
            </v-list>
          </v-menu>
          <v-btn
            size="small"
            :color="item.is_active ? 'error' : 'success'"
            variant="tonal"
            @click="toggleStatus(item)"
          >
            {{ item.is_active ? 'Khóa' : 'Mở' }}
          </v-btn>
        </div>
      </template>
    </AdminDataTable>
  </div>
</template>

<script setup lang="ts">
import type { RoleInfo, User } from '~/types'

definePageMeta({ layout: 'admin' })

const admin = useAdmin()
const dateFilter = useAdminDateFilter()
const snackbar = useSnackbar()
const { page, limit, total, applyMeta, resetPage } = useAdminServerTable(20)
const users = ref<User[]>([])
const roles = ref<RoleInfo[]>([])
const loading = ref(false)
const search = ref('')
const roleFilter = ref<string | null>(null)

const headers = [
  { title: 'Tài khoản', key: 'email', sortable: false },
  { title: 'Vai trò', key: 'role', sortable: false },
  { title: 'Trạng thái', key: 'is_active', sortable: false },
  { title: 'Ngày tạo', key: 'created_at' },
  { title: 'Thao tác', key: 'actions', sortable: false, align: 'end' as const, width: 200 },
]

const roleOptions = computed(() => roles.value.map((r) => ({ title: r.label, value: r.id })))

const roleLabel = (id: string) => roles.value.find((r) => r.id === id)?.label ?? id

const formatDate = (iso?: string) => (iso ? new Date(iso).toLocaleDateString('vi-VN') : '—')

async function load() {
  loading.value = true
  try {
    const result = await admin.fetchUsers(dateFilter.withDateQuery({
      page: page.value,
      limit: limit.value,
      ...(search.value ? { search: search.value } : {}),
      ...(roleFilter.value ? { role: roleFilter.value } : {}),
    }))
    users.value = result.items
    applyMeta(result)
  } finally {
    loading.value = false
  }
}

function onSearch() {
  resetPage()
  load()
}

function clearFilters() {
  search.value = ''
  roleFilter.value = null
  dateFilter.resetDates()
  resetPage()
  load()
}

async function changeRole(id: string, role: string) {
  try {
    await admin.updateUserRole(id, role)
    await load()
    snackbar.show('Đã cập nhật vai trò người dùng', 'success')
  } catch (e: unknown) {
    snackbar.show(e instanceof Error ? e.message : 'Không thể đổi vai trò', 'error')
  }
}

async function toggleStatus(user: User) {
  try {
    await admin.updateUserStatus(user.id, !user.is_active)
    await load()
    snackbar.show(user.is_active ? 'Đã khóa tài khoản' : 'Đã mở khóa tài khoản', 'success')
  } catch (e: unknown) {
    snackbar.show(e instanceof Error ? e.message : 'Không thể cập nhật trạng thái', 'error')
  }
}

onMounted(async () => {
  roles.value = await admin.fetchRoles()
})
</script>
