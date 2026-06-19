import type { ProductReview } from '~/types'

export type ReviewSummary = {
  average: number
  total: number
  distribution: Record<number, number>
}

const emptyDistribution = (): Record<number, number> => ({ 1: 0, 2: 0, 3: 0, 4: 0, 5: 0 })

const cache = new Map<string, ReviewSummary>()

const BATCH_SIZE = 100

export const useReviewSummary = () => {
  const { fetchReviews, fetchReviewSummariesBatch } = useProducts()

  const buildSummary = (items: ProductReview[], total: number): ReviewSummary => {
    const distribution: Record<number, number> = emptyDistribution()
    let sum = 0
    for (const r of items) {
      distribution[r.rating] = (distribution[r.rating] || 0) + 1
      sum += r.rating
    }
    const avg = items.length ? sum / items.length : 0
    return { average: Math.round(avg * 10) / 10, total, distribution }
  }

  const toSummary = (average: number, total: number): ReviewSummary => ({
    average,
    total,
    distribution: emptyDistribution(),
  })

  const getSummariesBatch = async (productIds: string[], force = false): Promise<Record<string, ReviewSummary>> => {
    const missing = force
      ? productIds
      : productIds.filter((id) => !cache.has(id))
    const out: Record<string, ReviewSummary> = {}

    for (const id of productIds) {
      const cached = cache.get(id)
      if (cached && !force) {
        out[id] = cached
      }
    }

    if (!missing.length) {
      return out
    }

    for (let i = 0; i < missing.length; i += BATCH_SIZE) {
      const chunk = missing.slice(i, i + BATCH_SIZE)
      try {
        const batch = await fetchReviewSummariesBatch(chunk)
        for (const id of chunk) {
          const row = batch[id]
          const summary = row ? toSummary(row.average, row.total) : toSummary(0, 0)
          cache.set(id, summary)
          out[id] = summary
        }
      } catch {
        for (const id of chunk) {
          const empty = toSummary(0, 0)
          cache.set(id, empty)
          out[id] = empty
        }
      }
    }

    return out
  }

  const getSummary = async (productId: string, force = false): Promise<ReviewSummary> => {
    if (!force && cache.has(productId)) {
      return cache.get(productId)!
    }
    const batch = await getSummariesBatch([productId], force)
    return batch[productId] ?? toSummary(0, 0)
  }

  const getCached = (productId: string) => cache.get(productId)

  return { getSummary, getSummariesBatch, getCached, buildSummary }
}
