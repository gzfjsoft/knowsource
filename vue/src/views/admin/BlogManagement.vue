<template>
  <div class="blog-management">
    <el-card>
      <template #header>
        <div class="header">
          <span>博客管理</span>
          <div class="actions">
            <el-input
              v-model="query.keyword"
              clearable
              placeholder="搜索标题/摘要/别名"
              style="width: 240px"
              @keyup.enter="loadData"
            />
            <el-select v-model="query.isPublished" style="width: 120px">
              <el-option :value="-1" label="全部" />
              <el-option :value="1" label="已发布" />
              <el-option :value="0" label="未发布" />
            </el-select>
            <el-button type="primary" @click="loadData">查询</el-button>
            <el-button type="success" @click="openCreate">新建</el-button>
          </div>
        </div>
      </template>

      <el-table :data="rows" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="title" label="标题" min-width="220" show-overflow-tooltip />
        <el-table-column prop="alias" label="别名" min-width="140">
          <template #default="{ row }">
            <span>{{ row.alias || "-" }}</span>
          </template>
        </el-table-column>
        <el-table-column label="外链" min-width="220">
          <template #default="{ row }">
            <a :href="toPublicUrl(row)" target="_blank">{{ toPublicPath(row) }}</a>
          </template>
        </el-table-column>
        <el-table-column prop="isPublished" label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.isPublished === 1 ? 'success' : 'info'">
              {{ row.isPublished === 1 ? "已发布" : "未发布" }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="updatedAt" label="更新时间" width="180">
          <template #default="{ row }">{{ formatTime(row.updatedAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openEdit(row.id)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pager">
        <el-pagination
          background
          layout="total, prev, pager, next, sizes"
          :total="total"
          :current-page="query.page"
          :page-size="query.pageSize"
          :page-sizes="[10, 20, 50]"
          @current-change="onPageChange"
          @size-change="onSizeChange"
        />
      </div>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑博客' : '新建博客'" width="980px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="标题">
          <el-input v-model="form.title" />
        </el-form-item>
        <el-form-item label="别名">
          <el-input v-model="form.alias" placeholder="例如 readme（唯一，可为空）" />
        </el-form-item>
        <el-form-item label="摘要">
          <el-input v-model="form.summary" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="标签">
          <el-input v-model="form.tags" placeholder="逗号分隔" />
        </el-form-item>
        <el-form-item label="分类">
          <el-input v-model="form.categories" placeholder="逗号分隔" />
        </el-form-item>
        <el-form-item label="作者">
          <el-input v-model="form.authors" />
        </el-form-item>
        <el-form-item label="Banner">
          <el-input v-model="form.banner" />
        </el-form-item>
        <el-form-item label="发布">
          <el-switch v-model="publishedBool" />
        </el-form-item>
        <el-form-item label="Markdown">
          <el-input v-model="form.content" type="textarea" :rows="14" />
        </el-form-item>
      </el-form>

      <el-divider>预览</el-divider>
      <div class="markdown-body preview" v-html="previewHtml"></div>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { parseMarkdown } from "@/utils/markdown";
import {
  blogAdminCreate,
  blogAdminDelete,
  blogAdminGet,
  blogAdminList,
  blogAdminUpdate,
} from "@/api/knowsource";

const loading = ref(false);
const rows = ref([]);
const total = ref(0);
const query = reactive({
  page: 1,
  pageSize: 10,
  keyword: "",
  isPublished: -1,
});

const dialogVisible = ref(false);
const isEdit = ref(false);
const form = reactive({
  id: 0,
  title: "",
  alias: "",
  summary: "",
  content: "",
  tags: "",
  categories: "",
  authors: "",
  banner: "",
  isPublished: 0,
});

const publishedBool = computed({
  get: () => form.isPublished === 1,
  set: (v) => {
    form.isPublished = v ? 1 : 0;
  },
});

const previewHtml = computed(() => {
  const md = form.content || "";
  return parseMarkdown(md);
});

const resetForm = () => {
  Object.assign(form, {
    id: 0,
    title: "",
    alias: "",
    summary: "",
    content: "",
    tags: "",
    categories: "",
    authors: "",
    banner: "",
    isPublished: 0,
  });
};

const loadData = async () => {
  loading.value = true;
  try {
    const res = await blogAdminList(query);
    if (res.code === 200 && res.data) {
      rows.value = res.data.list || [];
      total.value = res.data.total || 0;
    } else {
      ElMessage.error(res.message || "加载失败");
    }
  } catch (e) {
    ElMessage.error(`加载失败: ${e.message || "未知错误"}`);
  } finally {
    loading.value = false;
  }
};

const openCreate = () => {
  isEdit.value = false;
  resetForm();
  dialogVisible.value = true;
};

const openEdit = async (id) => {
  isEdit.value = true;
  resetForm();
  try {
    const res = await blogAdminGet({ id });
    if (res.code === 200 && res.data) {
      Object.assign(form, {
        id: res.data.id,
        title: res.data.title || "",
        alias: res.data.alias || "",
        summary: res.data.summary || "",
        content: res.data.content || "",
        tags: res.data.tags || "",
        categories: res.data.categories || "",
        authors: res.data.authors || "",
        banner: res.data.banner || "",
        isPublished: res.data.isPublished === 1 ? 1 : 0,
      });
      dialogVisible.value = true;
    } else {
      ElMessage.error(res.message || "加载详情失败");
    }
  } catch (e) {
    ElMessage.error(`加载详情失败: ${e.message || "未知错误"}`);
  }
};

const submitForm = async () => {
  if (!form.title.trim() || !form.content.trim()) {
    ElMessage.warning("标题和内容不能为空");
    return;
  }
  const payload = { ...form };
  try {
    const res = isEdit.value
      ? await blogAdminUpdate(payload)
      : await blogAdminCreate(payload);
    if (res.code === 200) {
      ElMessage.success("保存成功");
      dialogVisible.value = false;
      loadData();
    } else {
      ElMessage.error(res.message || "保存失败");
    }
  } catch (e) {
    ElMessage.error(`保存失败: ${e.message || "未知错误"}`);
  }
};

const handleDelete = async (id) => {
  try {
    await ElMessageBox.confirm("确认删除该博客？", "提示", { type: "warning" });
    const res = await blogAdminDelete({ id });
    if (res.code === 200) {
      ElMessage.success("删除成功");
      loadData();
    } else {
      ElMessage.error(res.message || "删除失败");
    }
  } catch {
    // ignore cancel
  }
};

const onPageChange = (p) => {
  query.page = p;
  loadData();
};

const onSizeChange = (s) => {
  query.pageSize = s;
  query.page = 1;
  loadData();
};

const formatTime = (ts) => {
  if (!ts) return "-";
  return new Date(ts * 1000).toLocaleString("zh-CN");
};

const toPublicPath = (row) => `/blog/${row.alias || row.id}`;
const toPublicUrl = (row) => `${window.location.origin}${toPublicPath(row)}`;

onMounted(loadData);
</script>

<style scoped>
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.actions {
  display: flex;
  align-items: center;
  gap: 8px;
}
.pager {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
.preview {
  max-height: 360px;
  overflow: auto;
  background: #fafafa;
  padding: 12px;
  border-radius: 6px;
}
</style>
