<template>
  <div class="install-page">
    <TechBackground />

    <div class="install-page__content">
      <header class="install-page__header">
        <el-button class="btn-ghost" @click="goIndex">
          <el-icon><ArrowLeft /></el-icon>
          返回首页
        </el-button>
        <h1 class="install-page__title">安装部署指南</h1>
      </header>

      <article
        v-loading="loading"
        class="install-doc markdown-body"
        v-html="htmlContent"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { ArrowLeft } from "@element-plus/icons-vue";
import { parseMarkdown } from "@/utils/markdown";
import TechBackground from "@/components/TechBackground.vue";
import installMd from "@/docs/INSTALL.md?raw";

const router = useRouter();
const loading = ref(true);
const htmlContent = ref("");

const goIndex = () => router.push("/index");

onMounted(() => {
  try {
    htmlContent.value = parseMarkdown(installMd);
  } finally {
    loading.value = false;
  }
});
</script>

<style scoped>
.install-page {
  position: relative;
  min-height: 100vh;
  color: rgba(226, 232, 240, 0.95);
}

.install-page__content {
  position: relative;
  z-index: 1;
  max-width: 920px;
  margin: 0 auto;
  padding: 28px 22px 48px;
}

.install-page__header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.install-page__title {
  margin: 0;
  font-size: 22px;
  font-weight: 700;
  letter-spacing: 0.2px;
}

.btn-ghost {
  border: 1px solid rgba(148, 163, 184, 0.28) !important;
  background: rgba(15, 23, 42, 0.42) !important;
  color: rgba(226, 232, 240, 0.95) !important;
  backdrop-filter: blur(10px);
}

.install-doc {
  padding: 28px 32px 36px;
  border-radius: 16px;
  border: 1px solid rgba(148, 163, 184, 0.24);
  background: rgba(2, 6, 23, 0.55);
  backdrop-filter: blur(12px);
  box-shadow: 0 30px 80px rgba(0, 0, 0, 0.45);
  line-height: 1.65;
  word-wrap: break-word;
}

.install-doc :deep(h1) {
  font-size: 1.6em;
  margin: 0 0 0.6em;
  padding-bottom: 0.35em;
  border-bottom: 1px solid rgba(148, 163, 184, 0.2);
}
.install-doc :deep(h2) {
  font-size: 1.25em;
  margin: 1.4em 0 0.5em;
  color: #38bdf8;
}
.install-doc :deep(h3) {
  font-size: 1.08em;
  margin: 1.1em 0 0.4em;
}
.install-doc :deep(p),
.install-doc :deep(li) {
  color: rgba(203, 213, 225, 0.95);
  font-size: 14px;
}
.install-doc :deep(ul),
.install-doc :deep(ol) {
  margin: 0.5em 0;
  padding-left: 1.5em;
}
.install-doc :deep(hr) {
  border: none;
  border-top: 1px solid rgba(148, 163, 184, 0.2);
  margin: 1.5em 0;
}
.install-doc :deep(a) {
  color: #38bdf8;
}
.install-doc :deep(code) {
  background: rgba(15, 23, 42, 0.65);
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 4px;
  padding: 0.15em 0.35em;
  font-size: 0.88em;
}
.install-doc :deep(pre) {
  background: rgba(15, 23, 42, 0.75);
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 8px;
  padding: 14px 16px;
  overflow: auto;
  margin: 0.75em 0;
}
.install-doc :deep(pre code) {
  background: none;
  border: none;
  padding: 0;
  font-size: 12px;
  line-height: 1.5;
}
.install-doc :deep(table) {
  border-collapse: collapse;
  width: 100%;
  margin: 0.75em 0;
  font-size: 13px;
}
.install-doc :deep(th),
.install-doc :deep(td) {
  border: 1px solid rgba(148, 163, 184, 0.25);
  padding: 8px 10px;
  text-align: left;
}
.install-doc :deep(th) {
  background: rgba(15, 23, 42, 0.5);
}
.install-doc :deep(blockquote) {
  margin: 0.75em 0;
  padding: 0.5em 1em;
  border-left: 3px solid #38bdf8;
  background: rgba(56, 189, 248, 0.08);
  color: rgba(203, 213, 225, 0.9);
}

@media (max-width: 620px) {
  .install-page__content {
    padding: 20px 14px 32px;
  }
  .install-doc {
    padding: 20px 16px 28px;
  }
}
</style>
