<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'

const settings = ref<Record<string, string>>({})
const siteSettings = ref({
  name: '',
  logo: '',
  description: '',
  keywords: '',
  theme: 'default',
  primary_color: '#6366f1',
  favicon: '',
  footer: '',
  tos: '',
  privacy: '',
  currency: 'CNY',
  currency_symbol: '¥'
})
const telegramSettings = ref({
  enable: false,
  bot_token: '',
  chat_id: ''
})
const loading = ref(false)
const saving = ref(false)
const activeTab = ref('site')

const tabs = [
  { key: 'site', name: '站点设置', icon: '🌐' },
  { key: 'register', name: '注册设置', icon: '📝' },
  { key: 'mail', name: '邮件设置', icon: '📧' },
  { key: 'telegram', name: 'Telegram', icon: '📱' },
  { key: 'subscribe', name: '订阅设置', icon: '🔗' },
  { key: 'other', name: '其他设置', icon: '⚙️' },
]

const settingGroups: Record<string, Array<{ key: string; label: string; type: string; placeholder?: string; options?: any[] }>> = {
  register: [
    { key: 'register_enable', label: '开放注册', type: 'checkbox' },
    { key: 'register_invite_only', label: '仅限邀请注册', type: 'checkbox' },
    { key: 'mail_verify', label: '邮箱验证', type: 'checkbox' },
    { key: 'register_ip_limit', label: 'IP 注册限制 (0=不限)', type: 'number', placeholder: '0' },
    { key: 'register_trial', label: '新用户试用', type: 'checkbox' },
    { key: 'register_trial_days', label: '试用天数', type: 'number', placeholder: '1' },
    { key: 'register_trial_traffic', label: '试用流量 (GB)', type: 'number', placeholder: '10' },
    { key: 'invite_commission', label: '邀请佣金比例 (%)', type: 'number', placeholder: '10' },
  ],
  mail: [
    { key: 'mail_enable', label: '启用邮件', type: 'checkbox' },
    { key: 'mail_host', label: 'SMTP 服务器', type: 'text', placeholder: 'smtp.example.com' },
    { key: 'mail_port', label: 'SMTP 端口', type: 'text', placeholder: '587' },
    { key: 'mail_username', label: 'SMTP 用户名', type: 'text' },
    { key: 'mail_password', label: 'SMTP 密码', type: 'password' },
    { key: 'mail_encryption', label: '加密方式', type: 'select', options: [{ value: 'tls', label: 'TLS' }, { value: 'ssl', label: 'SSL' }, { value: '', label: '无' }] },
    { key: 'mail_from_address', label: '发件人地址', type: 'text' },
    { key: 'mail_from_name', label: '发件人名称', type: 'text' },
  ],
  subscribe: [
    { key: 'subscribe_url', label: '订阅地址', type: 'text', placeholder: '留空使用站点地址' },
    { key: 'subscribe_single_mode', label: '单节点模式', type: 'checkbox' },
  ],
  other: [
    { key: 'server_push_interval', label: '节点推送间隔 (秒)', type: 'number', placeholder: '60' },
    { key: 'server_pull_interval', label: '节点拉取间隔 (秒)', type: 'number', placeholder: '60' },
    { key: 'traffic_reset_day', label: '流量重置日', type: 'number', placeholder: '1' },
  ],
}

const fetchSettings = async () => {
  loading.value = true
  try {
    const [settingsRes, siteRes, telegramRes] = await Promise.all([
      api.get('/api/v2/admin/settings'),
      api.get('/api/v2/admin/site/settings'),
      api.get('/api/v2/admin/telegram/settings')
    ])
    settings.value = settingsRes.data.data || {}
    if (siteRes.data.data) siteSettings.value = { ...siteSettings.value, ...siteRes.data.data }
    if (telegramRes.data.data) telegramSettings.value = { ...telegramSettings.value, ...telegramRes.data.data }
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const saveSettings = async () => {
  saving.value = true
  try {
    if (activeTab.value === 'site') {
      await api.post('/api/v2/admin/site/settings', siteSettings.value)
    } else if (activeTab.value === 'telegram') {
      await api.post('/api/v2/admin/telegram/settings', telegramSettings.value)
    } else {
      await api.post('/api/v2/admin/settings', settings.value)
    }
    alert('保存成功')
  } catch (e: any) {
    alert(e.response?.data?.error || '保存失败')
  } finally {
    saving.value = false
  }
}

const setTelegramWebhook = async () => {
  const webhookUrl = prompt('请输入 Webhook URL', `${window.location.origin}/api/v1/telegram/webhook`)
  if (!webhookUrl) return
  try {
    await api.post('/api/v2/admin/telegram/webhook', { webhook_url: webhookUrl })
    alert('Webhook 设置成功')
  } catch (e: any) {
    alert(e.response?.data?.error || '设置失败')
  }
}

onMounted(fetchSettings)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">系统设置</h1>
        <p class="text-gray-500 mt-1">配置系统参数</p>
      </div>
      <button @click="saveSettings" :disabled="saving" class="px-4 py-2 bg-indigo-500 text-white rounded-xl hover:bg-indigo-600 disabled:opacity-50">
        {{ saving ? '保存中...' : '保存设置' }}
      </button>
    </div>

    <div v-if="loading" class="text-center py-12 text-gray-500">加载中...</div>

    <div v-else class="flex gap-6">
      <!-- Tabs -->
      <div class="w-48 flex-shrink-0">
        <div class="bg-white rounded-xl shadow-sm p-2 space-y-1">
          <button v-for="tab in tabs" :key="tab.key" @click="activeTab = tab.key" :class="['w-full flex items-center gap-2 px-4 py-3 rounded-lg text-sm transition-colors', activeTab === tab.key ? 'bg-indigo-50 text-indigo-600' : 'text-gray-600 hover:bg-gray-50']">
            <span>{{ tab.icon }}</span>
            <span>{{ tab.name }}</span>
          </button>
        </div>
      </div>

      <!-- Content -->
      <div class="flex-1 bg-white rounded-xl shadow-sm p-6">
        <!-- Site Settings -->
        <div v-if="activeTab === 'site'" class="space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">站点名称</label>
              <input v-model="siteSettings.name" type="text" placeholder="dashGO" class="w-full px-4 py-2 border border-gray-200 rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent"/>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">主题色</label>
              <div class="flex gap-2">
                <input v-model="siteSettings.primary_color" type="color" class="w-12 h-10 rounded-lg border border-gray-200 cursor-pointer"/>
                <input v-model="siteSettings.primary_color" type="text" class="flex-1 px-4 py-2 border border-gray-200 rounded-xl"/>
              </div>
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">站点 Logo URL</label>
            <input v-model="siteSettings.logo" type="text" placeholder="https://example.com/logo.png" class="w-full px-4 py-2 border border-gray-200 rounded-xl"/>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">站点描述</label>
            <textarea v-model="siteSettings.description" rows="2" class="w-full px-4 py-2 border border-gray-200 rounded-xl"/>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">货币单位</label>
              <input v-model="siteSettings.currency" type="text" placeholder="CNY" class="w-full px-4 py-2 border border-gray-200 rounded-xl"/>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">货币符号</label>
              <input v-model="siteSettings.currency_symbol" type="text" placeholder="¥" class="w-full px-4 py-2 border border-gray-200 rounded-xl"/>
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">页脚内容 (HTML)</label>
            <textarea v-model="siteSettings.footer" rows="3" class="w-full px-4 py-2 border border-gray-200 rounded-xl font-mono text-sm"/>
          </div>
        </div>

        <!-- Telegram Settings -->
        <div v-else-if="activeTab === 'telegram'" class="space-y-4">
          <div class="flex items-center justify-between p-4 bg-gray-50 rounded-xl">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-blue-100 flex items-center justify-center text-xl">📱</div>
              <div>
                <p class="font-medium text-gray-900">Telegram Bot</p>
                <p class="text-sm text-gray-500">启用后用户可通过 Bot 管理账户</p>
              </div>
            </div>
            <label class="relative inline-flex items-center cursor-pointer">
              <input v-model="telegramSettings.enable" type="checkbox" class="sr-only peer"/>
              <div class="w-11 h-6 bg-gray-200 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-indigo-500"></div>
            </label>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Bot Token</label>
            <input v-model="telegramSettings.bot_token" type="password" placeholder="从 @BotFather 获取" class="w-full px-4 py-2 border border-gray-200 rounded-xl"/>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">管理员 Chat ID</label>
            <input v-model="telegramSettings.chat_id" type="text" placeholder="用于接收通知" class="w-full px-4 py-2 border border-gray-200 rounded-xl"/>
          </div>
          <button @click="setTelegramWebhook" class="px-4 py-2 border border-gray-200 rounded-xl hover:bg-gray-50">设置 Webhook</button>
        </div>

        <!-- Other Settings -->
        <div v-else class="space-y-4">
          <div v-for="item in settingGroups[activeTab]" :key="item.key">
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ item.label }}</label>
            <input v-if="item.type === 'text' || item.type === 'password' || item.type === 'number'" v-model="settings[item.key]" :type="item.type" :placeholder="item.placeholder" class="w-full px-4 py-2 border border-gray-200 rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent"/>
            <textarea v-else-if="item.type === 'textarea'" v-model="settings[item.key]" rows="3" :placeholder="item.placeholder" class="w-full px-4 py-2 border border-gray-200 rounded-xl"/>
            <select v-else-if="item.type === 'select'" v-model="settings[item.key]" class="w-full px-4 py-2 border border-gray-200 rounded-xl">
              <option v-for="opt in item.options" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
            </select>
            <label v-else-if="item.type === 'checkbox'" class="flex items-center gap-2">
              <input v-model="settings[item.key]" type="checkbox" true-value="1" false-value="0" class="rounded"/>
              <span class="text-sm text-gray-600">启用</span>
            </label>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
