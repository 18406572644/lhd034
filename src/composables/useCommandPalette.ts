import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { cartridgeApi, playthroughApi, wishlistApi, borrowApi } from '@/api'
import { matchPinyin } from '@/lib/pinyin'
import type { Cartridge, Playthrough, WishlistItem, BorrowRecord } from '@/types'

export interface SearchItem {
  id: string
  label: string
  description: string
  icon: string
  category: 'cartridge' | 'playthrough' | 'wishlist' | 'borrow' | 'command' | 'navigation'
  route?: string
  action?: () => void
}

const HISTORY_KEY = 'command-palette-history'
const MAX_HISTORY = 10

const hotSearches = [
  '马里奥',
  '塞尔达',
  '宝可梦',
  '新增卡带',
  '通关记录',
  '待玩清单'
]

const safeArray = <T>(arr: T[] | null | undefined): T[] =>
  Array.isArray(arr) ? arr : []

const safeJoin = (arr: string[] | null | undefined, separator: string = ', '): string =>
  safeArray(arr).join(separator)

export function useCommandPalette() {
  const router = useRouter()
  const visible = ref(false)
  const query = ref('')
  const searchHistory = ref<string[]>([])
  const loading = ref(false)

  const cartridgeItems = ref<SearchItem[]>([])
  const playthroughItems = ref<SearchItem[]>([])
  const wishlistItems = ref<SearchItem[]>([])
  const borrowItems = ref<SearchItem[]>([])

  const commandItems: SearchItem[] = [
    {
      id: 'cmd-add-cartridge',
      label: '新增卡带',
      description: '添加新的游戏卡带',
      icon: '➕',
      category: 'command',
      route: '/cartridges/new'
    },
    {
      id: 'cmd-add-playthrough',
      label: '新增通关记录',
      description: '记录一次通关',
      icon: '🏆',
      category: 'command',
      route: '/playthroughs/new'
    },
    {
      id: 'cmd-goto-dashboard',
      label: '前往仪表盘',
      description: '查看数据概览',
      icon: '🎮',
      category: 'navigation',
      route: '/'
    },
    {
      id: 'cmd-goto-cartridges',
      label: '前往卡带档案',
      description: '浏览所有卡带',
      icon: '📚',
      category: 'navigation',
      route: '/cartridges'
    },
    {
      id: 'cmd-goto-playthroughs',
      label: '前往通关记录',
      description: '查看通关历史',
      icon: '🏆',
      category: 'navigation',
      route: '/playthroughs'
    },
    {
      id: 'cmd-goto-progress',
      label: '前往进度追踪',
      description: '追踪游戏进度',
      icon: '📊',
      category: 'navigation',
      route: '/progress'
    },
    {
      id: 'cmd-goto-wishlist',
      label: '前往待玩清单',
      description: '管理待玩游戏',
      icon: '⭐',
      category: 'navigation',
      route: '/wishlist'
    },
    {
      id: 'cmd-goto-showcase',
      label: '前往虚拟展柜',
      description: '展示收藏',
      icon: '🖼️',
      category: 'navigation',
      route: '/showcase'
    },
    {
      id: 'cmd-goto-statistics',
      label: '前往统计分析',
      description: '查看数据统计',
      icon: '📈',
      category: 'navigation',
      route: '/statistics'
    },
    {
      id: 'cmd-goto-borrows',
      label: '前往借还管理',
      description: '管理借还记录',
      icon: '🔄',
      category: 'navigation',
      route: '/borrows'
    },
    {
      id: 'cmd-goto-settings',
      label: '前往备份设置',
      description: '管理备份',
      icon: '⚙️',
      category: 'navigation',
      route: '/settings'
    }
  ]

  const loadHistory = () => {
    try {
      const raw = localStorage.getItem(HISTORY_KEY)
      if (raw) {
        searchHistory.value = JSON.parse(raw)
      }
    } catch {
      searchHistory.value = []
    }
  }

  const saveHistory = (term: string) => {
    const trimmed = term.trim()
    if (!trimmed) return
    const list = searchHistory.value.filter(h => h !== trimmed)
    list.unshift(trimmed)
    if (list.length > MAX_HISTORY) list.length = MAX_HISTORY
    searchHistory.value = list
    localStorage.setItem(HISTORY_KEY, JSON.stringify(list))
  }

  const removeHistory = (term: string) => {
    searchHistory.value = searchHistory.value.filter(h => h !== term)
    localStorage.setItem(HISTORY_KEY, JSON.stringify(searchHistory.value))
  }

  const clearHistory = () => {
    searchHistory.value = []
    localStorage.removeItem(HISTORY_KEY)
  }

  const fetchData = async () => {
    loading.value = true
    try {
      const [cartridgesRes, playthroughsRes, wishlistRes, borrowRes] = await Promise.allSettled([
        cartridgeApi.getList({ page: 1, pageSize: 200 }),
        playthroughApi.getList({ page: 1, pageSize: 200 }),
        wishlistApi.getList(),
        borrowApi.getList()
      ])

      if (cartridgesRes.status === 'fulfilled' && cartridgesRes.value?.code === 0) {
        const items = safeArray(cartridgesRes.value?.data?.items)
        cartridgeItems.value = items.map((c: Cartridge) => ({
          id: `cartridge-${c.id}`,
          label: c.title || `卡带 #${c.id}`,
          description: `${c.platform || ''} · ${c.publisher || ''}`,
          icon: '📚',
          category: 'cartridge' as const,
          route: `/cartridges/${c.id}`
        }))
      }

      if (playthroughsRes.status === 'fulfilled' && playthroughsRes.value?.code === 0) {
        const items = safeArray(playthroughsRes.value?.data?.items)
        playthroughItems.value = items.map((p: Playthrough) => ({
          id: `playthrough-${p.id}`,
          label: p.cartridge?.title || `通关记录 #${p.id}`,
          description: `${p.completionDate || ''} · ${p.playTimeHours ?? 0}h`,
          icon: '🏆',
          category: 'playthrough' as const,
          route: `/cartridges/${p.cartridgeId}`
        }))
      }

      if (wishlistRes.status === 'fulfilled' && wishlistRes.value?.code === 0) {
        const items = safeArray(wishlistRes.value?.data)
        wishlistItems.value = items.map((w: WishlistItem) => {
          const priorityLabel = w.priority === 'high' ? '高优先级' : w.priority === 'medium' ? '中优先级' : '低优先级'
          const tagsStr = safeJoin(w.tags)
          return {
            id: `wishlist-${w.id}`,
            label: w.cartridge?.title || `待玩 #${w.id}`,
            description: tagsStr ? `${priorityLabel} · ${tagsStr}` : priorityLabel,
            icon: '⭐',
            category: 'wishlist' as const,
            route: `/wishlist`
          }
        })
      }

      if (borrowRes.status === 'fulfilled' && borrowRes.value?.code === 0) {
        const items = safeArray(borrowRes.value?.data)
        borrowItems.value = items.map((b: BorrowRecord) => {
          const statusLabel = b.status === 'borrowed' ? '借出中' : b.status === 'returned' ? '已归还' : '已逾期'
          return {
            id: `borrow-${b.id}`,
            label: b.cartridge?.title || `借出 #${b.id}`,
            description: `${b.borrowerName || ''} · ${statusLabel}`,
            icon: '🔄',
            category: 'borrow' as const,
            route: `/borrows`
          }
        })
      }
    } finally {
      loading.value = false
    }
  }

  const allDataItems = computed<SearchItem[]>(() => [
    ...cartridgeItems.value,
    ...playthroughItems.value,
    ...wishlistItems.value,
    ...borrowItems.value,
    ...commandItems
  ])

  const filteredItems = computed<SearchItem[]>(() => {
    const q = query.value.trim()
    if (!q) return []

    return allDataItems.value.filter(item => {
      if (matchPinyin(q, item.label)) return true
      if (matchPinyin(q, item.description)) return true
      return false
    })
  })

  const suggestions = computed<SearchItem[]>(() => {
    const q = query.value.trim()
    if (q) return filteredItems.value.slice(0, 20)

    const historyItems = searchHistory.value.map((term, i) => ({
      id: `history-${i}`,
      label: term,
      description: '搜索历史',
      icon: '🕐',
      category: 'command' as const,
      action: () => {
        query.value = term
      }
    }))

    const hotItems = hotSearches.map((term, i) => ({
      id: `hot-${i}`,
      label: term,
      description: '热门搜索',
      icon: '🔥',
      category: 'command' as const,
      action: () => {
        query.value = term
      }
    }))

    return [...historyItems, ...hotItems]
  })

  const open = () => {
    visible.value = true
    query.value = ''
    loadHistory()
    fetchData()
  }

  const close = () => {
    visible.value = false
    query.value = ''
  }

  const selectItem = (item: SearchItem) => {
    saveHistory(item.label)
    close()
    if (item.action) {
      item.action()
      return
    }
    if (item.route) {
      router.push(item.route)
    }
  }

  return {
    visible,
    query,
    loading,
    suggestions,
    searchHistory,
    hotSearches,
    open,
    close,
    selectItem,
    removeHistory,
    clearHistory,
    saveHistory
  }
}
