<script setup lang="ts">
import { ref, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { useCommandPalette } from '@/composables/useCommandPalette'
import type { SearchItem } from '@/composables/useCommandPalette'

const {
  visible,
  query,
  loading,
  suggestions,
  searchHistory,
  open,
  close,
  selectItem,
  removeHistory,
  clearHistory
} = useCommandPalette()

const inputRef = ref<HTMLInputElement | null>(null)
const activeIndex = ref(0)
const listRef = ref<HTMLElement | null>(null)

const categoryLabels: Record<string, string> = {
  cartridge: '卡带',
  playthrough: '通关记录',
  wishlist: '待玩',
  borrow: '借还',
  command: '快捷操作',
  navigation: '导航'
}

const categoryColors: Record<string, string> = {
  cartridge: 'var(--neon-blue)',
  playthrough: 'var(--pixel-green)',
  wishlist: 'var(--bright-yellow)',
  borrow: 'var(--pixel-orange)',
  command: 'var(--pixel-pink)',
  navigation: 'var(--neon-blue-dark)'
}

watch(visible, async (val) => {
  if (val) {
    activeIndex.value = 0
    await nextTick()
    inputRef.value?.focus()
  }
})

watch(suggestions, () => {
  activeIndex.value = 0
})

const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === 'ArrowDown') {
    e.preventDefault()
    activeIndex.value = Math.min(activeIndex.value + 1, suggestions.value.length - 1)
    scrollToActive()
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    activeIndex.value = Math.max(activeIndex.value - 1, 0)
    scrollToActive()
  } else if (e.key === 'Enter') {
    e.preventDefault()
    const item = suggestions.value[activeIndex.value]
    if (item) selectItem(item)
  } else if (e.key === 'Escape') {
    close()
  }
}

const scrollToActive = () => {
  nextTick(() => {
    const list = listRef.value
    if (!list) return
    const active = list.children[activeIndex.value] as HTMLElement
    if (active) {
      active.scrollIntoView({ block: 'nearest' })
    }
  })
}

const handleOverlayClick = (e: MouseEvent) => {
  if ((e.target as HTMLElement).classList.contains('cmd-overlay')) {
    close()
  }
}

const globalKeyHandler = (e: KeyboardEvent) => {
  if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
    e.preventDefault()
    if (visible.value) {
      close()
    } else {
      open()
    }
  }
}

onMounted(() => {
  document.addEventListener('keydown', globalKeyHandler)
})

onUnmounted(() => {
  document.removeEventListener('keydown', globalKeyHandler)
})

const isHistoryItem = (item: SearchItem) => item.id.startsWith('history-')
const isHotItem = (item: SearchItem) => item.id.startsWith('hot-')
</script>

<template>
  <Teleport to="body">
    <Transition name="cmd-fade">
      <div
        v-if="visible"
        class="cmd-overlay fixed inset-0 z-50 flex items-start justify-center pt-[15vh]"
        @click="handleOverlayClick"
      >
        <div class="cmd-palette w-full max-w-[640px] mx-4">
          <div class="cmd-container pixel-border bg-dark-bg-2 overflow-hidden">
            <div class="cmd-input-wrapper flex items-center gap-3 px-4 py-3 border-b-4" style="border-color: var(--neon-blue);">
              <span class="text-xl">🔍</span>
              <input
                ref="inputRef"
                v-model="query"
                type="text"
                class="cmd-input flex-1 bg-transparent border-none outline-none text-text-primary font-body text-lg"
                placeholder="搜索卡带、通关记录、快捷操作..."
                @keydown="handleKeydown"
              />
              <div class="flex items-center gap-1">
                <span v-if="loading" class="text-neon-blue animate-pulse pixel-font text-xs">LOADING</span>
                <kbd class="cmd-kbd">ESC</kbd>
              </div>
            </div>

            <div
              v-if="!query.trim() && searchHistory.length > 0"
              class="cmd-history-header flex items-center justify-between px-4 py-2 border-b"
              style="border-color: rgba(0,240,255,0.2);"
            >
              <span class="pixel-font text-xs text-text-secondary">搜索历史</span>
              <button
                class="pixel-font text-xs text-pixel-red hover:glow-yellow transition-all cursor-pointer bg-transparent border-none"
                @click="clearHistory"
              >
                清空
              </button>
            </div>

            <div
              ref="listRef"
              class="cmd-list max-h-[400px] overflow-y-auto scroll-hidden"
            >
              <div
                v-for="(item, index) in suggestions"
                :key="item.id"
                class="cmd-item flex items-center gap-3 px-4 py-3 cursor-pointer transition-all duration-150"
                :class="{
                  'cmd-item-active': index === activeIndex,
                  'cmd-item-history': isHistoryItem(item),
                  'cmd-item-hot': isHotItem(item)
                }"
                @click="selectItem(item)"
                @mouseenter="activeIndex = index"
              >
                <span class="cmd-item-icon text-xl flex-shrink-0">{{ item.icon }}</span>
                <div class="cmd-item-content flex-1 min-w-0">
                  <div class="cmd-item-label text-text-primary truncate">{{ item.label }}</div>
                  <div class="cmd-item-desc text-text-secondary text-sm truncate">{{ item.description }}</div>
                </div>
                <span
                  class="cmd-item-category pixel-font text-[10px] px-2 py-0.5 flex-shrink-0"
                  :style="{
                    borderColor: categoryColors[item.category],
                    color: categoryColors[item.category]
                  }"
                >
                  {{ categoryLabels[item.category] }}
                </span>
                <button
                  v-if="isHistoryItem(item)"
                  class="cmd-remove-btn flex-shrink-0 text-text-secondary hover:text-pixel-red transition-colors bg-transparent border-none cursor-pointer text-sm"
                  @click.stop="removeHistory(item.label)"
                >
                  ✕
                </button>
              </div>

              <div
                v-if="query.trim() && suggestions.length === 0 && !loading"
                class="cmd-empty flex flex-col items-center justify-center py-12"
              >
                <span class="text-4xl mb-3">👾</span>
                <span class="pixel-font text-xs text-text-secondary">NO RESULTS FOUND</span>
                <span class="text-text-secondary text-sm mt-1">试试其他关键词</span>
              </div>
            </div>

            <div class="cmd-footer flex items-center justify-between px-4 py-2 border-t" style="border-color: rgba(0,240,255,0.2);">
              <div class="flex items-center gap-3">
                <span class="cmd-footer-hint">
                  <kbd class="cmd-kbd-sm">↑↓</kbd>
                  <span class="text-text-secondary text-xs ml-1">导航</span>
                </span>
                <span class="cmd-footer-hint">
                  <kbd class="cmd-kbd-sm">↵</kbd>
                  <span class="text-text-secondary text-xs ml-1">选择</span>
                </span>
                <span class="cmd-footer-hint">
                  <kbd class="cmd-kbd-sm">Esc</kbd>
                  <span class="text-text-secondary text-xs ml-1">关闭</span>
                </span>
              </div>
              <span class="pixel-font text-[10px] text-text-secondary">CTRL+K</span>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.cmd-overlay {
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(4px);
}

.cmd-container {
  background: linear-gradient(145deg, var(--dark-bg-2) 0%, var(--dark-bg) 100%);
}

.cmd-input {
  font-family: var(--font-body);
}

.cmd-input::placeholder {
  color: var(--text-secondary);
}

.cmd-kbd {
  font-family: var(--font-pixel);
  font-size: 8px;
  padding: 2px 6px;
  background: var(--dark-bg-3);
  border: 2px solid var(--neon-blue);
  color: var(--neon-blue);
}

.cmd-kbd-sm {
  font-family: var(--font-pixel);
  font-size: 7px;
  padding: 1px 4px;
  background: var(--dark-bg-3);
  border: 1px solid rgba(0, 240, 255, 0.4);
  color: var(--text-secondary);
}

.cmd-item {
  border-bottom: 1px solid rgba(0, 240, 255, 0.1);
}

.cmd-item:hover {
  background: rgba(0, 240, 255, 0.05);
}

.cmd-item-active {
  background: rgba(0, 240, 255, 0.1) !important;
  border-left: 3px solid var(--neon-blue);
}

.cmd-item-category {
  border: 1px solid;
  border-radius: 2px;
}

.cmd-fade-enter-active,
.cmd-fade-leave-active {
  transition: opacity 0.15s ease;
}

.cmd-fade-enter-from,
.cmd-fade-leave-to {
  opacity: 0;
}

.cmd-fade-enter-active .cmd-palette {
  animation: cmd-slide-in 0.2s ease;
}

@keyframes cmd-slide-in {
  from {
    opacity: 0;
    transform: translateY(-20px) scale(0.95);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}
</style>
