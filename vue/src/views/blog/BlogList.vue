<template>
  <div class="blog-list-page">
    <TechBackground />
    <div class="blog-list-page__content">
      <header class="page-header">
        <router-link class="brand-link" to="/index">
          <span class="brand-mark" />
          <span>知源智库 AI</span>
        </router-link>
        <div class="page-header__actions">
          <el-button class="btn-ghost" @click="$router.push('/index')">首页</el-button>
          <el-button class="btn-ghost" @click="$router.push('/login')">登录</el-button>
        </div>
      </header>

      <section class="hero-card">
        <h1>博客文章</h1>
        <p>产品动态、技术实践与落地案例，持续更新。</p>
        <div class="search-bar">
          <el-input
            v-model="keyword"
            placeholder="搜索标题/摘要"
            clearable
            @keyup.enter="loadData"
            @clear="loadData"
          />
          <el-button type="primary" @click="loadData" :loading="loading">搜索</el-button>
        </div>
      </section>

      <section class="cards" v-loading="loading">
        <el-empty v-if="!rows.length && !loading" description="暂无文章" />
        <article v-for="row in rows" :key="row.id" class="blog-card">
          <div class="blog-card__meta">
            <span>{{ formatTime(row.updatedAt) }}</span>
            <span v-if="row.alias" class="alias">/{{ row.alias }}</span>
          </div>
          <router-link class="blog-card__title" :to="`/blog/${row.alias || row.id}`">
            {{ row.title }}
          </router-link>
          <p class="blog-card__summary">{{ row.summary || "暂无摘要" }}</p>
          <div class="blog-card__footer">
            <router-link :to="`/blog/${row.alias || row.id}`">
              <el-button text type="primary">阅读全文</el-button>
            </router-link>
          </div>
        </article>
      </section>

      <div class="pager">
        <el-pagination
          background
          layout="total, prev, pager, next, sizes"
          :total="total"
          :current-page="page"
          :page-size="pageSize"
          :page-sizes="[10, 20, 50]"
          @current-change="handlePageChange"
          @size-change="handleSizeChange"
        />
      </div>

      <footer class="page-footer">
        <SiteIcpLine theme="dark" />
      </footer>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from "vue";
import { blogListPublic } from "@/api/knowsource";
import { ElMessage } from "element-plus";
import TechBackground from "@/components/TechBackground.vue";
import SiteIcpLine from "@/components/SiteIcpLine.vue";

const loading = ref(false);
const rows = ref([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(10);
const keyword = ref("");

const loadData = async () => {
  loading.value = true;
  try {
    const res = await blogListPublic({
      page: page.value,
      pageSize: pageSize.value,
      keyword: keyword.value,
    });
    if (res.code === 200 && res.data) {
      rows.value = res.data.list || [];
      total.value = res.data.total || 0;
    } else {
      ElMessage.error(res.message || "加载博客失败");
    }
  } catch (e) {
    ElMessage.error(`加载博客失败: ${e.message || "未知错误"}`);
  } finally {
    loading.value = false;
  }
};

const handlePageChange = (p) => {
  page.value = p;
  loadData();
};

const handleSizeChange = (s) => {
  pageSize.value = s;
  page.value = 1;
  loadData();
};

const formatTime = (ts) => {
  if (!ts) return "-";
  return new Date(ts * 1000).toLocaleString("zh-CN");
};

onMounted(loadData);
</script>

<style scoped>
.blog-list-page {
  position: relative;
  min-height: 100vh;
  color: rgba(226, 232, 240, 0.95);
}

.blog-list-page__content {
  position: relative;
  z-index: 1;
  max-width: 1120px;
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

.hero-card {
  margin-top: 14px;
  border-radius: 16px;
  border: 1px solid rgba(148, 163, 184, 0.2);
  background: rgba(2, 6, 23, 0.38);
  backdrop-filter: blur(10px);
  padding: 18px;
}

.hero-card h1 {
  margin: 0;
  font-size: 30px;
}

.hero-card p {
  margin: 8px 0 0;
  color: rgba(148, 163, 184, 0.95);
}

.search-bar {
  margin-top: 14px;
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 10px;
  max-width: 520px;
}

.cards {
  margin-top: 16px;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.blog-card {
  border-radius: 14px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  background: rgba(2, 6, 23, 0.36);
  backdrop-filter: blur(10px);
  padding: 14px;
  min-height: 190px;
  display: flex;
  flex-direction: column;
}

.blog-card__meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: rgba(148, 163, 184, 0.9);
  font-size: 12px;
}

.blog-card__title {
  margin-top: 10px;
  color: #bae6fd;
  text-decoration: none;
  font-size: 18px;
  font-weight: 700;
  line-height: 1.3;
}

.blog-card__title:hover {
  text-decoration: underline;
}

.blog-card__summary {
  margin-top: 10px;
  color: rgba(203, 213, 225, 0.92);
  font-size: 13px;
  line-height: 1.6;
  flex: 1;
}

.blog-card__footer {
  margin-top: 8px;
  display: flex;
  justify-content: flex-end;
}

.alias {
  color: #a78bfa;
}

.pager {
  margin-top: 14px;
  display: flex;
  justify-content: flex-end;
}

.page-footer {
  margin-top: 16px;
  text-align: center;
}

@media (max-width: 1024px) {
  .cards {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 720px) {
  .cards {
    grid-template-columns: 1fr;
  }
  .hero-card h1 {
    font-size: 24px;
  }
  .search-bar {
    grid-template-columns: 1fr;
  }
}
</style>
