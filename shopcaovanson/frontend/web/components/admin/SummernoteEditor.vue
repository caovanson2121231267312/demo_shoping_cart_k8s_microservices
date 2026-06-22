<template>
  <ClientOnly>
    <div class="summernote-editor">
      <div ref="editorEl" />
    </div>
    <template #fallback>
      <v-textarea :model-value="modelValue" label="Nội dung" rows="8" variant="outlined" readonly />
    </template>
  </ClientOnly>
</template>

<script setup lang="ts">
const props = defineProps<{
  modelValue: string
  height?: number
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const editorEl = ref<HTMLElement | null>(null)
let initialized = false

const loadScript = (src: string) =>
  new Promise<void>((resolve, reject) => {
    if (document.querySelector(`script[src="${src}"]`)) {
      resolve()
      return
    }
    const s = document.createElement('script')
    s.src = src
    s.onload = () => resolve()
    s.onerror = reject
    document.head.appendChild(s)
  })

const loadCss = (href: string) => {
  if (document.querySelector(`link[href="${href}"]`)) return
  const l = document.createElement('link')
  l.rel = 'stylesheet'
  l.href = href
  document.head.appendChild(l)
}

const initEditor = async () => {
  if (!editorEl.value || initialized) return
  loadCss('https://cdn.jsdelivr.net/npm/summernote@0.8.20/dist/summernote-lite.min.css')
  await loadScript('https://cdn.jsdelivr.net/npm/jquery@3.7.1/dist/jquery.min.js')
  await loadScript('https://cdn.jsdelivr.net/npm/summernote@0.8.20/dist/summernote-lite.min.js')

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const $ = (window as any).jQuery
  const $el = $(editorEl.value)
  $el.summernote({
    height: props.height ?? 320,
    placeholder: 'Nhập nội dung bài viết...',
    toolbar: [
      ['style', ['style']],
      ['font', ['bold', 'italic', 'underline', 'clear']],
      ['para', ['ul', 'ol', 'paragraph']],
      ['insert', ['link', 'picture', 'table']],
      ['view', ['fullscreen', 'codeview']],
    ],
    callbacks: {
      onChange: (contents: string) => emit('update:modelValue', contents),
    },
  })
  if (props.modelValue) {
    $el.summernote('code', props.modelValue)
  }
  initialized = true
}

watch(() => props.modelValue, (val) => {
  if (!initialized || !editorEl.value) return
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const $ = (window as any).jQuery
  const current = $(editorEl.value).summernote('code')
  if (current !== val) {
    $(editorEl.value).summernote('code', val || '')
  }
})

onMounted(() => nextTick(initEditor))
onBeforeUnmount(() => {
  if (editorEl.value && initialized) {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const $ = (window as any).jQuery
    $(editorEl.value).summernote('destroy')
  }
})
</script>

<style>
.summernote-editor .note-editor {
  border-radius: var(--admin-radius-xs);
  border-color: rgba(0, 0, 0, 0.12);
}
</style>
