export type AdminThemeId =
  | 'obsidian'
  | 'ocean'
  | 'violet'
  | 'emerald'
  | 'rose'
  | 'slate'
  | 'cyan'
  | 'amber'
  | 'indigo'
  | 'fuchsia'
  | 'orange'
  | 'teal'
  | 'crimson'
  | 'navy'
  | 'coffee'
  | 'mint'
  | 'sky'
  | 'lime'
export type AdminSidebarStyle = 'dark' | 'light' | 'gradient'
export type AdminHeaderStyle = 'glass' | 'solid' | 'minimal' | 'colored' | 'bordered'
export type AdminContentWidth = 'full' | 'boxed'
export type AdminDensity = 'comfortable' | 'compact'
export type AdminNavPosition = 'sidebar' | 'top'
export type AdminNavStyle = 'pill' | 'line' | 'soft'
export type AdminSpacingScale = 'compact' | 'balanced' | 'relaxed'
export type AdminSidebarWidth = 'narrow' | 'default' | 'wide'
export type AdminHeaderSize = 'sm' | 'md' | 'lg'
export type AdminFontSize = 'xs' | 'sm' | 'md' | 'lg' | 'xl'

export interface AdminThemePreset {
  id: AdminThemeId
  label: string
  swatch: string
  primary: string
  primaryDark: string
  primaryLight: string
  accent: string
}

export interface AdminLayoutPrefs {
  themeId: AdminThemeId
  sidebarStyle: AdminSidebarStyle
  headerStyle: AdminHeaderStyle
  contentWidth: AdminContentWidth
  density: AdminDensity
  sidebarCollapsed: boolean
  navPosition: AdminNavPosition
  navStyle: AdminNavStyle
  contentSpacing: AdminSpacingScale
  sidebarWidth: AdminSidebarWidth
  headerSize: AdminHeaderSize
  fontSize: AdminFontSize
  showBreadcrumb: boolean
  showHeaderBorder: boolean
}

export const ADMIN_LAYOUT_STORAGE_KEY = 'shop_admin_layout_prefs'

export const defaultAdminLayoutPrefs: AdminLayoutPrefs = {
  themeId: 'obsidian',
  sidebarStyle: 'dark',
  headerStyle: 'solid',
  contentWidth: 'boxed',
  density: 'comfortable',
  sidebarCollapsed: false,
  navPosition: 'sidebar',
  navStyle: 'line',
  contentSpacing: 'balanced',
  sidebarWidth: 'default',
  headerSize: 'md',
  fontSize: 'md',
  showBreadcrumb: true,
  showHeaderBorder: true,
}

export const adminThemePresets: AdminThemePreset[] = [
  {
    id: 'obsidian',
    label: 'Obsidian',
    swatch: '#0a0a0b',
    primary: '#3b82f6',
    primaryDark: '#2563eb',
    primaryLight: '#60a5fa',
    accent: '#71717a',
  },
  {
    id: 'ocean',
    label: 'Ocean Blue',
    swatch: '#1565c0',
    primary: '#1565c0',
    primaryDark: '#0d47a1',
    primaryLight: '#42a5f5',
    accent: '#0288d1',
  },
  {
    id: 'sky',
    label: 'Sky Light',
    swatch: '#0ea5e9',
    primary: '#0ea5e9',
    primaryDark: '#0369a1',
    primaryLight: '#38bdf8',
    accent: '#06b6d4',
  },
  {
    id: 'cyan',
    label: 'Cyan Fresh',
    swatch: '#0891b2',
    primary: '#0891b2',
    primaryDark: '#0e7490',
    primaryLight: '#22d3ee',
    accent: '#06b6d4',
  },
  {
    id: 'teal',
    label: 'Teal Wave',
    swatch: '#0d9488',
    primary: '#0d9488',
    primaryDark: '#0f766e',
    primaryLight: '#2dd4bf',
    accent: '#14b8a6',
  },
  {
    id: 'mint',
    label: 'Mint Cool',
    swatch: '#14b8a6',
    primary: '#14b8a6',
    primaryDark: '#0d9488',
    primaryLight: '#5eead4',
    accent: '#2dd4bf',
  },
  {
    id: 'emerald',
    label: 'Emerald Pro',
    swatch: '#059669',
    primary: '#059669',
    primaryDark: '#047857',
    primaryLight: '#34d399',
    accent: '#10b981',
  },
  {
    id: 'lime',
    label: 'Lime Energy',
    swatch: '#65a30d',
    primary: '#65a30d',
    primaryDark: '#4d7c0f',
    primaryLight: '#a3e635',
    accent: '#84cc16',
  },
  {
    id: 'amber',
    label: 'Amber Gold',
    swatch: '#d97706',
    primary: '#d97706',
    primaryDark: '#b45309',
    primaryLight: '#fbbf24',
    accent: '#f59e0b',
  },
  {
    id: 'orange',
    label: 'Sunset Orange',
    swatch: '#ea580c',
    primary: '#ea580c',
    primaryDark: '#c2410c',
    primaryLight: '#fb923c',
    accent: '#f97316',
  },
  {
    id: 'rose',
    label: 'Rose Gold',
    swatch: '#e11d48',
    primary: '#e11d48',
    primaryDark: '#be123c',
    primaryLight: '#fb7185',
    accent: '#f43f5e',
  },
  {
    id: 'crimson',
    label: 'Crimson Bold',
    swatch: '#dc2626',
    primary: '#dc2626',
    primaryDark: '#b91c1c',
    primaryLight: '#f87171',
    accent: '#ef4444',
  },
  {
    id: 'fuchsia',
    label: 'Fuchsia Pop',
    swatch: '#c026d3',
    primary: '#c026d3',
    primaryDark: '#a21caf',
    primaryLight: '#e879f9',
    accent: '#d946ef',
  },
  {
    id: 'violet',
    label: 'Violet Night',
    swatch: '#7c3aed',
    primary: '#7c3aed',
    primaryDark: '#5b21b6',
    primaryLight: '#a78bfa',
    accent: '#c084fc',
  },
  {
    id: 'indigo',
    label: 'Indigo Deep',
    swatch: '#4f46e5',
    primary: '#4f46e5',
    primaryDark: '#4338ca',
    primaryLight: '#818cf8',
    accent: '#6366f1',
  },
  {
    id: 'navy',
    label: 'Navy Classic',
    swatch: '#1e3a8a',
    primary: '#1e3a8a',
    primaryDark: '#1e40af',
    primaryLight: '#3b82f6',
    accent: '#2563eb',
  },
  {
    id: 'slate',
    label: 'Slate Minimal',
    swatch: '#475569',
    primary: '#475569',
    primaryDark: '#334155',
    primaryLight: '#64748b',
    accent: '#94a3b8',
  },
  {
    id: 'coffee',
    label: 'Coffee Warm',
    swatch: '#78350f',
    primary: '#78350f',
    primaryDark: '#572005',
    primaryLight: '#b45309',
    accent: '#92400e',
  },
]

export const spacingTokens: Record<AdminSpacingScale, {
  contentPad: string
  sectionGap: string
  cardPad: string
  gridGap: string
  headerPadX: string
}> = {
  compact: {
    contentPad: '16px',
    sectionGap: '16px',
    cardPad: '16px',
    gridGap: '12px',
    headerPadX: '16px',
  },
  balanced: {
    contentPad: '24px',
    sectionGap: '24px',
    cardPad: '20px',
    gridGap: '16px',
    headerPadX: '24px',
  },
  relaxed: {
    contentPad: '32px',
    sectionGap: '32px',
    cardPad: '24px',
    gridGap: '20px',
    headerPadX: '32px',
  },
}

export const sidebarWidthTokens: Record<AdminSidebarWidth, { width: string; rail: string }> = {
  narrow: { width: '240px', rail: '68px' },
  default: { width: '272px', rail: '76px' },
  wide: { width: '300px', rail: '84px' },
}

export const headerSizeTokens: Record<AdminHeaderSize, string> = {
  sm: '56px',
  md: '64px',
  lg: '72px',
}

export const fontSizeTokens: Record<AdminFontSize, {
  base: string
  sm: string
  xs: string
  lg: string
  xl: string
  header: string
  pageTitle: string
  stat: string
  nav: string
  caption: string
  label: string
  sample: string
}> = {
  xs: {
    base: '12px',
    sm: '11px',
    xs: '10px',
    lg: '13px',
    xl: '15px',
    header: '15px',
    pageTitle: '1.25rem',
    stat: '22px',
    nav: '12px',
    caption: '10px',
    label: '9px',
    sample: 'Aa 12px',
  },
  sm: {
    base: '13px',
    sm: '12px',
    xs: '11px',
    lg: '14px',
    xl: '16px',
    header: '16px',
    pageTitle: '1.35rem',
    stat: '24px',
    nav: '13px',
    caption: '11px',
    label: '10px',
    sample: 'Aa 13px',
  },
  md: {
    base: '14px',
    sm: '13px',
    xs: '12px',
    lg: '15px',
    xl: '18px',
    header: '18px',
    pageTitle: '1.5rem',
    stat: '28px',
    nav: '14px',
    caption: '12px',
    label: '10px',
    sample: 'Aa 14px',
  },
  lg: {
    base: '15px',
    sm: '14px',
    xs: '13px',
    lg: '16px',
    xl: '20px',
    header: '20px',
    pageTitle: '1.65rem',
    stat: '32px',
    nav: '15px',
    caption: '13px',
    label: '11px',
    sample: 'Aa 15px',
  },
  xl: {
    base: '16px',
    sm: '15px',
    xs: '14px',
    lg: '18px',
    xl: '22px',
    header: '22px',
    pageTitle: '1.85rem',
    stat: '36px',
    nav: '16px',
    caption: '14px',
    label: '12px',
    sample: 'Aa 16px',
  },
}

export function getAdminTheme(id: AdminThemeId | string): AdminThemePreset {
  return adminThemePresets.find((t) => t.id === id) ?? adminThemePresets[0]
}

/** CSS variables áp dụng lên .admin-shell — theme màu cho nội dung; navbar giữ palette đen/trắng cố định */
export function buildAdminShellCssVars(
  theme: AdminThemePreset,
  prefs: Pick<AdminLayoutPrefs, 'contentSpacing' | 'sidebarWidth' | 'headerSize' | 'fontSize'>,
): Record<string, string> {
  const sp = spacingTokens[prefs.contentSpacing]
  const sw = sidebarWidthTokens[prefs.sidebarWidth]
  const fs = fontSizeTokens[prefs.fontSize]
  const { primary, primaryDark, primaryLight, accent } = theme

  return {
    '--admin-font-base': fs.base,
    '--admin-font-sm': fs.sm,
    '--admin-font-xs': fs.xs,
    '--admin-font-lg': fs.lg,
    '--admin-font-xl': fs.xl,
    '--admin-font-header': fs.header,
    '--admin-font-page-title': fs.pageTitle,
    '--admin-font-stat': fs.stat,
    '--admin-font-nav': fs.nav,
    '--admin-font-caption': fs.caption,
    '--admin-font-label': fs.label,
    '--admin-primary': primary,
    '--admin-primary-dark': primaryDark,
    '--admin-primary-light': primaryLight,
    '--admin-accent': accent,
    '--admin-glow': `${primary}40`,
    '--admin-nav-bg': '#0a0a0b',
    '--admin-nav-bg-elevated': '#111113',
    '--admin-nav-border': 'rgba(255, 255, 255, 0.07)',
    '--admin-nav-text': 'rgba(255, 255, 255, 0.55)',
    '--admin-nav-text-hover': 'rgba(255, 255, 255, 0.92)',
    '--admin-nav-active-text': '#ffffff',
    '--admin-nav-active-bg': 'rgba(255, 255, 255, 0.08)',
    '--admin-nav-active-indicator': '#ffffff',
    '--admin-sidebar-gradient': 'linear-gradient(180deg, #111113 0%, #0a0a0b 45%, #050506 100%)',
    '--admin-sidebar-solid': '#0a0a0b',
    '--admin-sidebar-accent': 'rgba(255, 255, 255, 0.2)',
    '--admin-sidebar-accent-themed': primary,
    '--admin-sidebar-brand-glow': 'transparent',
    '--admin-sidebar-footer-accent': 'linear-gradient(90deg, rgba(255,255,255,0.15), rgba(255,255,255,0.05))',
    '--admin-nav-active-bg-pill': 'rgba(255, 255, 255, 0.1)',
    '--admin-header-colored-bg': `linear-gradient(135deg, ${primary} 0%, ${primaryDark} 100%)`,
    '--admin-header-glass-bg': 'rgba(255, 255, 255, 0.94)',
    '--admin-header-tint': '#ffffff',
    '--admin-header-border': 'rgba(15, 23, 42, 0.08)',
    '--admin-header-shadow': '0 1px 2px rgba(15, 23, 42, 0.04)',
    '--admin-sidebar-w': sw.width,
    '--admin-sidebar-w-rail': sw.rail,
    '--admin-header-h': headerSizeTokens[prefs.headerSize],
    '--admin-content-pad': sp.contentPad,
    '--admin-section-gap': sp.sectionGap,
    '--admin-card-pad': sp.cardPad,
    '--admin-grid-gap': sp.gridGap,
    '--admin-header-px': sp.headerPadX,
  }
}
