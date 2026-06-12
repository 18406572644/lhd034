<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
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

const DEBUG = true
const debugLog = (msg: string, ...args: any[]) => {
  if (DEBUG) console.log(`[Statistics] ${msg}`, ...args)
}

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

const safeInitChart = (
  refEl: HTMLDivElement | null | undefined,
  existingInstance: echarts.ECharts | null,
  chartName: string
): echarts.ECharts | null => {
  try {
    if (!refEl) {
      debugLog(`⚠️ ${chartName} ref is null, skipping init`)
      return existingInstance
    }
    if (existingInstance) {
      debugLog(`♻️ disposing old ${chartName} instance before re-init`)
      existingInstance.dispose()
    }
    const instance = echarts.init(refEl)
    debugLog(`✅ ${chartName} initialized OK`, {
      clientWidth: refEl.clientWidth,
      clientHeight: refEl.clientHeight
    })
    return instance
  } catch (e) {
    console.error(`[Statistics] ❌ Failed to init ${chartName}:`, e)
    return existingInstance
  }
}

const initCharts = () => {
  debugLog('initCharts() called, checking refs...', {
    annual: !!annualChartRef.value,
    platform: !!platformChartRef.value,
    publisher: !!publisherChartRef.value,
    condition: !!conditionChartRef.value,
    rating: !!ratingChartRef.value,
    playtime: !!playTimeChartRef.value,
    difficulty: !!difficultyChartRef.value,
    valuetrend: !!valueTrendChartRef.value,
    region: !!regionChartRef.value,
    completion: !!completionChartRef.value
  })
  annualChart = safeInitChart(annualChartRef.value, annualChart, 'annualChart')
  platformChart = safeInitChart(platformChartRef.value, platformChart, 'platformChart')
  publisherChart = safeInitChart(publisherChartRef.value, publisherChart, 'publisherChart')
  conditionChart = safeInitChart(conditionChartRef.value, conditionChart, 'conditionChart')
  ratingChart = safeInitChart(ratingChartRef.value, ratingChart, 'ratingChart')
  playTimeChart = safeInitChart(playTimeChartRef.value, playTimeChart, 'playTimeChart')
  difficultyChart = safeInitChart(difficultyChartRef.value, difficultyChart, 'difficultyChart')
  valueTrendChart = safeInitChart(valueTrendChartRef.value, valueTrendChart, 'valueTrendChart')
  regionChart = safeInitChart(regionChartRef.value, regionChart, 'regionChart')
  completionChart = safeInitChart(completionChartRef.value, completionChart, 'completionChart')
}

const disposeAllCharts = () => {
  const all = [
    annualChart, platformChart, publisherChart, conditionChart,
    ratingChart, playTimeChart, difficultyChart, valueTrendChart,
    regionChart, completionChart
  ]
  all.forEach(inst => {
    try { inst?.dispose() } catch (e) { /* noop */ }
  })
  annualChart = platformChart = publisherChart = conditionChart = null
  ratingChart = playTimeChart = difficultyChart = valueTrendChart = null
  regionChart = completionChart = null
  debugLog('🧹 all charts disposed')
}

const safeSetOption = (
  instance: echarts.ECharts | null,
  option: any,
  chartName: string
) => {
  if (!instance) {
    debugLog(`⚠️ ${chartName} instance is null, skip render`)
    return
  }
  try {
    instance.setOption(option, true)
    instance.resize()
    debugLog(`🎨 ${chartName} rendered OK`)
  } catch (e) {
    console.error(`[Statistics] ❌ Failed to render ${chartName}:`, e)
  }
}

const renderAnnualChart = () => {
  if (!annualChart) return
  const months = ['1月', '2月', '3月', '4月', '5月', '6月', '7月', '8月', '9月', '10月', '11月', '12月']
  const counts = (monthlyData.value || []).map(d => d.count || 0)
  debugLog('renderAnnualChart data:', counts)
  safeSetOption(annualChart, {
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
      barWidth: '50%',
      animationDuration: 1500,
      animationEasing: 'elasticOut'
    }]
  }, 'annualChart')
}

const renderPlatformChart = () => {
  if (!platformChart) return
  const data = (platformData.value || []).map(p => ({ name: p.platform || '未知', value: p.count || 0 }))
  debugLog('renderPlatformChart data:', data)
  safeSetOption(platformChart, {
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
      itemStyle: { borderRadius: 4, borderColor: '#1A1A2E', borderWidth: 3 },
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
  }, 'platformChart')
}

const renderPublisherChart = () => {
  if (!publisherChart) return
  const top10 = [...(publisherData.value || [])].sort((a, b) => (b.count || 0) - (a.count || 0)).slice(0, 10).reverse()
  const names = top10.map(p => p.publisher || '未知')
  const counts = top10.map(p => p.count || 0)
  debugLog('renderPublisherChart data:', { names, counts })
  safeSetOption(publisherChart, {
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
      animationEasing: 'quadraticOut' as any
    }]
  }, 'publisherChart')
}

const renderConditionChart = () => {
  if (!conditionChart) return
  const data = (conditionData.value || []).map(c => ({
    name: ConditionLabels[c.condition] || c.condition || '未知',
    value: c.count || 0
  }))
  debugLog('renderConditionChart data:', data)
  safeSetOption(conditionChart, {
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
      itemStyle: { borderRadius: 6, borderColor: '#1A1A2E', borderWidth: 3 },
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
  }, 'conditionChart')
}

const renderRatingChart = () => {
  if (!ratingChart) return
  const data = ratingData.value || []
  const labels = data.map(r => r.label || `${r.rating}星`)
  const counts = data.map(r => r.count || 0)
  debugLog('renderRatingChart data:', { labels, counts })
  safeSetOption(ratingChart, {
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
  }, 'ratingChart')
}

const renderPlayTimeChart = () => {
  if (!playTimeChart) return
  const topItems = [...(playTimeTopData.value || [])].reverse()
  const names = topItems.map(item => {
    const t = item.title || '未知'
    return t.length > 8 ? t.slice(0, 8) + '...' : t
  })
  const hours = topItems.map(item => Number(item.playTimeHours) || 0)
  debugLog('renderPlayTimeChart data:', { names, hours })
  safeSetOption(playTimeChart, {
    backgroundColor: 'transparent',
    grid: { left: '3%', right: '12%', bottom: '3%', top: '5%', containLabel: true },
    tooltip: {
      trigger: 'axis',
      fontFamily: 'VT323',
      fontSize: 16,
      formatter: (params: any) => {
        const idx = params[0]?.dataIndex
        if (idx == null) return ''
        const item = topItems[idx]
        if (!item) return ''
        return `<strong>${item.title}</strong><br/>平台: ${item.platform || '未知'}<br/>时长: ${item.playTimeHours || 0}h<br/>通关: ${item.completionDate || '未知'}`
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
      animationEasing: 'quadraticOut' as any
    }]
  }, 'playTimeChart')
}

const renderDifficultyChart = () => {
  if (!difficultyChart) return
  const data = difficultyData.value || []
  const labels = data.map(d => d.label || `难度${d.difficulty}`)
  const counts = data.map(d => d.count || 0)
  const avgHours = data.map(d => Number(d.avgPlayTimeHours) || 0)
  debugLog('renderDifficultyChart data:', { labels, counts, avgHours })
  safeSetOption(difficultyChart, {
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
  }, 'difficultyChart')
}

const renderValueTrendChart = () => {
  if (!valueTrendChart) return
  const data = valueTrendData.value || []
  const dates = data.map(d => d.date)
  const cumulative = data.map(d => Number(d.cumulative) || 0)
  const monthly = data.map(d => Number(d.value) || 0)
  debugLog('renderValueTrendChart data:', { dates, cumulative, monthly })
  safeSetOption(valueTrendChart, {
    backgroundColor: 'transparent',
    grid: { left: '3%', right: '4%', bottom: '8%', top: '12%', containLabel: true },
    tooltip: {
      trigger: 'axis',
      fontFamily: 'VT323',
      fontSize: 16,
      valueFormatter: (val: any) => `¥${Number(val).toFixed(2)}`
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
  }, 'valueTrendChart')
}

const renderRegionChart = () => {
  if (!regionChart) return
  const data = (regionData.value || []).map(r => ({
    name: r.label || r.region || '未知',
    value: r.count || 0
  }))
  debugLog('renderRegionChart data:', data)
  safeSetOption(regionChart, {
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
      itemStyle: { borderRadius: 4, borderColor: '#1A1A2E', borderWidth: 3 },
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
  }, 'regionChart')
}

const renderCompletionChart = () => {
  if (!completionChart) return
  const data = completionRate.value || {
    totalCartridges: 0, completedCartridges: 0, rate: 0,
    playingCount: 0, unstartedCount: 0, shelvedCount: 0
  }
  const pieData = [
    { value: data.completedCartridges || 0, name: '已通关', itemStyle: { color: '#6BCB77' } },
    { value: data.playingCount || 0, name: '进行中', itemStyle: { color: '#FFD93D' } },
    { value: data.unstartedCount || 0, name: '未开始', itemStyle: { color: '#00F0FF' } },
    { value: data.shelvedCount || 0, name: '搁置', itemStyle: { color: '#FF8C42' } }
  ]
  debugLog('renderCompletionChart data:', data)
  safeSetOption(completionChart, {
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
      text: `${(data.rate || 0).toFixed(1)}%`,
      subtext: '通关率',
      left: 'center',
      top: '38%',
      textAlign: 'center',
      textStyle: { color: '#FFD93D', fontFamily: 'Press Start 2P', fontSize: 28, fontWeight: 'bold' },
      subtextStyle: { color: '#E8E8E8', fontFamily: 'Press Start 2P', fontSize: 10 }
    },
    series: [{
      type: 'pie',
      radius: ['55%', '80%'],
      center: ['50%', '45%'],
      avoidLabelOverlap: false,
      itemStyle: { borderRadius: 6, borderColor: '#1A1A2E', borderWidth: 3 },
      label: { show: false },
      emphasis: {
        label: { show: true, fontSize: 14, fontWeight: 'bold', color: '#FFD93D', fontFamily: 'Press Start 2P' },
        itemStyle: { shadowBlur: 20, shadowOffsetX: 0, shadowColor: 'rgba(0, 0, 0, 0.5)' }
      },
      data: pieData,
      animationDuration: 1500,
      animationEasing: 'bounceOut'
    }]
  }, 'completionChart')
}

const renderAllCharts = () => {
  nextTick(() => {
    try {
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
      debugLog('🎯 renderAllCharts() completed')
    } catch (e) {
      console.error('[Statistics] ❌ renderAllCharts failed:', e)
    }
  })
}

const initAndRender = async () => {
  await nextTick()
  await nextTick()
  debugLog('initAndRender: DOM should be ready now')
  initCharts()
  renderAllCharts()
}

const loadData = async () => {
  loading.value = true
  debugLog('loadData: starting API requests...')
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

    debugLog('loadData: API responses received', {
      overviewCode: ovRes.code,
      overviewData: ovRes.data,
      annualCode: annualRes.code,
      annualLen: (annualRes.data || []).length,
      platLen: (platRes.data || []).length,
      ratingLen: (ratingRes.data || []).length,
      playtimeLen: (playtimeRes.data || []).length,
      regionData: regionRes.data,
      compData: compRes.data
    })

    if (ovRes.code === 0) overview.value = ovRes.data
    if (annualRes.code === 0) monthlyData.value = annualRes.data || []
    if (platRes.code === 0) platformData.value = platRes.data || []
    if (pubRes.code === 0) publisherData.value = pubRes.data || []
    if (condRes.code === 0) conditionData.value = condRes.data || []
    if (ratingRes.code === 0) ratingData.value = ratingRes.data || []
    if (playtimeRes.code === 0) playTimeTopData.value = playtimeRes.data || []
    if (diffRes.code === 0) difficultyData.value = diffRes.data || []
    if (valueRes.code === 0) valueTrendData.value = valueRes.data || []
    if (regionRes.code === 0) regionData.value = regionRes.data || []
    if (compRes.code === 0) completionRate.value = compRes.data || null
  } catch (e) {
    console.error('[Statistics] ❌ loadData API error:', e)
  } finally {
    loading.value = false
    debugLog('loadData: loading=false, will trigger initAndRender via watcher')
  }
}

const handleResize = () => {
  const all = [
    annualChart, platformChart, publisherChart, conditionChart,
    ratingChart, playTimeChart, difficultyChart, valueTrendChart,
    regionChart, completionChart
  ]
  all.forEach(inst => {
    try { inst?.resize() } catch (e) { /* noop */ }
  })
}

watch(loading, (newVal, oldVal) => {
  debugLog(`watch loading: ${oldVal} → ${newVal}`)
  if (oldVal === true && newVal === false) {
    initAndRender()
  }
})

watch(selectedYear, loadData)

onMounted(() => {
  debugLog('onMounted: initializing...')
  initYears()
  loadData()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  debugLog('onUnmounted: cleaning up')
  window.removeEventListener('resize', handleResize)
  disposeAllCharts()
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
