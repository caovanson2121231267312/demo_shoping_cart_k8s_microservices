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
      cursor-mode
      v-model:items-per-page="limit"
      :headers="headers"
      :items="users"
      :total-items="total"
      :count="total"
      :loading="loading"
      :has-more="hasMore"
      :has-prev="hasPrev"
      :cursor-page-number="pageIndex + 1"
      :range-offset="rangeOffset"
      title="Danh sách người dùng"
      @update:options="onTableOptions"
      @cursor-next="onNextPage"
      @cursor-prev="onPrevPage"
    >
      <template #item.avatar="{ item }">
        <UserAvatar
          :user-id="item.id"
          :name="item.full_name || item.email"
          :has-avatar="item.has_avatar"
          :size="40"
        />
      </template>
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
          <v-btn size="small" variant="tonal" color="primary" @click="openEdit(item)">Sửa</v-btn>
          <v-menu>
            <template #activator="{ props }">
              <v-btn v-bind="props" size="small" variant="tonal">Vai trò</v-btn>
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

    <v-dialog v-model="dialog" max-width="560" persistent>
      <v-card>
        <v-card-title class="d-flex align-center justify-space-between">
          <span>Chi tiết người dùng</span>
          <v-btn icon variant="text" @click="dialog = false">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-card-title>
        <v-card-text>
          <div class="user-edit__avatar-section">
            <div class="user-edit__avatar-preview">
              <v-avatar v-if="previewUrl" size="96" class="user-edit__avatar-ring">
                <v-img :src="previewUrl" cover alt="Avatar preview" />
              </v-avatar>
              <UserAvatar
                v-else
                :user-id="editUser?.id"
                :name="form.full_name || editUser?.email"
                :has-avatar="editUser?.has_avatar && !removeAvatar"
                :reload-key="avatarReloadKey"
                :size="96"
              />
            </div>
            <div class="user-edit__avatar-actions">
              <v-btn
                size="small"
                variant="tonal"
                color="primary"
                prepend-icon="mdi-camera-outline"
                @click="fileInput?.click()"
              >
                Chọn ảnh
              </v-btn>
              <v-btn
                v-if="editUser?.has_avatar || previewUrl"
                size="small"
                variant="text"
                color="error"
                @click="markRemoveAvatar"
              >
                Xóa avatar
              </v-btn>
              <p class="user-edit__avatar-hint">JPG, PNG, WebP — tối đa 2MB. Ảnh lưu bảo mật trên MinIO.</p>
            </div>
            <input
              ref="fileInput"
              type="file"
              accept="image/jpeg,image/png,image/webp,image/gif"
              class="d-none"
              @change="onFileSelected"
            >
          </div>

          <v-text-field
            v-model="form.full_name"
            label="Họ và tên"
            variant="outlined"
            class="mb-2"
            :rules="[rules.required]"
          />
          <v-text-field
            :model-value="editUser?.email"
            label="Email"
            variant="outlined"
            class="mb-2"
            disabled
          />
          <v-text-field
            :model-value="editUser ? roleLabel(editUser.role) : ''"
            label="Vai trò"
            variant="outlined"
            class="mb-2"
            disabled
          />
          <v-switch
            v-model="form.is_active"
            label="Tài khoản hoạt động"
            color="primary"
            hide-details
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="dialog = false">Hủy</v-btn>
          <v-btn color="primary" :loading="saving" @click="save">Lưu thay đổi</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import type { RoleInfo, User } from '~/types'

definePageMeta({ layout: 'admin' })

const admin = useAdmin()
const dateFilter = useAdminDateFilter()
const snackbar = useSnackbar()
const { limit, total, hasMore, hasPrev, pageIndex, rangeOffset, applyMeta, reset, goNext, goPrev, queryParams } = useAdminCursorTable(20)
const users = ref<User[]>([])
const roles = ref<RoleInfo[]>([])
const loading = ref(false)
const saving = ref(false)
const dialog = ref(false)
const editUser = ref<User | null>(null)
const search = ref('')
const roleFilter = ref<string | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)
const pendingFile = ref<File | null>(null)
const previewUrl = ref<string | null>(null)
const removeAvatar = ref(false)
const avatarReloadKey = ref(0)

const form = reactive({
  full_name: '',
  is_active: true,
})

const rules = {
  required: (v: string) => Boolean(v?.trim()) || 'Bắt buộc',
}

const headers = [
  { title: '', key: 'avatar', sortable: false, width: 56 },
  { title: 'Tài khoản', key: 'email', sortable: false },
  { title: 'Vai trò', key: 'role', sortable: false },
  { title: 'Trạng thái', key: 'is_active', sortable: false },
  { title: 'Ngày tạo', key: 'created_at' },
  { title: 'Thao tác', key: 'actions', sortable: false, align: 'end' as const, width: 260 },
]

const roleOptions = computed(() => roles.value.map((r) => ({ title: r.label, value: r.id })))

const roleLabel = (id: string) => roles.value.find((r) => r.id === id)?.label ?? id

const formatDate = (iso?: string) => (iso ? new Date(iso).toLocaleDateString('vi-VN') : '—')

function clearPreview() {
  if (previewUrl.value) {
    URL.revokeObjectURL(previewUrl.value)
    previewUrl.value = null
  }
}

function resetEditState() {
  clearPreview()
  pendingFile.value = null
  removeAvatar.value = false
  if (fileInput.value) fileInput.value.value = ''
}

async function load() {
  loading.value = true
  try {
    const result = await admin.fetchUsers(dateFilter.withDateQuery(queryParams({
      ...(search.value ? { search: search.value } : {}),
      ...(roleFilter.value ? { role: roleFilter.value } : {}),
    })))
    users.value = result.items
    applyMeta(result)
  } finally {
    loading.value = false
  }
}

function onTableOptions() {
  reset()
  load()
}

function onNextPage() {
  if (goNext()) load()
}

function onPrevPage() {
  if (goPrev()) load()
}

function onSearch() {
  reset()
  load()
}

function clearFilters() {
  search.value = ''
  roleFilter.value = null
  dateFilter.resetDates()
  reset()
  load()
}

async function openEdit(item: User) {
  resetEditState()
  try {
    const user = await admin.fetchUser(item.id)
    editUser.value = user
    form.full_name = user.full_name
    form.is_active = user.is_active ?? true
    dialog.value = true
  } catch (e: unknown) {
    snackbar.show(e instanceof Error ? e.message : 'Không thể tải thông tin người dùng', 'error')
  }
}

function onFileSelected(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  if (file.size > 2 * 1024 * 1024) {
    snackbar.show('Ảnh không được vượt quá 2MB', 'error')
    input.value = ''
    return
  }
  clearPreview()
  pendingFile.value = file
  removeAvatar.value = false
  previewUrl.value = URL.createObjectURL(file)
}

function markRemoveAvatar() {
  clearPreview()
  pendingFile.value = null
  removeAvatar.value = true
  if (fileInput.value) fileInput.value.value = ''
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

async function save() {
  if (!editUser.value || !form.full_name.trim()) {
    snackbar.show('Họ và tên là bắt buộc', 'error')
    return
  }

  saving.value = true
  try {
    let updated = await admin.updateUser(editUser.value.id, { full_name: form.full_name.trim() })

    if (form.is_active !== (editUser.value.is_active ?? true)) {
      updated = await admin.updateUserStatus(editUser.value.id, form.is_active)
    }

    if (removeAvatar.value && editUser.value.has_avatar) {
      updated = await admin.deleteUserAvatar(editUser.value.id)
    } else if (pendingFile.value) {
      updated = await admin.uploadUserAvatar(editUser.value.id, pendingFile.value)
    }

    editUser.value = updated
    avatarReloadKey.value += 1
    dialog.value = false
    resetEditState()
    await load()
    snackbar.show('Đã cập nhật người dùng', 'success')
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : 'Không thể lưu thay đổi'
    snackbar.show(msg.includes('storage') ? 'MinIO chưa được cấu hình — không thể tải avatar' : msg, 'error')
  } finally {
    saving.value = false
  }
}

onUnmounted(() => {
  clearPreview()
})

onMounted(async () => {
  roles.value = await admin.fetchRoles()
})
</script>

<style scoped>
.user-edit__avatar-section {
  display: flex;
  gap: 20px;
  align-items: center;
  margin-bottom: 20px;
  padding: 16px;
  border-radius: 12px;
  background: var(--color-surface-muted, #f8fafc);
  border: 1px solid var(--color-border-light, rgba(0, 0, 0, 0.06));
}

.user-edit__avatar-preview {
  flex-shrink: 0;
}

.user-edit__avatar-ring {
  box-shadow: 0 4px 14px rgba(21, 101, 192, 0.15);
}

.user-edit__avatar-actions {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 6px;
}

.user-edit__avatar-hint {
  margin: 4px 0 0;
  font-size: 12px;
  color: var(--color-text-muted, #9e9e9e);
  line-height: 1.4;
}
</style>
