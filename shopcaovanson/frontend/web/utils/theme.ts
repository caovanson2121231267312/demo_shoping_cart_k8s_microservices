/**
 * Theme constants — dùng khi cần giá trị màu trong JS/TS (inline style, chart…)
 * Luôn đồng bộ với assets/css/tokens.css
 */
export const theme = {
  primary: '#1565c0',
  primaryDark: '#0d47a1',
  primaryLight: '#1976d2',
  primary50: '#e3f2fd',
  primary100: '#bbdefb',

  sale: '#d32f2f',
  saleDark: '#c62828',
  saleBg: '#fff5f5',

  secondary: '#f57c00',
  secondaryDark: '#e65100',

  success: '#388e3c',
  warning: '#f57c00',
  info: '#0288d1',

  text: '#212121',
  textSecondary: '#616161',
  textMuted: '#9e9e9e',

  bg: '#f5f7fa',
  surface: '#ffffff',
  surfaceMuted: '#f8fafc',

  footerBg: '#1a2332',
  footerBgDark: '#141b26',

  gradientPrimary: 'linear-gradient(135deg, #1565c0 0%, #0d47a1 100%)',
  gradientPromo: 'linear-gradient(90deg, #0d47a1 0%, #1565c0 50%, #1976d2 100%)',
  gradientSecondary: 'linear-gradient(135deg, #f57c00 0%, #e65100 100%)',
} as const

/** Màu nền danh mục — palette xanh thống nhất brand */
export const categoryBgColors = [
  'var(--color-category-1)',
  'var(--color-category-2)',
  'var(--color-category-3)',
  'var(--color-category-4)',
  'var(--color-category-5)',
  'var(--color-category-6)',
] as const

export const categoryAvatarColors = [
  'var(--color-primary-100)',
  'var(--color-primary-200)',
  'var(--color-primary-50)',
  'var(--color-primary-100)',
  'var(--color-primary-200)',
  'var(--color-primary-50)',
] as const

export function categoryBg(index: number): string {
  return categoryBgColors[index % categoryBgColors.length]
}

export function categoryAvatarBg(index: number): string {
  return categoryAvatarColors[index % categoryAvatarColors.length]
}
