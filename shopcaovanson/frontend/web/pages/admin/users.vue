<template>
  <v-container class="page-container py-6">
    <div class="d-flex align-center justify-space-between mb-6">
      <h1 class="text-h4 font-weight-bold">Quản lý người dùng</h1>
      <v-btn to="/admin" variant="text" prepend-icon="mdi-arrow-left">Dashboard</v-btn>
    </div>

    <v-card class="mb-4 pa-4">
      <v-row dense>
        <v-col cols="12" md="4">
          <v-text-field
            v-model="search"
            label="Tìm email / tên"
            prepend-inner-icon="mdi-magnify"
            density="compact"
            variant="outlined"
            hide-details
            @keyup.enter="load"
          />
        </v-col>
        <v-col cols="12" md="3">
          <v-select
            v-model="roleFilter"
            :items="roleOptions"
            label="Vai trò"
            clearable
            density="compact"
            variant="outlined"
            hide-details
          />
        </v-col>
        <v-col cols="12" md="2">
          <v-btn color="primary" @click="load">Tìm</v-btn>
        </v-col>
      </v-row>
    </v-card>

    <LoadingSpinner v-if="loading" />

    <v-card v-else>
      <v-data-table
        :headers="headers"
        :items="users"
        :items-per-page="limit"
        class="elevation-0"
      >
        <template #item.role="{ item }">
          <v-chip size="small">{{ roleLabel(item.role) }}</v-chip>
        </template>
        <template #item.is_active="{ item }">
          <v-chip :color="item.is_active ? 'success' : 'error'" size="small" variant="tonal">
            {{ item.is_active ? 'Hoạt động' : 'Khóa' }}
          </v-chip>
        </template>
        <template #item.actions="{ item }">
          <template v-if="admin.canManageUsers.value">
            <v-menu>
              <template #activator="{ props }">
                <v-btn v-bind="props" size="small" variant="outlined">Vai trò</v-btn>
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
              class="ml-2"
              :color="item.is_active ? 'error' : 'success'"
              variant="tonal"
              @click="toggleStatus(item)"
            >
              {{ item.is_active ? 'Khóa' : 'Mở' }}
            </v-btn>
          </template>
        </template>
      </v-data-table>
      <div class="d-flex justify-center pa-4">
        <v-pagination v-model="page" :length="totalPages" @update:model-value="load" />
      </div>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import type { RoleInfo, User } from '~/types'

definePageMeta({ layout: 'admin' })

const admin = useAdmin()
const users = ref<User[]>([])
const roles = ref<RoleInfo[]>([])
const loading = ref(true)
const search = ref('')
const roleFilter = ref<string | null>(null)
const page = ref(1)
const limit = ref(20)
const totalPages = ref(1)

const headers = [
  { title: 'Email', key: 'email' },
  { title: 'Họ tên', key: 'full_name' },
  { title: 'Vai trò', key: 'role' },
  { title: 'Trạng thái', key: 'is_active' },
  { title: '', key: 'actions', sortable: false },
]

const roleOptions = computed(() => roles.value.map((r) => ({ title: r.label, value: r.id })))

const roleLabel = (id: string) => roles.value.find((r) => r.id === id)?.label ?? id

async function load() {
  loading.value = true
  try {
    const result = await admin.fetchUsers({
      page: page.value,
      limit: limit.value,
      search: search.value || undefined,
      role: roleFilter.value || undefined,
    } as Record<string, string | number>)
    users.value = result.items
    totalPages.value = result.total_pages
  } finally {
    loading.value = false
  }
}

async function changeRole(id: string, role: string) {
  await admin.updateUserRole(id, role)
  await load()
}

async function toggleStatus(user: User) {
  await admin.updateUserStatus(user.id, !user.is_active)
  await load()
}

onMounted(async () => {
  roles.value = await admin.fetchRoles()
  await load()
})
</script>
