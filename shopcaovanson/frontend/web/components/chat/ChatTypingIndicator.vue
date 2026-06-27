<template>
  <div class="chat-typing" role="status" aria-live="polite" aria-atomic="true">
    <div class="chat-typing__row">
      <div class="chat-typing__bubble" aria-hidden="true">
        <span class="chat-typing__dot" />
        <span class="chat-typing__dot" />
        <span class="chat-typing__dot" />
      </div>
      <span class="chat-typing__label">{{ label }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  label: string
}>()
</script>

<style scoped>
.chat-typing {
  padding: 4px 0 8px;
  min-height: 40px;
}

.chat-typing__row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.chat-typing__bubble {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 10px 14px;
  border-radius: 18px;
  background: var(--color-surface, #f0f2f5);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.06);
}

.chat-typing__dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #90949c;
  animation: chat-typing-bounce 1.4s ease-in-out infinite;
}

.chat-typing__dot:nth-child(2) {
  animation-delay: 0.15s;
}

.chat-typing__dot:nth-child(3) {
  animation-delay: 0.3s;
}

.chat-typing__label {
  font-size: 0.8125rem;
  color: var(--color-text-muted, #65676b);
  line-height: 1.3;
}

@keyframes chat-typing-bounce {
  0%,
  60%,
  100% {
    transform: translateY(0);
    opacity: 0.45;
  }
  30% {
    transform: translateY(-4px);
    opacity: 1;
  }
}
</style>
