<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const collapsed = ref(false)

const menuItems = [
  { path: '/', name: '仪表盘', icon: '🎮' },
  { path: '/cartridges', name: '卡带档案', icon: '📚' },
  { path: '/playthroughs', name: '通关记录', icon: '🏆' },
  { path: '/progress', name: '进度追踪', icon: '📊' },
  { path: '/wishlist', name: '待玩清单', icon: '⭐' },
  { path: '/showcase', name: '虚拟展柜', icon: '🖼️' },
  { path: '/statistics', name: '统计分析', icon: '📈' },
  { path: '/borrows', name: '借还管理', icon: '🔄' }
]

const isActive = (path: string) => {
  if (path === '/') return route.path === '/'
  return route.path.startsWith(path)
}

const navigate = (path: string) => {
  router.push(path)
}

const toggleSidebar = () => {
  collapsed.value = !collapsed.value
}

const sidebarWidth = computed(() => collapsed.value ? '0px' : '240px')
</script>

<template>
  <div class="min-h-screen flex scanline-overlay">
    <aside
      class="fixed left-0 top-0 h-full z-30 transition-all duration-300 overflow-hidden"
      :style="{ width: sidebarWidth }"
    >
      <div class="w-60 h-full bg-dark-bg-2 pixel-border flex flex-col">
        <div class="p-6 border-b-4 border-neon-blue">
          <h1 class="pixel-font text-neon-blue glow-blue text-center text-sm leading-relaxed">
            PIXEL<br />CARTRIDGE<br />ARCHIVE
          </h1>
        </div>

        <nav class="flex-1 py-4 overflow-y-auto scroll-hidden">
          <div
            v-for="item in menuItems"
            :key="item.path"
            class="mx-3 mb-2 cursor-pointer transition-all duration-200"
            @click="navigate(item.path)"
          >
            <div
              class="flex items-center gap-3 px-4 py-3 pixel-border"
              :class="isActive(item.path) ? 'bg-dark-bg-3 glow-blue' : 'hover:bg-dark-bg-3'"
            >
              <span class="text-2xl">{{ item.icon }}</span>
              <span
                class="pixel-font text-xs"
                :class="isActive(item.path) ? 'text-neon-blue glow-blue' : 'text-text-primary'"
              >
                {{ item.name }}
              </span>
            </div>
          </div>
        </nav>

        <div class="p-4 border-t-4 border-neon-blue">
          <div class="text-center text-text-secondary text-xs pixel-font">
            © 2026 PIXEL ARCHIVE
          </div>
        </div>
      </div>
    </aside>

    <div
      class="flex-1 flex flex-col transition-all duration-300"
      :style="{ marginLeft: sidebarWidth }"
    >
      <header class="h-16 bg-dark-bg-2 border-b-4 border-neon-blue flex items-center justify-between px-6 sticky top-0 z-20">
        <button
          class="pixel-btn !px-3 !py-2 lg:hidden"
          @click="toggleSidebar"
        >
          {{ collapsed ? '☰' : '✕' }}
        </button>

        <div class="hidden lg:block">
          <h2 class="pixel-font text-neon-blue glow-blue text-sm">
            {{ menuItems.find(m => isActive(m.path))?.name || '仪表盘' }}
          </h2>
        </div>

        <div class="flex items-center gap-4">
          <span class="pixel-badge">
            {{ route.path }}
          </span>
        </div>
      </header>

      <main class="flex-1 p-6 overflow-auto scroll-hidden">
        <router-view />
      </main>
    </div>

    <div
      v-if="!collapsed"
      class="fixed inset-0 bg-black/50 z-20 lg:hidden"
      @click="toggleSidebar"
    />
  </div>
</template>

<style scoped>
.border-neon-blue {
  border-color: var(--neon-blue);
}

.bg-dark-bg-2 {
  background-color: var(--dark-bg-2);
}

.bg-dark-bg-3 {
  background-color: var(--dark-bg-3);
}

.text-neon-blue {
  color: var(--neon-blue);
}

.text-text-primary {
  color: var(--text-primary);
}

.text-text-secondary {
  color: var(--text-secondary);
}
</style>
