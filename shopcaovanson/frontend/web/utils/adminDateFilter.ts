export type AdminDateRangeQuery = {
  created_from?: string
  created_to?: string
}

/** YYYY-MM-DD query params for list APIs */
export function dateRangeToQuery(from: string, to: string): AdminDateRangeQuery {
  const query: AdminDateRangeQuery = {}
  if (from?.trim()) query.created_from = from.trim()
  if (to?.trim()) query.created_to = to.trim()
  return query
}

function startOfDay(isoOrDate: string): number {
  const d = new Date(isoOrDate)
  d.setHours(0, 0, 0, 0)
  return d.getTime()
}

function endOfDay(isoOrDate: string): number {
  const d = new Date(isoOrDate)
  d.setHours(23, 59, 59, 999)
  return d.getTime()
}

/** Client-side fallback when API does not filter by date */
export function filterByCreatedAt<T extends { created_at?: string }>(
  items: T[],
  from: string,
  to: string,
): T[] {
  if (!from && !to) return items
  return items.filter((item) => {
    if (!item.created_at) return false
    const ts = new Date(item.created_at).getTime()
    if (Number.isNaN(ts)) return false
    if (from && ts < startOfDay(from)) return false
    if (to && ts > endOfDay(to)) return false
    return true
  })
}

export function mergeDateQuery(
  query: Record<string, string | number>,
  from: string,
  to: string,
): Record<string, string | number> {
  return { ...query, ...dateRangeToQuery(from, to) }
}
