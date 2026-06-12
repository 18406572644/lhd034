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
  ConditionStat,
  RatingStat,
  PlayTimeRankItem,
  DifficultyStat,
  ValueTrendItem,
  RegionStat,
  CompletionRate,
  BackupInfo,
  BackupConfig,
  PlayingSession,
  PlayingProgress,
  PlayingCartridgeProgress,
  BatchUpdateRequest,
  BatchUpdatePreview,
  BatchActionResult,
  DataQualityReport,
  DuplicateGroup,
  MissingFieldStat,
  AnomalyItem
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
    status?: string
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
  },
  getSessions(id: number) {
    return api.get<PlayingSession[]>(`/cartridges/${id}/sessions`)
  },
  getProgress(id: number) {
    return api.get<PlayingProgress>(`/cartridges/${id}/progress`)
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

export const sessionApi = {
  getList(params?: { cartridgeId?: number }) {
    return api.get<PlayingSession[]>('/sessions', { params })
  },
  getPlaying() {
    return api.get<PlayingCartridgeProgress[]>('/sessions/playing')
  },
  getProgress(id: number) {
    return api.get<PlayingProgress>(`/sessions/${id}`)
  },
  create(data: Partial<PlayingSession>) {
    return api.post<PlayingSession>('/sessions', data)
  },
  update(id: number, data: Partial<PlayingSession>) {
    return api.put<PlayingSession>(`/sessions/${id}`, data)
  },
  delete(id: number) {
    return api.delete(`/sessions/${id}`)
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
  getList(params?: { status?: string; cartridgeId?: number }) {
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
  },
  getRatings() {
    return api.get<RatingStat[]>('/statistics/ratings')
  },
  getPlayTimeTop10() {
    return api.get<PlayTimeRankItem[]>('/statistics/playtime-top10')
  },
  getDifficulty() {
    return api.get<DifficultyStat[]>('/statistics/difficulty')
  },
  getValueTrend() {
    return api.get<ValueTrendItem[]>('/statistics/value-trend')
  },
  getRegions() {
    return api.get<RegionStat[]>('/statistics/regions')
  },
  getCompletionRate() {
    return api.get<CompletionRate>('/statistics/completion-rate')
  }
}

export const backupApi = {
  getList() {
    return api.get<BackupInfo[]>('/backups')
  },
  createBackup() {
    return api.post<BackupInfo>('/backups')
  },
  deleteBackup(filename: string) {
    return api.delete(`/backups/${filename}`)
  },
  restoreBackup(filename: string) {
    return api.post<{ message: string; snapshotFilename: string }>(`/backups/${filename}/restore`)
  },
  getConfig() {
    return api.get<BackupConfig>('/backups/config')
  },
  updateConfig(config: Partial<BackupConfig>) {
    return api.put<BackupConfig>('/backups/config', config)
  }
}

export const batchApi = {
  updatePreview(data: BatchUpdateRequest) {
    return api.post<BatchUpdatePreview>('/batch/update/preview', data)
  },
  update(data: BatchUpdateRequest) {
    return api.post<BatchActionResult>('/batch/update', data)
  },
  addToWishlist(ids: number[], tags?: string[]) {
    return api.post<BatchActionResult>('/batch/wishlist', { ids, tags })
  },
  setTags(ids: number[], tags: string[], action: 'overwrite' | 'append' = 'append') {
    return api.post<BatchActionResult>('/batch/tags', { ids, tags, action })
  },
  deleteCartridges(ids: number[]) {
    return api.delete<BatchActionResult>('/batch/cartridges', { data: { ids } })
  }
}

export const dataQualityApi = {
  getReport() {
    return api.get<DataQualityReport>('/data-quality/report')
  },
  getDuplicates() {
    return api.get<DuplicateGroup[]>('/data-quality/duplicates')
  },
  getMissingFields() {
    return api.get<MissingFieldStat[]>('/data-quality/missing-fields')
  },
  getAnomalies() {
    return api.get<AnomalyItem[]>('/data-quality/anomalies')
  },
  fixDuplicates(keepId: number) {
    return api.post<BatchActionResult>('/data-quality/fix-duplicates', { keepId })
  },
  fixMissingField(field: string, value: any, ids?: number[]) {
    return api.post<BatchActionResult>('/data-quality/fix-missing', { field, value, ids })
  }
}
