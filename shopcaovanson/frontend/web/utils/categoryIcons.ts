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

export const categoryColors = [
  '#E3F2FD', '#FCE4EC', '#E8F5E9', '#FFF3E0', '#F3E5F5',
  '#E0F7FA', '#FFFDE7', '#EFEBE9', '#ECEFF1', '#F1F8E9',
]

export function categoryColor(index: number): string {
  return categoryColors[index % categoryColors.length]
}
