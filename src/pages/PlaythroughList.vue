<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { playthroughApi } from '@/api'
import type { Playthrough } from '@/types'

const router = useRouter()
const loading = ref(false)
const page = ref(1)
const pageSize = ref(9)
const total = ref(0)
const items = ref<Playthrough[]>([])
const deleteConfirmId = ref<number | null>(null)
const groupByYear = ref(false)

const filters = reactive({ year: '', difficulty: '' })

const difficultyOptions = [
  { value: '', label: '全部难度' },
  { value: '1', label: '★ 简单' },
  { value: '2', label: '★★ 较易' },
  { value: '3', label: '★★★ 普通' },
  { value: '4', label: '★★★★ 较难' },
  { value: '5', label: '★★★★★ 困难' }
]

const yearOptions = computed(() => {
  const years = new Set(items.value.map(p => new Date(p.completionDate).getFullYear().toString()))
  return Array.from(years).sort((a, b) => Number(b) - Number(a))
})

const totalPages = computed(() => Math.ceil(total.value / pageSize.value))
const totalPlayTime = computed(() => items.value.reduce((sum, p) => sum + p.playTimeHours, 0))

const groupedItems = computed(() => {
  if (!groupByYear.value) return [{ year: '', data: items.value }]
  const groups: Record<string, Playthrough[]> = {}
  items.value.forEach(p => {
    const y = new Date(p.completionDate).getFullYear().toString()
    if (!groups[y]) groups[y] = []
    groups[y].push(p)
  })
  return Object.entries(groups).map(([year, data]) => ({ year, data })).sort((a, b) => Number(b.year) - Number(a.year))
})

const loadList = async () => {
  loading.value = true
  try {
    const params = { page: page.value, pageSize: pageSize.value, ...filters }
    Object.keys(params).forEach(k => !params[k as keyof typeof params] && delete params[k as keyof typeof params])
    const res = await playthroughApi.getList(params)
    items.value = res.data.items
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

const handleFilterChange = () => { page.value = 1; loadList() }

const handleDelete = async (id: number) => {
  try {
    await playthroughApi.delete(id)
    deleteConfirmId.value = null
    if (items.value.length === 1 && page.value > 1) page.value--
    loadList()
  } catch (e) {
    console.error('删除失败', e)
  }
}

const goToPage = (p: number) => { if (p >= 1 && p <= totalPages.value) { page.value = p; loadList() } }

const visiblePages = computed(() => {
  const pages: number[] = []
  const total = totalPages.value, current = page.value
  if (total <= 7) for (let i = 1; i <= total; i++) pages.push(i)
  else if (current <= 4) { for (let i = 1; i <= 5; i++) pages.push(i); pages.push(-1, total) }
  else if (current >= total - 3) { pages.push(1, -1); for (let i = total - 4; i <= total; i++) pages.push(i) }
  else pages.push(1, -1, current - 1, current, current + 1, -1, total)
  return pages
})

const formatDate = (d: string) => {
  const dt = new Date(d)
  return `${dt.getFullYear()}.${String(dt.getMonth() + 1).padStart(2, '0')}.${String(dt.getDate()).padStart(2, '0')}`
}

const endingBadgeClass = (t: string) => {
  const map: Record<string, string> = { '真结局': 'pixel-badge-success', '好结局': 'pixel-badge-success', '坏结局': 'pixel-badge-danger', '隐藏结局': 'pixel-badge-warning' }
  return map[t] || ''
}

onMounted(loadList)
</script>

<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
      <div>
        <h1 class="text-bright-yellow glow-yellow">通关记录中心</h1>
        <p class="text-text-secondary">共 {{ total }} 条通关记录</p>
      </div>
      <button class="pixel-btn pixel-btn-primary" @click="router.push('/playthroughs/new')">➕ 添加通关记录</button>
    </div>

    <div class="pixel-card p-4">
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div>
          <label class="block text-sm text-text-secondary mb-1 pixel-font">年份</label>
          <select v-model="filters.year" class="pixel-input w-full" @change="handleFilterChange">
            <option value="">全部年份</option>
            <option v-for="y in yearOptions" :key="y" :value="y">{{ y }}年</option>
          </select>
        </div>
        <div>
          <label class="block text-sm text-text-secondary mb-1 pixel-font">难度</label>
          <select v-model="filters.difficulty" class="pixel-input w-full" @change="handleFilterChange">
            <option v-for="d in difficultyOptions" :key="d.value" :value="d.value">{{ d.label }}</option>
          </select>
        </div>
        <div class="flex items-end">
          <label class="pixel-btn flex items-center gap-2 cursor-pointer !py-3">
            <input type="checkbox" v-model="groupByYear" class="w-4 h-4" />
            <span>按年份分组</span>
          </label>
        </div>
      </div>
    </div>

    <div v-if="loading" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div v-for="i in 6" :key="i" class="pixel-card p-4 opacity-50">
        <div class="h-4 bg-dark-bg-3 w-3/4 animate-pulse mb-4"></div>
        <div class="h-3 bg-dark-bg-3 w-1/2 animate-pulse mb-2"></div>
        <div class="h-3 bg-dark-bg-3 w-2/3 animate-pulse"></div>
      </div>
    </div>

    <div v-else-if="items.length === 0" class="text-center py-16">
      <div class="text-6xl mb-4">🏆</div>
      <h3 class="text-bright-yellow mb-2">暂无通关记录</h3>
      <p class="text-text-secondary mb-6">点击右上角按钮添加你的第一条通关记录</p>
      <button class="pixel-btn pixel-btn-primary" @click="router.push('/playthroughs/new')">添加第一条记录</button>
    </div>

    <div v-else class="space-y-8">
      <template v-for="group in groupedItems" :key="group.year">
        <h2 v-if="groupByYear && group.year" class="text-neon-blue pixel-font text-lg border-b-2 border-neon-blue pb-2">
          📅 {{ group.year }}年
        </h2>
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <div v-for="item in group.data" :key="item.id" class="pixel-card p-4 relative group">
            <div class="absolute top-2 right-2 flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity z-10">
              <button class="pixel-btn !p-1.5 !text-xs" title="编辑" @click="router.push(`/cartridges/${item.cartridgeId}`)">✏️</button>
              <button class="pixel-btn pixel-btn-danger !p-1.5 !text-xs" title="删除" @click="deleteConfirmId = item.id">🗑️</button>
            </div>
            <div class="pr-16">
              <h4 class="pixel-font text-bright-yellow text-sm mb-3 truncate" :title="item.cartridge?.title">
                {{ item.cartridge?.title || '未知游戏' }}
              </h4>
              <div class="flex items-center gap-1 mb-3">
                <span v-for="i in 5" :key="i" class="text-lg" :class="i <= item.difficultyRating ? 'text-pixel-orange' : 'text-text-secondary opacity-30'">★</span>
                <span class="text-xs text-text-secondary ml-2">难度</span>
              </div>
              <div class="space-y-2 text-sm">
                <div class="flex items-center gap-2"><span class="text-text-secondary">📅</span><span>{{ formatDate(item.completionDate) }}</span></div>
                <div class="flex items-center gap-2"><span class="text-text-secondary">⏱️</span><span>{{ item.playTimeHours }} 小时</span></div>
                <div class="flex items-center gap-2 flex-wrap">
                  <span class="pixel-badge !text-[10px]" :class="endingBadgeClass(item.endingType)">{{ item.endingType }}</span>
                  <span v-if="item.multipleEndings" class="pixel-badge pixel-badge-warning !text-[10px]">多结局</span>
                </div>
                <p v-if="item.notes" class="text-text-secondary text-xs mt-2 line-clamp-2" :title="item.notes">💬 {{ item.notes }}</p>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>

    <div v-if="totalPages > 1 && !loading" class="flex justify-center items-center gap-2 pt-4">
      <button class="pixel-btn !py-2 !px-3" :disabled="page === 1" @click="goToPage(page - 1)">◀</button>
      <template v-for="p in visiblePages" :key="p">
        <span v-if="p === -1" class="text-text-secondary px-2">...</span>
        <button v-else class="pixel-btn !py-2 !px-4" :class="{ 'pixel-btn-primary': p === page }" @click="goToPage(p)">{{ p }}</button>
      </template>
      <button class="pixel-btn !py-2 !px-3" :disabled="page === totalPages" @click="goToPage(page + 1)">▶</button>
    </div>

    <div class="pixel-card p-4 flex flex-col md:flex-row justify-around items-center gap-4">
      <div class="text-center">
        <div class="stat-number">{{ total }}</div>
        <div class="text-text-secondary pixel-font text-xs">总通关数</div>
      </div>
      <div class="w-px h-12 bg-neon-blue hidden md:block"></div>
      <div class="text-center">
        <div class="stat-number">{{ totalPlayTime }}</div>
        <div class="text-text-secondary pixel-font text-xs">总游玩时长 (小时)</div>
      </div>
    </div>

    <div v-if="deleteConfirmId !== null" class="fixed inset-0 bg-black/70 flex items-center justify-center z-50" @click.self="deleteConfirmId = null">
      <div class="pixel-card p-6 max-w-sm w-full mx-4">
        <h3 class="text-bright-yellow mb-4">确认删除</h3>
        <p class="text-text-secondary mb-6">确定要删除这条通关记录吗？此操作无法撤销。</p>
        <div class="flex gap-4 justify-end">
          <button class="pixel-btn" @click="deleteConfirmId = null">取消</button>
          <button class="pixel-btn pixel-btn-danger" @click="handleDelete(deleteConfirmId)">确认删除</button>
        </div>
      </div>
    </div>
  </div>
</template>
