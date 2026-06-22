<template>
  <div class="admin-chart-wrap">
    <div v-if="!series.length" class="admin-chart-wrap__empty">
      <v-icon size="40" color="grey-lighten-1">mdi-chart-areaspline</v-icon>
      <span>Chưa có dữ liệu</span>
    </div>
    <Bar v-else :data="chartData" :options="chartOptions" />
  </div>
</template>

<script setup lang="ts">
import { Bar } from 'vue-chartjs'
import type { ChartData, ChartOptions } from 'chart.js'
import type { AnalyticsSeriesPoint } from '~/types'
import { baseCartesianOptions, ensureChartJs, hexToRgba, readAdminChartColors } from '~/utils/adminChart'

const props = defineProps<{
  series: AnalyticsSeriesPoint[]
  formatRevenue?: (v: number) => string
}>()

ensureChartJs()

const colors = computed(() => readAdminChartColors())

const chartData = computed<ChartData<'bar'>>(() => ({
  labels: props.series.map((p) => p.label),
  datasets: [
    {
      type: 'bar' as const,
      label: 'Đơn hàng',
      data: props.series.map((p) => Number(p.orders ?? 0)),
      backgroundColor: hexToRgba(colors.value.accent, 0.75),
      hoverBackgroundColor: colors.value.accent,
      borderRadius: 6,
      yAxisID: 'y',
      order: 2,
      maxBarThickness: 36,
    },
    {
      type: 'line' as const,
      label: 'Doanh thu',
      data: props.series.map((p) => Number(p.revenue ?? 0)),
      borderColor: colors.value.primary,
      backgroundColor: hexToRgba(colors.value.primary, 0.08),
      fill: true,
      tension: 0.4,
      pointRadius: 3,
      pointHoverRadius: 5,
      pointBackgroundColor: '#fff',
      pointBorderColor: colors.value.primary,
      pointBorderWidth: 2,
      borderWidth: 2.5,
      yAxisID: 'y1',
      order: 1,
    },
  ],
}))

const chartOptions = computed<ChartOptions<'bar'>>(() => {
  const base = baseCartesianOptions()
  const fmtRev = props.formatRevenue ?? ((v: number) => String(v))
  return {
    ...base,
    plugins: {
      ...base.plugins,
      tooltip: {
        ...base.plugins?.tooltip,
        callbacks: {
          label: (ctx) => {
            const val = Number(ctx.parsed.y)
            if (ctx.dataset.label === 'Doanh thu') {
              return ` Doanh thu: ${fmtRev(val)}`
            }
            return ` Đơn hàng: ${val}`
          },
        },
      },
    },
    scales: {
      x: base.scales?.x,
      y: {
        ...base.scales?.y,
        position: 'left',
        title: { display: true, text: 'Đơn hàng', font: { size: 11 } },
        grid: { color: 'rgba(148, 163, 184, 0.12)' },
      },
      y1: {
        position: 'right',
        beginAtZero: true,
        grid: { drawOnChartArea: false },
        border: { display: false },
        title: { display: true, text: 'Doanh thu', font: { size: 11 } },
        ticks: {
          font: { size: 11 },
          callback: (v) => fmtRev(Number(v)),
        },
      },
    },
  }
})
</script>

<style scoped>
.admin-chart-wrap {
  position: relative;
  height: 320px;
  width: 100%;
}

.admin-chart-wrap__empty {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: var(--admin-text-muted, #94a3b8);
  font-size: 13px;
  background: rgba(148, 163, 184, 0.06);
  border-radius: var(--admin-radius-sm);
}
</style>
