export const useCoupons = () => {
  const { apiFetch } = useAuth()
  const config = useRuntimeConfig()
  const apiBase = (config.public.apiUrl as string) || ''

  const validateCoupon = (code: string, amount: number) =>
    $fetch<import('~/types').ValidateCouponResult>(`${apiBase}/api/coupons/validate`, {
      method: 'POST',
      body: { code, amount },
    })

  const fetchCoupons = (query: Record<string, string | number> = {}) =>
    apiFetch<import('~/types').CouponListResult>('/api/admin/coupons', { query })

  const createCoupon = (body: Record<string, unknown>) =>
    apiFetch<import('~/types').Coupon>('/api/admin/coupons', { method: 'POST', body })

  const updateCoupon = (id: string, body: Record<string, unknown>) =>
    apiFetch<import('~/types').Coupon>(`/api/admin/coupons/${id}`, { method: 'PUT', body })

  const deleteCoupon = (id: string) =>
    apiFetch(`/api/admin/coupons/${id}`, { method: 'DELETE' })

  return { validateCoupon, fetchCoupons, createCoupon, updateCoupon, deleteCoupon }
}
