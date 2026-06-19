import type {
  Category,
  CreateReviewInput,
  Product,
  ProductFilters,
  ProductListResult,
  ReviewListResult,
} from '~/types'

export const useProducts = () => {
  const { apiFetch } = useAuth()
  const config = useRuntimeConfig()

  const publicFetch = async <T>(path: string, options: Record<string, unknown> = {}) => {
    const base = (config.public.apiUrl as string) || ''
    const url = path.startsWith('http') ? path : `${base}${path}`
    return await $fetch<T>(url, options)
  }

  const fetchProducts = async (filters: ProductFilters = {}) => {
    const query: Record<string, string | number> = {}
    if (filters.page) query.page = filters.page
    if (filters.limit) query.limit = filters.limit
    if (filters.category) query.category = filters.category
    if (filters.search) query.search = filters.search
    if (filters.min_price !== undefined) query.min_price = filters.min_price
    if (filters.max_price !== undefined) query.max_price = filters.max_price
    if (filters.sort) query.sort = filters.sort

    return await publicFetch<ProductListResult>('/api/products', { query })
  }

  const fetchProductBySlug = async (slug: string) => {
    return await publicFetch<Product>(`/api/products/${slug}`)
  }

  const fetchCategories = async () => {
    return await publicFetch<Category[]>('/api/categories')
  }

  const createCategory = async (input: Partial<Category>) =>
    apiFetch<Category>('/api/admin/categories', { method: 'POST', body: input })

  const updateCategory = async (id: string, input: Partial<Category>) =>
    apiFetch<Category>(`/api/admin/categories/${id}`, { method: 'PUT', body: input })

  const deleteCategory = async (id: string) =>
    apiFetch(`/api/admin/categories/${id}`, { method: 'DELETE' })

  const fetchReviews = async (productId: string, page = 1, limit = 10) => {
    return await publicFetch<ReviewListResult>(`/api/products/${productId}/reviews`, {
      query: { page, limit },
    })
  }

  const fetchReviewSummariesBatch = async (productIds: string[]) => {
    const res = await publicFetch<{ data: Record<string, { average: number; total: number }> }>(
      '/api/products/reviews/summary',
      {
        method: 'POST',
        body: { product_ids: productIds },
      },
    )
    return res.data || {}
  }

  const createReview = async (productId: string, input: CreateReviewInput) => {
    return await apiFetch(`/api/products/${productId}/reviews`, {
      method: 'POST',
      body: input,
    })
  }

  const createProduct = async (product: Partial<Product>) => {
    return await apiFetch<Product>('/api/products', {
      method: 'POST',
      body: product,
    })
  }

  const updateProduct = async (id: string, product: Partial<Product>) => {
    return await apiFetch<Product>(`/api/products/${id}`, {
      method: 'PUT',
      body: product,
    })
  }

  const deleteProduct = async (id: string) => {
    return await apiFetch(`/api/products/${id}`, { method: 'DELETE' })
  }

  const getEffectivePrice = (product: Product) => {
    if (product.sale_price != null && product.sale_price < product.price) {
      return product.sale_price
    }
    return product.price
  }

  const isOnSale = (product: Product) => {
    return product.sale_price != null && product.sale_price < product.price
  }

  return {
    fetchProducts,
    fetchProductBySlug,
    fetchCategories,
    createCategory,
    updateCategory,
    deleteCategory,
    fetchReviews,
    fetchReviewSummariesBatch,
    createReview,
    createProduct,
    updateProduct,
    deleteProduct,
    getEffectivePrice,
    isOnSale,
  }
}
