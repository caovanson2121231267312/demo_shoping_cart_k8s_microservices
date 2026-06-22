<template>
  <div class="admin-chart-wrap">
    <div v-if="!series.length" class="admin-chart-wrap__empty">
      <v-icon size="40" color="grey-lighten-1">mdi-chart-bar</v-icon>
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
  valueKey: 'revenue' | 'orders' | 'count' | 'avg_rating'
  color?: string
  label?: string
  formatValue?: (v: number) => string
}>()

ensureChartJs()

const chartColor = computed(() => props.color || readAdminChartColors().primary)

const chartData = computed<ChartData<'bar'>>(() => ({
  labels: props.series.map((p) => p.label),
  datasets: [
    {
      label: props.label || datasetLabel.value,
      data: props.series.map((p) => Number(p[props.valueKey] ?? 0)),
      backgroundColor: hexToRgba(chartColor.value, 0.85),
      hoverBackgroundColor: chartColor.value,
      borderRadius: 8,
      borderSkipped: false,
      maxBarThickness: 48,
    },
  ],
}))

const datasetLabel = computed(() => {
  const map: Record<string, string> = {
    revenue: 'Doanh thu',
    orders: 'Đơn hàng',
    count: 'Số lượng',
    avg_rating: 'Đánh giá TB',
  }
  return map[props.valueKey] ?? 'Giá trị'
})

const chartOptions = computed<ChartOptions<'bar'>>(() => {
  const base = baseCartesianOptions()
  const formatter = props.formatValue ?? ((v: number) => String(v))
  return {
    ...base,
    plugins: {
      ...base.plugins,
      legend: { display: !!props.label, ...base.plugins?.legend },
      tooltip: {
        ...base.plugins?.tooltip,
        callbacks: {
          label: (ctx) => `${ctx.dataset.label}: ${formatter(Number(ctx.parsed.y))}`,
        },
      },
    },
  }
})
</script>

<style scoped>
.admin-chart-wrap {
  position: relative;
  height: 280px;
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
