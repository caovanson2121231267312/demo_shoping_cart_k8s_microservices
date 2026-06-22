export interface PaginatedResult {
  total: number
  page?: number
  limit?: number
  total_pages?: number
}

/** Server-side pagination state for admin tables (API max limit is typically 100). */
export function useAdminServerTable(initialLimit = 20) {
  const page = ref(1)
  const limit = ref(initialLimit)
  const total = ref(0)
  const totalPages = ref(1)

  function applyMeta(result: PaginatedResult) {
    total.value = result.total
    if (result.total_pages != null) {
      totalPages.value = result.total_pages
    } else {
      const perPage = result.limit ?? limit.value
      totalPages.value = Math.max(1, Math.ceil(result.total / perPage))
    }
  }

  function resetPage() {
    page.value = 1
  }

  return { page, limit, total, totalPages, applyMeta, resetPage }
}
