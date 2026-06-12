<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { wishlistApi, cartridgeApi } from '@/api'
import { PriorityLabels, PriorityOptions } from '@/types'
import type { WishlistItem, Cartridge } from '@/types'

const router = useRouter()
const loading = ref(false)
const items = ref<WishlistItem[]>([])
const cartridges = ref<Cartridge[]>([])
const deleteConfirmId = ref<number | null>(null)

const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const form = reactive({
  cartridgeId: 0,
  priority: 'medium' as 'high' | 'medium' | 'low',
  plannedStartDate: '',
  tags: '' as string | string[],
  notes: ''
})

const loadData = async () => {
  loading.value = true
  try {
    const [wishRes, cartRes] = await Promise.all([
      wishlistApi.getList(),
      cartridgeApi.getList({ pageSize: 100 })
    ])
    items.value = wishRes.data
    cartridges.value = cartRes.data.items
  } finally {
    loading.value = false
  }
}

const formatDate = (d: string) => {
  const dt = new Date(d)
  return `${dt.getFullYear()}.${String(dt.getMonth() + 1).padStart(2, '0')}.${String(dt.getDate()).padStart(2, '0')}`
}

const priorityBadgeClass = (p: string) => {
  const map: Record<string, string> = { high: 'pixel-badge-danger', medium: 'pixel-badge-warning', low: '' }
  return map[p] || ''
}

const groupedItems = computed(() => {
  const groups: Record<string, WishlistItem[]> = { high: [], medium: [], low: [] }
  items.value.forEach(item => {
    if (groups[item.priority]) groups[item.priority].push(item)
  })
  return [
    { key: 'high', label: '🔴 高优先级', data: groups.high },
    { key: 'medium', label: '🟡 中优先级', data: groups.medium },
    { key: 'low', label: '🔵 低优先级', data: groups.low }
  ]
})

const availableCartridges = computed(() => {
  const usedIds = items.value.map(i => i.cartridgeId)
  return cartridges.value.filter(c => !usedIds.includes(c.id) || c.id === form.cartridgeId)
})

const openDialog = (item?: WishlistItem) => {
  editingId.value = item?.id || null
  if (item) {
    form.cartridgeId = item.cartridgeId
    form.priority = item.priority
    form.plannedStartDate = item.plannedStartDate
    form.tags = item.tags.join(', ')
    form.notes = item.notes
  } else {
    form.cartridgeId = availableCartridges.value[0]?.id || 0
    form.priority = 'medium'
    form.plannedStartDate = new Date().toISOString().split('T')[0]
    form.tags = ''
    form.notes = ''
  }
  dialogVisible.value = true
}

const saveItem = async () => {
  try {
    const tags = Array.isArray(form.tags)
      ? form.tags
      : (form.tags as string).split(',').map(t => t.trim()).filter(Boolean)
    const data = { ...form, tags }
    if (editingId.value) {
      await wishlistApi.update(editingId.value, data)
    } else {
      await wishlistApi.create(data)
    }
    dialogVisible.value = false
    loadData()
  } catch (e) {
    console.error('保存失败', e)
  }
}

const handleDelete = async (id: number) => {
  try {
    await wishlistApi.delete(id)
    deleteConfirmId.value = null
    loadData()
  } catch (e) {
    console.error('删除失败', e)
  }
}

const startPlaying = (_item: WishlistItem) => router.push('/progress')

const getCartridge = (id: number) => cartridges.value.find(c => c.id === id)

onMounted(loadData)
</script>

<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
      <div>
        <h1 class="text-neon-blue glow-blue">待玩清单</h1>
        <p class="text-text-secondary">共 {{ items.length }} 款游戏等待游玩</p>
      </div>
      <button class="pixel-btn pixel-btn-primary" @click="openDialog()">
        ➕ 添加到待玩
      </button>
    </div>

    <div v-if="loading" class="space-y-6">
      <div v-for="g in 3" :key="g" class="space-y-3">
        <div class="h-4 bg-dark-bg-3 w-32 animate-pulse"></div>
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <div v-for="i in 2" :key="i" class="pixel-card p-4 opacity-50">
            <div class="h-4 bg-dark-bg-3 w-3/4 animate-pulse mb-4"></div>
            <div class="h-3 bg-dark-bg-3 w-1/2 animate-pulse mb-2"></div>
            <div class="h-3 bg-dark-bg-3 w-2/3 animate-pulse"></div>
          </div>
        </div>
      </div>
    </div>

    <div v-else-if="items.length === 0" class="text-center py-16">
      <div class="text-6xl mb-4">📋</div>
      <h3 class="text-bright-yellow mb-2">待玩清单是空的</h3>
      <p class="text-text-secondary mb-6">添加你想玩的游戏，开始规划你的游戏之旅！</p>
      <button class="pixel-btn pixel-btn-primary" @click="openDialog()">
        添加第一款游戏
      </button>
    </div>

    <div v-else class="space-y-8">
      <template v-for="group in groupedItems" :key="group.key">
        <div v-if="group.data.length > 0">
          <h2 class="pixel-font text-lg mb-4 border-b-2 border-neon-blue pb-2">
            <span :class="{ 'text-pixel-red': group.key === 'high', 'text-bright-yellow': group.key === 'medium', 'text-neon-blue': group.key === 'low' }">
              {{ group.label }}
            </span>
            <span class="text-text-secondary text-sm ml-2">({{ group.data.length }})</span>
          </h2>
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <div v-for="item in group.data" :key="item.id" class="pixel-card p-4 relative group">
              <div class="absolute top-2 right-2 flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity z-10">
                <button class="pixel-btn !p-1.5 !text-xs" title="编辑" @click="openDialog(item)">✏️</button>
                <button class="pixel-btn pixel-btn-danger !p-1.5 !text-xs" title="删除" @click="deleteConfirmId = item.id">🗑️</button>
              </div>
              <div class="pr-16">
                <h4 class="pixel-font text-bright-yellow text-sm mb-2 truncate">
                  {{ getCartridge(item.cartridgeId)?.title || '未知游戏' }}
                </h4>
                <div class="flex items-center gap-2 mb-3">
                  <span class="pixel-badge !text-[10px]">
                    {{ getCartridge(item.cartridgeId)?.platform || '未知平台' }}
                  </span>
                  <span class="pixel-badge !text-[10px]" :class="priorityBadgeClass(item.priority)">
                    {{ PriorityLabels[item.priority] }}优先级
                  </span>
                </div>
                <div class="space-y-1 text-sm">
                  <div class="flex items-center gap-2">
                    <span class="text-text-secondary">📅</span>
                    <span>预计开始: {{ formatDate(item.plannedStartDate) }}</span>
                  </div>
                  <div v-if="item.tags.length > 0" class="flex items-center gap-1 flex-wrap">
                    <span v-for="tag in item.tags" :key="tag" class="pixel-badge !text-[8px] !px-1.5 !py-0.5">
                      #{{ tag }}
                    </span>
                  </div>
                  <p v-if="item.notes" class="text-text-secondary text-xs mt-2 line-clamp-2" :title="item.notes">
                    💬 {{ item.notes }}
                  </p>
                </div>
                <button class="pixel-btn pixel-btn-primary w-full mt-4 !text-xs" @click="startPlaying(item)">
                  🎮 开始游玩
                </button>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>

    <div v-if="dialogVisible" class="fixed inset-0 bg-black/70 flex items-center justify-center z-50" @click.self="dialogVisible = false">
      <div class="pixel-card p-6 max-w-lg w-full mx-4 max-h-[90vh] overflow-y-auto">
        <h3 class="text-bright-yellow mb-4">{{ editingId ? '编辑待玩项目' : '添加到待玩' }}</h3>
        <div class="space-y-4">
          <div>
            <label class="block text-sm text-text-secondary mb-1 pixel-font">选择卡带</label>
            <select v-model="form.cartridgeId" class="pixel-input w-full" :disabled="!!editingId">
              <option v-for="c in availableCartridges" :key="c.id" :value="c.id">
                {{ c.title }} ({{ c.platform }})
              </option>
            </select>
          </div>
          <div>
            <label class="block text-sm text-text-secondary mb-1 pixel-font">优先级</label>
            <select v-model="form.priority" class="pixel-input w-full">
              <option v-for="p in PriorityOptions" :key="p.value" :value="p.value">{{ p.label }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm text-text-secondary mb-1 pixel-font">预计开始日期</label>
            <input v-model="form.plannedStartDate" type="date" class="pixel-input w-full" />
          </div>
          <div>
            <label class="block text-sm text-text-secondary mb-1 pixel-font">标签 (逗号分隔)</label>
            <input v-model="form.tags" type="text" class="pixel-input w-full" placeholder="如: RPG, 神作, 必玩" />
          </div>
          <div>
            <label class="block text-sm text-text-secondary mb-1 pixel-font">备注</label>
            <textarea v-model="form.notes" class="pixel-input w-full min-h-[80px]" placeholder="添加一些备注..."></textarea>
          </div>
        </div>
        <div class="flex gap-4 justify-end mt-6">
          <button class="pixel-btn" @click="dialogVisible = false">取消</button>
          <button class="pixel-btn pixel-btn-primary" @click="saveItem">保存</button>
        </div>
      </div>
    </div>

    <div v-if="deleteConfirmId !== null" class="fixed inset-0 bg-black/70 flex items-center justify-center z-50" @click.self="deleteConfirmId = null">
      <div class="pixel-card p-6 max-w-sm w-full mx-4">
        <h3 class="text-bright-yellow mb-4">确认删除</h3>
        <p class="text-text-secondary mb-6">确定要从待玩清单中移除这款游戏吗？</p>
        <div class="flex gap-4 justify-end">
          <button class="pixel-btn" @click="deleteConfirmId = null">取消</button>
          <button class="pixel-btn pixel-btn-danger" @click="handleDelete(deleteConfirmId)">确认删除</button>
        </div>
      </div>
    </div>
  </div>
</template>
