<template>
  <div>
    <v-btn
      v-if="!chat.panelOpen.value"
      class="chat-widget-fab"
      color="accent"
      size="large"
      icon
      elevation="8"
      @click="handleOpen"
    >
      <v-badge
        :content="chat.unreadCount.value"
        :model-value="chat.unreadCount.value > 0"
        color="error"
      >
        <v-icon size="28" color="white">mdi-chat-processing</v-icon>
      </v-badge>
    </v-btn>

    <v-navigation-drawer
      :model-value="chat.panelOpen.value"
      location="right"
      temporary
      width="380"
      class="chat-panel"
      @update:model-value="onPanelChange"
    >
      <div class="d-flex flex-column h-100">
        <div class="d-flex align-center pa-3 bg-grey-lighten-3">
          <span class="text-h6">Hỗ trợ trực tuyến</span>
          <v-spacer />
          <v-btn icon variant="text" @click="chat.closeChat()">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </div>
        <ChatRoom class="flex-grow-1" />
      </div>
    </v-navigation-drawer>
  </div>
</template>

<script setup lang="ts">
const auth = useAuth()
const chat = useChat()
const router = useRouter()

const handleOpen = async () => {
  if (!auth.isLoggedIn.value) {
    await router.push('/auth/login')
    return
  }
  await chat.openChat()
}

const onPanelChange = (open: boolean) => {
  if (!open) {
    chat.closeChat()
  }
}
</script>
