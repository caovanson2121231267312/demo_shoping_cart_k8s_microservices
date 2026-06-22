<template>
  <Teleport to="body">
    <Transition name="admin-customize">
      <div v-if="modelValue" class="admin-customize-root" @keydown.esc="close">
        <div class="admin-customize-backdrop" aria-hidden="true" @click="close" />
        <aside class="admin-customize-sheet" role="dialog" aria-label="Tùy chỉnh giao diện" :style="layout.shellStyle">
          <div class="admin-customize-panel">
          <div class="admin-customize-panel__head">
            <div>
              <div class="admin-customize-panel__title">Tùy chỉnh giao diện</div>
              <div class="admin-customize-panel__sub">Theme, navbar, header & khoảng cách</div>
            </div>
            <v-btn icon="mdi-close" variant="text" size="small" @click="close" />
          </div>

          <v-tabs v-model="tab" density="compact" color="primary" class="admin-customize-tabs mb-4">
            <v-tab value="theme" class="text-none">Theme</v-tab>
            <v-tab value="nav" class="text-none">Navbar</v-tab>
            <v-tab value="header" class="text-none">Header</v-tab>
            <v-tab value="spacing" class="text-none">Bố cục</v-tab>
          </v-tabs>

          <v-window v-model="tab">
            <v-window-item value="theme">
            <section class="admin-customize__section">
              <div class="admin-customize__label">Màu chủ đạo ({{ layout.themes.length }} tuỳ chọn)</div>
              <p class="admin-customize-hint">Áp dụng ngay: sidebar, menu đang chọn, logo, header kiểu &quot;Màu&quot;</p>
              <div class="admin-theme-grid">
                  <button
                    v-for="t in layout.themes"
                    :key="t.id"
                    type="button"
                    class="admin-theme-chip"
                    :class="{ 'admin-theme-chip--active': layout.prefs.value.themeId === t.id }"
                    @click="layout.setTheme(t.id)"
                  >
                    <span class="admin-theme-chip__swatch" :style="{ background: t.swatch }" />
                    {{ t.label }}
                  </button>
                </div>
              </section>
            </v-window-item>

            <v-window-item value="nav">
              <section class="admin-customize__section">
                <div class="admin-customize__label">Vị trí menu</div>
                <div class="admin-option-cards">
                  <button
                    v-for="opt in navPositionOptions"
                    :key="opt.value"
                    type="button"
                    class="admin-option-card"
                    :class="{ 'admin-option-card--active': layout.prefs.value.navPosition === opt.value }"
                    @click="layout.setNavPosition(opt.value)"
                  >
                    <v-icon :icon="opt.icon" size="22" />
                    <span class="admin-option-card__title">{{ opt.label }}</span>
                    <span class="admin-option-card__desc">{{ opt.desc }}</span>
                  </button>
                </div>
              </section>

              <section v-if="layout.prefs.value.navPosition === 'sidebar'" class="admin-customize__section">
                <div class="admin-customize__label">Kiểu sidebar</div>
                <div class="admin-layout-options">
                  <v-btn
                    v-for="opt in sidebarOptions"
                    :key="opt.value"
                    size="small"
                    :variant="layout.prefs.value.sidebarStyle === opt.value ? 'flat' : 'outlined'"
                    :color="layout.prefs.value.sidebarStyle === opt.value ? 'primary' : undefined"
                    class="text-none"
                    @click="layout.setSidebarStyle(opt.value)"
                  >
                    {{ opt.label }}
                  </v-btn>
                </div>
                <div class="admin-customize__label mt-4">Độ rộng sidebar</div>
                <div class="admin-layout-options">
                  <v-btn
                    v-for="opt in sidebarWidthOptions"
                    :key="opt.value"
                    size="small"
                    :variant="layout.prefs.value.sidebarWidth === opt.value ? 'flat' : 'outlined'"
                    :color="layout.prefs.value.sidebarWidth === opt.value ? 'primary' : undefined"
                    class="text-none"
                    @click="layout.setSidebarWidth(opt.value)"
                  >
                    {{ opt.label }}
                  </v-btn>
                </div>
                <v-switch
                  :model-value="layout.prefs.value.sidebarCollapsed"
                  label="Thu gọn sidebar (rail)"
                  color="primary"
                  density="compact"
                  hide-details
                  class="mt-3"
                  @update:model-value="(v) => v !== layout.prefs.value.sidebarCollapsed && layout.toggleSidebarCollapsed()"
                />
              </section>

              <section class="admin-customize__section">
                <div class="admin-customize__label">Kiểu menu active</div>
                <div class="admin-layout-options">
                  <v-btn
                    v-for="opt in navStyleOptions"
                    :key="opt.value"
                    size="small"
                    :variant="layout.prefs.value.navStyle === opt.value ? 'flat' : 'outlined'"
                    :color="layout.prefs.value.navStyle === opt.value ? 'primary' : undefined"
                    class="text-none"
                    @click="layout.setNavStyle(opt.value)"
                  >
                    {{ opt.label }}
                  </v-btn>
                </div>
              </section>
            </v-window-item>

            <v-window-item value="header">
              <section class="admin-customize__section">
                <div class="admin-customize__label">Kiểu header</div>
                <p class="admin-customize-hint">Mỗi kiểu dùng màu chủ đạo đang chọn — đổi theme ở tab Theme để thấy rõ hơn</p>
                <div class="admin-header-previews">
                  <button
                    v-for="opt in headerOptions"
                    :key="opt.value"
                    type="button"
                    class="admin-header-preview"
                    :class="[
                      `admin-header-preview--${opt.value}`,
                      { 'admin-header-preview--active': layout.prefs.value.headerStyle === opt.value },
                    ]"
                    @click="layout.setHeaderStyle(opt.value)"
                  >
                    <span class="admin-header-preview__bar" />
                    <span class="admin-header-preview__label">{{ opt.label }}</span>
                    <span class="admin-header-preview__desc">{{ opt.desc }}</span>
                  </button>
                </div>
              </section>

              <section class="admin-customize__section">
                <div class="admin-customize__label">Chiều cao header</div>
                <div class="admin-layout-options">
                  <v-btn
                    v-for="opt in headerSizeOptions"
                    :key="opt.value"
                    size="small"
                    :variant="layout.prefs.value.headerSize === opt.value ? 'flat' : 'outlined'"
                    :color="layout.prefs.value.headerSize === opt.value ? 'primary' : undefined"
                    class="text-none"
                    @click="layout.setHeaderSize(opt.value)"
                  >
                    {{ opt.label }}
                  </v-btn>
                </div>
              </section>

              <section class="admin-customize__section">
                <v-switch
                  :model-value="layout.prefs.value.showBreadcrumb"
                  label="Hiện breadcrumb / mô tả"
                  color="primary"
                  density="compact"
                  hide-details
                  class="mb-3"
                  @update:model-value="layout.setShowBreadcrumb(!!$event)"
                />
                <v-switch
                  :model-value="layout.prefs.value.showHeaderBorder"
                  label="Viền dưới header"
                  color="primary"
                  density="compact"
                  hide-details
                  @update:model-value="layout.setShowHeaderBorder(!!$event)"
                />
              </section>
            </v-window-item>

            <v-window-item value="spacing">
              <section class="admin-customize__section">
                <div class="admin-customize__label">Cỡ chữ</div>
                <div class="admin-font-size-options">
                  <button
                    v-for="opt in fontSizeOptions"
                    :key="opt.value"
                    type="button"
                    class="admin-font-size-btn"
                    :class="{ 'admin-font-size-btn--active': layout.prefs.value.fontSize === opt.value }"
                    @click="layout.setFontSize(opt.value)"
                  >
                    <span class="admin-font-size-btn__sample" :style="{ fontSize: fontSizeSamples[opt.value] }">Aa</span>
                    <span class="admin-font-size-btn__label">{{ opt.label }}</span>
                    <span class="admin-font-size-btn__desc">{{ opt.desc }}</span>
                  </button>
                </div>
                <div class="admin-font-size-preview" :style="{ fontSize: fontSizeSamples[layout.prefs.value.fontSize] }">
                  <strong>Tiêu đề trang</strong>
                  <span>Mô tả và nội dung văn bản trong admin panel.</span>
                </div>
              </section>

              <section class="admin-customize__section">
                <div class="admin-customize__label">Khoảng cách nội dung</div>
                <div class="admin-option-cards admin-option-cards--row3">
                  <button
                    v-for="opt in spacingOptions"
                    :key="opt.value"
                    type="button"
                    class="admin-option-card admin-option-card--compact"
                    :class="{ 'admin-option-card--active': layout.prefs.value.contentSpacing === opt.value }"
                    @click="layout.setContentSpacing(opt.value)"
                  >
                    <span class="admin-option-card__title">{{ opt.label }}</span>
                    <span class="admin-option-card__desc">{{ opt.desc }}</span>
                  </button>
                </div>
              </section>

              <section class="admin-customize__section">
                <div class="admin-customize__label">Chiều rộng nội dung</div>
                <div class="admin-layout-options">
                  <v-btn
                    v-for="opt in widthOptions"
                    :key="opt.value"
                    size="small"
                    :variant="layout.prefs.value.contentWidth === opt.value ? 'flat' : 'outlined'"
                    :color="layout.prefs.value.contentWidth === opt.value ? 'primary' : undefined"
                    class="text-none"
                    @click="layout.setContentWidth(opt.value)"
                  >
                    {{ opt.label }}
                  </v-btn>
                </div>
              </section>

              <section class="admin-customize__section">
                <div class="admin-customize__label">Mật độ component</div>
                <div class="admin-layout-options">
                  <v-btn
                    v-for="opt in densityOptions"
                    :key="opt.value"
                    size="small"
                    :variant="layout.prefs.value.density === opt.value ? 'flat' : 'outlined'"
                    :color="layout.prefs.value.density === opt.value ? 'primary' : undefined"
                    class="text-none"
                    @click="layout.setDensity(opt.value)"
                  >
                    {{ opt.label }}
                  </v-btn>
                </div>
              </section>
            </v-window-item>
          </v-window>

          <v-btn block variant="outlined" color="grey" class="text-none mt-4" @click="layout.resetPrefs()">
            Đặt lại mặc định
          </v-btn>
          </div>
        </aside>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import type {
  AdminContentWidth,
  AdminDensity,
  AdminFontSize,
  AdminHeaderSize,
  AdminHeaderStyle,
  AdminNavPosition,
  AdminNavStyle,
  AdminSidebarStyle,
  AdminSidebarWidth,
  AdminSpacingScale,
} from '~/utils/adminThemes'
import { fontSizeTokens } from '~/utils/adminThemes'

defineProps<{ modelValue: boolean }>()
const emit = defineEmits<{ 'update:modelValue': [value: boolean] }>()

const close = () => emit('update:modelValue', false)

const layout = useAdminLayout()
const tab = ref('theme')

const navPositionOptions: { value: AdminNavPosition; label: string; desc: string; icon: string }[] = [
  { value: 'sidebar', label: 'Sidebar trái', desc: 'Menu dọc cổ điển', icon: 'mdi-dock-left' },
  { value: 'top', label: 'Navbar trên', desc: 'Menu ngang trên header', icon: 'mdi-dock-top' },
]

const sidebarOptions: { value: AdminSidebarStyle; label: string }[] = [
  { value: 'gradient', label: 'Gradient' },
  { value: 'dark', label: 'Tối' },
  { value: 'light', label: 'Sáng' },
]

const sidebarWidthOptions: { value: AdminSidebarWidth; label: string }[] = [
  { value: 'narrow', label: 'Hẹp' },
  { value: 'default', label: 'Vừa' },
  { value: 'wide', label: 'Rộng' },
]

const navStyleOptions: { value: AdminNavStyle; label: string }[] = [
  { value: 'pill', label: 'Pill' },
  { value: 'line', label: 'Gạch' },
  { value: 'soft', label: 'Soft' },
]

const headerOptions: { value: AdminHeaderStyle; label: string; desc: string }[] = [
  { value: 'glass', label: 'Glass', desc: 'Nền pha màu + blur' },
  { value: 'solid', label: 'Solid', desc: 'Nền tint + viền màu' },
  { value: 'minimal', label: 'Minimal', desc: 'Tiêu đề màu theme' },
  { value: 'colored', label: 'Màu', desc: 'Header full gradient' },
  { value: 'bordered', label: 'Viền', desc: 'Viền dưới đậm màu' },
]

const headerSizeOptions: { value: AdminHeaderSize; label: string }[] = [
  { value: 'sm', label: 'Thấp' },
  { value: 'md', label: 'Vừa' },
  { value: 'lg', label: 'Cao' },
]

const spacingOptions: { value: AdminSpacingScale; label: string; desc: string }[] = [
  { value: 'compact', label: 'Gọn', desc: '16px pad' },
  { value: 'balanced', label: 'Cân bằng', desc: '24px pad' },
  { value: 'relaxed', label: 'Thoáng', desc: '32px pad' },
]

const widthOptions: { value: AdminContentWidth; label: string }[] = [
  { value: 'boxed', label: 'Boxed' },
  { value: 'full', label: 'Full width' },
]

const densityOptions: { value: AdminDensity; label: string }[] = [
  { value: 'comfortable', label: 'Thoáng' },
  { value: 'compact', label: 'Gọn' },
]

const fontSizeOptions: { value: AdminFontSize; label: string; desc: string }[] = [
  { value: 'xs', label: 'Rất nhỏ', desc: '12px' },
  { value: 'sm', label: 'Nhỏ', desc: '13px' },
  { value: 'md', label: 'Vừa', desc: '14px' },
  { value: 'lg', label: 'Lớn', desc: '15px' },
  { value: 'xl', label: 'Rất lớn', desc: '16px' },
]

const fontSizeSamples = Object.fromEntries(
  (Object.keys(fontSizeTokens) as AdminFontSize[]).map((key) => [key, fontSizeTokens[key].base]),
) as Record<AdminFontSize, string>
</script>

<style scoped>
.admin-customize-root {
  position: fixed;
  inset: 0;
  z-index: 2500;
  display: flex;
  justify-content: flex-end;
}

.admin-customize-backdrop {
  position: absolute;
  inset: 0;
  background: rgba(15, 23, 42, 0.42);
  backdrop-filter: blur(2px);
}

.admin-customize-sheet {
  position: relative;
  z-index: 1;
  width: min(400px, 100vw);
  height: 100%;
  background: #fff;
  box-shadow: -12px 0 40px rgba(15, 23, 42, 0.14);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  font-size: var(--admin-font-base, 14px);
}

.admin-customize-panel {
  padding: 20px;
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
}

.admin-customize-panel__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 8px;
}

.admin-customize-panel__title {
  font-size: 18px;
  font-weight: 800;
}

.admin-customize-panel__sub {
  font-size: 12px;
  color: #64748b;
  margin-top: 2px;
}

.admin-customize-tabs :deep(.v-tab) {
  min-width: 0;
  font-size: 13px;
}

.admin-option-cards {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

.admin-option-cards--row3 {
  grid-template-columns: repeat(3, 1fr);
}

.admin-option-card {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 4px;
  padding: 14px;
  border-radius: var(--admin-radius-sm);
  border: 2px solid rgba(15, 23, 42, 0.08);
  background: #fff;
  cursor: pointer;
  text-align: left;
  transition: border-color 0.18s ease, box-shadow 0.18s ease;
}

.admin-option-card--compact {
  padding: 12px 10px;
  align-items: center;
  text-align: center;
}

.admin-option-card:hover {
  border-color: rgba(var(--v-theme-primary), 0.35);
}

.admin-option-card--active {
  border-color: rgb(var(--v-theme-primary));
  box-shadow: 0 0 0 3px rgba(var(--v-theme-primary), 0.12);
}

.admin-option-card__title {
  font-size: 13px;
  font-weight: 700;
}

.admin-option-card__desc {
  font-size: 11px;
  color: #64748b;
  line-height: 1.3;
}

.admin-customize-hint {
  font-size: 12px;
  color: #64748b;
  margin: 0 0 12px;
  line-height: 1.4;
}

.admin-header-previews {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

.admin-header-preview {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 12px;
  border-radius: var(--admin-radius-sm);
  border: 2px solid rgba(15, 23, 42, 0.08);
  background: #fff;
  cursor: pointer;
  text-align: left;
  transition: border-color 0.18s ease, box-shadow 0.18s ease;
}

.admin-header-preview--active {
  border-color: rgb(var(--v-theme-primary));
  box-shadow: 0 0 0 3px rgba(var(--v-theme-primary), 0.12);
}

.admin-header-preview__bar {
  display: block;
  height: 28px;
  border-radius: var(--admin-radius-xs);
  border: 1px solid rgba(15, 23, 42, 0.06);
}

.admin-header-preview--glass .admin-header-preview__bar {
  background: var(--admin-header-glass-bg);
  box-shadow: var(--admin-header-shadow);
}

.admin-header-preview--solid .admin-header-preview__bar {
  background: var(--admin-header-tint);
  border-bottom: 2px solid var(--admin-header-border);
}

.admin-header-preview--minimal .admin-header-preview__bar {
  background: transparent;
  border-bottom: 1px solid color-mix(in srgb, var(--admin-primary) 30%, transparent);
}

.admin-header-preview--colored .admin-header-preview__bar {
  background: var(--admin-header-colored-bg);
  border: none;
}

.admin-header-preview--bordered .admin-header-preview__bar {
  background: #fff;
  border-bottom: 3px solid var(--admin-primary);
}

.admin-header-preview__label {
  font-size: 13px;
  font-weight: 700;
}

.admin-header-preview__desc {
  font-size: 11px;
  color: #64748b;
  line-height: 1.3;
}

.admin-font-size-options {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 8px;
  margin-bottom: 12px;
}

.admin-font-size-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 10px 6px;
  border-radius: var(--admin-radius-sm);
  border: 2px solid rgba(15, 23, 42, 0.08);
  background: #fff;
  cursor: pointer;
  transition: border-color 0.18s ease, box-shadow 0.18s ease;
}

.admin-font-size-btn:hover {
  border-color: rgba(var(--v-theme-primary), 0.35);
}

.admin-font-size-btn--active {
  border-color: rgb(var(--v-theme-primary));
  box-shadow: 0 0 0 3px rgba(var(--v-theme-primary), 0.12);
}

.admin-font-size-btn__sample {
  font-weight: 800;
  line-height: 1;
  color: rgb(var(--v-theme-primary));
}

.admin-font-size-btn__label {
  font-size: 11px;
  font-weight: 700;
  line-height: 1.2;
}

.admin-font-size-btn__desc {
  font-size: 10px;
  color: #64748b;
}

.admin-font-size-preview {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 14px 16px;
  border-radius: var(--admin-radius-sm);
  border: 1px dashed rgba(15, 23, 42, 0.12);
  background: rgba(15, 23, 42, 0.02);
  line-height: 1.45;
}

.admin-font-size-preview strong {
  font-size: var(--admin-font-page-title, 1.5rem);
  font-weight: 800;
  letter-spacing: -0.02em;
}

.admin-font-size-preview span {
  color: #64748b;
}
</style>

<style>
.admin-customize-enter-active .admin-customize-sheet,
.admin-customize-leave-active .admin-customize-sheet {
  transition: transform 0.28s cubic-bezier(0.4, 0, 0.2, 1);
}

.admin-customize-enter-active .admin-customize-backdrop,
.admin-customize-leave-active .admin-customize-backdrop {
  transition: opacity 0.28s ease;
}

.admin-customize-enter-from .admin-customize-sheet,
.admin-customize-leave-to .admin-customize-sheet {
  transform: translateX(100%);
}

.admin-customize-enter-from .admin-customize-backdrop,
.admin-customize-leave-to .admin-customize-backdrop {
  opacity: 0;
}
</style>
