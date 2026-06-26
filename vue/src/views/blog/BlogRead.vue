<template>
  <div class="blog-read-page">
    <TechBackground />
    <div class="blog-read-page__content" v-loading="loading">
      <header class="page-header">
        <router-link class="brand-link" to="/index">
          <span class="brand-mark" />
          <span>知源智库 AI</span>
        </router-link>
        <div class="page-header__actions">
          <router-link to="/blogs">
            <el-button class="btn-ghost">博客列表</el-button>
          </router-link>
          <router-link to="/index">
            <el-button class="btn-ghost">首页</el-button>
          </router-link>
        </div>
      </header>

      <el-card v-if="blog" class="article-card">
        <h1 class="title">{{ blog.title }}</h1>
        <div class="meta">
          <span>作者：{{ blog.authors || "匿名" }}</span>
          <span>更新时间：{{ formatTime(blog.updatedAt) }}</span>
          <span v-if="blog.alias">别名：{{ blog.alias }}</span>
        </div>
        <p v-if="blog.summary" class="summary">{{ blog.summary }}</p>
        <div class="markdown-body article-content" v-html="contentHtml"></div>
      </el-card>
      <el-empty v-else description="博客不存在或未发布" />

      <footer class="page-footer">
        <SiteIcpLine theme="dark" />
      </footer>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { parseMarkdown } from "@/utils/markdown";
import { blogRead } from "@/api/knowsource";
import { ElMessage } from "element-plus";
import TechBackground from "@/components/TechBackground.vue";
import SiteIcpLine from "@/components/SiteIcpLine.vue";

const route = useRoute();
const loading = ref(false);
const blog = ref(null);

const contentHtml = computed(() => {
  const md = blog.value?.content || "";
  return parseMarkdown(md);
});

const loadData = async () => {
  const slug = String(route.params.slug || "").trim();
  if (!slug) return;
  loading.value = true;
  try {
    const res = await blogRead(slug);
    if (res.code === 200 && res.data) {
      blog.value = res.data;
    } else {
      blog.value = null;
      ElMessage.warning(res.message || "文章不存在");
    }
  } catch (e) {
    blog.value = null;
    ElMessage.error(`加载失败: ${e.message || "未知错误"}`);
  } finally {
    loading.value = false;
  }
};

const formatTime = (ts) => {
  if (!ts) return "-";
  return new Date(ts * 1000).toLocaleString("zh-CN");
};

watch(() => route.params.slug, loadData);
onMounted(loadData);
</script>

<style scoped>
.blog-read-page {
  position: relative;
  min-height: 100vh;
  color: rgba(226, 232, 240, 0.95);
}

.blog-read-page__content {
  position: relative;
  z-index: 1;
  max-width: 1000px;
  margin: 0 auto;
  padding: 24px 16px 18px;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.brand-link {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: rgba(226, 232, 240, 0.95);
  text-decoration: none;
  font-weight: 600;
}

.brand-mark {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  background: linear-gradient(120deg, #38bdf8, #a78bfa);
  box-shadow: 0 0 18px rgba(56, 189, 248, 0.55);
}

.page-header__actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn-ghost {
  border: 1px solid rgba(148, 163, 184, 0.28) !important;
  background: rgba(15, 23, 42, 0.35) !important;
  color: rgba(226, 232, 240, 0.95) !important;
}

.article-card {
  margin-top: 14px;
  border-radius: 16px;
  border: 1px solid rgba(148, 163, 184, 0.2);
  background: rgba(2, 6, 23, 0.42);
  backdrop-filter: blur(10px);
}

.title {
  margin: 0;
  font-size: 34px;
  line-height: 1.2;
}
.meta {
  margin-top: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
  color: rgba(148, 163, 184, 0.9);
  font-size: 13px;
}
.summary {
  margin-top: 14px;
  color: rgba(203, 213, 225, 0.95);
  background: rgba(15, 23, 42, 0.5);
  padding: 12px 14px;
  border-radius: 8px;
  border: 1px solid rgba(148, 163, 184, 0.18);
}

.article-content {
  margin-top: 20px;
  padding-top: 18px;
  border-top: 1px solid rgba(148, 163, 184, 0.22);
  color: rgba(226, 232, 240, 0.94);
  line-height: 1.75;
}

.article-content :deep(h1),
.article-content :deep(h2),
.article-content :deep(h3) {
  color: #e2e8f0;
}

.article-content :deep(p),
.article-content :deep(li) {
  color: rgba(226, 232, 240, 0.92);
}

.article-content :deep(p) {
  margin: 0 0 14px;
  text-indent: 2em;
}

.article-content :deep(ul),
.article-content :deep(ol) {
  margin: 10px 0 14px;
  padding-left: 1.6em;
}

.article-content :deep(li) {
  margin: 6px 0;
  padding-left: 0.2em;
}

.article-content :deep(h1),
.article-content :deep(h2),
.article-content :deep(h3),
.article-content :deep(h4),
.article-content :deep(h5),
.article-content :deep(h6),
.article-content :deep(pre),
.article-content :deep(blockquote),
.article-content :deep(table),
.article-content :deep(ul),
.article-content :deep(ol) {
  text-indent: 0;
}

.article-content :deep(a) {
  color: #7dd3fc;
}

.article-content :deep(blockquote) {
  border-left: 3px solid rgba(56, 189, 248, 0.55);
  margin: 12px 0;
  padding: 8px 12px;
  color: rgba(186, 230, 253, 0.94);
  background: rgba(15, 23, 42, 0.42);
}

.article-content :deep(pre) {
  background: rgba(15, 23, 42, 0.62);
  border: 1px solid rgba(148, 163, 184, 0.24);
  padding: 12px;
  border-radius: 8px;
  overflow: auto;
}

.article-content :deep(code) {
  background: rgba(15, 23, 42, 0.62);
  padding: 2px 4px;
  border-radius: 4px;
}

.article-content :deep(table) {
  border-collapse: collapse;
  width: 100%;
}

.article-content :deep(th),
.article-content :deep(td) {
  border: 1px solid rgba(148, 163, 184, 0.24);
  padding: 8px 10px;
}

.page-footer {
  margin-top: 16px;
  text-align: center;
}

@media (max-width: 720px) {
  .title {
    font-size: 28px;
  }
}
</style>
