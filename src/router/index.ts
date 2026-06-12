import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    component: () => import('@/layouts/MainLayout.vue'),
    children: [
      { path: '', name: 'dashboard', component: () => import('@/pages/Dashboard.vue') },
      { path: 'cartridges', name: 'cartridges', component: () => import('@/pages/CartridgeList.vue') },
      { path: 'cartridges/new', name: 'cartridge-new', component: () => import('@/pages/CartridgeForm.vue') },
      { path: 'cartridges/:id', name: 'cartridge-detail', component: () => import('@/pages/CartridgeDetail.vue') },
      { path: 'cartridges/:id/edit', name: 'cartridge-edit', component: () => import('@/pages/CartridgeForm.vue') },
      { path: 'playthroughs', name: 'playthroughs', component: () => import('@/pages/PlaythroughList.vue') },
      { path: 'playthroughs/new', name: 'playthrough-new', component: () => import('@/pages/PlaythroughForm.vue') },
      { path: 'playthroughs/:id/edit', name: 'playthrough-edit', component: () => import('@/pages/PlaythroughForm.vue') },
      { path: 'progress', name: 'progress', component: () => import('@/pages/ProgressTracker.vue') },
      { path: 'wishlist', name: 'wishlist', component: () => import('@/pages/Wishlist.vue') },
      { path: 'showcase', name: 'showcase', component: () => import('@/pages/Showcase.vue') },
      { path: 'statistics', name: 'statistics', component: () => import('@/pages/Statistics.vue') },
      { path: 'borrows', name: 'borrows', component: () => import('@/pages/BorrowList.vue') }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
