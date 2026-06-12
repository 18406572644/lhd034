<script setup lang="ts">
import { ref, reactive, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import { useCartridgeStore } from '@/stores/cartridge'
import { ConditionLabels, ConditionOptions, PlayStatusLabels, PlayStatusBadgeClass } from '@/types'
import type { Cartridge, BatchUpdatePreview } from '@/types'
import { sessionApi, cartridgeApi, batchApi } from '@/api'

const router = useRouter()
const store = useCartridgeStore()

const page = ref(1)
const pageSize = ref(12)
const deleteConfirmId = ref<number | null>(null)
const statusDialog = ref(false)
const statusTarget = ref<Cartridge | null>(null)
const selectedStatus = ref<'unstarted' | 'playing' | 'completed' | 'shelved'>('unstarted')
const newSessionDialog = ref(false)
const newSessionTarget = ref<Cartridge | null>(null)

const newSessionForm = reactive({
  sessionDate: new Date().toISOString().split('T')[0],
  durationMinutes: 60,
  progressPercent: 0,
  notes: ''
})

const filters = reactive({
  search: '',
  platform: '',
  publisher: '',
  condition: '',
  year: '',
  status: ''
})

const selectMode = ref(false)
const selectedIds = ref<number[]>([])
const lastSelectedIndex = ref<number | null>(null)

const batchEditDialog = ref(false)
const batchEditPreview = ref<BatchUpdatePreview | null>(null)
const batchEditLoading = ref(false)
const batchEditMode = ref<'overwrite' | 'append' | 'increment' | 'percentage'>('overwrite')
const batchEditFields = reactive<Record<string, any>>({})
const batchEditFieldOptions = [
  { value: 'platform', label: '平台', type: 'text' },
  { value: 'condition', label: '品相', type: 'select', options: ConditionOptions },
  { value: 'region', label: '区域', type: 'text' },
  { value: 'purchasePrice', label: '购买价格', type: 'number' },
  { value: 'publisher', label: '发行商', type: 'text' },
  { value: 'releaseYear', label: '发行年份', type: 'number' },
  { value: 'status', label: '游玩状态', type: 'select', options: [
    { value: 'unstarted', label: '未开始' },
    { value: 'playing', label: '进行中' },
    { value: 'completed', label: '已通关' },
    { value: 'shelved', label: '搁置' }
  ]},
  { value: 'notes', label: '备注', type: 'textarea' }
]
const selectedBatchField = ref('platform')
const batchEditValue = ref<any>('')
const batchEditIncrement = ref(0)

const batchWishlistDialog = ref(false)
const batchTagsDialog = ref(false)
const batchTagsInput = ref('')
const batchTagsAction = ref<'overwrite' | 'append'>('append')

const batchDeleteConfirm = ref(false)

const conditionBadgeClass = (condition: string) => {
  const map: Record<string, string> = {
    mint: 'pixel-badge-success',
    excellent: 'pixel-badge-success',
    good: '',
    fair: 'pixel-badge-warning',
    poor: 'pixel-badge-danger'
  }
  return map[condition] || ''
}

const totalPages = computed(() => Math.ceil(store.total / pageSize.value))

const hasPlaythroughs = (cartridge: Cartridge) => {
  return cartridge.playthroughs && cartridge.playthroughs.length > 0
}

const getCurrentProgress = (cartridge: Cartridge | null | undefined) => {
  if (!cartridge || !cartridge.sessions || cartridge.sessions.length === 0) return 0
  const percents = cartridge.sessions.map(s => s?.progressPercent ?? 0)
  if (percents.length === 0) return 0
  return Math.max(...percents)
}

const isSelected = (id: number) => selectedIds.value.includes(id)

const toggleSelect = (index: number, cartridge: Cartridge, event: MouseEvent) => {
  if (!selectMode.value) return
  
  event.stopPropagation()
  
  if (event.shiftKey && lastSelectedIndex.value !== null) {
    const start = Math.min(lastSelectedIndex.value, index)
    const end = Math.max(lastSelectedIndex.value, index)
    const newSelected: number[] = []
    for (let i = start; i <= end; i++) {
      const id = store.cartridges[i]?.id
      if (id) newSelected.push(id)
    }
    if (event.ctrlKey || event.metaKey) {
      const existingSet = new Set(selectedIds.value)
      newSelected.forEach(id => existingSet.add(id))
      selectedIds.value = Array.from(existingSet)
    } else {
      selectedIds.value = newSelected
    }
  } else if (event.ctrlKey || event.metaKey) {
    const id = cartridge.id
    if (isSelected(id)) {
      selectedIds.value = selectedIds.value.filter(i => i !== id)
    } else {
      selectedIds.value.push(id)
    }
    lastSelectedIndex.value = index
  } else {
    selectedIds.value = [cartridge.id]
    lastSelectedIndex.value = index
  }
}

const toggleSelectAll = () => {
  if (selectedIds.value.length === store.cartridges.length) {
    selectedIds.value = []
  } else {
    selectedIds.value = store.cartridges.map(c => c.id).filter(Boolean) as number[]
  }
}

const allSelected = computed(() => {
  return store.cartridges.length > 0 && selectedIds.value.length === store.cartridges.length
})

const someSelected = computed(() => {
  return selectedIds.value.length > 0 && selectedIds.value.length < store.cartridges.length
})

const openStatusDialog = (cartridge: Cartridge) => {
  if (!cartridge) return
  statusTarget.value = cartridge
  selectedStatus.value = (cartridge.status || 'unstarted') as typeof selectedStatus.value
  statusDialog.value = true
}

const saveStatus = async () => {
  if (!statusTarget.value) return
  try {
    await cartridgeApi.update(statusTarget.value.id, { status: selectedStatus.value })
    Message.success('状态已更新')
    statusDialog.value = false
    loadList()
  } catch (e) {
    Message.error('更新失败')
  }
}

const setStatus = (key: string) => {
  selectedStatus.value = key as typeof selectedStatus.value
}

const openNewSessionDialog = async (cartridge: Cartridge) => {
  if (!cartridge) return
  newSessionTarget.value = cartridge
  newSessionForm.sessionDate = new Date().toISOString().split('T')[0]
  newSessionForm.durationMinutes = 60
  try {
    const progress = await cartridgeApi.getProgress(cartridge.id)
    newSessionForm.progressPercent = progress?.data?.currentProgress ?? getCurrentProgress(cartridge)
  } catch {
    newSessionForm.progressPercent = getCurrentProgress(cartridge)
  }
  newSessionForm.notes = ''
  newSessionDialog.value = true
}

const saveNewSession = async () => {
  if (!newSessionTarget.value) return
  try {
    await sessionApi.create({
      cartridgeId: newSessionTarget.value.id,
      sessionDate: newSessionForm.sessionDate,
      durationMinutes: newSessionForm.durationMinutes,
      progressPercent: newSessionForm.progressPercent,
      notes: newSessionForm.notes
    })
    Message.success('会话记录已保存')
    newSessionDialog.value = false
    loadList()
  } catch (e) {
    Message.error('保存失败')
  }
}

const loadList = () => {
  const params = {
    page: page.value,
    pageSize: pageSize.value,
    ...filters
  }
  Object.keys(params).forEach(key => {
    if (!params[key as keyof typeof params]) delete params[key as keyof typeof params]
  })
  store.fetchList(params)
}

const handleSearch = () => {
  page.value = 1
  loadList()
}

const handleFilterChange = () => {
  page.value = 1
  loadList()
}

const handleDelete = async (id: number) => {
  if (await store.deleteCartridge(id)) {
    deleteConfirmId.value = null
    if (store.cartridges.length === 0 && page.value > 1) {
      page.value--
    }
    loadList()
  }
}

const goToPage = (p: number) => {
  if (p >= 1 && p <= totalPages.value) {
    page.value = p
    loadList()
  }
}

const visiblePages = computed(() => {
  const pages: number[] = []
  const total = totalPages.value
  const current = page.value

  if (total <= 7) {
    for (let i = 1; i <= total; i++) pages.push(i)
  } else {
    if (current <= 4) {
      for (let i = 1; i <= 5; i++) pages.push(i)
      pages.push(-1, total)
    } else if (current >= total - 3) {
      pages.push(1, -1)
      for (let i = total - 4; i <= total; i++) pages.push(i)
    } else {
      pages.push(1, -1, current - 1, current, current + 1, -1, total)
    }
  }
  return pages
})

const openBatchEdit = () => {
  if (selectedIds.value.length === 0) {
    Message.warning('请先选择要编辑的卡带')
    return
  }
  Object.keys(batchEditFields).forEach(key => {
    delete batchEditFields[key]
  })
  batchEditMode.value = 'overwrite'
  selectedBatchField.value = 'platform'
  batchEditValue.value = ''
  batchEditIncrement.value = 0
  batchEditPreview.value = null
  batchEditDialog.value = true
}

const addBatchEditField = () => {
  const field = selectedBatchField.value
  if (batchEditFields[field]) {
    Message.warning('该字段已添加')
    return
  }
  
  const fieldOption = batchEditFieldOptions.find(f => f.value === field)
  if (fieldOption?.type === 'select') {
    batchEditFields[field] = fieldOption.options?.[0]?.value || ''
  } else if (fieldOption?.type === 'number') {
    batchEditFields[field] = 0
  } else {
    batchEditFields[field] = ''
  }
}

const removeBatchEditField = (field: string) => {
  delete batchEditFields[field]
}

const previewBatchEdit = async () => {
  if (Object.keys(batchEditFields).length === 0) {
    Message.warning('请至少添加一个要修改的字段')
    return
  }
  
  batchEditLoading.value = true
  try {
    const fields: Record<string, any> = {}
    for (const key in batchEditFields) {
      fields[key] = batchEditFields[key]
    }
    
    const res = await batchApi.updatePreview({
      ids: selectedIds.value,
      fields,
      mode: batchEditMode.value,
      increment: batchEditIncrement.value
    })
    
    if (res.code === 0) {
      batchEditPreview.value = res.data
    }
  } catch (e) {
    Message.error('预览失败')
  } finally {
    batchEditLoading.value = false
  }
}

const executeBatchEdit = async () => {
  if (Object.keys(batchEditFields).length === 0) {
    Message.warning('请至少添加一个要修改的字段')
    return
  }
  
  batchEditLoading.value = true
  try {
    const fields: Record<string, any> = {}
    for (const key in batchEditFields) {
      fields[key] = batchEditFields[key]
    }
    
    const res = await batchApi.update({
      ids: selectedIds.value,
      fields,
      mode: batchEditMode.value,
      increment: batchEditIncrement.value
    })
    
    if (res.code === 0) {
      Message.success(`成功更新 ${res.data.updatedCount} 条记录`)
      batchEditDialog.value = false
      selectedIds.value = []
      loadList()
    }
  } catch (e) {
    Message.error('批量更新失败')
  } finally {
    batchEditLoading.value = false
  }
}

const openBatchWishlist = () => {
  if (selectedIds.value.length === 0) {
    Message.warning('请先选择卡带')
    return
  }
  batchWishlistDialog.value = true
}

const executeBatchWishlist = async () => {
  try {
    const res = await batchApi.addToWishlist(selectedIds.value)
    if (res.code === 0) {
      Message.success(`成功添加 ${res.data.addedCount} 张，跳过 ${res.data.skippedCount} 张`)
      batchWishlistDialog.value = false
    }
  } catch (e) {
    Message.error('添加失败')
  }
}

const openBatchTags = () => {
  if (selectedIds.value.length === 0) {
    Message.warning('请先选择卡带')
    return
  }
  batchTagsInput.value = ''
  batchTagsAction.value = 'append'
  batchTagsDialog.value = true
}

const executeBatchTags = async () => {
  const tags = batchTagsInput.value.split(',').map(t => t.trim()).filter(Boolean)
  if (tags.length === 0) {
    Message.warning('请输入标签')
    return
  }
  
  try {
    const res = await batchApi.setTags(selectedIds.value, tags, batchTagsAction.value)
    if (res.code === 0) {
      Message.success(`成功更新 ${res.data.updatedCount} 条记录`)
      batchTagsDialog.value = false
    }
  } catch (e) {
    Message.error('设置标签失败')
  }
}

const openBatchDelete = () => {
  if (selectedIds.value.length === 0) {
    Message.warning('请先选择要删除的卡带')
    return
  }
  batchDeleteConfirm.value = true
}

const executeBatchDelete = async () => {
  try {
    const res = await batchApi.deleteCartridges(selectedIds.value)
    if (res.code === 0) {
      Message.success(`成功删除 ${res.data.deletedCount} 条记录`)
      batchDeleteConfirm.value = false
      selectedIds.value = []
      loadList()
    }
  } catch (e) {
    Message.error('删除失败')
  }
}

watch(page, loadList)

watch(selectMode, (newVal) => {
  if (!newVal) {
    selectedIds.value = []
    lastSelectedIndex.value = null
  }
})

watch(store.cartridges, () => {
  if (selectMode.value) {
    const validIds = new Set(store.cartridges.map(c => c.id))
    selectedIds.value = selectedIds.value.filter(id => validIds.has(id))
  }
})

onMounted(async () => {
  await Promise.all([store.fetchPlatforms(), store.fetchPublishers()])
  loadList()
})
</script>

<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
      <div>
        <h1 class="text-neon-blue glow-blue">卡带档案</h1>
        <p class="text-text-secondary">共 {{ store.total }} 张卡带</p>
      </div>
      <div class="flex gap-3 flex-wrap">
        <button 
          class="pixel-btn" 
          :class="{ 'pixel-btn-primary': selectMode }"
          @click="selectMode = !selectMode"
        >
          {{ selectMode ? '✓ 退出多选' : '☐ 多选模式' }}
        </button>
        <button class="pixel-btn pixel-btn-primary" @click="router.push('/cartridges/new')">
          ➕ 添加卡带
        </button>
      </div>
    </div>

    <div v-if="selectMode && selectedIds.length > 0" class="pixel-card p-4 bg-dark-bg-2 border-bright-yellow">
      <div class="flex flex-wrap items-center justify-between gap-4">
        <div class="flex items-center gap-3">
          <span class="pixel-font text-bright-yellow">
            已选择 <span class="text-neon-blue">{{ selectedIds.length }}</span> 张卡带
          </span>
        </div>
        <div class="flex flex-wrap gap-2">
          <button class="pixel-btn !py-1 !px-3 !text-sm" @click="openBatchEdit">
            ✏️ 批量编辑
          </button>
          <button class="pixel-btn !py-1 !px-3 !text-sm" @click="openBatchWishlist">
            📋 加入待玩
          </button>
          <button class="pixel-btn !py-1 !px-3 !text-sm" @click="openBatchTags">
            🏷️ 设置标签
          </button>
          <button class="pixel-btn pixel-btn-danger !py-1 !px-3 !text-sm" @click="openBatchDelete">
            🗑️ 批量删除
          </button>
        </div>
      </div>
    </div>

    <div class="pixel-card p-4">
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-6 gap-4">
        <div>
          <label class="block text-sm text-text-secondary mb-1 pixel-font">搜索</label>
          <input
            v-model="filters.search"
            type="text"
            class="pixel-input w-full"
            placeholder="游戏标题..."
            @keyup.enter="handleSearch"
          />
        </div>
        <div>
          <label class="block text-sm text-text-secondary mb-1 pixel-font">平台</label>
          <select v-model="filters.platform" class="pixel-input w-full" @change="handleFilterChange">
            <option value="">全部平台</option>
            <option v-for="p in store.platforms" :key="p" :value="p">{{ p }}</option>
          </select>
        </div>
        <div>
          <label class="block text-sm text-text-secondary mb-1 pixel-font">发行商</label>
          <select v-model="filters.publisher" class="pixel-input w-full" @change="handleFilterChange">
            <option value="">全部发行商</option>
            <option v-for="p in store.publishers" :key="p" :value="p">{{ p }}</option>
          </select>
        </div>
        <div>
          <label class="block text-sm text-text-secondary mb-1 pixel-font">品相</label>
          <select v-model="filters.condition" class="pixel-input w-full" @change="handleFilterChange">
            <option value="">全部品相</option>
            <option v-for="c in ConditionOptions" :key="c.value" :value="c.value">{{ c.label }}</option>
          </select>
        </div>
        <div>
          <label class="block text-sm text-text-secondary mb-1 pixel-font">游玩状态</label>
          <select v-model="filters.status" class="pixel-input w-full" @change="handleFilterChange">
            <option value="">全部状态</option>
            <option value="unstarted">未开始</option>
            <option value="playing">进行中</option>
            <option value="completed">已通关</option>
            <option value="shelved">搁置</option>
          </select>
        </div>
        <div>
          <label class="block text-sm text-text-secondary mb-1 pixel-font">年份</label>
          <input
            v-model="filters.year"
            type="text"
            class="pixel-input w-full"
            placeholder="如: 1998"
            @keyup.enter="handleFilterChange"
            @blur="handleFilterChange"
          />
        </div>
      </div>
    </div>

    <div v-if="selectMode" class="flex items-center gap-2">
      <button
        class="pixel-btn !py-1 !px-2 !text-sm"
        @click="toggleSelectAll"
      >
        {{ allSelected ? '☑ 全选' : someSelected ? '☒ 部分选中' : '☐ 全选' }}
      </button>
      <span class="text-sm text-text-secondary">
        点击选择，按住 Shift 连续选择，按住 Ctrl 追加选择
      </span>
    </div>

    <div v-if="store.loading" class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">
      <div v-for="i in 8" :key="i" class="cartridge-case opacity-50">
        <div class="cartridge-label">
          <div class="h-4 bg-dark-bg-3 w-3/4 animate-pulse"></div>
        </div>
      </div>
    </div>

    <div v-else-if="store.cartridges.length === 0" class="text-center py-16">
      <div class="text-6xl mb-4">🎮</div>
      <h3 class="text-bright-yellow mb-2">暂无卡带</h3>
      <p class="text-text-secondary mb-6">点击右上角按钮添加你的第一张卡带</p>
      <button class="pixel-btn pixel-btn-primary" @click="router.push('/cartridges/new')">
        添加第一张卡带
      </button>
    </div>

    <div v-else class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">
      <div
        v-for="(cartridge, index) in store.cartridges"
        :key="cartridge?.id ?? Math.random()"
        class="relative group"
        :class="{ 'ring-2 ring-bright-yellow ring-offset-2 ring-offset-dark-bg rounded-lg': isSelected(cartridge?.id ?? 0) }"
        @click="selectMode ? toggleSelect(index, cartridge, $event as MouseEvent) : router.push(`/cartridges/${cartridge?.id}`)"
      >
        <div v-if="selectMode" class="absolute -top-2 -left-2 z-20">
          <div 
            class="w-6 h-6 border-2 flex items-center justify-center bg-dark-bg"
            :class="isSelected(cartridge?.id ?? 0) ? 'border-bright-yellow bg-bright-yellow' : 'border-neon-blue'"
          >
            <span v-if="isSelected(cartridge?.id ?? 0)" class="text-dark-bg text-sm">✓</span>
          </div>
        </div>

        <div class="cartridge-case" :class="{ 'cursor-pointer': !selectMode }">
          <div class="cartridge-label">
            <img v-if="cartridge?.coverImage" :src="cartridge.coverImage" :alt="cartridge?.title || '卡带封面'" />
            <div v-else class="flex flex-col items-center justify-center h-full">
              <div class="text-xs pixel-font text-bright-yellow leading-relaxed">{{ cartridge?.title || '未知游戏' }}</div>
              <div class="text-xs text-neon-blue mt-1">{{ cartridge?.platform || '—' }}</div>
            </div>
          </div>
        </div>

        <div v-if="!selectMode" class="absolute top-2 right-2 flex flex-col gap-1 opacity-0 group-hover:opacity-100 transition-opacity z-10">
          <button
            class="pixel-btn !p-2 !text-xs"
            title="记录游玩会话"
            @click.stop="openNewSessionDialog(cartridge)"
          >
            ➕
          </button>
          <button
            class="pixel-btn !p-2 !text-xs"
            title="修改状态"
            @click.stop="openStatusDialog(cartridge)"
          >
            🏷️
          </button>
          <button
            class="pixel-btn !p-2 !text-xs"
            title="编辑"
            @click.stop="router.push(`/cartridges/${cartridge?.id}/edit`)"
          >
            ✏️
          </button>
          <button
            class="pixel-btn pixel-btn-danger !p-2 !text-xs"
            title="删除"
            @click.stop="deleteConfirmId = cartridge?.id ?? null"
          >
            🗑️
          </button>
        </div>

        <div class="mt-3 space-y-2">
          <div class="flex items-start justify-between gap-2">
            <h4 class="pixel-font text-bright-yellow text-xs leading-tight truncate">{{ cartridge?.title || '未知游戏' }}</h4>
            <span
              class="pixel-badge !text-[8px] shrink-0"
              :class="PlayStatusBadgeClass[cartridge?.status || 'unstarted']"
            >
              {{ PlayStatusLabels[cartridge?.status || 'unstarted'] }}
            </span>
          </div>
          <div
            v-if="cartridge?.status === 'playing' && getCurrentProgress(cartridge) > 0"
            class="pixel-progress !h-2"
          >
            <div
              class="pixel-progress-bar"
              :style="{ width: `${getCurrentProgress(cartridge)}%` }"
            ></div>
          </div>
          <div class="flex items-center gap-2 flex-wrap">
            <span class="pixel-badge !text-[8px]">{{ cartridge?.platform || '—' }}</span>
            <span class="pixel-badge !text-[8px]" :class="conditionBadgeClass(cartridge?.condition || 'good')">
              {{ ConditionLabels[cartridge?.condition || 'good'] }}
            </span>
          </div>
          <div class="flex items-center justify-between text-sm text-text-secondary">
            <span>📅 {{ cartridge?.releaseYear ?? '—' }}</span>
            <span>💰 ¥{{ cartridge?.purchasePrice ?? 0 }}</span>
          </div>
        </div>
      </div>
    </div>

    <div v-if="totalPages > 1 && !store.loading" class="flex justify-center items-center gap-2 pt-4">
      <button
        class="pixel-btn !py-2 !px-3"
        :disabled="page === 1"
        @click="goToPage(page - 1)"
      >
        ◀
      </button>
      <template v-for="p in visiblePages" :key="p">
        <span v-if="p === -1" class="text-text-secondary px-2">...</span>
        <button
          v-else
          class="pixel-btn !py-2 !px-4"
          :class="{ 'pixel-btn-primary': p === page }"
          @click="goToPage(p)"
        >
          {{ p }}
        </button>
      </template>
      <button
        class="pixel-btn !py-2 !px-3"
        :disabled="page === totalPages"
        @click="goToPage(page + 1)"
      >
        ▶
      </button>
    </div>

    <div
      v-if="deleteConfirmId !== null"
      class="fixed inset-0 bg-black/70 flex items-center justify-center z-50"
      @click.self="deleteConfirmId = null"
    >
      <div class="pixel-card p-6 max-w-sm w-full mx-4">
        <h3 class="text-bright-yellow mb-4">确认删除</h3>
        <p class="text-text-secondary mb-6">确定要删除这张卡带吗？此操作无法撤销。</p>
        <div class="flex gap-4 justify-end">
          <button class="pixel-btn" @click="deleteConfirmId = null">取消</button>
          <button class="pixel-btn pixel-btn-danger" @click="handleDelete(deleteConfirmId)">
            确认删除
          </button>
        </div>
      </div>
    </div>

    <div
      v-if="statusDialog"
      class="fixed inset-0 bg-black/70 flex items-center justify-center z-50"
      @click.self="statusDialog = false"
    >
      <div class="pixel-card p-6 max-w-md w-full mx-4">
        <h3 class="text-bright-yellow mb-4">修改游玩状态</h3>
        <p class="text-text-secondary mb-4">
          当前卡带: <span class="text-neon-blue">{{ statusTarget?.title }}</span>
        </p>
        <div class="space-y-2 mb-6">
          <label
            v-for="(label, key) in PlayStatusLabels"
            :key="key"
            class="flex items-center gap-3 cursor-pointer p-3 pixel-border"
            :class="{ '!border-bright-yellow': selectedStatus === key }"
            @click="setStatus(key)"
          >
            <div
              class="w-4 h-4 border-2"
              :class="selectedStatus === key ? 'bg-bright-yellow border-bright-yellow' : 'border-neon-blue'"
            ></div>
            <span class="pixel-badge" :class="PlayStatusBadgeClass[key]">{{ label }}</span>
          </label>
        </div>
        <div class="flex gap-4 justify-end">
          <button class="pixel-btn" @click="statusDialog = false">取消</button>
          <button class="pixel-btn pixel-btn-primary" @click="saveStatus">保存</button>
        </div>
      </div>
    </div>

    <div
      v-if="newSessionDialog"
      class="fixed inset-0 bg-black/70 flex items-center justify-center z-50"
      @click.self="newSessionDialog = false"
    >
      <div class="pixel-card p-6 max-w-md w-full mx-4">
        <h3 class="text-bright-yellow mb-4">记录游玩会话</h3>
        <p class="text-text-secondary mb-4">
          当前卡带: <span class="text-neon-blue">{{ newSessionTarget?.title }}</span>
        </p>
        <div class="space-y-4">
          <div>
            <label class="block text-sm text-text-secondary mb-1 pixel-font">游玩日期</label>
            <input
              v-model="newSessionForm.sessionDate"
              type="date"
              class="pixel-input w-full"
            />
          </div>
          <div>
            <label class="block text-sm text-text-secondary mb-1 pixel-font">游玩时长（分钟）</label>
            <input
              v-model.number="newSessionForm.durationMinutes"
              type="number"
              min="0"
              class="pixel-input w-full"
              placeholder="如: 90"
            />
          </div>
          <div>
            <label class="block text-sm text-text-secondary mb-1 pixel-font">当前进度百分比</label>
            <input
              v-model.number="newSessionForm.progressPercent"
              type="range"
              min="0"
              max="100"
              class="w-full"
            />
            <div class="text-center pixel-font text-neon-blue mt-1">
              {{ newSessionForm.progressPercent }}%
            </div>
          </div>
          <div>
            <label class="block text-sm text-text-secondary mb-1 pixel-font">备注</label>
            <textarea
              v-model="newSessionForm.notes"
              class="pixel-input w-full"
              rows="3"
              placeholder="记录本次游玩的内容、感受..."
            ></textarea>
          </div>
        </div>
        <div class="flex gap-4 justify-end mt-6">
          <button class="pixel-btn" @click="newSessionDialog = false">取消</button>
          <button class="pixel-btn pixel-btn-primary" @click="saveNewSession">保存</button>
        </div>
      </div>
    </div>

    <div
      v-if="batchEditDialog"
      class="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4"
      @click.self="batchEditDialog = false"
    >
      <div class="pixel-card p-6 max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        <h3 class="text-bright-yellow mb-4">批量编辑字段</h3>
        <p class="text-text-secondary mb-4">
          已选择 <span class="text-neon-blue">{{ selectedIds.length }}</span> 张卡带
        </p>

        <div class="space-y-4 mb-6">
          <div>
            <label class="block text-sm text-text-secondary mb-2 pixel-font">修改模式</label>
            <div class="flex flex-wrap gap-2">
              <button
                class="pixel-btn !py-1 !px-3 !text-sm"
                :class="{ 'pixel-btn-primary': batchEditMode === 'overwrite' }"
                @click="batchEditMode = 'overwrite'"
              >
                覆盖
              </button>
              <button
                class="pixel-btn !py-1 !px-3 !text-sm"
                :class="{ 'pixel-btn-primary': batchEditMode === 'append' }"
                @click="batchEditMode = 'append'"
              >
                追加
              </button>
              <button
                class="pixel-btn !py-1 !px-3 !text-sm"
                :class="{ 'pixel-btn-primary': batchEditMode === 'increment' }"
                @click="batchEditMode = 'increment'"
              >
                增量修改
              </button>
              <button
                class="pixel-btn !py-1 !px-3 !text-sm"
                :class="{ 'pixel-btn-primary': batchEditMode === 'percentage' }"
                @click="batchEditMode = 'percentage'"
              >
                百分比
              </button>
            </div>
          </div>

          <div v-if="batchEditMode === 'increment' || batchEditMode === 'percentage'" class="pixel-border p-3">
            <label class="block text-sm text-text-secondary mb-1 pixel-font">
              {{ batchEditMode === 'increment' ? '增量值' : '百分比值' }}
            </label>
            <input
              v-model.number="batchEditIncrement"
              type="number"
              class="pixel-input w-full"
              :placeholder="batchEditMode === 'increment' ? '输入增量值，如 10' : '输入百分比，如 10 表示涨价10%'"
            />
            <p class="text-xs text-text-secondary mt-1">
              {{ batchEditMode === 'percentage' ? '正数为涨价，负数为降价' : '正数为增加，负数为减少' }}
            </p>
          </div>

          <div>
            <label class="block text-sm text-text-secondary mb-2 pixel-font">添加字段</label>
            <div class="flex gap-2">
              <select v-model="selectedBatchField" class="pixel-input flex-1">
                <option v-for="f in batchEditFieldOptions" :key="f.value" :value="f.value">
                  {{ f.label }}
                </option>
              </select>
              <button class="pixel-btn pixel-btn-primary" @click="addBatchEditField">
                ➕ 添加
              </button>
            </div>
          </div>

          <div v-if="Object.keys(batchEditFields).length > 0" class="space-y-3">
            <label class="block text-sm text-text-secondary pixel-font">已选字段</label>
            <div
              v-for="fieldOption in batchEditFieldOptions.filter(f => batchEditFields[f.value] !== undefined)"
              :key="fieldOption.value"
              class="pixel-border p-3 flex items-center gap-3"
            >
              <span class="pixel-badge shrink-0">{{ fieldOption.label }}</span>
              <div class="flex-1">
                <select 
                  v-if="fieldOption.type === 'select'"
                  v-model="batchEditFields[fieldOption.value]"
                  class="pixel-input w-full"
                >
                  <option v-for="opt in fieldOption.options" :key="opt.value" :value="opt.value">
                    {{ opt.label }}
                  </option>
                </select>
                <textarea
                  v-else-if="fieldOption.type === 'textarea'"
                  v-model="batchEditFields[fieldOption.value]"
                  class="pixel-input w-full"
                  rows="2"
                ></textarea>
                <input
                  v-else
                  v-model="batchEditFields[fieldOption.value]"
                  :type="fieldOption.type"
                  class="pixel-input w-full"
                />
              </div>
              <button 
                class="pixel-btn pixel-btn-danger !p-1 !text-xs"
                @click="removeBatchEditField(fieldOption.value)"
              >
                ✕
              </button>
            </div>
          </div>
        </div>

        <div class="flex gap-3 justify-end">
          <button class="pixel-btn" @click="previewBatchEdit" :disabled="batchEditLoading">
            🔍 预览
          </button>
          <button class="pixel-btn" @click="batchEditDialog = false">取消</button>
          <button class="pixel-btn pixel-btn-primary" @click="executeBatchEdit" :disabled="batchEditLoading">
            确认修改
          </button>
        </div>

        <div v-if="batchEditPreview" class="mt-6 pt-6 border-t border-dark-bg-3">
          <h4 class="text-neon-blue mb-4 pixel-font">预览结果</h4>
          <p class="text-text-secondary mb-4">
            将修改 <span class="text-bright-yellow">{{ batchEditPreview.totalRecords }}</span> 条记录
          </p>
          
          <div class="space-y-3 mb-4">
            <div
              v-for="change in batchEditPreview.fieldChanges"
              :key="change.field"
              class="pixel-border p-3"
            >
              <div class="flex justify-between items-center mb-2">
                <span class="pixel-font text-bright-yellow">{{ change.label }}</span>
                <span class="text-sm text-text-secondary">
                  {{ change.changeCount }} 条有变化
                </span>
              </div>
              <div class="flex items-center gap-3 text-sm">
                <span class="text-red-400 line-through">{{ change.beforeValue ?? '空' }}</span>
                <span class="text-neon-blue">→</span>
                <span class="text-green-400">{{ change.afterValue ?? '空' }}</span>
              </div>
            </div>
          </div>

          <div v-if="batchEditPreview.sampleBefore.length > 0" class="mt-4">
            <h5 class="text-sm text-text-secondary mb-2 pixel-font">示例对比 (前3条)</h5>
            <div class="space-y-2">
              <div
                v-for="(before, idx) in batchEditPreview.sampleBefore"
                :key="before.id"
                class="pixel-border p-2 text-xs"
              >
                <div class="text-bright-yellow mb-1">{{ before.title }}</div>
                <div class="grid grid-cols-2 gap-2">
                  <div>
                    <span class="text-text-secondary">修改前:</span>
                    <span
                      v-for="change in batchEditPreview.fieldChanges"
                      :key="change.field"
                      class="text-red-400 ml-2"
                    >
                      {{ (before as any)[change.field] ?? '空' }}
                    </span>
                  </div>
                  <div>
                    <span class="text-text-secondary">修改后:</span>
                    <span
                      v-for="change in batchEditPreview.fieldChanges"
                      :key="change.field"
                      class="text-green-400 ml-2"
                    >
                      {{ batchEditPreview.sampleAfter[idx]?.[change.field] ?? '空' }}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div
      v-if="batchWishlistDialog"
      class="fixed inset-0 bg-black/70 flex items-center justify-center z-50"
      @click.self="batchWishlistDialog = false"
    >
      <div class="pixel-card p-6 max-w-sm w-full mx-4">
        <h3 class="text-bright-yellow mb-4">批量加入待玩清单</h3>
        <p class="text-text-secondary mb-6">
          确定将 <span class="text-neon-blue">{{ selectedIds.length }}</span> 张卡带加入待玩清单吗？
        </p>
        <div class="flex gap-4 justify-end">
          <button class="pixel-btn" @click="batchWishlistDialog = false">取消</button>
          <button class="pixel-btn pixel-btn-primary" @click="executeBatchWishlist">确认添加</button>
        </div>
      </div>
    </div>

    <div
      v-if="batchTagsDialog"
      class="fixed inset-0 bg-black/70 flex items-center justify-center z-50"
      @click.self="batchTagsDialog = false"
    >
      <div class="pixel-card p-6 max-w-md w-full mx-4">
        <h3 class="text-bright-yellow mb-4">批量设置标签</h3>
        <p class="text-text-secondary mb-4">
          已选择 <span class="text-neon-blue">{{ selectedIds.length }}</span> 张卡带
        </p>
        
        <div class="space-y-4 mb-6">
          <div>
            <label class="block text-sm text-text-secondary mb-2 pixel-font">操作方式</label>
            <div class="flex gap-2">
              <button
                class="pixel-btn !py-1 !px-3 !text-sm"
                :class="{ 'pixel-btn-primary': batchTagsAction === 'append' }"
                @click="batchTagsAction = 'append'"
              >
                追加标签
              </button>
              <button
                class="pixel-btn !py-1 !px-3 !text-sm"
                :class="{ 'pixel-btn-primary': batchTagsAction === 'overwrite' }"
                @click="batchTagsAction = 'overwrite'"
              >
                覆盖标签
              </button>
            </div>
          </div>
          
          <div>
            <label class="block text-sm text-text-secondary mb-1 pixel-font">标签（用逗号分隔）</label>
            <input
              v-model="batchTagsInput"
              type="text"
              class="pixel-input w-full"
              placeholder="如: 神作, 必玩, 怀旧"
            />
          </div>
        </div>
        
        <div class="flex gap-4 justify-end">
          <button class="pixel-btn" @click="batchTagsDialog = false">取消</button>
          <button class="pixel-btn pixel-btn-primary" @click="executeBatchTags">确认</button>
        </div>
      </div>
    </div>

    <div
      v-if="batchDeleteConfirm"
      class="fixed inset-0 bg-black/70 flex items-center justify-center z-50"
      @click.self="batchDeleteConfirm = false"
    >
      <div class="pixel-card p-6 max-w-sm w-full mx-4">
        <h3 class="text-bright-yellow mb-4">确认批量删除</h3>
        <p class="text-text-secondary mb-2">
          确定要删除选中的 <span class="text-neon-blue">{{ selectedIds.length }}</span> 张卡带吗？
        </p>
        <p class="text-red-400 text-sm mb-6">此操作无法撤销，请谨慎操作！</p>
        <div class="flex gap-4 justify-end">
          <button class="pixel-btn" @click="batchDeleteConfirm = false">取消</button>
          <button class="pixel-btn pixel-btn-danger" @click="executeBatchDelete">
            确认删除
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
