import type { Article, ArticleListResult } from '~/types'

export const useArticles = () => {
  const { apiFetch } = useAuth()

  const fetchArticles = (query: Record<string, string | number | boolean> = {}) =>
    apiFetch<ArticleListResult>('/api/articles', { query })

  const fetchArticleBySlug = (slug: string) =>
    apiFetch<Article>(`/api/articles/${encodeURIComponent(slug)}`)

  const adminFetchArticles = (query: Record<string, string | number> = {}) =>
    apiFetch<ArticleListResult>('/api/admin/articles', { query })

  const createArticle = (body: Partial<Article>) =>
    apiFetch<Article>('/api/admin/articles', { method: 'POST', body })

  const updateArticle = (id: string, body: Partial<Article>) =>
    apiFetch<Article>(`/api/admin/articles/${id}`, { method: 'PUT', body })

  const deleteArticle = (id: string) =>
    apiFetch(`/api/admin/articles/${id}`, { method: 'DELETE' })

  return {
    fetchArticles,
    fetchArticleBySlug,
    adminFetchArticles,
    createArticle,
    updateArticle,
    deleteArticle,
  }
}
