import { categoryAvatarBg, categoryBg } from '~/utils/theme'

const iconMap: Record<string, string> = {
  electronics: 'mdi-cellphone',
  fashion: 'mdi-tshirt-crew',
  food: 'mdi-food-apple',
  beauty: 'mdi-face-woman-shimmer',
  sports: 'mdi-basketball',
  home: 'mdi-sofa',
  books: 'mdi-book-open-page-variant',
  toys: 'mdi-toy-brick',
  automotive: 'mdi-car',
  garden: 'mdi-flower',
}

export function categoryIcon(slug: string): string {
  const root = slug.split('-')[0]
  return iconMap[root] || iconMap[slug] || 'mdi-tag-outline'
}

/** @deprecated dùng categoryBg() từ utils/theme */
export function categoryColor(index: number): string {
  return categoryBg(index)
}

export { categoryBg, categoryAvatarBg }
