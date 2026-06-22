<template>
  <Teleport to="body">
    <Transition name="admin-toast">
      <div
        v-if="visible"
        class="admin-toast"
        :class="`admin-toast--${tone}`"
        role="status"
        aria-live="polite"
      >
        <div class="admin-toast__icon-wrap">
          <v-icon :icon="icon" size="22" />
        </div>
        <div class="admin-toast__content">
          <div class="admin-toast__title">{{ title }}</div>
          <div class="admin-toast__message">{{ message }}</div>
        </div>
        <button type="button" class="admin-toast__close" aria-label="Đóng" @click="hide">
          <v-icon icon="mdi-close" size="18" />
        </button>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
const { message, title, visible, tone, icon, hide } = useSnackbar()
</script>

<style scoped>
.admin-toast {
  position: fixed;
  top: 20px;
  right: 20px;
  z-index: 9999;
  display: flex;
  align-items: flex-start;
  gap: 12px;
  min-width: 300px;
  max-width: min(420px, calc(100vw - 32px));
  padding: 14px 14px 14px 16px;
  border-radius: 10px;
  background: #ffffff;
  border: 1px solid rgba(15, 23, 42, 0.08);
  box-shadow:
    0 4px 6px rgba(15, 23, 42, 0.04),
    0 16px 40px rgba(15, 23, 42, 0.12);
}

.admin-toast__icon-wrap {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.admin-toast--success .admin-toast__icon-wrap {
  background: #ecfdf5;
  color: #059669;
}

.admin-toast--error .admin-toast__icon-wrap {
  background: #fef2f2;
  color: #dc2626;
}

.admin-toast--warning .admin-toast__icon-wrap {
  background: #fffbeb;
  color: #d97706;
}

.admin-toast--info .admin-toast__icon-wrap,
.admin-toast--primary .admin-toast__icon-wrap {
  background: #eff6ff;
  color: #2563eb;
}

.admin-toast__content {
  flex: 1;
  min-width: 0;
  padding-top: 1px;
}

.admin-toast__title {
  font-size: 14px;
  font-weight: 700;
  color: #0f172a;
  line-height: 1.3;
  margin-bottom: 2px;
}

.admin-toast__message {
  font-size: 13px;
  color: #64748b;
  line-height: 1.45;
}

.admin-toast__close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: #94a3b8;
  cursor: pointer;
  flex-shrink: 0;
  transition: background 0.15s ease, color 0.15s ease;
}

.admin-toast__close:hover {
  background: #f1f5f9;
  color: #475569;
}

.admin-toast-enter-active,
.admin-toast-leave-active {
  transition:
    opacity 0.22s ease,
    transform 0.22s cubic-bezier(0.4, 0, 0.2, 1);
}

.admin-toast-enter-from,
.admin-toast-leave-to {
  opacity: 0;
  transform: translateX(24px) translateY(-8px);
}

@media (max-width: 600px) {
  .admin-toast {
    top: 12px;
    right: 12px;
    left: 12px;
    min-width: 0;
    max-width: none;
  }
}
</style>
