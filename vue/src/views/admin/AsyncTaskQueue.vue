<template>
  <div class="async-task-queue">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>异步任务队列</span>
          <div class="header-actions">
            <el-switch v-model="autoRefresh" active-text="自动刷新" inactive-text="手动" />
            <el-button type="primary" :loading="loading" @click="loadFull">刷新</el-button>
          </div>
        </div>
      </template>

      <el-form :inline="true" class="search-form">
        <el-form-item label="任务类型">
          <el-input v-model="filters.taskType" placeholder="如 raw_documents_audit_in" clearable style="width: 240px" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" clearable placeholder="全部" style="width: 200px">
            <el-option label="init 未开始" value="init" />
            <el-option label="running 运行中" value="running" />
            <el-option label="canceled 已取消" value="canceled" />
            <el-option label="failed 失败" value="failed" />
            <el-option label="success 已成功（历史）" value="success" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
          <el-button
            type="danger"
            plain
            :disabled="loading || selectedRows.length === 0"
            @click="handleBatchDelete"
          >
            批量删除
          </el-button>
        </el-form-item>
      </el-form>

      <el-table
        :data="tableData"
        border
        stripe
        v-loading="loading"
        style="width: 100%"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="44" />
        <el-table-column prop="id" label="ID" width="90" />
        <el-table-column prop="taskType" label="任务类型" width="220" show-overflow-tooltip />
        <el-table-column prop="taskDesc" label="描述" min-width="220" show-overflow-tooltip />
        <el-table-column prop="sourceId" label="源ID" width="100" />
        <el-table-column prop="sourceKey" label="触发者/Key" width="140" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">
              {{ statusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="executeResult" label="结果" min-width="200" show-overflow-tooltip />
        <el-table-column prop="createdAt" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column prop="updatedAt" label="更新时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.updatedAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="170" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="isStoppable(row.status)"
              type="warning"
              link
              :disabled="loading"
              @click="handleStop(row)"
            >
              停止
            </el-button>
            <el-button
              v-if="isDeletable(row.status)"
              type="danger"
              link
              :disabled="loading"
              @click="handleDelete(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100, 200]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadFull"
          @current-change="loadFull"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { onBeforeUnmount, onMounted, reactive, ref, watch } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { listAsyncTask, getAsyncTaskWatermark, cancelAsyncTask, deleteAsyncTask } from "@/api/knowsource";

const loading = ref(false);
const tableData = ref([]);
const selectedRows = ref([]);
const autoRefresh = ref(true);
const lastWatermark = ref("");
let timer = null;

const filters = reactive({
  taskType: "",
  status: undefined,
});

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
});

const formatTime = (ts) => {
  if (!ts) return "-";
  const d = new Date(ts * 1000);
  return d.toLocaleString("zh-CN");
};

const statusLabel = (s) => {
  const v = String(s || "").toLowerCase();
  const map = {
    init: "未开始",
    running: "运行中",
    canceled: "已取消",
    success: "成功",
    failed: "失败",
  };
  return map[v] || String(s || "-");
};

const statusTagType = (s) => {
  const v = String(s || "").toLowerCase();
  switch (v) {
    case "init":
      return "info";
    case "canceled":
      return "warning";
    case "success":
      return "success";
    case "failed":
      return "danger";
    case "running":
      return "primary";
    default:
      return "info";
  }
};

const isStoppable = (s) => {
  const v = String(s || "").toLowerCase();
  return v === "init" || v === "running";
};

const isDeletable = (s) => {
  const v = String(s || "").toLowerCase();
  return v !== "init" && v !== "running";
};

const listParams = () => ({
  page: pagination.page,
  pageSize: pagination.pageSize,
  taskType: filters.taskType || "",
  status: filters.status || "",
});

/** 同步服务端水印（列表拉取后调用，避免与 Redis 不一致） */
const syncWatermarkFromServer = async () => {
  try {
    const wmRes = await getAsyncTaskWatermark({});
    if (wmRes.code === 200 && wmRes.data && wmRes.data.watermark != null) {
      lastWatermark.value = String(wmRes.data.watermark);
    }
  } catch (_) {
    /* 忽略 */
  }
};

/** 仅拉列表，不显示 loading（自动刷新用，避免表格闪烁） */
const fetchListSilent = async () => {
  try {
    const res = await listAsyncTask(listParams());
    if (res.code === 200 && res.data) {
      tableData.value = res.data.list || [];
      pagination.total = res.data.total || 0;
    }
  } catch (_) {
    /* 自动刷新静默失败 */
  }
};

/** 完整加载：表格 loading + 列表 + 同步水印 */
const loadFull = async () => {
  loading.value = true;
  try {
    const res = await listAsyncTask(listParams());
    if (res.code === 200 && res.data) {
      tableData.value = res.data.list || [];
      pagination.total = res.data.total || 0;
    } else {
      ElMessage.error(res.message || "查询失败");
    }
  } catch (e) {
    ElMessage.error(e?.message || "查询失败");
  } finally {
    loading.value = false;
    await syncWatermarkFromServer();
  }
};

/** 轮询：先比 Redis 水印，变化才请求列表 */
const pollByWatermark = async () => {
  if (!autoRefresh.value) return;
  try {
    const wmRes = await getAsyncTaskWatermark({});
    if (wmRes.code !== 200) return;
    const w = wmRes.data?.watermark != null ? String(wmRes.data.watermark) : "";
    if (w === lastWatermark.value) return;
    lastWatermark.value = w;
    await fetchListSilent();
  } catch (_) {
    /* 静默 */
  }
};

const handleSearch = () => {
  pagination.page = 1;
  loadFull();
};

const handleReset = () => {
  filters.taskType = "";
  filters.status = undefined;
  handleSearch();
};

const handleSelectionChange = (rows) => {
  selectedRows.value = Array.isArray(rows) ? rows : [];
};

const handleStop = async (row) => {
  const id = Number(row?.id || 0);
  if (!id) return;
  try {
    await ElMessageBox.confirm(`确认停止任务 #${id} 吗？`, "提示", {
      type: "warning",
      confirmButtonText: "确认停止",
      cancelButtonText: "取消",
    });
    const res = await cancelAsyncTask({ id });
    if (res.code === 200) {
      ElMessage.success(res.message || "停止成功");
      await loadFull();
      return;
    }
    ElMessage.error(res.message || "停止失败");
  } catch (_) {
    // 用户取消
  }
};

const handleBatchDelete = async () => {
  const rows = selectedRows.value || [];
  if (rows.length === 0) return;

  const deletableRows = rows.filter((r) => isDeletable(r?.status));
  if (deletableRows.length === 0) {
    ElMessage.warning("所选任务均在执行中，请先停止后再删除");
    return;
  }

  if (deletableRows.length !== rows.length) {
    ElMessage.warning("已自动跳过执行中任务，仅删除可删除项");
  }

  const ids = deletableRows
    .map((r) => Number(r?.id || 0))
    .filter((id) => id > 0);
  if (ids.length === 0) return;

  try {
    await ElMessageBox.confirm(`确认批量删除 ${ids.length} 条任务吗？该操作不可恢复。`, "提示", {
      type: "warning",
      confirmButtonText: "确认删除",
      cancelButtonText: "取消",
    });

    let success = 0;
    let failed = 0;
    for (const id of ids) {
      try {
        const res = await deleteAsyncTask({ id });
        if (res.code === 200) {
          success += 1;
        } else {
          failed += 1;
        }
      } catch (_) {
        failed += 1;
      }
    }

    if (failed === 0) {
      ElMessage.success(`已删除 ${success} 条任务`);
    } else if (success === 0) {
      ElMessage.error(`删除失败：${failed} 条`);
    } else {
      ElMessage.warning(`删除完成：成功 ${success} 条，失败 ${failed} 条`);
    }
    selectedRows.value = [];
    await loadFull();
  } catch (_) {
    // 用户取消
  }
};

const handleDelete = async (row) => {
  const id = Number(row?.id || 0);
  if (!id) return;
  try {
    await ElMessageBox.confirm(`确认删除任务 #${id} 吗？该操作不可恢复。`, "提示", {
      type: "warning",
      confirmButtonText: "确认删除",
      cancelButtonText: "取消",
    });
    const res = await deleteAsyncTask({ id });
    if (res.code === 200) {
      ElMessage.success(res.message || "删除成功");
      await loadFull();
      return;
    }
    ElMessage.error(res.message || "删除失败");
  } catch (_) {
    // 用户取消
  }
};

const startTimer = () => {
  stopTimer();
  timer = window.setInterval(() => {
    pollByWatermark();
  }, 2000);
};

const stopTimer = () => {
  if (timer) {
    window.clearInterval(timer);
    timer = null;
  }
};

watch(autoRefresh, () => startTimer());

onMounted(() => {
  loadFull();
  startTimer();
});

onBeforeUnmount(() => stopTimer());
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 16px;
  font-weight: 600;
}
.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}
.search-form {
  margin-bottom: 14px;
}
.pagination {
  margin-top: 14px;
  display: flex;
  justify-content: flex-end;
}
</style>
