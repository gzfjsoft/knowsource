<template>
  <div :class="['site-icp-line', `site-icp-line--${theme}`]">
    <span>广州丰捷软件技术有限公司 版权所有(C) 2026 电话:15622302051</span>
     
    <span class="sep">·</span>
    <a
      href="https://beian.miit.gov.cn/#/Integrated/index"
      target="_blank"
      rel="noopener noreferrer"
    >粤ICP备15110461号-6</a>
  </div>
</template>

<script setup>
defineProps({
  /** light: 浅色背景页；dark: 落地页深色背景 */
  theme: {
    type: String,
    default: 'light',
    validator: (v) => ['light', 'dark'].includes(v)
  }
})
</script>

<style scoped>
.site-icp-line {
  font-size: 12px;
  line-height: 1.6;
  text-align: center;
}

.site-icp-line .sep {
  margin: 0 6px;
  opacity: 0.7;
}

.site-icp-line a {
  text-decoration: none;
}

.site-icp-line a:hover {
  text-decoration: underline;
}

.site-icp-line--light {
  color: #64748b;
}

.site-icp-line--light a {
  color: #409eff;
}

.site-icp-line--dark {
  color: rgba(148, 163, 184, 0.95);
}

.site-icp-line--dark a {
  color: rgba(186, 230, 253, 0.95);
}
</style>
