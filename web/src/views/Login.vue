<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import api from '@/api'

const router = useRouter()
const userStore = useUserStore()

const email = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')
const siteName = ref('dashGO')

const fetchSettings = async () => {
  try {
    const res = await api.get('/api/v1/guest/settings')
    if (res.data.data?.site_name) siteName.value = res.data.data.site_name
  } catch (e) {}
}

const handleLogin = async () => {
  if (!email.value || !password.value) {
    error.value = '请填写邮箱和密码'
    return
  }
  loading.value = true
  error.value = ''
  try {
    await userStore.login(email.value, password.value)
    // 根据用户角色跳转
    if (userStore.isAdmin) {
      router.push('/admin')
    } else {
      router.push('/')
    }
  } catch (e: any) {
    error.value = e.response?.data?.error || '登录失败'
  } finally {
    loading.value = false
  }
}

onMounted(fetchSettings)
</script>

<template>
  <div class="min-h-screen flex items-center justify-center p-4 bg-gray-50">
    <div class="w-full max-w-sm">
      <!-- Logo -->
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-14 h-14 rounded-2xl bg-gradient-to-br from-indigo-500 to-purple-600 text-white text-xl font-bold shadow-lg shadow-indigo-500/30 mb-4">
          X
        </div>
        <h1 class="text-2xl font-bold text-gray-900">{{ siteName }}</h1>
        <p class="text-gray-500 text-sm mt-1">登录您的账户</p>
      </div>

      <!-- Form -->
      <div class="bg-white rounded-2xl shadow-sm p-6">
        <form @submit.prevent="handleLogin" class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">邮箱</label>
            <input v-model="email" type="email" placeholder="your@email.com" class="w-full px-4 py-3 rounded-xl border border-gray-200 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 transition-all" autocomplete="email"/>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">密码</label>
            <input v-model="password" type="password" placeholder="••••••••" class="w-full px-4 py-3 rounded-xl border border-gray-200 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 transition-all" autocomplete="current-password"/>
          </div>
          <div v-if="error" class="p-3 rounded-xl bg-red-50 text-red-600 text-sm flex items-center gap-2">
            <svg class="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
            {{ error }}
          </div>
          <button type="submit" :disabled="loading" class="w-full py-3 bg-indigo-500 text-white rounded-xl font-medium hover:bg-indigo-600 disabled:opacity-50 transition-colors active:scale-[0.98]">
            {{ loading ? '登录中...' : '登录' }}
          </button>
        </form>
        <div class="mt-5 text-center text-sm text-gray-500">
          还没有账户？
          <RouterLink to="/register" class="text-indigo-600 hover:text-indigo-700 font-medium">立即注册</RouterLink>
        </div>
      </div>
    </div>
  </div>
</template>
