<template>
  <div class="admin-chart">
    <div v-if="!series.length" class="admin-chart__empty text-caption text-muted text-center py-8">
      Chưa có dữ liệu
    </div>
    <canvas v-else ref="canvasRef" class="admin-chart__canvas" />
  </div>
</template>

<script setup lang="ts">
import type { AnalyticsSeriesPoint } from '~/types'

const props = defineProps<{
  series: AnalyticsSeriesPoint[]
  valueKey: 'revenue' | 'orders' | 'count' | 'avg_rating'
  color?: string
  formatValue?: (v: number) => string
}>()

const canvasRef = ref<HTMLCanvasElement | null>(null)

const draw = () => {
  const canvas = canvasRef.value
  if (!canvas || !props.series.length) return

  const dpr = window.devicePixelRatio || 1
  const width = canvas.clientWidth || 600
  const height = 220
  canvas.width = width * dpr
  canvas.height = height * dpr

  const ctx = canvas.getContext('2d')
  if (!ctx) return
  ctx.scale(dpr, dpr)

  const values = props.series.map((p) => Number(p[props.valueKey] ?? 0))
  const max = Math.max(...values, 1)
  const barColor = props.color || '#1565c0'
  const pad = { l: 8, r: 8, t: 16, b: 36 }
  const chartW = width - pad.l - pad.r
  const chartH = height - pad.t - pad.b
  const gap = 4
  const barW = Math.max(4, (chartW - gap * (values.length - 1)) / values.length)

  ctx.clearRect(0, 0, width, height)

  values.forEach((v, i) => {
    const h = (v / max) * chartH
    const x = pad.l + i * (barW + gap)
    const y = pad.t + chartH - h
    ctx.fillStyle = barColor
    ctx.beginPath()
    ctx.roundRect(x, y, barW, h, 3)
    ctx.fill()

    if (values.length <= 14) {
      ctx.fillStyle = '#757575'
      ctx.font = '10px Segoe UI, sans-serif'
      ctx.textAlign = 'center'
      const label = props.series[i]?.label ?? ''
      ctx.fillText(label.length > 8 ? `${label.slice(0, 7)}…` : label, x + barW / 2, height - 8)
    }
  })
}

watch(() => props.series, () => nextTick(draw), { deep: true })
onMounted(() => {
  draw()
  window.addEventListener('resize', draw)
})
onUnmounted(() => window.removeEventListener('resize', draw))
</script>

<style scoped>
.admin-chart__canvas {
  width: 100%;
  height: 220px;
  display: block;
}

.admin-chart__empty {
  background: var(--color-surface-muted, #f5f5f5);
  border-radius: 8px;
}
</style>
