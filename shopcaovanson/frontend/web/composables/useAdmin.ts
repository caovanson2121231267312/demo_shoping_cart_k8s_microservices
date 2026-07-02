import type {
  AdminUserStats,
  OrderStats,
  RoleInfo,
  User,
  UserListResult,
} from '~/types'

export const useAdmin = () => {
  const { apiFetch } = useAuth()
  const authStore = useAuthStore()

  const canManageUsers = computed(() => authStore.can('admin'))
  const canManageProducts = computed(() => authStore.can('manager'))
  const canManageOrders = computed(() => authStore.can('staff'))
  const canViewUsers = computed(() => authStore.can('support'))

  const fetchUserStats = () => apiFetch<AdminUserStats>('/api/admin/stats')
  const fetchOrderStats = () => apiFetch<OrderStats>('/api/admin/orders/stats')
  const fetchRoles = () => apiFetch<RoleInfo[]>('/api/admin/roles')

  const fetchUsers = (query: Record<string, string | number> = {}) =>
    apiFetch<UserListResult>('/api/admin/users', { query })

  const updateUserRole = (id: string, role: string) =>
    apiFetch<User>(`/api/admin/users/${id}/role`, { method: 'PUT', body: { role } })

  const updateUserStatus = (id: string, isActive: boolean) =>
    apiFetch<User>(`/api/admin/users/${id}/status`, { method: 'PUT', body: { is_active: isActive } })

  const fetchUser = (id: string) => apiFetch<User>(`/api/admin/users/${id}`)

  const updateUser = (id: string, body: { full_name: string }) =>
    apiFetch<User>(`/api/admin/users/${id}`, { method: 'PUT', body })

  const uploadUserAvatar = async (id: string, file: File) => {
    const form = new FormData()
    form.append('avatar', file)
    return apiFetch<User>(`/api/admin/users/${id}/avatar`, { method: 'POST', body: form })
  }

  const deleteUserAvatar = (id: string) =>
    apiFetch<User>(`/api/admin/users/${id}/avatar`, { method: 'DELETE' })

  const fetchCoupons = (query: Record<string, string | number> = {}) =>
    apiFetch<import('~/types').CouponListResult>('/api/admin/coupons', { query })

  const createCoupon = (body: Record<string, unknown>) =>
    apiFetch<import('~/types').Coupon>('/api/admin/coupons', { method: 'POST', body })

  const updateCoupon = (id: string, body: Record<string, unknown>) =>
    apiFetch<import('~/types').Coupon>(`/api/admin/coupons/${id}`, { method: 'PUT', body })

  const deleteCoupon = (id: string) =>
    apiFetch(`/api/admin/coupons/${id}`, { method: 'DELETE' })

  return {
    canManageUsers,
    canManageProducts,
    canManageOrders,
    canViewUsers,
    fetchUserStats,
    fetchOrderStats,
    fetchRoles,
    fetchUsers,
    updateUserRole,
    updateUserStatus,
    fetchUser,
    updateUser,
    uploadUserAvatar,
    deleteUserAvatar,
    fetchCoupons,
    createCoupon,
    updateCoupon,
    deleteCoupon,
  }
}
