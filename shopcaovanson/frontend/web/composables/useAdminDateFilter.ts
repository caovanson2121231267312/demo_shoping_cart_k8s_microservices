import { dateRangeToQuery, filterByCreatedAt, mergeDateQuery } from '~/utils/adminDateFilter'

export function useAdminDateFilter() {
  const state = reactive({
    createdFrom: '',
    createdTo: '',
    resetDates() {
      state.createdFrom = ''
      state.createdTo = ''
    },
    withDateQuery(query: Record<string, string | number> = {}) {
      return mergeDateQuery(query, state.createdFrom, state.createdTo)
    },
    filterItems<T extends { created_at?: string }>(items: T[]) {
      return filterByCreatedAt(items, state.createdFrom, state.createdTo)
    },
  })

  const hasDateFilter = computed(() => !!(state.createdFrom || state.createdTo))
  const queryParams = computed(() => dateRangeToQuery(state.createdFrom, state.createdTo))

  return Object.assign(state, { hasDateFilter, queryParams })
}
