<template>
  <div class="client-management">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>租户与企业</span>
          <el-button type="primary" @click="openCreate">
            <el-icon><Plus /></el-icon>
            新增
          </el-button>
        </div>
      </template>

      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="企业账户ID">
          <el-input
            v-model="searchForm.clientId"
            placeholder="模糊查询"
            clearable
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="loadList">
            <el-icon><Search /></el-icon>
            搜索
          </el-button>
          <el-button @click="resetSearch">
            <el-icon><Refresh /></el-icon>
            重置
          </el-button>
        </el-form-item>
      </el-form>

      <el-table
        v-loading="loading"
        :data="tableData"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column prop="clientId" label="企业账户ID" width="220" />
        <el-table-column prop="name" label="名称" width="220" />
        <el-table-column
          prop="desp"
          label="描述"
          min-width="240"
          show-overflow-tooltip
        />
        <el-table-column prop="updatedAt" label="更新时间" width="180" />
        <el-table-column prop="createdAt" label="创建时间" width="180" />
        <el-table-column
          prop="clientJsonInfo"
          label="JSON"
          min-width="320"
          show-overflow-tooltip
        />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)"
              >编辑</el-button
            >
            <el-button
              type="danger"
              link
              size="small"
              @click="handleDelete(row)"
              >删除</el-button
            >
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadList"
          @current-change="loadList"
        />
      </div>
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      title="新增/修改 Client"
      width="700px"
      :close-on-click-modal="false"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
        <el-form-item label="企业账户ID" prop="clientId">
          <el-input v-model="form.clientId" placeholder="例如: default" />
        </el-form-item>
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="例如：Demo 平台" />
        </el-form-item>
        <el-form-item label="描述" prop="desp">
          <el-input v-model="form.desp" placeholder="可选" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submit"
          >保存</el-button
        >
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { Plus, Search, Refresh } from "@element-plus/icons-vue";
import {
  adminClientCreate,
  adminClientDelete,
  adminClientList,
} from "@/api/knowsource";
import { useUserStore } from "@/stores/user";

const loading = ref(false);
const saving = ref(false);
const dialogVisible = ref(false);
const formRef = ref(null);

const searchForm = reactive({
  clientId: "",
});

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
});

const tableData = ref([]);

const form = reactive({
  clientId: "",
  name: "",
  desp: "",
});

const rules = {
  clientId: [
    { required: true, message: "请输入企业账户名", trigger: "blur" },
    { pattern: /^[a-zA-Z0-9_-]+$/, message: "仅支持字母、数字、下划线和横线", trigger: "blur" }
  ],
  name: [{ required: true, message: "请输入名称", trigger: "blur" }],
};

const loadList = async () => {
  loading.value = true;
  try {
    const res = await adminClientList({
      clientId: searchForm.clientId || "",
      page: pagination.page,
      pageSize: pagination.pageSize,
    });
    if (res.code === 200 && res.data) {
      tableData.value = res.data.list || [];
      pagination.total = res.data.total || 0;
    } else {
      ElMessage.error(res.message || "加载失败");
    }
  } catch (e) {
    ElMessage.error(e?.message || "加载失败");
  } finally {
    loading.value = false;
  }
};

const resetSearch = () => {
  searchForm.clientId = "";
  pagination.page = 1;
  loadList();
};

const openCreate = () => {
  form.clientId = "";
  form.name = "";
  form.desp = "";
  dialogVisible.value = true;
};

const handleEdit = (row) => {
  form.clientId = row.clientId || "";
  form.name = row.name || "";
  form.desp = row.desp || "";
  dialogVisible.value = true;
};

const submit = async () => {
  if (!formRef.value) return;
  await formRef.value.validate(async (valid) => {
    if (!valid) return;
    saving.value = true;
    try {
      const res = await adminClientCreate({
        clientId: form.clientId,
        name: form.name,
        desp: form.desp,
      });
      if (res.code === 200) {
        ElMessage.success("保存成功");
        dialogVisible.value = false;
        await loadList();
      } else {
        ElMessage.error(res.message || "保存失败");
      }
    } catch (e) {
      ElMessage.error(e?.message || "保存失败");
    } finally {
      saving.value = false;
    }
  });
};

const userStore = useUserStore();

const handleDelete = async (row) => {
  // 获取当前登录用户的clientId
  const currentClientId = localStorage.getItem('clientId');
  
  // 检查是否为当前登录用户的clientId
  if (currentClientId !== row.clientId) {
    ElMessage.warning('只能删除当前登录租户的clientId');
    return;
  }
  
  // 检查是否为系统保护的租户
  if (row.clientId === 'demo' || row.clientId === 'admin') {
    ElMessage.warning('禁止删除系统保护的租户');
    return;
  }
  
  // 要求用户输入AGREE确认删除
  const confirmText = await ElMessageBox.prompt(
    `删除操作将不可恢复！请输入 "AGREE" 确认删除 clientId=${row.clientId}`, 
    "删除确认", {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: "danger",
      inputPattern: /^AGREE$/,
      inputErrorMessage: '请输入 "AGREE" 确认删除',
    }
  );
  
  try {
    // 用户确认输入正确，执行删除
    const res = await adminClientDelete({ clientId: row.clientId });
    if (res.code === 200) {
      ElMessage.success("删除成功");
      await loadList();
      // 删除成功后，执行退出登录
      userStore.logout();
      // 清除clientId
      localStorage.removeItem('clientId');
      // 跳转到登录页面
      setTimeout(() => {
        window.location.href = '/login';
      }, 1500);
    } else {
      ElMessage.error(res.message || "删除失败");
    }
  } catch (e) {
    if (e !== "cancel") ElMessage.error(e?.message || "删除失败");
  }
};

onMounted(() => {
  loadList();
});
</script>

<style scoped>
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.search-form {
  margin-bottom: 20px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>
