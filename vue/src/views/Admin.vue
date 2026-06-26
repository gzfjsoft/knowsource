<template>
  <div class="admin-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>平台工具集</span>
        </div>
      </template>
      <p class="admin-intro">
        以下为未在左侧菜单单独列出的管理入口；对话日志、调用统计、依赖检查等请从「AI 与运维」「超级管理」进入。
      </p>

      <el-row :gutter="16">
        <el-col :xs="24" :sm="12">
          <el-card shadow="hover" class="nav-card" @click="go('/doc-path-tree')">
            <div class="nav-title">文档路径文件树</div>
            <div class="nav-desc">管理/查看服务器文档路径树</div>
          </el-card>
        </el-col>
        <el-col :xs="24" :sm="12">
          <el-card shadow="hover" class="nav-card" @click="go('/permission-management')">
            <div class="nav-title">权限管理</div>
            <div class="nav-desc">角色权限与权限点配置</div>
          </el-card>
        </el-col>
        <el-col :xs="24" :sm="12">
          <el-card shadow="hover" class="nav-card" @click="go('/check-raw-documents-file-exists')">
            <div class="nav-title">检查文档文件是否存在</div>
            <div class="nav-desc">检查所有原始文档文件是否存在于硬盘</div>
          </el-card>
        </el-col>
        <el-col :xs="24" :sm="12">
          <el-card shadow="hover" class="nav-card" @click="go('/regenerate-summaries')">
            <div class="nav-title">重新生成文档概要</div>
            <div class="nav-desc">重新生成所有已审核文档的概要并更新到 Qdrant</div>
          </el-card>
        </el-col>
      </el-row>
    </el-card>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()
const go = (path) => router.push(path)
</script>

<style scoped>
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 18px;
  font-weight: 500;
}

.admin-intro {
  margin: 0 0 16px;
  font-size: 13px;
  color: #909399;
  line-height: 1.5;
}

.nav-card {
  cursor: pointer;
  user-select: none;
  margin-bottom: 16px;
}

.nav-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.nav-desc {
  margin-top: 6px;
  font-size: 13px;
  color: #909399;
}
</style>
