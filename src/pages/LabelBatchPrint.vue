<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import { cartridgeApi } from '@/api'
import { ConditionLabels, ConditionOptions } from '@/types'
import type { Cartridge } from '@/types'
import QRCodeLabel from '@/components/QRCodeLabel.vue'
import { generateQRDataURL } from '@/lib/qrcode'

const router = useRouter()

const loading = ref(true)
const allCartridges = ref<Cartridge[]>([])
const selectedIds = ref<Set<number>>(new Set())
const filters = ref({
  search: '',
  platform: '',
  condition: ''
})
const platforms = ref<string[]>([])
const labelsPerRow = ref(4)
const labelSize = ref<'small' | 'medium'>('small')

const filteredCartridges = computed(() => {
  return allCartridges.value.filter(c => {
    if (filters.value.search) {
      const search = filters.value.search.toLowerCase()
      if (!c.title.toLowerCase().includes(search) &&
          !String(c.id).includes(search)) {
        return false
      }
    }
    if (filters.value.platform && c.platform !== filters.value.platform) {
      return false
    }
    if (filters.value.condition && c.condition !== filters.value.condition) {
      return false
    }
    return true
  })
})

const selectedCartridges = computed(() => {
  return filteredCartridges.value.filter(c => selectedIds.value.has(c.id))
})

const isAllSelected = computed(() => {
  return filteredCartridges.value.length > 0 &&
    filteredCartridges.value.every(c => selectedIds.value.has(c.id))
})

const rows = computed(() => {
  const result: Cartridge[][] = []
  const perRow = labelsPerRow.value
  for (let i = 0; i < selectedCartridges.value.length; i += perRow) {
    result.push(selectedCartridges.value.slice(i, i + perRow))
  }
  return result
})

const loadData = async () => {
  try {
    loading.value = true
    const [listRes, platformsRes] = await Promise.all([
      cartridgeApi.getList({ page: 1, pageSize: 1000 }),
      cartridgeApi.getPlatforms()
    ])
    allCartridges.value = listRes.data?.items || []
    platforms.value = platformsRes.data || []
  } catch (e) {
    console.error(e)
    Message.error('加载卡带列表失败')
  } finally {
    loading.value = false
  }
}

const toggleSelect = (id: number) => {
  if (selectedIds.value.has(id)) {
    selectedIds.value.delete(id)
  } else {
    selectedIds.value.add(id)
  }
  selectedIds.value = new Set(selectedIds.value)
}

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedIds.value = new Set()
  } else {
    selectedIds.value = new Set(filteredCartridges.value.map(c => c.id))
  }
}

const clearSelection = () => {
  selectedIds.value = new Set()
}

const printLabels = async () => {
  if (selectedCartridges.value.length === 0) {
    Message.warning('请先选择要打印的卡带')
    return
  }
  const printWindow = window.open('', '_blank')
  if (!printWindow) {
    Message.error('无法打开打印窗口，请检查浏览器设置')
    return
  }
  const sizePx = labelSize.value === 'small' ? 140 : 200
  const qrSize = labelSize.value === 'small' ? 100 : 140
  const gap = 12

  const escapeHtml = (str: string) => {
    const div = document.createElement('div')
    div.textContent = str
    return div.innerHTML
  }

  const labelData = await Promise.all(
    selectedCartridges.value.map(async (c) => {
      const qrDataUrl = await generateQRDataURL(c.id, qrSize)
      return {
        cartridge: c,
        qrDataUrl
      }
    })
  )

  const labelsHtml = labelData.map(({ cartridge: c, qrDataUrl }) => `
    <div class="label-item" style="width:${sizePx}px;">
      <div class="label-inner">
        <div class="qr-wrapper">
          <img src="${qrDataUrl}" alt="QR" style="width:100%;display:block;" />
        </div>
        <div class="label-info">
          <div class="label-title">${escapeHtml(c.title)}</div>
          <div class="label-meta">
            <span class="tag">${escapeHtml(c.platform)}</span>
            <span class="tag">${escapeHtml(ConditionLabels[c.condition] || c.condition)}</span>
          </div>
          <div class="label-id">ID: ${c.id}</div>
        </div>
      </div>
    </div>
  `).join('')

  printWindow.document.write(`
    <!DOCTYPE html>
    <html>
    <head>
      <title>批量打印标签 - 共 ${selectedCartridges.value.length} 张</title>
      <style>
        @page {
          size: A4;
          margin: 10mm;
        }
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
          font-family: 'Courier New', monospace;
          background: white;
          color: #111;
        }
        .print-page {
          width: 190mm;
          padding: 5mm;
        }
        .labels-grid {
          display: flex;
          flex-wrap: wrap;
          gap: ${gap}px;
          justify-content: flex-start;
        }
        .label-item {
          flex: 0 0 auto;
          page-break-inside: avoid;
        }
        .label-inner {
          border: 2px solid #333;
          background: white;
          padding: 8px;
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 6px;
        }
        .qr-wrapper {
          width: ${labelSize.value === 'small' ? 100 : 140}px;
          height: ${labelSize.value === 'small' ? 100 : 140}px;
          padding: 4px;
          background: white;
          border: 1px solid #999;
        }
        .label-info {
          width: 100%;
          text-align: center;
        }
        .label-title {
          font-size: ${labelSize.value === 'small' ? 10 : 12}px;
          font-weight: bold;
          color: #111;
          line-height: 1.3;
          max-height: 2.6em;
          overflow: hidden;
          word-break: break-all;
        }
        .label-meta {
          display: flex;
          justify-content: center;
          gap: 4px;
          margin-top: 4px;
          flex-wrap: wrap;
        }
        .tag {
          font-size: ${labelSize.value === 'small' ? 8 : 10}px;
          padding: 1px 4px;
          border: 1px solid #666;
          color: #333;
        }
        .label-id {
          font-size: ${labelSize.value === 'small' ? 8 : 10}px;
          color: #666;
          margin-top: 4px;
          letter-spacing: 0.5px;
        }
        .page-header {
          text-align: center;
          padding: 10px 0 20px 0;
          border-bottom: 2px dashed #ccc;
          margin-bottom: 20px;
          font-size: 14px;
          color: #666;
        }
        @media print {
          body { background: white; }
          .print-page { page-break-after: always; }
          .print-page:last-child { page-break-after: auto; }
        }
      </style>
    </head>
    <body>
      <div class="print-page">
        <div class="page-header">
          卡带标签 - 共 ${selectedCartridges.value.length} 张 | 生成时间: ${new Date().toLocaleString('zh-CN')}
        </div>
        <div class="labels-grid">
          ${labelsHtml}
        </div>
      </div>
      <script>
        window.onload = function() {
          setTimeout(() => {
            window.print();
          }, 1000);
        };
      <\/script>
    </body>
    </html>
  `)
  printWindow.document.close()
}

onMounted(loadData)
</script>

<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
      <div>
        <h1 class="text-neon-blue glow-blue">批量标签生成</h1>
        <p class="text-text-secondary">
          选择多张卡带，生成 A4 排版的标签页，可直接打印贴在实体卡带盒上
        </p>
      </div>
      <div class="flex gap-3">
        <button
          class="pixel-btn pixel-btn-primary"
          :disabled="selectedCartridges.length === 0"
          @click="printLabels"
        >
          🖨️ 打印 {{ selectedCartridges.length }} 张标签
        </button>
        <button
          v-if="selectedIds.size > 0"
          class="pixel-btn"
          @click="clearSelection"
        >
          清空选择
        </button>
      </div>
    </div>

    <div class="pixel-card p-4">
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
        <div class="lg:col-span-2">
          <label class="block text-sm text-text-secondary mb-1 pixel-font">搜索</label>
          <input
            v-model="filters.search"
            type="text"
            class="pixel-input w-full"
            placeholder="搜索游戏标题或 ID..."
          />
        </div>
        <div>
          <label class="block text-sm text-text-secondary mb-1 pixel-font">平台</label>
          <select v-model="filters.platform" class="pixel-input w-full">
            <option value="">全部平台</option>
            <option v-for="p in platforms" :key="p" :value="p">{{ p }}</option>
          </select>
        </div>
        <div>
          <label class="block text-sm text-text-secondary mb-1 pixel-font">品相</label>
          <select v-model="filters.condition" class="pixel-input w-full">
            <option value="">全部品相</option>
            <option v-for="c in ConditionOptions" :key="c.value" :value="c.value">{{ c.label }}</option>
          </select>
        </div>
        <div>
          <label class="block text-sm text-text-secondary mb-1 pixel-font">标签尺寸</label>
          <select v-model="labelSize" class="pixel-input w-full">
            <option value="small">小 (推荐 4列/行)</option>
            <option value="medium">中 (推荐 3列/行)</option>
          </select>
        </div>
      </div>

      <div class="mt-4 flex flex-wrap items-center gap-4 pt-4 border-t-2 border-neon-blue/30">
        <div class="flex items-center gap-2">
          <span class="pixel-font text-xs text-text-secondary">每页列数:</span>
          <div class="flex gap-1">
            <button
              v-for="n in [2, 3, 4, 5]"
              :key="n"
              class="pixel-btn !px-3 !py-1 !text-[10px]"
              :class="{ 'pixel-btn-primary': labelsPerRow === n }"
              @click="labelsPerRow = n"
            >
              {{ n }}列
            </button>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <span class="pixel-font text-xs text-text-secondary">
            已筛选: {{ filteredCartridges.length }} 张
          </span>
          <span class="pixel-font text-xs text-bright-yellow">
            已选择: {{ selectedIds.size }} 张
          </span>
        </div>
        <button
          class="pixel-btn !px-3 !py-1 !text-[10px] ml-auto"
          @click="toggleSelectAll"
        >
          {{ isAllSelected ? '取消全选' : '全选当前筛选' }}
        </button>
      </div>
    </div>

    <div v-if="loading" class="text-center py-16">
      <div class="pixel-font text-neon-blue text-xl animate-pulse">LOADING...</div>
    </div>

    <div v-else-if="filteredCartridges.length === 0" class="text-center py-16">
      <div class="text-6xl mb-4">🎮</div>
      <h3 class="text-bright-yellow mb-2">没有匹配的卡带</h3>
      <p class="text-text-secondary">请调整筛选条件</p>
    </div>

    <div v-else class="pixel-card p-4">
      <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
        <div
          v-for="cartridge in filteredCartridges"
          :key="cartridge.id"
          class="relative cursor-pointer transition-all"
          :class="[
            selectedIds.has(cartridge.id)
              ? 'ring-4 ring-bright-yellow ring-offset-2 ring-offset-dark-bg-2'
              : ''
          ]"
          @click="toggleSelect(cartridge.id)"
        >
          <div
            class="absolute top-2 left-2 z-10 w-6 h-6 flex items-center justify-center text-xs font-bold"
            :class="[
              selectedIds.has(cartridge.id)
                ? 'bg-bright-yellow text-dark-bg'
                : 'bg-dark-bg-3 border-2 border-neon-blue text-neon-blue'
            ]"
          >
            {{ selectedIds.has(cartridge.id) ? '✓' : '' }}
          </div>
          <div class="cartridge-case !aspect-[3/4]">
            <div class="cartridge-label">
              <img
                v-if="cartridge.coverImage"
                :src="cartridge.coverImage"
                :alt="cartridge.title"
              />
              <div v-else class="flex flex-col items-center justify-center h-full">
                <div class="text-xs pixel-font text-bright-yellow leading-relaxed">
                  {{ cartridge.title }}
                </div>
                <div class="text-xs text-neon-blue mt-1">{{ cartridge.platform }}</div>
              </div>
            </div>
          </div>
          <div class="mt-2 space-y-1 px-1">
            <div class="pixel-font text-[10px] text-bright-yellow truncate">
              {{ cartridge.title }}
            </div>
            <div class="flex items-center gap-1 flex-wrap">
              <span class="pixel-badge !text-[8px] !px-1.5 !py-0.5">
                {{ cartridge.platform }}
              </span>
              <span class="pixel-badge !text-[8px] !px-1.5 !py-0.5">
                {{ ConditionLabels[cartridge.condition] }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="selectedCartridges.length > 0" class="pixel-card p-4">
      <h3 class="text-bright-yellow mb-4">
        📄 打印预览 ({{ labelsPerRow }}列/行)
      </h3>
      <div
        class="bg-white p-8 space-y-4 max-h-[600px] overflow-y-auto scroll-hidden"
        style="box-shadow: inset 0 0 20px rgba(0,0,0,0.1);"
      >
        <div
          v-for="(row, rowIdx) in rows"
          :key="rowIdx"
          class="flex justify-start gap-3"
        >
          <div
            v-for="cartridge in row"
            :key="cartridge.id"
            class="shrink-0"
            :style="{ width: labelSize === 'small' ? '160px' : '220px' }"
          >
            <QRCodeLabel
              :cartridge="cartridge"
              :size="labelSize"
              :show-title="true"
              :for-print="true"
            />
          </div>
        </div>
      </div>
      <div class="mt-4 flex justify-end">
        <button
          class="pixel-btn pixel-btn-primary"
          @click="printLabels"
        >
          🖨️ 打印全部 {{ selectedCartridges.length }} 张标签
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.ring-offset-dark-bg-2 {
  --tw-ring-offset-color: var(--dark-bg-2);
}
</style>
