import { createRouter, createWebHistory, type RouteLocationNormalized, type NavigationGuardNext } from 'vue-router'
import { useUserStore } from '@/stores/user'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: () => import('@/layouts/MainLayout.vue'),
      children: [
        {
          path: '',
          name: 'Dashboard',
          component: () => import('@/views/Dashboard.vue'),
          meta: { requiresAuth: true }
        },
        {
          path: 'subscribe',
          name: 'Subscribe',
          component: () => import('@/views/Subscribe.vue'),
          meta: { requiresAuth: true }
        },
        {
          path: 'plans',
          name: 'Plans',
          component: () => import('@/views/Plans.vue'),
          meta: { requiresAuth: true }
        },
        {
          path: 'orders',
          name: 'Orders',
          component: () => import('@/views/Orders.vue'),
          meta: { requiresAuth: true }
        },
        {
          path: 'tickets',
          name: 'Tickets',
          component: () => import('@/views/Tickets.vue'),
          meta: { requiresAuth: true }
        },
        {
          path: 'settings',
          name: 'Settings',
          component: () => import('@/views/Settings.vue'),
          meta: { requiresAuth: true }
        },
        {
          path: 'invite',
          name: 'Invite',
          component: () => import('@/views/Invite.vue'),
          meta: { requiresAuth: true }
        },
        {
          path: 'knowledge',
          name: 'Knowledge',
          component: () => import('@/views/Knowledge.vue'),
          meta: { requiresAuth: true }
        },
      ]
    },
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/Login.vue')
    },
    {
      path: '/register',
      name: 'Register',
      component: () => import('@/views/Register.vue')
    },
    // Admin routes
    {
      path: '/admin',
      component: () => import('@/layouts/AdminLayout.vue'),
      meta: { requiresAuth: true, requiresAdmin: true },
      children: [
        {
          path: '',
          name: 'AdminDashboard',
          component: () => import('@/views/admin/Dashboard.vue')
        },
        {
          path: 'users',
          name: 'AdminUsers',
          component: () => import('@/views/admin/Users.vue')
        },
        {
          path: 'servers',
          name: 'AdminServers',
          component: () => import('@/views/admin/Servers.vue')
        },
        {
          path: 'plans',
          name: 'AdminPlans',
          component: () => import('@/views/admin/Plans.vue')
        },
        {
          path: 'orders',
          name: 'AdminOrders',
          component: () => import('@/views/admin/Orders.vue')
        },
        {
          path: 'tickets',
          name: 'AdminTickets',
          component: () => import('@/views/admin/Tickets.vue')
        },
        {
          path: 'settings',
          name: 'AdminSettings',
          component: () => import('@/views/admin/Settings.vue')
        },
        {
          path: 'coupons',
          name: 'AdminCoupons',
          component: () => import('@/views/admin/Coupons.vue')
        },
        {
          path: 'notices',
          name: 'AdminNotices',
          component: () => import('@/views/admin/Notices.vue')
        },
        {
          path: 'hosts',
          name: 'AdminHosts',
          component: () => import('@/views/admin/Hosts.vue')
        },
        {
          path: 'traffic',
          name: 'AdminTraffic',
          component: () => import('@/views/admin/Traffic.vue')
        },
        {
          path: 'server-groups',
          name: 'AdminServerGroups',
          component: () => import('@/views/admin/ServerGroups.vue')
        },
        {
          path: 'user-groups',
          name: 'AdminUserGroups',
          component: () => import('@/views/admin/UserGroups.vue')
        },
        {
          path: 'traffic-management',
          name: 'AdminTrafficManagement',
          component: () => import('@/views/admin/TrafficManagement.vue')
        },
      ]
    }
  ]
})

router.beforeEach(async (to: RouteLocationNormalized, _from: RouteLocationNormalized, next: NavigationGuardNext) => {
  const userStore = useUserStore()
  
  // 如果访问登录或注册页，且已登录，重定向到首页
  if ((to.path === '/login' || to.path === '/register') && userStore.isLoggedIn) {
    next('/')
    return
  }
  
  // 如果有 token 但没有用户信息，先获取用户信息
  if (userStore.token && !userStore.user) {
    try {
      await userStore.fetchUser()
    } catch (e) {
      // 获取失败，清除 token（已在 fetchUser 中处理）
      console.error('Failed to fetch user:', e)
    }
  }
  
  // 需要认证的页面
  if (to.meta.requiresAuth && !userStore.isLoggedIn) {
    next({ path: '/login', query: { redirect: to.fullPath } })
    return
  }
  
  // 需要管理员权限的页面
  if (to.meta.requiresAdmin && !userStore.isAdmin) {
    next('/')
    return
  }
  
  next()
})

export default router
