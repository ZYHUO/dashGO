<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import api from '@/api'

const userStore = useUserStore()
const route = useRoute()
const router = useRouter()
const isSidebarOpen = ref(false)
const siteSettings = ref<any>({})

const navItems = [
  { path: '/', name: '仪表盘', icon: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6' },
  { path: '/subscribe', name: '订阅', icon: 'M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1' },
  { path: '/plans', name: '套餐', icon: 'M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10' },
  { path: '/orders', name: '订单', icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2' },
  { path: '/tickets', name: '工单', icon: 'M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z' },
  { path: '/invite', name: '邀请', icon: 'M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z' },
  { path: '/knowledge', name: '帮助', icon: 'M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253' },
  { path: '/settings', name: '设置', icon: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z' },
]

const isActive = (path: string) => {
  if (path === '/') return route.path === '/'
  return route.path.startsWith(path)
}

const siteName = computed(() => siteSettings.value.site_name || 'dashGO')
const primaryColor = computed(() => siteSettings.value.primary_color || '#6366f1')

const fetchSettings = async () => {
  try {
    const res = await api.get('/api/v1/guest/settings')
    siteSettings.value = res.data.data || {}
  } catch (e) {}
}

onMounted(fetchSettings)

watch(() => route.path, () => {
  isSidebarOpen.value = false
})
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Mobile Header -->
    <header class="lg:hidden fixed top-0 left-0 right-0 z-50 bg-white/95 backdrop-blur-md border-b border-gray-100 safe-area-top">
      <div class="flex items-center justify-between px-4 h-14">
        <button @click="isSidebarOpen = true" class="p-2 -ml-2 rounded-lg hover:bg-gray-100 active:bg-gray-200 transition-colors">
          <svg class="w-6 h-6 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"/>
          </svg>
        </button>
        <span class="font-semibold text-gray-900">{{ siteName }}</span>
        <RouterLink to="/settings" class="p-2 -mr-2 rounded-lg hover:bg-gray-100">
          <div class="w-8 h-8 rounded-full bg-gradient-to-br from-indigo-500 to-purple-500 flex items-center justify-center text-white text-sm font-medium">
            {{ userStore.user?.email?.charAt(0).toUpperCase() }}
          </div>
        </RouterLink>
      </div>
    </header>

    <!-- Sidebar Overlay -->
    <Transition name="fade">
      <div v-if="isSidebarOpen" class="fixed inset-0 z-40 bg-black/30 lg:hidden" @click="isSidebarOpen = false"/>
    </Transition>

    <!-- Sidebar -->
    <aside :class="['fixed top-0 left-0 z-50 h-full w-72 bg-white transition-transform duration-300 ease-out lg:translate-x-0 shadow-xl lg:shadow-none', isSidebarOpen ? 'translate-x-0' : '-translate-x-full']">
      <div class="flex flex-col h-full">
        <!-- Logo -->
        <div class="flex items-center gap-3 px-5 h-16 border-b border-gray-100">
          <div class="w-9 h-9 rounded-xl bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center text-white font-bold text-sm shadow-lg shadow-indigo-500/30">
            X
          </div>
          <span class="text-lg font-bold text-gray-900">{{ siteName }}</span>
        </div>

        <!-- Navigation -->
        <nav class="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
          <RouterLink
            v-for="item in navItems"
            :key="item.path"
            :to="item.path"
            :class="['flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all duration-200', isActive(item.path) ? 'bg-indigo-50 text-indigo-600' : 'text-gray-600 hover:bg-gray-50']"
          >
            <svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" :d="item.icon"/>
            </svg>
            <span class="font-medium text-sm">{{ item.name }}</span>
          </RouterLink>
          
          <!-- Admin -->
          <RouterLink
            v-if="userStore.isAdmin"
            to="/admin"
            :class="['flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all duration-200 mt-4 pt-4 border-t border-gray-100', route.path.startsWith('/admin') ? 'bg-red-50 text-red-600' : 'text-gray-600 hover:bg-gray-50']"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
            </svg>
            <span class="font-medium text-sm">管理后台</span>
          </RouterLink>
        </nav>

        <!-- User Info -->
        <div class="p-3 border-t border-gray-100">
          <div class="flex items-center gap-3 p-3 rounded-xl bg-gray-50">
            <div class="w-10 h-10 rounded-full bg-gradient-to-br from-indigo-500 to-purple-500 flex items-center justify-center text-white font-medium shadow-md">
              {{ userStore.user?.email?.charAt(0).toUpperCase() }}
            </div>
            <div class="flex-1 min-w-0">
              <p class="text-sm font-medium text-gray-900 truncate">{{ userStore.user?.email }}</p>
              <p class="text-xs text-gray-500">{{ userStore.isAdmin ? '管理员' : '用户' }}</p>
            </div>
          </div>
          <button @click="userStore.logout()" class="w-full mt-2 px-4 py-2 text-sm text-gray-500 hover:text-red-500 hover:bg-red-50 rounded-xl transition-colors">
            退出登录
          </button>
        </div>
      </div>
    </aside>

    <!-- Main Content -->
    <main class="lg:ml-72 pt-14 lg:pt-0 min-h-screen">
      <div class="p-4 lg:p-6 max-w-6xl mx-auto">
        <RouterView />
      </div>
    </main>

    <!-- Mobile Bottom Nav -->
    <nav class="lg:hidden fixed bottom-0 left-0 right-0 bg-white/95 backdrop-blur-md border-t border-gray-100 safe-area-bottom z-40">
      <div class="flex items-center justify-around h-16">
        <RouterLink v-for="item in navItems.slice(0, 5)" :key="item.path" :to="item.path" :class="['flex flex-col items-center gap-1 px-3 py-2 rounded-lg transition-colors', isActive(item.path) ? 'text-indigo-600' : 'text-gray-400']">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" :d="item.icon"/>
          </svg>
          <span class="text-xs">{{ item.name }}</span>
        </RouterLink>
      </div>
    </nav>
  </div>
</template>

<style scoped>
.fade-enter-active, .fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
}
.safe-area-top {
  padding-top: env(safe-area-inset-top);
}
.safe-area-bottom {
  padding-bottom: env(safe-area-inset-bottom);
}
</style>
