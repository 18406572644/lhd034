import axios from 'axios'
import type { AxiosRequestConfig } from 'axios'
import type {
  ApiResponse,
  PagedResponse,
  Cartridge,
  Playthrough,
  Review,
  WishlistItem,
  BorrowRecord,
  OverviewStats,
  MonthlyData,
  PlatformStat,
  PublisherStat,
  ConditionStat
} from '@/types'

const baseURL = import.meta.env.VITE_API_BASE_URL || '/api'

const request = axios.create({
  baseURL,
  timeout: 10000
})

request.interceptors.response.use(
  (response) => response.data,
  (error) => Promise.reject(error)
)

export const api = {
  async get<T>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return request.get(url, config)
  },
  async post<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return request.post(url, data, config)
  },
  async put<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return request.put(url, data, config)
  },
  async delete<T>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return request.delete(url, config)
  }
}

export const cartridgeApi = {
  getList(params: {
    page?: number
    pageSize?: number
    platform?: string
    publisher?: string
    condition?: string
    year?: string
    search?: string
  }) {
    return api.get<PagedResponse<Cartridge>>('/cartridges', { params })
  },
  getById(id: number) {
    return api.get<Cartridge>(`/cartridges/${id}`)
  },
  create(data: Partial<Cartridge>) {
    return api.post<Cartridge>('/cartridges', data)
  },
  update(id: number, data: Partial<Cartridge>) {
    return api.put<Cartridge>(`/cartridges/${id}`, data)
  },
  delete(id: number) {
    return api.delete(`/cartridges/${id}`)
  },
  upload(file: File) {
    const formData = new FormData()
    formData.append('file', file)
    return api.post<{ url: string; filename: string }>('/cartridges/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },
  getPlatforms() {
    return api.get<string[]>('/cartridges/platforms')
  },
  getPublishers() {
    return api.get<string[]>('/cartridges/publishers')
  },
  getPlaythroughs(id: number) {
    return api.get<Playthrough[]>(`/cartridges/${id}/playthroughs`)
  },
  getReview(id: number) {
    return api.get<Review | null>(`/cartridges/${id}/review`)
  }
}

export const playthroughApi = {
  getList(params?: { page?: number; pageSize?: number; cartridgeId?: number; year?: string; difficulty?: string }) {
    return api.get<PagedResponse<Playthrough>>('/playthroughs', { params })
  },
  getById(id: number) {
    return api.get<Playthrough>(`/playthroughs/${id}`)
  },
  create(data: Partial<Playthrough>) {
    return api.post<Playthrough>('/playthroughs', data)
  },
  update(id: number, data: Partial<Playthrough>) {
    return api.put<Playthrough>(`/playthroughs/${id}`, data)
  },
  delete(id: number) {
    return api.delete(`/playthroughs/${id}`)
  }
}

export const reviewApi = {
  getList() {
    return api.get<Review[]>('/reviews')
  },
  create(data: Partial<Review>) {
    return api.post<Review>('/reviews', data)
  },
  update(id: number, data: Partial<Review>) {
    return api.put<Review>(`/reviews/${id}`, data)
  }
}

export const wishlistApi = {
  getList() {
    return api.get<WishlistItem[]>('/wishlist')
  },
  create(data: Partial<WishlistItem>) {
    return api.post<WishlistItem>('/wishlist', data)
  },
  update(id: number, data: Partial<WishlistItem>) {
    return api.put<WishlistItem>(`/wishlist/${id}`, data)
  },
  delete(id: number) {
    return api.delete(`/wishlist/${id}`)
  }
}

export const borrowApi = {
  getList(params?: { status?: string }) {
    return api.get<BorrowRecord[]>('/borrows', { params })
  },
  getById(id: number) {
    return api.get<BorrowRecord>(`/borrows/${id}`)
  },
  create(data: Partial<BorrowRecord>) {
    return api.post<BorrowRecord>('/borrows', data)
  },
  update(id: number, data: Partial<BorrowRecord>) {
    return api.put<BorrowRecord>(`/borrows/${id}`, data)
  },
  returnRecord(id: number) {
    return api.put<BorrowRecord>(`/borrows/${id}/return`)
  },
  delete(id: number) {
    return api.delete(`/borrows/${id}`)
  }
}

export const statisticsApi = {
  getOverview() {
    return api.get<OverviewStats>('/statistics/overview')
  },
  getAnnual(year?: number) {
    return api.get<MonthlyData[]>('/statistics/annual', { params: { year } })
  },
  getPlatforms() {
    return api.get<PlatformStat[]>('/statistics/platforms')
  },
  getPublishers() {
    return api.get<PublisherStat[]>('/statistics/publishers')
  },
  getConditions() {
    return api.get<ConditionStat[]>('/statistics/conditions')
  }
}
