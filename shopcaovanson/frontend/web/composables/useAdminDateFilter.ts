import { dateRangeToQuery, filterByCreatedAt, mergeDateQuery } from '~/utils/adminDateFilter'

export function useAdminDateFilter() {
  const createdFrom = ref('')
  const createdTo = ref('')

  const queryParams = computed(() => dateRangeToQuery(createdFrom.value, createdTo.value))

  const resetDates = () => {
    createdFrom.value = ''
    createdTo.value = ''
  }

  const withDateQuery = (query: Record<string, string | number> = {}) =>
    mergeDateQuery(query, createdFrom.value, createdTo.value)

  const filterItems = <T extends { created_at?: string }>(items: T[]) =>
    filterByCreatedAt(items, createdFrom.value, createdTo.value)

  const hasDateFilter = computed(() => !!(createdFrom.value || createdTo.value))

  return {
    createdFrom,
    createdTo,
    queryParams,
    hasDateFilter,
    resetDates,
    withDateQuery,
    filterItems,
  }
}
