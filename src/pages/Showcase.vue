<script setup lang="ts">
import { ref, reactive, computed, onMounted, h } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import draggable from 'vuedraggable'
import { Message, Modal } from '@arco-design/web-vue'
import { cartridgeApi } from '@/api'
import { ConditionLabels, ConditionOptions } from '@/types'
import type { Cartridge } from '@/types'

const router = useRouter()
const route = useRoute()

const cartridges = ref<Cartridge[]>([])
const platforms = ref<string[]>([])
const publishers = ref<string[]>([])
const loading = ref(true)

const viewMode = ref<'grid' | 'shelf' | 'rotate3d'>('shelf')
const groupBy = ref<'platform' | 'era' | 'none'>('platform')
const isEditMode = ref(false)
const showShareModal = ref(false)

const showcaseConfig = reactive({
  title: '我的虚拟展柜',
  description: '珍藏经典，回味童年'
})

const filters = reactive({
  platform: '',
  publisher: '',
  condition: ''
})

const filteredCartridges = computed(() => {
  return cartridges.value.filter(c => {
    if (filters.platform && c.platform !== filters.platform) return false
    if (filters.publisher && c.publisher !== filters.publisher) return false
    if (filters.condition && c.condition !== filters.condition) return false
    return true
  })
})

const getEra = (year: number) => {
  if (year < 1985) return '70s-80s初'
  if (year < 1990) return '80s末'
  if (year < 1995) return '90s初'
  if (year < 2000) return '90s末'
  if (year < 2005) return '00s初'
  if (year < 2010) return '00s末'
  if (year < 2015) return '10s初'
  if (year < 2020) return '10s末'
  return '20s+'
}

const groupKey = (c: Cartridge) => {
  if (groupBy.value === 'platform') return c.platform
  if (groupBy.value === 'era') return getEra(c.releaseYear)
  return '全部收藏'
}

const groupedCartridges = computed(() => {
  const groups: Record<string, Cartridge[]> = {}
  filteredCartridges.value.forEach(c => {
    const key = groupKey(c)
    if (!groups[key]) groups[key] = []
    groups[key].push(c)
  })
  return groups
})

const groupOrder = computed(() => {
  const keys = Object.keys(groupedCartridges.value)
  if (groupBy.value === 'platform') {
    return platforms.value.filter(p => keys.includes(p))
  }
  return keys.sort()
})

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

const handleFilterChange = () => {}

const onDragEnd = () => {
  Message.success('排序已更新')
}

const onGroupDragEnd = (groupName: string, evt: any) => {
  const { oldIndex, newIndex } = evt
  if (oldIndex === newIndex) return

  const groupItems = groupedCartridges.value[groupName]
  const movedItem = groupItems[oldIndex]
  const movedId = movedItem.id

  const indicesInAll: number[] = []
  filteredCartridges.value.forEach((c, i) => {
    if (groupKey(c) === groupName) {
      indicesInAll.push(i)
    }
  })

  const oldGlobalIdx = indicesInAll[oldIndex]
  const newGlobalIdx = indicesInAll[newIndex]

  const arr = cartridges.value
  const actualOldIdx = arr.findIndex(c => c.id === movedId)
  if (actualOldIdx === -1) return

  const [item] = arr.splice(actualOldIdx, 1)

  let insertIdx = 0
  let count = 0
  for (let i = 0; i < arr.length; i++) {
    if (groupKey(arr[i]) === groupName) {
      if (count === newIndex) {
        insertIdx = i
        break
      }
      count++
    }
  }
  if (count === newIndex) {
    insertIdx = arr.length
  }

  arr.splice(insertIdx, 0, item)
  Message.success('排序已更新')
}

const generateShareLink = () => {
  const encodedTitle = encodeURIComponent(showcaseConfig.title)
  const encodedDesc = encodeURIComponent(showcaseConfig.description)
  const ids = cartridges.value.map(c => c.id).join(',')
  const shareUrl = `${window.location.origin}${route.path}?share=${encodedTitle}&desc=${encodedDesc}&ids=${ids}&view=${viewMode.value}&group=${groupBy.value}`
  return shareUrl
}

const copyShareLink = async () => {
  try {
    await navigator.clipboard.writeText(generateShareLink())
    Message.success('分享链接已复制到剪贴板')
  } catch {
    Message.error('复制失败，请手动复制')
  }
}

const goToDetail = (id: number) => {
  router.push(`/cartridges/${id}`)
}

const loadData = async () => {
  loading.value = true
  try {
    const [cartRes, platRes, pubRes] = await Promise.all([
      cartridgeApi.getList({ pageSize: 100 }),
      cartridgeApi.getPlatforms(),
      cartridgeApi.getPublishers()
    ])
    if (cartRes.code === 0) cartridges.value = cartRes.data.items
    if (platRes.code === 0) platforms.value = platRes.data
    if (pubRes.code === 0) publishers.value = pubRes.data
  } finally {
    loading.value = false
  }
}

const CartridgeCard = (props: { cartridge: Cartridge; viewMode: string }, { emit }: any) => {
  const c = props.cartridge
  const badgeClass = conditionBadgeClass(c.condition)

  return h('div', {
    class: `cartridge-card ${props.viewMode}`,
    onClick: () => emit('click')
  }, [
    h('div', { class: 'flip-card-inner' }, [
      h('div', { class: 'flip-card-front' }, [
        h('div', { class: 'cartridge-case' }, [
          h('div', { class: 'cartridge-label' }, [
            c.coverImage
              ? h('img', { src: c.coverImage, alt: c.title })
              : h('div', { class: 'no-cover-text' }, c.title)
          ])
        ])
      ]),
      h('div', { class: 'flip-card-back' }, [
        h('div', { class: 'cartridge-back' }, [
          h('div', { class: 'back-title pixel-font' }, c.title),
          h('div', { class: 'back-info' }, [
            h('div', null, `平台: ${c.platform}`),
            h('div', null, `年份: ${c.releaseYear}`),
            h('div', null, `发行: ${c.publisher}`),
            h('div', { class: 'back-condition' }, [
              h('span', { class: `pixel-badge !text-[8px] ${badgeClass}` }, ConditionLabels[c.condition])
            ])
          ])
        ])
      ])
    ]),
    h('div', { class: 'card-footer' }, [
      h('div', { class: 'card-title' }, c.title)
    ])
  ])
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="p-6 space-y-6 scanline-overlay showcase-page">
    <div class="text-center py-4 crt-effect">
      <div v-if="isEditMode" class="mb-4 space-y-2">
        <input
          v-model="showcaseConfig.title"
          class="pixel-input text-center w-full max-w-md pixel-font"
          placeholder="输入展柜标题"
        />
        <input
          v-model="showcaseConfig.description"
          class="pixel-input text-center w-full max-w-md text-sm"
          placeholder="输入展柜描述"
        />
      </div>
      <template v-else>
        <h1 class="text-neon-blue glow-blue mb-2">{{ showcaseConfig.title }}</h1>
        <p class="text-text-secondary pixel-font text-xs">{{ showcaseConfig.description }}</p>
      </template>
    </div>

    <div class="pixel-card p-4 space-y-4">
      <div class="flex flex-wrap items-center gap-4">
        <div class="flex items-center gap-2">
          <span class="text-text-secondary pixel-font text-xs">视图:</span>
          <div class="flex gap-1">
            <button
              :class="['view-btn', { active: viewMode === 'grid' }]"
              @click="viewMode = 'grid'"
              title="网格视图"
            >
              ⊞
            </button>
            <button
              :class="['view-btn', { active: viewMode === 'shelf' }]"
              @click="viewMode = 'shelf'"
              title="木质展架"
            >
              ⊟
            </button>
            <button
              :class="['view-btn', { active: viewMode === 'rotate3d' }]"
              @click="viewMode = 'rotate3d'"
              title="3D旋转"
            >
              ↻
            </button>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <span class="text-text-secondary pixel-font text-xs">分组:</span>
          <select v-model="groupBy" class="pixel-input !py-1 !text-sm">
            <option value="platform">按平台</option>
            <option value="era">按年代</option>
            <option value="none">不分组</option>
          </select>
        </div>

        <div class="flex items-center gap-2 ml-auto">
          <button
            :class="['pixel-btn !py-2 !px-3 !text-[10px]', isEditMode ? 'pixel-btn-success' : '']"
            @click="isEditMode = !isEditMode"
          >
            {{ isEditMode ? '完成' : '编辑' }}
          </button>
          <button class="pixel-btn !py-2 !px-3 !text-[10px]" @click="showShareModal = true">
            分享
          </button>
        </div>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div>
          <label class="block text-sm text-text-secondary mb-1 pixel-font">平台</label>
          <select v-model="filters.platform" class="pixel-input w-full" @change="handleFilterChange">
            <option value="">全部平台</option>
            <option v-for="p in platforms" :key="p" :value="p">{{ p }}</option>
          </select>
        </div>
        <div>
          <label class="block text-sm text-text-secondary mb-1 pixel-font">发行商</label>
          <select v-model="filters.publisher" class="pixel-input w-full" @change="handleFilterChange">
            <option value="">全部发行商</option>
            <option v-for="p in publishers" :key="p" :value="p">{{ p }}</option>
          </select>
        </div>
        <div>
          <label class="block text-sm text-text-secondary mb-1 pixel-font">品相</label>
          <select v-model="filters.condition" class="pixel-input w-full" @change="handleFilterChange">
            <option value="">全部品相</option>
            <option v-for="c in ConditionOptions" :key="c.value" :value="c.value">{{ c.label }}</option>
          </select>
        </div>
      </div>
    </div>

    <div v-if="loading" class="text-center py-16">
      <div class="text-4xl mb-4 animate-pulse">🎮</div>
      <p class="text-text-secondary pixel-font">加载中...</p>
    </div>

    <div v-else-if="groupOrder.length === 0" class="text-center py-16">
      <div class="text-6xl mb-4">📦</div>
      <h3 class="text-bright-yellow mb-2">展柜空空如也</h3>
      <p class="text-text-secondary">添加一些卡带点亮你的展柜吧！</p>
    </div>

    <div v-else class="space-y-8">
      <div
        v-for="group in groupOrder"
        :key="group"
        :class="['shelf-container', { 'wooden-shelf': viewMode === 'shelf', 'rotate3d-container': viewMode === 'rotate3d' }]"
      >
        <div v-if="groupBy !== 'none'" class="flex items-center justify-between mb-6 group-header">
          <h2 class="text-bright-yellow glow-yellow pixel-font text-sm group-title">
            ▣ {{ group }}
          </h2>
          <span class="pixel-badge !text-[10px]">
            {{ groupedCartridges[group].length }} 张
          </span>
        </div>

        <draggable
          v-if="isEditMode"
          :list="groupedCartridges[group]"
          item-key="id"
          class="cartridge-grid"
          :class="{
            'grid-style': viewMode === 'grid',
            'shelf-style': viewMode === 'shelf',
            'rotate3d-style': viewMode === 'rotate3d'
          }"
          ghost-class="drag-ghost"
          chosen-class="drag-chosen"
          drag-class="drag-dragging"
          @end="onGroupDragEnd(group, $event)"
        >
          <template #item="{ element }">
            <div class="cartridge-item drag-item">
              <div class="drag-handle" title="拖拽排序">⋮⋮</div>
              <component :is="CartridgeCard" :cartridge="element" :view-mode="viewMode" @click="goToDetail(element.id)" />
            </div>
          </template>
        </draggable>

        <div
          v-else
          class="cartridge-grid"
          :class="{
            'grid-style': viewMode === 'grid',
            'shelf-style': viewMode === 'shelf',
            'rotate3d-style': viewMode === 'rotate3d'
          }"
        >
          <div v-for="cartridge in groupedCartridges[group]" :key="cartridge.id" class="cartridge-item">
            <component :is="CartridgeCard" :cartridge="cartridge" :view-mode="viewMode" @click="goToDetail(cartridge.id)" />
          </div>
        </div>

        <div v-if="viewMode === 'shelf'" class="shelf-board"></div>
      </div>
    </div>

    <Modal v-model:visible="showShareModal" title="分享展柜" width="480px">
      <div class="space-y-4">
        <div>
          <label class="block text-sm text-text-secondary mb-1 pixel-font">展柜标题</label>
          <input v-model="showcaseConfig.title" class="pixel-input w-full" placeholder="输入标题" />
        </div>
        <div>
          <label class="block text-sm text-text-secondary mb-1 pixel-font">展柜描述</label>
          <textarea v-model="showcaseConfig.description" class="pixel-input w-full" rows="2" placeholder="输入描述"></textarea>
        </div>
        <div>
          <label class="block text-sm text-text-secondary mb-1 pixel-font">分享链接</label>
          <div class="flex gap-2">
            <input :value="generateShareLink()" class="pixel-input w-full !text-xs" readonly />
            <button class="pixel-btn !py-2 !px-3 !text-[10px]" @click="copyShareLink">复制</button>
          </div>
        </div>
      </div>
      <template #footer>
        <button class="pixel-btn" @click="showShareModal = false">关闭</button>
      </template>
    </Modal>
  </div>
</template>

<style scoped>
.showcase-page {
  position: relative;
}

.view-btn {
  width: 32px;
  height: 32px;
  background: var(--dark-bg-2);
  border: 2px solid var(--neon-blue);
  color: var(--neon-blue);
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.view-btn:hover {
  background: var(--dark-bg-3);
  box-shadow: 0 0 10px rgba(0, 240, 255, 0.5);
}

.view-btn.active {
  background: var(--neon-blue);
  color: var(--dark-bg);
  box-shadow: 0 0 10px var(--neon-blue);
}

.shelf-container {
  position: relative;
  padding: 16px;
  transition: all 0.3s ease;
}

.wooden-shelf {
  background: linear-gradient(180deg, #3d2914 0%, #2a1d0d 50%, #1a1208 100%);
  border: 4px solid #5a3d1f;
  box-shadow:
    inset 0 8px 16px rgba(0, 0, 0, 0.5),
    0 4px 0 #2a1d0d,
    0 8px 0 #1a1208;
  padding: 24px;
  padding-top: 40px;
}

.wooden-shelf::before {
  content: '';
  position: absolute;
  top: -8px;
  left: 0;
  right: 0;
  height: 8px;
  background: linear-gradient(180deg, var(--neon-blue) 0%, var(--neon-blue-dark) 100%);
  box-shadow: 0 0 20px var(--neon-blue);
  animation: neon-pulse 2s ease-in-out infinite;
}

.wooden-shelf .group-header {
  margin-bottom: 20px;
}

.wooden-shelf .group-title {
  color: var(--bright-yellow);
  text-shadow: 0 0 5px var(--bright-yellow), 0 0 10px var(--bright-yellow);
}

.shelf-board {
  position: relative;
  height: 16px;
  background: linear-gradient(180deg, #5a3d1f 0%, #3d2914 50%, #2a1d0d 100%);
  margin-top: -4px;
  border-radius: 0 0 4px 4px;
  box-shadow:
    0 4px 8px rgba(0, 0, 0, 0.4),
    inset 0 -2px 4px rgba(0, 0, 0, 0.3);
}

.cartridge-grid {
  display: grid;
  gap: 16px;
  min-height: 100px;
}

.grid-style {
  grid-template-columns: repeat(2, 1fr);
}

.shelf-style {
  grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
}

.rotate3d-style {
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  perspective: 1000px;
}

@media (min-width: 640px) {
  .grid-style { grid-template-columns: repeat(3, 1fr); }
}
@media (min-width: 768px) {
  .grid-style { grid-template-columns: repeat(4, 1fr); }
}
@media (min-width: 1024px) {
  .grid-style { grid-template-columns: repeat(6, 1fr); }
}
@media (min-width: 1280px) {
  .grid-style { grid-template-columns: repeat(8, 1fr); }
}

.cartridge-item {
  position: relative;
  cursor: pointer;
}

.drag-item {
  cursor: grab;
}

.drag-item:active {
  cursor: grabbing;
}

.drag-handle {
  position: absolute;
  top: -8px;
  right: -8px;
  width: 24px;
  height: 24px;
  background: var(--neon-blue);
  color: var(--dark-bg);
  font-size: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 2px;
  z-index: 10;
  opacity: 0;
  transition: opacity 0.2s ease;
  cursor: grab;
}

.cartridge-item:hover .drag-handle {
  opacity: 1;
}

.drag-ghost {
  opacity: 0.4;
  background: var(--dark-bg-3);
}

.drag-chosen {
  transform: scale(1.05);
}

.drag-dragging {
  transform: rotate(3deg) scale(1.05);
  z-index: 100;
}

.cartridge-card {
  width: 100%;
  perspective: 1000px;
}

.flip-card-inner {
  position: relative;
  width: 100%;
  aspect-ratio: 3 / 4;
  transition: transform 0.6s;
  transform-style: preserve-3d;
}

.cartridge-card:hover .flip-card-inner {
  transform: rotateY(180deg);
}

.flip-card-front,
.flip-card-back {
  position: absolute;
  width: 100%;
  height: 100%;
  backface-visibility: hidden;
  -webkit-backface-visibility: hidden;
}

.flip-card-back {
  transform: rotateY(180deg);
}

.cartridge-back {
  width: 100%;
  height: 100%;
  background: linear-gradient(145deg, var(--dark-bg-3) 0%, var(--dark-bg-2) 100%);
  border: 4px solid var(--neon-blue);
  padding: 12px 8px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  box-shadow:
    0 0 20px rgba(0, 240, 255, 0.3),
    inset 0 0 20px rgba(0, 240, 255, 0.1);
  box-sizing: border-box;
}

.back-title {
  font-size: 10px;
  color: var(--bright-yellow);
  text-align: center;
  padding-bottom: 6px;
  border-bottom: 2px solid var(--pixel-pink);
  text-shadow: 0 0 5px var(--bright-yellow);
}

.back-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 11px;
  color: var(--text-primary);
}

.back-condition {
  margin-top: auto;
  text-align: center;
}

.no-cover-text {
  font-size: 10px;
  font-family: var(--font-pixel);
  color: var(--bright-yellow);
  line-height: 1.2;
  padding: 0 4px;
  text-align: center;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  word-break: break-all;
}

.card-footer {
  margin-top: 8px;
  text-align: center;
}

.card-title {
  font-size: 12px;
  color: var(--text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  padding: 0 4px;
}

.rotate3d-container {
  perspective: 1500px;
}

.rotate3d-style .cartridge-card {
  transition: transform 0.5s ease;
}

.rotate3d-style .cartridge-card:hover {
  transform: translateY(-10px) rotateY(15deg) rotateX(-5deg);
}

.rotate3d-style .flip-card-inner {
  transition: transform 0.6s cubic-bezier(0.4, 0, 0.2, 1);
}

.rotate3d-style .cartridge-card:hover .flip-card-inner {
  transform: rotateY(180deg) translateZ(20px);
}

@keyframes neon-pulse {
  0%, 100% {
    opacity: 1;
    box-shadow: 0 0 20px var(--neon-blue);
  }
  50% {
    opacity: 0.8;
    box-shadow: 0 0 30px var(--neon-blue), 0 0 40px var(--pixel-pink);
  }
}

.group-header {
  position: relative;
  z-index: 1;
}
</style>
