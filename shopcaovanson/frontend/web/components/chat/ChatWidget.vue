<template>
  <Teleport to="body">
    <button
      v-if="!chat.panelOpen.value"
      type="button"
      class="chat-widget__fab"
      aria-label="Mở chat hỗ trợ"
      @click="handleOpen"
    >
      <v-icon size="26" color="white">mdi-chat-outline</v-icon>
      <span
        v-if="chat.unreadCount.value > 0"
        class="chat-widget__badge"
      >
        {{ chat.unreadCount.value > 99 ? '99+' : chat.unreadCount.value }}
      </span>
    </button>

    <Transition name="chat-widget-scrim">
      <div
        v-if="chat.panelOpen.value"
        class="chat-widget__scrim"
        aria-hidden="true"
        @click="chat.closeChat()"
      />
    </Transition>

    <Transition name="chat-widget-panel">
      <aside
        v-if="chat.panelOpen.value"
        class="chat-widget__panel"
        role="dialog"
        aria-label="Chat hỗ trợ"
      >
        <ChatRoom class="h-100" />
      </aside>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
const chat = useChat()

const handleOpen = async () => {
  await chat.openChat()
}
</script>

<style scoped>
.chat-widget__fab {
  position: fixed;
  right: max(20px, env(safe-area-inset-right, 0px));
  bottom: max(24px, env(safe-area-inset-bottom, 0px));
  z-index: var(--chat-widget-z);
  width: 56px;
  height: 56px;
  margin: 0;
  padding: 0;
  border: none;
  border-radius: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  background: rgb(var(--v-theme-accent));
  color: #fff;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.18);
  overflow: visible;
  transform: translateZ(0);
  transition: box-shadow 0.2s ease, filter 0.2s ease;
}

.chat-widget__fab:hover {
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.22);
  filter: brightness(1.04);
}

.chat-widget__badge {
  position: absolute;
  top: -2px;
  right: -2px;
  min-width: 20px;
  height: 20px;
  padding: 0 5px;
  border-radius: 10px;
  background: rgb(var(--v-theme-error));
  color: #fff;
  font-size: 11px;
  font-weight: 700;
  line-height: 20px;
  text-align: center;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.2);
  pointer-events: none;
}

.chat-widget__scrim {
  position: fixed;
  inset: 0;
  z-index: var(--chat-widget-z);
  background: rgba(0, 0, 0, 0.28);
}

.chat-widget__panel {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  width: min(380px, 100vw);
  z-index: calc(var(--chat-widget-z) + 1);
  background: var(--color-bg, #fff);
  box-shadow: -4px 0 24px rgba(0, 0, 0, 0.12);
  contain: strict;
  isolation: isolate;
  transform: translateZ(0);
  will-change: transform;
}

.chat-widget-scrim-enter-active,
.chat-widget-scrim-leave-active {
  transition: opacity 0.22s ease;
}

.chat-widget-scrim-enter-from,
.chat-widget-scrim-leave-to {
  opacity: 0;
}

.chat-widget-panel-enter-active,
.chat-widget-panel-leave-active {
  transition: transform 0.28s cubic-bezier(0.32, 0.72, 0, 1);
}

.chat-widget-panel-enter-from,
.chat-widget-panel-leave-to {
  transform: translateX(100%);
}
</style>
