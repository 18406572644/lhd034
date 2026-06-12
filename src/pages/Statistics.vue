<script setup lang="ts">
import { ref, onMounted, watch, nextTick } from 'vue'
import * as echarts from 'echarts'
import { statisticsApi } from '@/api'
import { ConditionLabels } from '@/types'
import type {
  OverviewStats,
  MonthlyData,
  PlatformStat,
  PublisherStat,
  ConditionStat,
  RatingStat,
  PlayTimeRankItem,
  DifficultyStat,
  ValueTrendItem,
  RegionStat,
  CompletionRate
} from '@/types'

const loading = ref(true)
const selectedYear = ref(new Date().getFullYear())
const years = ref<number[]>([])

const overview = ref<OverviewStats | null>(null)
const monthlyData = ref<MonthlyData[]>([])
const platformData = ref<PlatformStat[]>([])
const publisherData = ref<PublisherStat[]>([])
const conditionData = ref<ConditionStat[]>([])
const ratingData = ref<RatingStat[]>([])
const playTimeTopData = ref<PlayTimeRankItem[]>([])
const difficultyData = ref<DifficultyStat[]>([])
const valueTrendData = ref<ValueTrendItem[]>([])
const regionData = ref<RegionStat[]>([])
const completionRate = ref<CompletionRate | null>(null)

const annualChartRef = ref<HTMLDivElement | null>(null)
const platformChartRef = ref<HTMLDivElement | null>(null)
const publisherChartRef = ref<HTMLDivElement | null>(null)
const conditionChartRef = ref<HTMLDivElement | null>(null)
const ratingChartRef = ref<HTMLDivElement | null>(null)
const playTimeChartRef = ref<HTMLDivElement | null>(null)
const difficultyChartRef = ref<HTMLDivElement | null>(null)
const valueTrendChartRef = ref<HTMLDivElement | null>(null)
const regionChartRef = ref<HTMLDivElement | null>(null)
const completionChartRef = ref<HTMLDivElement | null>(null)

let annualChart: echarts.ECharts | null = null
let platformChart: echarts.ECharts | null = null
let publisherChart: echarts.ECharts | null = null
let conditionChart: echarts.ECharts | null = null
let ratingChart: echarts.ECharts | null = null
let playTimeChart: echarts.ECharts | null = null
let difficultyChart: echarts.ECharts | null = null
let valueTrendChart: echarts.ECharts | null = null
let regionChart: echarts.ECharts | null = null
let completionChart: echarts.ECharts | null = null

const neonColors = ['#00F0FF', '#FFD93D', '#FF6B9D', '#6BCB77', '#FF8C42', '#FF4757', '#A78BFA', '#22D3EE']

const getCompletionColor = () => {
  const rate = completionRate.value?.rate ?? 0
  if (rate >= 80) return '#6BCB77'
  if (rate >= 50) return '#FFD93D'
  if (rate >= 20) return '#FF8C42'
  return '#FF4757'
}

const initYears = () => {
  const currentYear = new Date().getFullYear()
  years.value = Array.from({ length: 5 }, (_, i) => currentYear - i)
}

const initCharts = () => {
  if (annualChartRef.value) annualChart = echarts.init(annualChartRef.value)
  if (platformChartRef.value) platformChart = echarts.init(platformChartRef.value)
  if (publisherChartRef.value) publisherChart = echarts.init(publisherChartRef.value)
  if (conditionChartRef.value) conditionChart = echarts.init(conditionChartRef.value)
  if (ratingChartRef.value) ratingChart = echarts.init(ratingChartRef.value)
  if (playTimeChartRef.value) playTimeChart = echarts.init(playTimeChartRef.value)
  if (difficultyChartRef.value) difficultyChart = echarts.init(difficultyChartRef.value)
  if (valueTrendChartRef.value) valueTrendChart = echarts.init(valueTrendChartRef.value)
  if (regionChartRef.value) regionChart = echarts.init(regionChartRef.value)
  if (completionChartRef.value) completionChart = echarts.init(completionChartRef.value)
}

const renderAnnualChart = () => {
  if (!annualChart) return
  const months = ['1月', '2月', '3月', '4月', '5月', '6月', '7月', '8月', '9月', '10月', '11月', '12月']
  const counts = monthlyData.value.map(d => d.count)

  annualChart.setOption({
    backgroundColor: 'transparent',
    grid: { left: '3%', right: '4%', bottom: '3%', top: '10%', containLabel: true },
    xAxis: {
      type: 'category',
      data: months,
      axisLine: { lineStyle: { color: '#00F0FF' } },
      axisLabel: { color: '#E8E8E8', fontFamily: 'Press Start 2P', fontSize: 10 },
      axisTick: { show: false }
    },
    yAxis: {
      type: 'value',
      axisLine: { lineStyle: { color: '#00F0FF' } },
      axisLabel: { color: '#E8E8E8', fontFamily: 'Press Start 2P', fontSize: 10 },
      splitLine: { lineStyle: { color: 'rgba(0, 240, 255, 0.1)' } }
    },
    series: [{
      type: 'bar',
      data: counts,
      itemStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: '#00F0FF' },
          { offset: 1, color: '#00B8C4' }
        ]),
        borderRadius: [4, 4, 0, 0],
        borderColor: '#FFD93D',
        borderWidth: 2
      },
      emphasis: {
        itemStyle: {
          shadowBlur: 20,
          shadowColor: '#00F0FF'
        }
      },
      barWidth: '50%',
      animationDuration: 1500,
      animationEasing: 'elasticOut'
    }]
  })
}

const renderPlatformChart = () => {
  if (!platformChart) return
  const data = platformData.value.map(p => ({ name: p.platform, value: p.count }))

  platformChart.setOption({
    backgroundColor: 'transparent',
    tooltip: { trigger: 'item', fontFamily: 'VT323', fontSize: 16 },
    legend: {
      orient: 'vertical',
      right: '5%',
      top: 'center',
      textStyle: { color: '#E8E8E8', fontFamily: 'Press Start 2P', fontSize: 10 },
      itemWidth: 12,
      itemHeight: 12
    },
    series: [{
      type: 'pie',
      radius: ['30%', '60%'],
      center: ['35%', '50%'],
      avoidLabelOverlap: false,
      itemStyle: {
        borderRadius: 4,
        borderColor: '#1A1A2E',
        borderWidth: 3
      },
      label: { show: false },
      emphasis: {
        label: { show: true, fontSize: 14, fontWeight: 'bold', color: '#FFD93D', fontFamily: 'Press Start 2P' },
        itemStyle: { shadowBlur: 20, shadowOffsetX: 0, shadowColor: 'rgba(0, 0, 0, 0.5)' }
      },
      data: data,
      color: neonColors,
      animationDuration: 1500,
      animationEasing: 'bounceOut'
    }]
  })
}

const renderPublisherChart = () => {
  if (!publisherChart) return
  const top10 = [...publisherData.value].sort((a, b) => b.count - a.count).slice(0, 10).reverse()
  const names = top10.map(p => p.publisher)
  const counts = top10.map(p => p.count)

  publisherChart.setOption({
    backgroundColor: 'transparent',
    grid: { left: '3%', right: '10%', bottom: '3%', top: '5%', containLabel: true },
    xAxis: {
      type: 'value',
      axisLine: { lineStyle: { color: '#00F0FF' } },
      axisLabel: { color: '#E8E8E8', fontFamily: 'Press Start 2P', fontSize: 10 },
      splitLine: { lineStyle: { color: 'rgba(0, 240, 255, 0.1)' } }
    },
    yAxis: {
      type: 'category',
      data: names,
      axisLine: { lineStyle: { color: '#00F0FF' } },
      axisLabel: { color: '#E8E8E8', fontFamily: 'Press Start 2P', fontSize: 9 },
      axisTick: { show: false }
    },
    series: [{
      type: 'bar',
      data: counts,
      itemStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 1, 0, [
          { offset: 0, color: '#FF6B9D' },
          { offset: 1, color: '#FF8C42' }
        ]),
        borderRadius: [0, 4, 4, 0],
        borderColor: '#FFD93D',
        borderWidth: 2
      },
      barWidth: '60%',
      animationDuration: 1500,
      animationEasing: 'quadOut'
    }]
  })
}

const renderConditionChart = () => {
  if (!conditionChart) return
  const data = conditionData.value.map(c => ({
    name: ConditionLabels[c.condition] || c.condition,
    value: c.count
  }))

  conditionChart.setOption({
    backgroundColor: 'transparent',
    tooltip: { trigger: 'item', fontFamily: 'VT323', fontSize: 16 },
    legend: {
      bottom: '0',
      textStyle: { color: '#E8E8E8', fontFamily: 'Press Start 2P', fontSize: 10 },
      itemWidth: 12,
      itemHeight: 12
    },
    series: [{
      type: 'pie',
      radius: ['40%', '70%'],
      center: ['50%', '40%'],
      avoidLabelOverlap: false,
      itemStyle: {
        borderRadius: 6,
        borderColor: '#1A1A2E',
        borderWidth: 3
      },
      label: { show: false },
      emphasis: {
        label: { show: true, fontSize: 14, fontWeight: 'bold', color: '#FFD93D', fontFamily: 'Press Start 2P' },
        itemStyle: { shadowBlur: 20, shadowOffsetX: 0, shadowColor: 'rgba(0, 0, 0, 0.5)' }
      },
      data: data,
      color: ['#6BCB77', '#00F0FF', '#FFD93D', '#FF8C42', '#FF4757'],
      animationDuration: 1500,
      animationEasing: 'bounceOut'
    }]
  })
}

const renderRatingChart = () => {
  if (!ratingChart) return
  const data = ratingData.value || []
  const labels = data.map(r => r.label || `${r.rating}星`)
  const counts = data.map(r => r.count)

  ratingChart.setOption({
    backgroundColor: 'transparent',
    grid: { left: '3%', right: '4%', bottom: '3%', top: '10%', containLabel: true },
    tooltip: { trigger: 'axis', fontFamily: 'VT323', fontSize: 16 },
    xAxis: {
      type: 'category',
      data: labels,
      axisLine: { lineStyle: { color: '#00F0FF' } },
      axisLabel: { color: '#E8E8E8', fontFamily: 'Press Start 2P', fontSize: 10 },
      axisTick: { show: false }
    },
    yAxis: {
      type: 'value',
      axisLine: { lineStyle: { color: '#00F0FF' } },
      axisLabel: { color: '#E8E8E8', fontFamily: 'Press Start 2P', fontSize: 10 },
      splitLine: { lineStyle: { color: 'rgba(0, 240, 255, 0.1)' } }
    },
    series: [{
      type: 'bar',
      data: counts,
      itemStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: '#FF6B9D' },
          { offset: 1, color: '#A78BFA' }
        ]),
        borderRadius: [4, 4, 0, 0],
        borderColor: '#FFD93D',
        borderWidth: 2
      },
      barWidth: '50%',
      animationDuration: 1500,
      animationEasing: 'elasticOut',
      markLine: {
        silent: true,
        lineStyle: { color: '#FFD93D', type: 'dashed', width: 2 },
        label: { color: '#FFD93D', fontFamily: 'Press Start 2P', fontSize: 10, formatter: '均值' },
        data: [{ type: 'average', name: '平均' }]
      }
    }]
  })
}

const renderPlayTimeChart = () => {
  if (!playTimeChart) return
  const topItems = [...playTimeTopData.value].reverse()
  const names = topItems.map(item => item.title.length > 8 ? item.title.slice(0, 8) + '...' : item.title)
  const hours = topItems.map(item => item.playTimeHours)

  playTimeChart.setOption({
    backgroundColor: 'transparent',
    grid: { left: '3%', right: '12%', bottom: '3%', top: '5%', containLabel: true },
    tooltip: {
      trigger: 'axis',
      fontFamily: 'VT323',
      fontSize: 16,
      formatter: (params: any) => {
        const idx = params[0].dataIndex
        const item = topItems[idx]
        return `<strong>${item.title}</strong><br/>平台: ${item.platform}<br/>时长: ${item.playTimeHours}h<br/>通关: ${item.completionDate || '未知'}`
      }
    },
    xAxis: {
      type: 'value',
      axisLine: { lineStyle: { color: '#00F0FF' } },
      axisLabel: { color: '#E8E8E8', fontFamily: 'Press Start 2P', fontSize: 10 },
      splitLine: { lineStyle: { color: 'rgba(0, 240, 255, 0.1)' } }
    },
    yAxis: {
      type: 'category',
      data: names,
      axisLine: { lineStyle: { color: '#00F0FF' } },
      axisLabel: { color: '#E8E8E8', fontFamily: 'Press Start 2P', fontSize: 9 },
      axisTick: { show: false }
    },
    series: [{
      type: 'bar',
      data: hours,
      itemStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 1, 0, [
          { offset: 0, color: '#22D3EE' },
          { offset: 1, color: '#6BCB77' }
        ]),
        borderRadius: [0, 4, 4, 0],
        borderColor: '#FFD93D',
        borderWidth: 2
      },
      label: {
        show: true,
        position: 'right',
        color: '#FFD93D',
        fontFamily: 'Press Start 2P',
        fontSize: 9,
        formatter: '{c}h'
      },
      barWidth: '60%',
      animationDuration: 1500,
      animationEasing: 'quadOut'
    }]
  })
}

const renderDifficultyChart = () => {
  if (!difficultyChart) return
  const data = difficultyData.value || []
  const labels = data.map(d => d.label)
  const counts = data.map(d => d.count)
  const avgHours = data.map(d => d.avgPlayTimeHours)

  difficultyChart.setOption({
    backgroundColor: 'transparent',
    grid: { left: '3%', right: '4%', bottom: '3%', top: '12%', containLabel: true },
    tooltip: { trigger: 'axis', fontFamily: 'VT323', fontSize: 16 },
    legend: {
      data: ['通关数量', '平均时长(h)'],
      top: '0',
      textStyle: { color: '#E8E8E8', fontFamily: 'Press Start 2P', fontSize: 10 },
      itemWidth: 12,
      itemHeight: 12
    },
    xAxis: {
      type: 'category',
      data: labels,
      axisLine: { lineStyle: { color: '#00F0FF' } },
      axisLabel: { color: '#E8E8E8', fontFamily: 'Press Start 2P', fontSize: 9 },
      axisTick: { show: false }
    },
    yAxis: [
      {
        type: 'value',
        name: '数量',
        nameTextStyle: { color: '#E8E8E8', fontFamily: 'Press Start 2P', fontSize: 10 },
        axisLine: { lineStyle: { color: '#00F0FF' } },
        axisLabel: { color: '#E8E8E8', fontFamily: 'Press Start 2P', fontSize: 10 },
        splitLine: { lineStyle: { color: 'rgba(0, 240, 255, 0.1)' } }
      },
      {
        type: 'value',
        name: '时长(h)',
        nameTextStyle: { color: '#E8E8E8', fontFamily: 'Press Start 2P', fontSize: 10 },
        axisLine: { lineStyle: { color: '#FFD93D' } },
        axisLabel: { color: '#E8E8E8', fontFamily: 'Press Start 2P', fontSize: 10 },
        splitLine: { show: false }
      }
    ],
    series: [
      {
        name: '通关数量',
        type: 'bar',
        data: counts,
        itemStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: '#FF6B9D' },
            { offset: 1, color: '#FF8C42' }
          ]),
          borderRadius: [4, 4, 0, 0],
          borderColor: '#FFD93D',
          borderWidth: 2
        },
        barWidth: '35%',
        animationDuration: 1500,
        animationEasing: 'elasticOut'
      },
      {
        name: '平均时长(h)',
        type: 'line',
        yAxisIndex: 1,
        data: avgHours,
        smooth: true,
        symbol: 'circle',
        symbolSize: 8,
        lineStyle: { color: '#FFD93D', width: 3 },
        itemStyle: { color: '#FFD93D', borderColor: '#1A1A2E', borderWidth: 2 },
        animationDuration: 1500
      }
    ]
  })
}

const renderValueTrendChart = () => {
  if (!valueTrendChart) return
  const data = valueTrendData.value || []
  const dates = data.map(d => d.date)
  const cumulative = data.map(d => d.cumulative)
  const monthly = data.map(d => d.value)

  valueTrendChart.setOption({
    backgroundColor: 'transparent',
    grid: { left: '3%', right: '4%', bottom: '8%', top: '12%', containLabel: true },
    tooltip: {
      trigger: 'axis',
      fontFamily: 'VT323',
      fontSize: 16,
      valueFormatter: (val: number) => `¥${val.toFixed(2)}`
    },
    legend: {
      data: ['累计投入', '月度投入'],
      top: '0',
      textStyle: { color: '#E8E8E8', fontFamily: 'Press Start 2P', fontSize: 10 },
      itemWidth: 12,
      itemHeight: 12
    },
    xAxis: {
      type: 'category',
      data: dates,
      boundaryGap: false,
      axisLine: { lineStyle: { color: '#00F0FF' } },
      axisLabel: { color: '#E8E8E8', fontFamily: 'Press Start 2P', fontSize: 9, rotate: 45 },
      axisTick: { show: false }
    },
    yAxis: {
      type: 'value',
      axisLine: { lineStyle: { color: '#00F0FF' } },
      axisLabel: {
        color: '#E8E8E8',
        fontFamily: 'Press Start 2P',
        fontSize: 10,
        formatter: (val: number) => `¥${val}`
      },
      splitLine: { lineStyle: { color: 'rgba(0, 240, 255, 0.1)' } }
    },
    series: [
      {
        name: '累计投入',
        type: 'line',
        data: cumulative,
        smooth: true,
        symbol: 'circle',
        symbolSize: 6,
        lineStyle: { color: '#6BCB77', width: 3 },
        itemStyle: { color: '#6BCB77', borderColor: '#1A1A2E', borderWidth: 2 },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(107, 203, 119, 0.4)' },
            { offset: 1, color: 'rgba(107, 203, 119, 0.05)' }
          ])
        },
        animationDuration: 1500
      },
      {
        name: '月度投入',
        type: 'line',
        data: monthly,
        smooth: true,
        symbol: 'diamond',
        symbolSize: 6,
        lineStyle: { color: '#FFD93D', width: 2, type: 'dashed' },
        itemStyle: { color: '#FFD93D', borderColor: '#1A1A2E', borderWidth: 2 },
        animationDuration: 1500
      }
    ]
  })
}

const renderRegionChart = () => {
  if (!regionChart) return
  const data = regionData.value.map(r => ({ name: r.label || r.region, value: r.count }))

  regionChart.setOption({
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'item',
      fontFamily: 'VT323',
      fontSize: 16,
      formatter: '{b}: {c} ({d}%)'
    },
    legend: {
      orient: 'vertical',
      right: '5%',
      top: 'center',
      textStyle: { color: '#E8E8E8', fontFamily: 'Press Start 2P', fontSize: 10 },
      itemWidth: 12,
      itemHeight: 12
    },
    series: [{
      type: 'pie',
      radius: ['30%', '65%'],
      center: ['35%', '50%'],
      roseType: 'radius',
      itemStyle: {
        borderRadius: 4,
        borderColor: '#1A1A2E',
        borderWidth: 3
      },
      label: {
        show: true,
        color: '#E8E8E8',
        fontFamily: 'Press Start 2P',
        fontSize: 9,
        formatter: '{d}%'
      },
      labelLine: { lineStyle: { color: '#00F0FF' } },
      emphasis: {
        label: { fontSize: 12, fontWeight: 'bold', color: '#FFD93D' },
        itemStyle: { shadowBlur: 20, shadowOffsetX: 0, shadowColor: 'rgba(0, 0, 0, 0.5)' }
      },
      data: data,
      color: neonColors,
      animationDuration: 1500,
      animationEasing: 'bounceOut'
    }]
  })
}

const renderCompletionChart = () => {
  if (!completionChart || !completionRate.value) return
  const data = completionRate.value
  const pieData = [
    { value: data.completedCartridges, name: '已通关', itemStyle: { color: '#6BCB77' } },
    { value: data.playingCount, name: '进行中', itemStyle: { color: '#FFD93D' } },
    { value: data.unstartedCount, name: '未开始', itemStyle: { color: '#00F0FF' } },
    { value: data.shelvedCount, name: '搁置', itemStyle: { color: '#FF8C42' } }
  ]

  completionChart.setOption({
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'item',
      fontFamily: 'VT323',
      fontSize: 16,
      formatter: '{b}: {c} ({d}%)'
    },
    legend: {
      bottom: '0',
      textStyle: { color: '#E8E8E8', fontFamily: 'Press Start 2P', fontSize: 10 },
      itemWidth: 12,
      itemHeight: 12
    },
    title: {
      text: `${data.rate.toFixed(1)}%`,
      subtext: '通关率',
      left: 'center',
      top: '38%',
      textAlign: 'center',
      textStyle: {
        color: '#FFD93D',
        fontFamily: 'Press Start 2P',
        fontSize: 28,
        fontWeight: 'bold'
      },
      subtextStyle: {
        color: '#E8E8E8',
        fontFamily: 'Press Start 2P',
        fontSize: 10
      }
    },
    series: [{
      type: 'pie',
      radius: ['55%', '80%'],
      center: ['50%', '45%'],
      avoidLabelOverlap: false,
      itemStyle: {
        borderRadius: 6,
        borderColor: '#1A1A2E',
        borderWidth: 3
      },
      label: { show: false },
      emphasis: {
        label: { show: true, fontSize: 14, fontWeight: 'bold', color: '#FFD93D', fontFamily: 'Press Start 2P' },
        itemStyle: { shadowBlur: 20, shadowOffsetX: 0, shadowColor: 'rgba(0, 0, 0, 0.5)' }
      },
      data: pieData,
      animationDuration: 1500,
      animationEasing: 'bounceOut'
    }]
  })
}

const renderAllCharts = () => {
  nextTick(() => {
    renderAnnualChart()
    renderPlatformChart()
    renderPublisherChart()
    renderConditionChart()
    renderRatingChart()
    renderPlayTimeChart()
    renderDifficultyChart()
    renderValueTrendChart()
    renderRegionChart()
    renderCompletionChart()
  })
}

const loadData = async () => {
  loading.value = true
  try {
    const [
      ovRes, annualRes, platRes, pubRes, condRes,
      ratingRes, playtimeRes, diffRes, valueRes, regionRes, compRes
    ] = await Promise.all([
      statisticsApi.getOverview(),
      statisticsApi.getAnnual(selectedYear.value),
      statisticsApi.getPlatforms(),
      statisticsApi.getPublishers(),
      statisticsApi.getConditions(),
      statisticsApi.getRatings(),
      statisticsApi.getPlayTimeTop10(),
      statisticsApi.getDifficulty(),
      statisticsApi.getValueTrend(),
      statisticsApi.getRegions(),
      statisticsApi.getCompletionRate()
    ])
    if (ovRes.code === 0) overview.value = ovRes.data
    if (annualRes.code === 0) monthlyData.value = annualRes.data
    if (platRes.code === 0) platformData.value = platRes.data
    if (pubRes.code === 0) publisherData.value = pubRes.data
    if (condRes.code === 0) conditionData.value = condRes.data
    if (ratingRes.code === 0) ratingData.value = ratingRes.data
    if (playtimeRes.code === 0) playTimeTopData.value = playtimeRes.data
    if (diffRes.code === 0) difficultyData.value = diffRes.data
    if (valueRes.code === 0) valueTrendData.value = valueRes.data
    if (regionRes.code === 0) regionData.value = regionRes.data
    if (compRes.code === 0) completionRate.value = compRes.data
    renderAllCharts()
  } finally {
    loading.value = false
  }
}

const handleResize = () => {
  annualChart?.resize()
  platformChart?.resize()
  publisherChart?.resize()
  conditionChart?.resize()
  ratingChart?.resize()
  playTimeChart?.resize()
  difficultyChart?.resize()
  valueTrendChart?.resize()
  regionChart?.resize()
  completionChart?.resize()
}

watch(selectedYear, loadData)

onMounted(() => {
  initYears()
  initCharts()
  loadData()
  window.addEventListener('resize', handleResize)
})
</script>

<template>
  <div class="p-6 space-y-6">
    <div class="text-center py-4">
      <h1 class="text-neon-blue glow-blue mb-2">年度游玩报告</h1>
      <p class="text-text-secondary pixel-font text-xs">ANNUAL PLAY REPORT</p>
    </div>

    <div class="flex justify-center">
      <select v-model="selectedYear" class="pixel-input text-center pixel-font">
        <option v-for="y in years" :key="y" :value="y">{{ y }} 年</option>
      </select>
    </div>

    <div v-if="loading" class="text-center py-16">
      <div class="text-4xl mb-4 animate-pulse">📊</div>
      <p class="text-text-secondary pixel-font">加载中...</p>
    </div>

    <template v-else>
      <div class="grid grid-cols-2 md:grid-cols-5 gap-4">
        <div class="stat-card">
          <div class="text-2xl mb-2">🎮</div>
          <div class="stat-number text-neon-blue">{{ overview?.totalCartridges ?? 0 }}</div>
          <div class="text-text-secondary mt-2 text-sm">总卡带</div>
        </div>
        <div class="stat-card">
          <div class="text-2xl mb-2">🏆</div>
          <div class="stat-number text-pixel-green">{{ overview?.totalPlaythroughs ?? 0 }}</div>
          <div class="text-text-secondary mt-2 text-sm">已通关</div>
        </div>
        <div class="stat-card">
          <div class="text-2xl mb-2">⏱️</div>
          <div class="stat-number text-bright-yellow">{{ overview?.totalPlayTime ?? 0 }}h</div>
          <div class="text-text-secondary mt-2 text-sm">总时长</div>
        </div>
        <div class="stat-card">
          <div class="text-2xl mb-2">💰</div>
          <div class="stat-number text-pixel-pink">¥{{ overview?.totalValue ?? 0 }}</div>
          <div class="text-text-secondary mt-2 text-sm">总价值</div>
        </div>
        <div class="stat-card">
          <div class="text-2xl mb-2">📈</div>
          <div class="stat-number" :style="{ color: getCompletionColor() }">{{ completionRate?.rate?.toFixed(1) ?? '0.0' }}%</div>
          <div class="text-text-secondary mt-2 text-sm">通关率</div>
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div class="pixel-card p-4">
          <h3 class="text-neon-blue glow-blue mb-4 text-center">年度通关统计</h3>
          <div ref="annualChartRef" class="w-full h-64"></div>
        </div>

        <div class="pixel-card p-4">
          <h3 class="text-pixel-pink mb-4 text-center">平台占比</h3>
          <div ref="platformChartRef" class="w-full h-64"></div>
        </div>

        <div class="pixel-card p-4">
          <h3 class="text-bright-yellow glow-yellow mb-4 text-center">发行商 Top 10</h3>
          <div ref="publisherChartRef" class="w-full h-64"></div>
        </div>

        <div class="pixel-card p-4">
          <h3 class="text-pixel-green mb-4 text-center">品相分布</h3>
          <div ref="conditionChartRef" class="w-full h-64"></div>
        </div>

        <div class="pixel-card p-4">
          <h3 class="mb-4 text-center" style="color: #A78BFA">⭐ 评分分布</h3>
          <div ref="ratingChartRef" class="w-full h-64"></div>
        </div>

        <div class="pixel-card p-4">
          <h3 class="mb-4 text-center" style="color: #22D3EE">🏅 游玩时长 TOP 10</h3>
          <div ref="playTimeChartRef" class="w-full h-64"></div>
        </div>

        <div class="pixel-card p-4">
          <h3 class="mb-4 text-center" style="color: #FF8C42">🎯 难度分析</h3>
          <div ref="difficultyChartRef" class="w-full h-64"></div>
        </div>

        <div class="pixel-card p-4">
          <h3 class="text-pixel-green mb-4 text-center">💎 收藏价值趋势</h3>
          <div ref="valueTrendChartRef" class="w-full h-64"></div>
        </div>

        <div class="pixel-card p-4">
          <h3 class="mb-4 text-center" style="color: #6BCB77">🌍 区域分布</h3>
          <div ref="regionChartRef" class="w-full h-64"></div>
        </div>

        <div class="pixel-card p-4">
          <h3 class="text-bright-yellow glow-yellow mb-4 text-center">📊 通关率分析</h3>
          <div ref="completionChartRef" class="w-full h-64"></div>
        </div>
      </div>
    </template>
  </div>
</template>
