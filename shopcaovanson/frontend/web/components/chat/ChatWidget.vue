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

    <v-navigation-drawer
      :model-value="chat.panelOpen.value"
      location="right"
      temporary
      width="380"
      class="chat-widget__panel"
      @update:model-value="onPanelChange"
    >
      <ChatRoom class="h-100" />
    </v-navigation-drawer>
  </Teleport>
</template>

<script setup lang="ts">
const chat = useChat()

const handleOpen = async () => {
  await chat.openChat()
}

const onPanelChange = (open: boolean) => {
  if (!open) {
    chat.closeChat()
  }
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
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.22);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  overflow: visible;
}

.chat-widget__fab:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.28);
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

.chat-widget__panel {
  z-index: calc(var(--chat-widget-z) + 1) !important;
}
</style>
