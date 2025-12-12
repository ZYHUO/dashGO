<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'
import api from '@/api'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const inviteCode = ref('')
const emailCode = ref('')
const loading = ref(false)
const sendingCode = ref(false)
const error = ref('')
const siteName = ref('dashGO')
const emailVerifyEnabled = ref(false)
const cooldown = ref(0)

let cooldownTimer: any = null

const fetchSettings = async () => {
  try {
    const res = await api.get('/api/v1/guest/settings')
    if (res.data.data?.site_name) siteName.value = res.data.data.site_name
    emailVerifyEnabled.value = res.data.data?.mail_verify || false
  } catch (e) {}
}

onMounted(() => {
  fetchSettings()
  const code = route.query.code as string
  if (code) inviteCode.value = code
})

const sendEmailCode = async () => {
  if (!email.value) {
    error.value = '请先输入邮箱'
    return
  }
  sendingCode.value = true
  error.value = ''
  try {
    await api.post('/api/v1/guest/send_email_code', { email: email.value })
    cooldown.value = 60
    cooldownTimer = setInterval(() => {
      cooldown.value--
      if (cooldown.value <= 0) {
        clearInterval(cooldownTimer)
      }
    }, 1000)
  } catch (e: any) {
    error.value = e.response?.data?.error || '发送失败'
  } finally {
    sendingCode.value = false
  }
}

const handleRegister = async () => {
  if (!email.value || !password.value) {
    error.value = '请填写邮箱和密码'
    return
  }
  if (password.value !== confirmPassword.value) {
    error.value = '两次输入的密码不一致'
    return
  }
  if (password.value.length < 6) {
    error.value = '密码长度至少6位'
    return
  }
  if (emailVerifyEnabled.value && !emailCode.value) {
    error.value = '请输入邮箱验证码'
    return
  }
  loading.value = true
  error.value = ''
  try {
    await userStore.registerWithCode(email.value, password.value, inviteCode.value || undefined, emailCode.value || undefined)
    router.push('/')
  } catch (e: any) {
    error.value = e.response?.data?.error || '注册失败'
  } finally {
    loading.value = false
  }
}
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
        <p class="text-gray-500 text-sm mt-1">创建新账户</p>
      </div>

      <!-- Form -->
      <div class="bg-white rounded-2xl shadow-sm p-6">
        <form @submit.prevent="handleRegister" class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">邮箱</label>
            <input v-model="email" type="email" placeholder="your@email.com" class="w-full px-4 py-3 rounded-xl border border-gray-200 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 transition-all" autocomplete="email"/>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">密码</label>
            <input v-model="password" type="password" placeholder="至少6位" class="w-full px-4 py-3 rounded-xl border border-gray-200 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 transition-all" autocomplete="new-password"/>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">确认密码</label>
            <input v-model="confirmPassword" type="password" placeholder="再次输入密码" class="w-full px-4 py-3 rounded-xl border border-gray-200 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 transition-all" autocomplete="new-password"/>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">邀请码 <span class="text-gray-400">(可选)</span></label>
            <input v-model="inviteCode" type="text" placeholder="输入邀请码" class="w-full px-4 py-3 rounded-xl border border-gray-200 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 transition-all"/>
          </div>
          <div v-if="emailVerifyEnabled">
            <label class="block text-sm font-medium text-gray-700 mb-1.5">邮箱验证码</label>
            <div class="flex gap-2">
              <input v-model="emailCode" type="text" placeholder="输入验证码" class="flex-1 px-4 py-3 rounded-xl border border-gray-200 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 transition-all"/>
              <button type="button" @click="sendEmailCode" :disabled="sendingCode || cooldown > 0" class="px-4 py-3 bg-gray-100 text-gray-700 rounded-xl text-sm font-medium hover:bg-gray-200 disabled:opacity-50 whitespace-nowrap">
                {{ cooldown > 0 ? `${cooldown}s` : (sendingCode ? '发送中...' : '获取验证码') }}
              </button>
            </div>
          </div>
          <div v-if="error" class="p-3 rounded-xl bg-red-50 text-red-600 text-sm flex items-center gap-2">
            <svg class="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
            {{ error }}
          </div>
          <button type="submit" :disabled="loading" class="w-full py-3 bg-indigo-500 text-white rounded-xl font-medium hover:bg-indigo-600 disabled:opacity-50 transition-colors active:scale-[0.98]">
            {{ loading ? '注册中...' : '注册' }}
          </button>
        </form>
        <div class="mt-5 text-center text-sm text-gray-500">
          已有账户？
          <RouterLink to="/login" class="text-indigo-600 hover:text-indigo-700 font-medium">立即登录</RouterLink>
        </div>
      </div>
    </div>
  </div>
</template>
