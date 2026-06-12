import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Cartridge, OverviewStats } from '@/types'
import { cartridgeApi, statisticsApi } from '@/api'

export const useCartridgeStore = defineStore('cartridge', () => {
  const cartridges = ref<Cartridge[]>([])
  const total = ref(0)
  const loading = ref(false)
  const stats = ref<OverviewStats | null>(null)
  const platforms = ref<string[]>([])
  const publishers = ref<string[]>([])

  const fetchList = async (params?: {
    page?: number
    pageSize?: number
    platform?: string
    publisher?: string
    condition?: string
    year?: string
    search?: string
  }) => {
    loading.value = true
    try {
      const res = await cartridgeApi.getList(params || {})
      if (res.code === 0) {
        cartridges.value = res.data.items
        total.value = res.data.total
      }
    } finally {
      loading.value = false
    }
  }

  const fetchStats = async () => {
    try {
      const res = await statisticsApi.getOverview()
      if (res.code === 0) {
        stats.value = res.data
      }
    } catch (error) {
      console.error('Failed to fetch stats:', error)
    }
  }

  const fetchPlatforms = async () => {
    try {
      const res = await cartridgeApi.getPlatforms()
      if (res.code === 0) {
        platforms.value = res.data
      }
    } catch (error) {
      console.error('Failed to fetch platforms:', error)
    }
  }

  const fetchPublishers = async () => {
    try {
      const res = await cartridgeApi.getPublishers()
      if (res.code === 0) {
        publishers.value = res.data
      }
    } catch (error) {
      console.error('Failed to fetch publishers:', error)
    }
  }

  const deleteCartridge = async (id: number) => {
    const res = await cartridgeApi.delete(id)
    if (res.code === 0) {
      cartridges.value = cartridges.value.filter(c => c.id !== id)
      total.value--
      return true
    }
    return false
  }

  return {
    cartridges,
    total,
    loading,
    stats,
    platforms,
    publishers,
    fetchList,
    fetchStats,
    fetchPlatforms,
    fetchPublishers,
    deleteCartridge
  }
})
