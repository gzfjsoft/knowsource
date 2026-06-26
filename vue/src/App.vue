<template>
  <router-view />
</template>

<script setup>
import { onMounted, onUnmounted, watch } from 'vue'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
let refreshTokenTimer = null

// 刷新 token
const refreshToken = async () => {
  if (!userStore.isLoggedIn) {
    return
  }
  
  try {
    const result = await userStore.refreshToken()
    if (result.success) {
      console.log('Token refreshed successfully')
    } else {
      console.warn('Token refresh failed:', result.message)
      // 如果刷新失败，可能是 token 已过期，清除登录状态
      if (result.message && result.message.includes('未登录') || result.message.includes('过期')) {
        userStore.logout()
        window.location.href = '/login'
      }
    }
  } catch (error) {
    console.error('Token refresh error:', error)
  }
}

// 启动定时刷新 token（每30分钟一次）
const refreshTokenInterval = 60 * 1000 * 30

const startTokenRefresh = () => {
  // 清除之前的定时器
  if (refreshTokenTimer) {
    clearInterval(refreshTokenTimer)
  }
  
  // 如果用户已登录，启动定时刷新
  if (userStore.isLoggedIn) {
    // 每分钟（30000 毫秒）刷新一次
    refreshTokenTimer = setInterval(refreshToken, refreshTokenInterval)
    console.log('Token refresh timer started (every 30 minute)')
  }
}

// 更新浏览器标题
const updateBrowserTitle = () => {
  const baseTitle = '知源智库 AI'
  if (userStore.isLoggedIn && userStore.userInfo?.ClientName) {
    document.title = `${baseTitle} - ${userStore.userInfo.ClientName}`
  } else {
    document.title = baseTitle
  }
}

// 监听用户登录事件，登录成功后启动定时器并更新标题
const handleUserLoggedIn = () => {
  if (userStore.isLoggedIn && !refreshTokenTimer) {
    startTokenRefresh()
    updateBrowserTitle()
  }
}

onMounted(() => {
  // 页面加载时启动定时刷新
  startTokenRefresh()
  
  // 页面加载时更新标题
  updateBrowserTitle()
  
  // 监听用户登录事件
  window.addEventListener('user-logged-in', handleUserLoggedIn)
  
  // 监听用户信息变化，更新标题
  watch(
    () => userStore.userInfo,
    () => {
      updateBrowserTitle()
    },
    { deep: true }
  )
  
  // 监听登录状态变化，更新标题
  watch(
    () => userStore.isLoggedIn,
    () => {
      updateBrowserTitle()
    }
  )
})

onUnmounted(() => {
  // 组件卸载时清除定时器和移除事件监听
  if (refreshTokenTimer) {
    clearInterval(refreshTokenTimer)
    refreshTokenTimer = null
  }
  window.removeEventListener('user-logged-in', handleUserLoggedIn)
})
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

#app {
  width: 100%;
  height: 100vh;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
}
</style>

