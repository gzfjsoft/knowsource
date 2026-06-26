<template>
  <div class="landing">
    <TechBackground />

    <div class="landing__content">
      <header class="landing__header">
        <div class="brand">
          <div class="brand__mark" />
          <div class="brand__text">
            <div class="brand__name">知源智库 AI</div>
            <div class="brand__tag">企业知识沉淀 · 检索 · 问答 · 运营</div>
          </div>
        </div>

        <div class="actions">
          <el-button class="btn-ghost" size="large" @click="goBlogs">博客</el-button>
          <el-button class="btn-ghost" size="large" @click="goInstall">安装指南</el-button>
          <el-button class="btn-ghost" size="large" @click="goLogin">立即登录</el-button>
        </div>
      </header>

      <main class="hero">
        <div class="hero__left">
          <div class="kicker">现代科技企业级知识平台</div>
          <h1 class="title">
            把文档变成可搜索、可对话、可治理的
            <span class="title__accent">企业知识资产</span>
          </h1>
          <p class="subtitle">
            统一管理原始文档与知识库类型，支持库内全文检索与 AI 问答，配套权限与运维能力，面向企业的安全与可控。
          </p>

          <div class="cta">
            <el-button type="primary" size="large" class="btn-primary" @click="goLogin">
              开始使用
            </el-button>
            <el-button size="large" class="btn-secondary" @click="scrollToFeatures">
              了解功能
            </el-button>
          </div>

          <div class="highlights">
            <div class="pill"><span class="dot dot--cyan" />权限与角色</div>
            <div class="pill"><span class="dot dot--violet" />全文检索</div>
            <div class="pill"><span class="dot dot--green" />AI 问答</div>
            <div class="pill"><span class="dot dot--amber" />运维监控</div>
          </div>
        </div>

        <div class="hero__right" role="presentation">
          <div class="glass">
            <div class="glass__top">
              <div class="signal">
                <span class="signal__dot dot--red" />
                <span class="signal__dot dot--amber" />
                <span class="signal__dot dot--green" />
              </div>
              <div class="glass__title">知识库实时预览</div>
            </div>
            <div class="glass__body">
              <div class="line shimmer" style="width: 72%" />
              <div class="line shimmer" style="width: 56%" />
              <div class="line shimmer" style="width: 84%" />
              <div class="spacer" />
              <div class="chip-row">
                <span class="chip">文档管理</span>
                <span class="chip">标签</span>
                <span class="chip">知识库类型</span>
                <span class="chip">检索</span>
              </div>
              <div class="spacer" />
              <div class="bubble bubble--user">
                这份制度里关于请假审批流程怎么写？
              </div>
              <div class="bubble bubble--ai">
                我已为你从“人事制度/流程”中定位到相关段落，并给出审批路径与注意事项。
              </div>
            </div>
          </div>
        </div>
      </main>

      <section ref="featuresRef" class="features">
        <div class="features__head">
          <h2 class="features__title">产品功能</h2>
          <p class="features__desc">从沉淀到检索，从对话到治理，覆盖企业知识全链路。</p>
        </div>

        <div class="grid">
          <div class="card" v-for="f in features" :key="f.title">
            <div class="card__icon">{{ f.icon }}</div>
            <div class="card__title">{{ f.title }}</div>
            <div class="card__desc">{{ f.desc }}</div>
            <div class="card__meta">{{ f.meta }}</div>
          </div>
        </div>
      </section>

      <footer class="footer">
        <div class="footer__line" />
        <SiteIcpLine theme="dark" />
      </footer>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import TechBackground from "@/components/TechBackground.vue";
import SiteIcpLine from "@/components/SiteIcpLine.vue";

const router = useRouter();
const featuresRef = ref(null);

const features = [
  {
    icon: "📁",
    title: "文档管理",
    desc: "集中管理原始文档、内容审核与维护。",
    meta: "面向多租户与企业场景",
  },
  {
    icon: "🔎",
    title: "库内全文检索",
    desc: "快速定位段落、附件与关键字段，降低搜索成本。",
    meta: "支持大规模内容",
  },
  {
    icon: "💬",
    title: "知识库问答",
    desc: "自然语言提问，结合知识库内容给出可追溯的答案。",
    meta: "更快获得结论",
  },
  {
    icon: "🧩",
    title: "标签与分类",
    desc: "通过标签与类型管理体系，提升检索与治理效率。",
    meta: "结构化沉淀",
  },
  {
    icon: "🛡️",
    title: "权限与角色",
    desc: "按角色、员工、部门授权，确保安全与可控。",
    meta: "企业级权限模型",
  },
  {
    icon: "🧰",
    title: "运维与统计",
    desc: "AI 调用统计、对话日志与依赖检查，便于稳定运行。",
    meta: "可观测与可运营",
  },
];

const goLogin = () => router.push("/login");
const goBlogs = () => router.push("/blogs");
const goInstall = () => router.push("/install");
const scrollToFeatures = () =>
  featuresRef.value?.scrollIntoView({ behavior: "smooth", block: "start" });
</script>

<style scoped>
.landing {
  position: relative;
  min-height: 100vh;
  color: rgba(226, 232, 240, 0.95);
}

.landing__content {
  position: relative;
  z-index: 1;
  padding: 28px 22px 14px;
  max-width: 1120px;
  margin: 0 auto;
}

.landing__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.brand {
  display: flex;
  align-items: center;
  gap: 12px;
}

.brand__mark {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  background:
    radial-gradient(circle at 30% 30%, rgba(56, 189, 248, 0.9), rgba(167, 139, 250, 0.65) 55%, rgba(34, 197, 94, 0.35) 100%);
  box-shadow: 0 10px 30px rgba(56, 189, 248, 0.2);
}

.brand__name {
  font-size: 16px;
  font-weight: 700;
  letter-spacing: 0.2px;
}

.brand__tag {
  margin-top: 2px;
  font-size: 12px;
  color: rgba(148, 163, 184, 0.92);
}

.actions {
  display: flex;
  gap: 10px;
  align-items: center;
}

.btn-ghost {
  border: 1px solid rgba(148, 163, 184, 0.28) !important;
  background: rgba(15, 23, 42, 0.42) !important;
  color: rgba(226, 232, 240, 0.95) !important;
  backdrop-filter: blur(10px);
}

.hero {
  display: grid;
  grid-template-columns: 1.1fr 0.9fr;
  gap: 22px;
  padding: 36px 0 18px;
  align-items: center;
}

.kicker {
  display: inline-flex;
  gap: 8px;
  align-items: center;
  padding: 8px 12px;
  border-radius: 999px;
  border: 1px solid rgba(148, 163, 184, 0.24);
  background: rgba(15, 23, 42, 0.38);
  backdrop-filter: blur(10px);
  color: rgba(148, 163, 184, 0.95);
  font-size: 12px;
  letter-spacing: 0.4px;
  width: fit-content;
}

.title {
  margin-top: 14px;
  font-size: 44px;
  line-height: 1.08;
  letter-spacing: -0.6px;
}

.title__accent {
  background: linear-gradient(90deg, #38bdf8, #a78bfa 55%, #22c55e);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

.subtitle {
  margin-top: 14px;
  color: rgba(148, 163, 184, 0.95);
  font-size: 14px;
  line-height: 1.65;
  max-width: 520px;
}

.cta {
  margin-top: 18px;
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.btn-primary {
  box-shadow: 0 18px 50px rgba(56, 189, 248, 0.22);
}

.btn-secondary {
  border: 1px solid rgba(148, 163, 184, 0.28) !important;
  background: rgba(15, 23, 42, 0.22) !important;
  color: rgba(226, 232, 240, 0.95) !important;
}

.highlights {
  margin-top: 18px;
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.pill {
  padding: 8px 10px;
  border-radius: 999px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  background: rgba(2, 6, 23, 0.34);
  backdrop-filter: blur(10px);
  font-size: 12px;
  color: rgba(226, 232, 240, 0.92);
  display: inline-flex;
  gap: 8px;
  align-items: center;
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 999px;
  display: inline-block;
}
.dot--cyan {
  background: #38bdf8;
  box-shadow: 0 0 18px rgba(56, 189, 248, 0.6);
}
.dot--violet {
  background: #a78bfa;
  box-shadow: 0 0 18px rgba(167, 139, 250, 0.55);
}
.dot--green {
  background: #22c55e;
  box-shadow: 0 0 18px rgba(34, 197, 94, 0.45);
}
.dot--amber {
  background: #f59e0b;
  box-shadow: 0 0 18px rgba(245, 158, 11, 0.45);
}
.dot--red {
  background: #f43f5e;
}

.hero__right {
  display: flex;
  justify-content: flex-end;
}

.glass {
  width: min(420px, 100%);
  border-radius: 18px;
  border: 1px solid rgba(148, 163, 184, 0.24);
  background: rgba(2, 6, 23, 0.42);
  backdrop-filter: blur(12px);
  box-shadow: 0 30px 80px rgba(0, 0, 0, 0.45);
  overflow: hidden;
  transform: translateY(0);
  animation: float 6s ease-in-out infinite;
}

.glass__top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 14px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.14);
}

.glass__title {
  font-size: 12px;
  color: rgba(148, 163, 184, 0.95);
}

.signal {
  display: flex;
  gap: 8px;
  align-items: center;
}

.signal__dot {
  width: 8px;
  height: 8px;
  border-radius: 999px;
  opacity: 0.9;
}

.glass__body {
  padding: 14px;
}

.line {
  height: 10px;
  border-radius: 8px;
  background: rgba(148, 163, 184, 0.14);
  margin-bottom: 10px;
  position: relative;
  overflow: hidden;
}

.shimmer::after {
  content: "";
  position: absolute;
  inset: 0;
  transform: translateX(-80%);
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(56, 189, 248, 0.16) 40%,
    rgba(167, 139, 250, 0.18) 60%,
    transparent 100%
  );
  animation: shimmer 2.2s ease-in-out infinite;
}

.chip-row {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.chip {
  font-size: 12px;
  padding: 6px 10px;
  border-radius: 999px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  background: rgba(15, 23, 42, 0.28);
  color: rgba(226, 232, 240, 0.9);
}

.bubble {
  padding: 10px 12px;
  border-radius: 14px;
  font-size: 12px;
  line-height: 1.55;
  margin-top: 10px;
  border: 1px solid rgba(148, 163, 184, 0.16);
}

.bubble--user {
  background: rgba(56, 189, 248, 0.10);
}

.bubble--ai {
  background: rgba(167, 139, 250, 0.10);
}

.spacer {
  height: 10px;
}

.features {
  margin-top: 26px;
  padding: 20px 0 6px;
}

.features__head {
  margin-bottom: 14px;
}

.features__title {
  font-size: 20px;
  letter-spacing: 0.2px;
}

.features__desc {
  margin-top: 8px;
  color: rgba(148, 163, 184, 0.95);
  font-size: 13px;
}

.grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.card {
  border-radius: 16px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  background: rgba(2, 6, 23, 0.38);
  backdrop-filter: blur(10px);
  padding: 14px 14px 12px;
  transition: transform 200ms ease, border-color 200ms ease, box-shadow 200ms ease;
  position: relative;
  overflow: hidden;
}

.card::before {
  content: "";
  position: absolute;
  inset: -1px;
  background: radial-gradient(700px 200px at 10% 0%, rgba(56, 189, 248, 0.12), transparent 55%);
  opacity: 0;
  transition: opacity 200ms ease;
  pointer-events: none;
}

.card:hover {
  transform: translateY(-3px);
  border-color: rgba(56, 189, 248, 0.35);
  box-shadow: 0 22px 60px rgba(0, 0, 0, 0.35);
}

.card:hover::before {
  opacity: 1;
}

.card__icon {
  font-size: 20px;
}

.card__title {
  margin-top: 10px;
  font-weight: 700;
  letter-spacing: 0.2px;
}

.card__desc {
  margin-top: 8px;
  color: rgba(148, 163, 184, 0.95);
  font-size: 12px;
  line-height: 1.6;
}

.card__meta {
  margin-top: 10px;
  font-size: 12px;
  color: rgba(226, 232, 240, 0.72);
}

.footer {
  padding: 22px 0 10px;
  margin-top: 10px;
  text-align: center;
}

.footer__line {
  height: 1px;
  width: 100%;
  background: linear-gradient(90deg, transparent, rgba(148, 163, 184, 0.26), transparent);
  margin-bottom: 10px;
}

@keyframes float {
  0%,
  100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-10px);
  }
}

@keyframes shimmer {
  0% {
    transform: translateX(-80%);
  }
  100% {
    transform: translateX(80%);
  }
}

@media (max-width: 980px) {
  .hero {
    grid-template-columns: 1fr;
  }
  .hero__right {
    justify-content: flex-start;
  }
  .title {
    font-size: 36px;
  }
  .grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 620px) {
  .landing__content {
    padding: 20px 14px 10px;
  }
  .title {
    font-size: 30px;
  }
  .grid {
    grid-template-columns: 1fr;
  }
}

@media (prefers-reduced-motion: reduce) {
  .glass,
  .shimmer::after {
    animation: none !important;
  }
}
</style>

