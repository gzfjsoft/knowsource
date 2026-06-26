import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import { ElMessage } from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import 'element-plus/dist/index.css'
import '@/assets/markdown-body.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import { loader } from '@guolao/vue-monaco-editor'
import App from './App.vue'
import router from './router'

// 错误信息保留时间翻倍（默认约 3s，改为 6s）
const defaultErrorDuration = 6000
const _error = ElMessage.error
ElMessage.error = (options) => {
  const opts = typeof options === 'string' ? { message: options } : { ...options }
  return _error({ ...opts, duration: opts.duration ?? defaultErrorDuration })
}

// 配置 Monaco 从本地路径加载，而不是从 CDN
loader.config({
  paths: {
    vs: '/monaco/vs'
  }
})

const app = createApp(App)
const pinia = createPinia()

// 注册所有图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.use(pinia)
app.use(router)
app.use(ElementPlus, { locale: zhCn })

app.mount('#app')

