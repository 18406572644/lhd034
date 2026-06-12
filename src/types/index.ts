export interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

export interface PagedResponse<T> {
  items: T[]
  total: number
  page: number
  pageSize: number
}

export interface Cartridge {
  id: number
  title: string
  platform: string
  publisher: string
  releaseYear: number
  condition: 'mint' | 'excellent' | 'good' | 'fair' | 'poor'
  purchasePrice: number
  purchaseDate: string
  coverImage: string
  screenshots: string[]
  region: string
  notes: string
  createdAt: string
  updatedAt: string
  playthroughs?: Playthrough[]
  review?: Review | null
  wishlist?: WishlistItem | null
  borrowRecords?: BorrowRecord[]
}

export interface Playthrough {
  id: number
  cartridgeId: number
  startDate: string
  completionDate: string
  playTimeHours: number
  difficultyRating: 1 | 2 | 3 | 4 | 5
  endingType: string
  multipleEndings: boolean
  achievedEndings: string[]
  notes: string
  createdAt: string
  cartridge?: Cartridge
}

export interface Review {
  id: number
  cartridgeId: number
  contentRating: 1 | 2 | 3 | 4 | 5
  gameplayRating: 1 | 2 | 3 | 4 | 5
  graphicsRating: 1 | 2 | 3 | 4 | 5
  soundRating: 1 | 2 | 3 | 4 | 5
  overallRating: number
  reviewText: string
  storyNotes: string
  easterEggs: string[]
  createdAt: string
  cartridge?: Cartridge
}

export interface WishlistItem {
  id: number
  cartridgeId: number
  priority: 'high' | 'medium' | 'low'
  plannedStartDate: string
  tags: string[]
  notes: string
  addedAt: string
  cartridge?: Cartridge
}

export interface BorrowRecord {
  id: number
  cartridgeId: number
  borrowerName: string
  borrowerContact: string
  borrowDate: string
  expectedReturnDate: string
  actualReturnDate: string | null
  status: 'borrowed' | 'returned' | 'overdue'
  notes: string
  createdAt: string
  cartridge?: Cartridge
}

export interface OverviewStats {
  totalCartridges: number
  totalPlaythroughs: number
  totalPlayTime: number
  wishlistCount: number
  borrowedCount: number
  totalValue: number
  newThisMonth: number
  completedThisMonth: number
}

export interface MonthlyData {
  month: string
  count: number
  playTimeHours: number
}

export interface PlatformStat {
  platform: string
  count: number
}

export interface PublisherStat {
  publisher: string
  count: number
}

export interface ConditionStat {
  condition: string
  count: number
}

export const ConditionLabels: Record<string, string> = {
  mint: '全新',
  excellent: '优秀',
  good: '良好',
  fair: '一般',
  poor: '较差'
}

export const PriorityLabels: Record<string, string> = {
  high: '高',
  medium: '中',
  low: '低'
}

export const StatusLabels: Record<string, string> = {
  borrowed: '借出中',
  returned: '已归还',
  overdue: '已逾期'
}

export const PlatformOptions = [
  'NES', 'SNES', 'N64', 'GameCube', 'Wii', 'Wii U', 'Switch',
  'GameBoy', 'GBA', 'DS', '3DS',
  'PS1', 'PS2', 'PS3', 'PS4', 'PS5', 'PSP', 'PS Vita',
  'Xbox', 'Xbox 360', 'Xbox One', 'Xbox Series',
  'MD', 'SS', 'DC',
  'PC Engine', 'Neo Geo', 'Atari', '其他'
]

export const ConditionOptions = [
  { value: 'mint', label: '全新' },
  { value: 'excellent', label: '优秀' },
  { value: 'good', label: '良好' },
  { value: 'fair', label: '一般' },
  { value: 'poor', label: '较差' }
]

export const PriorityOptions = [
  { value: 'high', label: '高优先级' },
  { value: 'medium', label: '中优先级' },
  { value: 'low', label: '低优先级' }
]

export const EndingTypeOptions = [
  '标准结局', '好结局', '坏结局', '真结局', 'A结局', 'B结局', 'C结局', '隐藏结局', '其他'
]
