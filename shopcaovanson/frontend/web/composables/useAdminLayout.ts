import { useTheme } from 'vuetify'
import {
  ADMIN_LAYOUT_STORAGE_KEY,
  buildAdminShellCssVars,
  defaultAdminLayoutPrefs,
  getAdminTheme,
  type AdminContentWidth,
  type AdminDensity,
  type AdminFontSize,
  type AdminHeaderSize,
  type AdminHeaderStyle,
  type AdminLayoutPrefs,
  type AdminNavPosition,
  type AdminNavStyle,
  type AdminSidebarStyle,
  type AdminSidebarWidth,
  type AdminSpacingScale,
  type AdminThemeId,
  adminThemePresets,
} from '~/utils/adminThemes'

function loadPrefs(): AdminLayoutPrefs {
  if (!import.meta.client) {
    return { ...defaultAdminLayoutPrefs }
  }
  try {
    const raw = localStorage.getItem(ADMIN_LAYOUT_STORAGE_KEY)
    if (!raw) {
      return { ...defaultAdminLayoutPrefs }
    }
    const merged = { ...defaultAdminLayoutPrefs, ...JSON.parse(raw) } as AdminLayoutPrefs
    if (!adminThemePresets.some((t) => t.id === merged.themeId)) {
      merged.themeId = defaultAdminLayoutPrefs.themeId
    }
    const validHeader: AdminHeaderStyle[] = ['glass', 'solid', 'minimal', 'colored', 'bordered']
    if (!validHeader.includes(merged.headerStyle)) {
      merged.headerStyle = defaultAdminLayoutPrefs.headerStyle
    }
    const validFontSize: AdminFontSize[] = ['xs', 'sm', 'md', 'lg', 'xl']
    if (!validFontSize.includes(merged.fontSize)) {
      merged.fontSize = defaultAdminLayoutPrefs.fontSize
    }
    return merged
  } catch {
    return { ...defaultAdminLayoutPrefs }
  }
}

function savePrefs(prefs: AdminLayoutPrefs) {
  if (!import.meta.client) {
    return
  }
  localStorage.setItem(ADMIN_LAYOUT_STORAGE_KEY, JSON.stringify(prefs))
}

export const useAdminLayout = () => {
  const prefs = useState<AdminLayoutPrefs>('admin-layout-prefs', () => loadPrefs())
  const customizeOpen = useState('admin-customize-open', () => false)
  const mobileDrawer = useState('admin-mobile-drawer', () => false)
  const vuetifyTheme = useTheme()

  const theme = computed(() => getAdminTheme(prefs.value.themeId))

  const shellStyle = computed(() => buildAdminShellCssVars(theme.value, prefs.value))

  const sidebarChromeStyle = computed(() => {
    const style = prefs.value.sidebarStyle
    if (style === 'gradient') {
      return {
        background: 'var(--admin-sidebar-gradient)',
        borderRight: '1px solid color-mix(in srgb, var(--admin-sidebar-accent) 35%, transparent)',
      }
    }
    if (style === 'dark') {
      return {
        background: 'var(--admin-nav-bg, #0a0a0b)',
        borderRight: '1px solid var(--admin-nav-border, rgba(255, 255, 255, 0.07))',
      }
    }
    return {
      background: 'var(--admin-surface)',
      borderRight: '1px solid var(--admin-border)',
      borderLeft: '4px solid var(--admin-sidebar-accent-themed, var(--admin-primary))',
      boxShadow: 'var(--admin-shadow)',
    }
  })

  const shellClasses = computed(() => [
    `admin-shell--theme-${prefs.value.themeId}`,
    `admin-shell--sidebar-${prefs.value.sidebarStyle}`,
    `admin-shell--header-${prefs.value.headerStyle}`,
    `admin-shell--width-${prefs.value.contentWidth}`,
    `admin-shell--density-${prefs.value.density}`,
    `admin-shell--nav-${prefs.value.navPosition}`,
    `admin-shell--navstyle-${prefs.value.navStyle}`,
    `admin-shell--spacing-${prefs.value.contentSpacing}`,
    `admin-shell--sidebar-w-${prefs.value.sidebarWidth}`,
    `admin-shell--header-size-${prefs.value.headerSize}`,
    `admin-shell--font-${prefs.value.fontSize}`,
    {
      'admin-shell--rail': prefs.value.sidebarCollapsed && prefs.value.navPosition === 'sidebar',
      'admin-shell--no-header-border': !prefs.value.showHeaderBorder,
      'admin-shell--no-breadcrumb': !prefs.value.showBreadcrumb,
    },
  ])

  const applyThemeVars = (el?: HTMLElement | null) => {
    if (!import.meta.client) {
      return
    }
    const target = el ?? document.querySelector('.admin-shell')
    if (!(target instanceof HTMLElement)) {
      return
    }
    const vars = shellStyle.value
    for (const [key, value] of Object.entries(vars)) {
      target.style.setProperty(key, value)
    }
  }

  const syncVuetifyTheme = () => {
    const t = theme.value
    const colors = vuetifyTheme.themes.value.light.colors
    colors.primary = t.primary
    colors.info = t.accent
    colors.secondary = t.primaryLight
  }

  const persist = () => {
    savePrefs(prefs.value)
    applyThemeVars()
    syncVuetifyTheme()
  }

  const init = () => {
    prefs.value = loadPrefs()
    applyThemeVars()
    syncVuetifyTheme()
  }

  const setTheme = (themeId: AdminThemeId) => {
    prefs.value.themeId = themeId
    persist()
  }

  const setSidebarStyle = (style: AdminSidebarStyle) => {
    prefs.value.sidebarStyle = style
    persist()
  }

  const setHeaderStyle = (style: AdminHeaderStyle) => {
    prefs.value.headerStyle = style
    persist()
  }

  const setContentWidth = (width: AdminContentWidth) => {
    prefs.value.contentWidth = width
    persist()
  }

  const setDensity = (density: AdminDensity) => {
    prefs.value.density = density
    persist()
  }

  const setNavPosition = (position: AdminNavPosition) => {
    prefs.value.navPosition = position
    if (position === 'top') {
      prefs.value.sidebarCollapsed = false
    }
    persist()
  }

  const setNavStyle = (style: AdminNavStyle) => {
    prefs.value.navStyle = style
    persist()
  }

  const setContentSpacing = (spacing: AdminSpacingScale) => {
    prefs.value.contentSpacing = spacing
    persist()
  }

  const setSidebarWidth = (width: AdminSidebarWidth) => {
    prefs.value.sidebarWidth = width
    persist()
  }

  const setHeaderSize = (size: AdminHeaderSize) => {
    prefs.value.headerSize = size
    persist()
  }

  const setFontSize = (size: AdminFontSize) => {
    prefs.value.fontSize = size
    persist()
  }

  const setShowBreadcrumb = (show: boolean) => {
    prefs.value.showBreadcrumb = show
    persist()
  }

  const setShowHeaderBorder = (show: boolean) => {
    prefs.value.showHeaderBorder = show
    persist()
  }

  const toggleSidebarCollapsed = () => {
    prefs.value.sidebarCollapsed = !prefs.value.sidebarCollapsed
    persist()
  }

  const resetPrefs = () => {
    prefs.value = { ...defaultAdminLayoutPrefs }
    persist()
  }

  return {
    prefs,
    theme,
    themes: adminThemePresets,
    customizeOpen,
    mobileDrawer,
    shellClasses,
    shellStyle,
    sidebarChromeStyle,
    init,
    applyThemeVars,
    setTheme,
    setSidebarStyle,
    setHeaderStyle,
    setContentWidth,
    setDensity,
    setNavPosition,
    setNavStyle,
    setContentSpacing,
    setSidebarWidth,
    setHeaderSize,
    setFontSize,
    setShowBreadcrumb,
    setShowHeaderBorder,
    toggleSidebarCollapsed,
    resetPrefs,
    persist,
  }
}
