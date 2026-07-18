export interface CursorPaginatedResult {
  total: number
  limit?: number
  next_cursor?: string
  has_more?: boolean
}

/** Cursor-based server table state for admin lists (no OFFSET). */
export function useAdminCursorTable(initialLimit = 20) {
  const limit = ref(initialLimit)
  const total = ref(0)
  const hasMore = ref(false)
  const nextCursor = ref('')

  /** Stack of cursors used to reach each page (index 0 = first page, no cursor). */
  const cursorStack = ref<(string | undefined)[]>([undefined])
  const pageIndex = ref(0)

  const hasPrev = computed(() => pageIndex.value > 0)
  const currentCursor = computed(() => cursorStack.value[pageIndex.value])

  function applyMeta(result: CursorPaginatedResult) {
    total.value = result.total
    hasMore.value = Boolean(result.has_more)
    nextCursor.value = result.next_cursor ?? ''
    if (result.limit) {
      limit.value = result.limit
    }
  }

  function reset() {
    cursorStack.value = [undefined]
    pageIndex.value = 0
    nextCursor.value = ''
    hasMore.value = false
  }

  function goNext() {
    if (!hasMore.value || !nextCursor.value) return false
    pageIndex.value += 1
    if (cursorStack.value.length <= pageIndex.value) {
      cursorStack.value.push(nextCursor.value)
    }
    return true
  }

  function goPrev() {
    if (pageIndex.value <= 0) return false
    pageIndex.value -= 1
    return true
  }

  function queryParams(extra: Record<string, string | number> = {}) {
    const query: Record<string, string | number> = {
      limit: limit.value,
      ...extra,
    }
    const cursor = currentCursor.value
    if (cursor) {
      query.cursor = cursor
    }
    return query
  }

  const rangeOffset = computed(() => pageIndex.value * limit.value)

  return {
    limit,
    total,
    hasMore,
    hasPrev,
    nextCursor,
    pageIndex,
    currentCursor,
    rangeOffset,
    applyMeta,
    reset,
    goNext,
    goPrev,
    queryParams,
  }
}
