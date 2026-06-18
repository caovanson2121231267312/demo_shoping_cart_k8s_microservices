import type { ProductReview } from '~/types'

export type ReviewSummary = {
  average: number
  total: number
  distribution: Record<number, number>
}

const cache = new Map<string, ReviewSummary>()

export const useReviewSummary = () => {
  const { fetchReviews } = useProducts()

  const buildSummary = (items: ProductReview[], total: number): ReviewSummary => {
    const distribution: Record<number, number> = { 1: 0, 2: 0, 3: 0, 4: 0, 5: 0 }
    let sum = 0
    for (const r of items) {
      distribution[r.rating] = (distribution[r.rating] || 0) + 1
      sum += r.rating
    }
    const avg = items.length ? sum / items.length : 0
    return { average: Math.round(avg * 10) / 10, total, distribution }
  }

  const getSummary = async (productId: string, force = false): Promise<ReviewSummary> => {
    if (!force && cache.has(productId)) {
      return cache.get(productId)!
    }
    try {
      const result = await fetchReviews(productId, 1, 100)
      const summary = buildSummary(result.items, result.total)
      cache.set(productId, summary)
      return summary
    } catch {
      const empty = { average: 0, total: 0, distribution: { 1: 0, 2: 0, 3: 0, 4: 0, 5: 0 } }
      cache.set(productId, empty)
      return empty
    }
  }

  const getCached = (productId: string) => cache.get(productId)

  return { getSummary, getCached, buildSummary }
}
