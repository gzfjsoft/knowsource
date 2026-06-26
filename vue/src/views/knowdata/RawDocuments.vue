<template>
  <div class="raw-documents">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>文档管理</span>
          <el-button v-if="hasPermission('功能-上传文档')" type="primary" @click="handleUpload">
            <el-icon><Upload /></el-icon>
            上传文档
          </el-button>
        </div>
      </template>
      
      <!-- 搜索表单 -->
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="知识库">
          <el-select
            v-model="searchForm.documentCode"
            placeholder="请选择知识库"
            clearable
            filterable
            style="width: 200px"
          >
            <el-option
              v-for="item in documentsTypeList"
              :key="item.code"
              :label="item.name"
              :value="item.code"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="文件名">
          <el-input v-model="searchForm.fileName" placeholder="请输入文件名" clearable />
        </el-form-item>
        <el-form-item label="标签">
          <el-select
            v-model="searchForm.tag"
            placeholder="请选择标签"
            clearable
            filterable
            style="width: 200px"
          >
            <el-option
              v-for="tag in availableTags"
              :key="tag"
              :label="tag"
              :value="tag"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="审核状态">
          <el-select
            v-model="searchForm.isAudit"
            placeholder="全部"
            clearable
            style="width: 140px"
          >
            <el-option label="已审核" value="1" />
            <el-option label="未审核" value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch" :loading="loading">
            <el-icon><Search /></el-icon>
            搜索
          </el-button>
          <el-button @click="handleReset">
            <el-icon><Refresh /></el-icon>
            重置
          </el-button>
        </el-form-item>
      </el-form>

      <!-- 表格 -->
      <el-table
        v-loading="loading"
        :data="tableData"
        border
        stripe
        style="width: 100%"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="35" :selectable="rawDocRowSelectable" />
        <!-- <el-table-column prop="id" label="ID" width="50" /> -->
        <el-table-column prop="documentCode" label="知识库名称" width="150">
          <template #default="{ row }">
            {{ getDocumentTypeName(row.documentCode) }}
          </template>
        </el-table-column>
        <el-table-column prop="fileName" label="文件名" width="250">
          <template #default="{ row }">
            <div class="file-name-cell">
              <span class="file-name-link" @click="handleFileNameClick(row)">
                {{ row.fileName }}
              </span>
              <el-tooltip
                v-if="row.status"
                :content="row.statusMsg || row.status"
                :disabled="!row.statusMsg"
                placement="top"
              >
                <el-tag
                  class="file-status-tag"
                  size="small"
                  :type="rawDocStatusTagType(row)"
                  effect="plain"
                >
                  {{ row.status }}
                </el-tag>
              </el-tooltip>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="fileSize" label="文件大小" width="100">
          <template #default="{ row }">
            {{ formatFileSize(row.fileSize) }}
          </template>
        </el-table-column>
        <el-table-column prop="tag" label="标签" width="120">
          <template #default="{ row }">
            <el-tag v-if="row.tag" size="small" type="info">{{ row.tag }}</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="isToMd" label="转MD" width="70">
          <template #default="{ row }">
            <el-tag :type="row.isToMd === 1 ? 'success' : 'info'">
              {{ row.isToMd === 1 ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <!-- <el-table-column prop="isToAi" label="转AI" width="60">
          <template #default="{ row }">
            <el-tag :type="row.isToAi === 1 ? 'success' : 'info'">
              {{ row.isToAi === 1 ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column> -->
        <el-table-column prop="uploadUser" label="上传人" width="120">
          <template #default="{ row }">
            {{ row.uploadUser || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="isAudit" label="审核状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.isAudit === 1 ? 'success' : 'info'">
              {{ row.isAudit === 1 ? '已审核' : '未审核' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="auditUser" label="审核人" width="120">
          <template #default="{ row }">
            {{ row.auditUser || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="auditedAt" label="审核时间" width="180">
          <template #default="{ row }">
            {{ row.auditedAt ? formatTime(row.auditedAt) : '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.createdAt) }}
          </template>
        </el-table-column>
        <!-- <el-table-column prop="fileMd5" label="MD5" width="150" /> -->
        <!-- <el-table-column prop="zipFileName" label="压缩文件名" width="200" /> -->
        <!-- <el-table-column prop="zipFileSize" label="压缩文件大小" width="120">
          <template #default="{ row }">
            {{ row.zipFileSize ? formatFileSize(row.zipFileSize) : '-' }}
          </template>
        </el-table-column> -->
        <el-table-column label="操作" width="580" fixed="right">
          <template #default="{ row }">
            <div class="op-rows">
              <div class="op-row">
                <el-button
                  type="primary"
                  size="small"
                  @click="handleDownload(row)"
                  :disabled="rawDocRowOpsLocked(row)"
                >
                  下载
                </el-button>
                <el-button
                  v-if="row.isAudit !== 1"
                  type="warning"
                  size="small"
                  @click="handleConvertToZIP(row)"
                  :loading="convertingIds.includes(row.id)"
                  :disabled="rawDocRowOpsLocked(row)"
                >
                  识别文字
                </el-button>
                <el-button
                  v-if="row.isAudit !== 1 && rawDocStatusBusyExtracting(row)"
                  type="danger"
                  size="small"
                  plain
                  @click="handleCancelRecognize(row)"
                >
                  中断识别
                </el-button>
                <el-button
                  v-if="hasPermission('功能-审核文档') && !rawDocStatusBusyInserting(row)"
                  type="info"
                  size="small"
                  @click="handleAudit(row)"
                  :disabled="rawDocRowOpsLocked(row)"
                >
                  {{ row.isAudit === 1 ? '取消审核' : '审核' }}
                </el-button>
                <el-button
                  v-if="hasPermission('功能-审核文档') && rawDocStatusBusyInserting(row)"
                  type="danger"
                  size="small"
                  plain
                  @click="handleCancelAudit(row)"
                >
                  中断审核入库
                </el-button>
                <el-button
                  v-if="hasPermission('功能-删除文档') && !rawDocStatusBusyInserting(row)"
                  type="danger"
                  size="small"
                  @click="handleDelete(row)"
                  :disabled="row.isAudit === 1 || rawDocRowOpsLocked(row)"
                >
                  删除
                </el-button>
              </div>
              <div class="op-row">
                
                <el-button
                  v-if="hasPermission('功能-更新文档内容') && row.isAudit !== 1 && row.isToMd === 1"
                  type="primary"
                  size="small"
                  @click="handleEdit(row)"
                  :disabled="rawDocRowOpsLocked(row)"
                >
                  <el-icon><Edit /></el-icon>
                  编辑
                </el-button>
                <!-- <el-button
                  v-if="hasPermission('功能-更新文档内容') && row.isAudit !== 1 && row.isToMd === 1"
                  type="warning"
                  size="small"
                  :loading="normalizeLoadingId === row.id"
                  @click="handleMarkdownNormalizePreviewRow(row)"
                >
                  <el-icon><MagicStick /></el-icon>
                  LLM 规范化
                </el-button> -->
                <el-button
                  v-if="hasPermission('功能-修改文档标签')"
                  type="warning"
                  size="small"
                  @click="handleChangeTag(row)"
                  :disabled="row.isAudit === 1 || rawDocRowOpsLocked(row)"
                >
                <el-icon><Edit /></el-icon>
                  更改标签
                </el-button>
                <el-button
                  v-if="hasPermission('功能-变更文档类型')"
                  type="primary"
                  size="small"
                  @click="handleChangeDocumentType(row)"
                  :disabled="row.isAudit === 1 || rawDocRowOpsLocked(row)"
                >
                  更改知识库
                </el-button>
                <el-button
                  type="info"
                  size="small"
                  @click="handleCompare(row)"
                  :loading="compareLoadingId === row.id"
                  :disabled="rawDocRowOpsLocked(row)"
                >
                  <el-icon><Files /></el-icon>
                  比较
                </el-button>
                <el-button
                  v-if="showQdrantChunksButton && hasPermission('功能-审核文档') && row.isAudit === 1"
                  type="success"
                  size="small"
                  plain
                  @click="handleViewQdrantChunks(row)"
                  :loading="qdrantChunksLoadingId === row.id"
                  :disabled="rawDocRowOpsLocked(row)"
                >
                  分块
                </el-button>
                <el-button
                  v-if="hasPermission('功能-审核文档') && row.isAudit === 1"
                  type="primary"
                  size="small"
                  plain
                  @click="handleViewQaPairs(row)"
                  :loading="qaPairsLoadingId === row.id"
                  :disabled="rawDocRowOpsLocked(row)"
                >
                  问答
                </el-button>
                <el-button
                  v-if="0  == 1"
                  type="success"
                  size="small"
                  @click="handleConvertToMD(row)"
                  :loading="convertingIds.includes(row.id)"
                >
                  转MD2
                </el-button>
              </div>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 批量操作 -->
      <div v-if="selectedRows.length > 0" class="batch-actions">
        <el-button
          v-if="batchRecognizeEligibleCount > 0"
          type="warning"
          @click="handleBatchRecognize"
        >
          批量识别 ({{ batchRecognizeEligibleCount }})
        </el-button>
        <el-button
          v-if="hasPermission('功能-审核文档') && batchAuditEligibleCount > 0"
          type="info"
          @click="handleBatchAudit"
        >
          批量审核 ({{ batchAuditEligibleCount }})
        </el-button>
        <el-button
          v-if="hasPermission('功能-审核文档') && batchUnAuditEligibleCount > 0"
          type="info"
          plain
          @click="handleBatchUnAudit"
        >
          批量取消审核 ({{ batchUnAuditEligibleCount }})
        </el-button>
        <el-button
          v-if="hasPermission('功能-删除文档')"
          type="danger"
          @click="handleBatchDelete"
        >
          批量删除 ({{ selectedRows.length }})
        </el-button>
      </div>

      <!-- 分页 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>

    <!-- 更改知识库对话框 -->
    <el-dialog
      v-model="changeTypeDialogVisible"
      title="更改知识库"
      width="500px"
    >
      <el-form
        ref="changeTypeFormRef"
        :model="changeTypeForm"
        :rules="changeTypeRules"
        label-width="120px"
      >
        <el-form-item label="文件名">
          <el-input v-model="changeTypeForm.fileName" disabled />
        </el-form-item>
        <el-form-item label="原知识库">
          <el-input :value="getDocumentTypeName(changeTypeForm.oldDocumentCode)" disabled />
        </el-form-item>
        <el-form-item label="新知识库" prop="newDocumentCode">
          <el-select
            v-model="changeTypeForm.newDocumentCode"
            placeholder="请选择新知识库"
            filterable
            style="width: 100%"
          >
            <el-option
              v-for="item in documentsTypeList"
              :key="item.code"
              :label="item.name"
              :value="item.code"
              :disabled="item.code === changeTypeForm.oldDocumentCode"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="changeTypeDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitChangeType" :loading="changeTypeLoading">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 更改标签对话框 -->
    <el-dialog
      v-model="changeTagDialogVisible"
      title="更改标签"
      width="500px"
    >
      <el-form
        ref="changeTagFormRef"
        :model="changeTagForm"
        :rules="changeTagRules"
        label-width="120px"
      >
        <el-form-item label="文件名">
          <el-input v-model="changeTagForm.fileName" disabled />
        </el-form-item>
        <el-form-item label="标签" prop="tag">
          <el-select
            v-model="changeTagForm.tag"
            placeholder="请选择标签"
            filterable
            clearable
            style="width: 100%"
          >
            <el-option
              v-for="tag in availableTags"
              :key="tag"
              :label="tag"
              :value="tag"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="changeTagDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitChangeTag" :loading="changeTagLoading">
          确定
        </el-button>
      </template>
    </el-dialog>


    <!-- MD 内容确认对话框 -->
    <el-dialog
      v-model="mdContentDialogVisible"
      title="Markdown 内容确认"
      width="80%"
      :close-on-click-modal="false"
    >
      <div class="md-content-header">
        <el-tag type="info">文件名：{{ mdContentData.fileName }}</el-tag>
        <el-tag type="success">转换成功</el-tag>
      </div>
      <el-divider />
      <div class="md-content-container">
        <el-input
          v-model="mdContentData.content"
          type="textarea"
          :rows="20"
          readonly
          placeholder="Markdown 内容将显示在这里"
        />
      </div>
      <template #footer>
        <el-button @click="mdContentDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="handleConfirmMDContent">
          确认并保存
        </el-button>
      </template>
    </el-dialog>

    <!-- ZIP 转换结果对话框 -->
    <el-dialog
      v-model="zipResultDialogVisible"
      title="转换成功"
      class="zip-result-dialog"
      :close-on-click-modal="false"
    >
      <el-divider />
      <div class="zip-result-header">
        <el-icon><Document /></el-icon>
        <span class="zip-filename">{{ zipResultData.fileName }}</span>
      </div>

      <div class="zip-result-container">
        <el-alert
          title="识别成功"
          type="success"
          :closable="false"
          show-icon
        />
      </div>
      <template #footer>
        <el-button @click="zipResultDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 批量识别进度对话框 -->
    <el-dialog
      v-model="batchRecognizeDialogVisible"
      title="批量识别"
      width="480px"
      class="batch-recognize-dialog"
      :close-on-click-modal="false"
      :show-close="batchRecognizeDone"
    >
      <div class="batch-recognize-content">
        <template v-if="!batchRecognizeDone">
          <div class="batch-recognize-current">
            正在识别：{{ batchRecognizeFileName || '...' }}
          </div>
          <div class="batch-recognize-stats">
            {{ batchRecognizeCurrent }} / {{ batchRecognizeTotal }}
          </div>
          <el-progress
            :percentage="batchRecognizeProgress"
            :stroke-width="14"
          />
        </template>
        <template v-else>
          <el-alert
            :title="`批量识别完成：成功 ${batchRecognizeSuccessCount} 个，失败 ${batchRecognizeFailCount} 个`"
            :type="batchRecognizeFailCount > 0 ? 'warning' : 'success'"
            :closable="false"
            show-icon
          />
          <div v-if="batchRecognizeFailList.length > 0" class="batch-recognize-fail-list">
            <div class="fail-list-title">失败文件：</div>
            <ul>
              <li v-for="(item, idx) in batchRecognizeFailList" :key="idx">
                {{ item.fileName }}：{{ item.message }}
              </li>
            </ul>
          </div>
        </template>
      </div>
      <template #footer>
        <el-button v-if="batchRecognizeDone" type="primary" @click="batchRecognizeDialogVisible = false">
          关闭
        </el-button>
      </template>
    </el-dialog>

    <!-- 批量审核进度对话框 -->
    <el-dialog
      v-model="batchAuditDialogVisible"
      title="批量审核"
      width="480px"
      class="batch-audit-dialog"
      :close-on-click-modal="false"
      :show-close="batchAuditDone"
    >
      <div class="batch-recognize-content">
        <template v-if="!batchAuditDone">
          <div class="batch-recognize-current">
            正在审核：{{ batchAuditFileName || '...' }}
          </div>
          <div class="batch-recognize-stats">
            {{ batchAuditCurrent }} / {{ batchAuditTotal }}
          </div>
          <el-progress
            :percentage="batchAuditProgress"
            :stroke-width="14"
          />
        </template>
        <template v-else>
          <el-alert
            :title="`批量审核完成：成功 ${batchAuditSuccessCount} 个，失败 ${batchAuditFailCount} 个`"
            :type="batchAuditFailCount > 0 ? 'warning' : 'success'"
            :closable="false"
            show-icon
          />
          <div v-if="batchAuditFailList.length > 0" class="batch-recognize-fail-list">
            <div class="fail-list-title">失败文件：</div>
            <ul>
              <li v-for="(item, idx) in batchAuditFailList" :key="idx">
                {{ item.fileName }}：{{ item.message }}
              </li>
            </ul>
          </div>
        </template>
      </div>
      <template #footer>
        <el-button v-if="batchAuditDone" type="primary" @click="batchAuditDialogVisible = false">
          关闭
        </el-button>
      </template>
    </el-dialog>

    <!-- 批量取消审核进度对话框 -->
    <el-dialog
      v-model="batchUnAuditDialogVisible"
      title="批量取消审核"
      width="480px"
      class="batch-audit-dialog"
      :close-on-click-modal="false"
      :show-close="batchUnAuditDone"
    >
      <div class="batch-recognize-content">
        <template v-if="!batchUnAuditDone">
          <div class="batch-recognize-current">
            正在取消审核：{{ batchUnAuditFileName || '...' }}
          </div>
          <div class="batch-recognize-stats">
            {{ batchUnAuditCurrent }} / {{ batchUnAuditTotal }}
          </div>
          <el-progress
            :percentage="batchUnAuditProgress"
            :stroke-width="14"
          />
        </template>
        <template v-else>
          <el-alert
            :title="`批量取消审核完成：成功 ${batchUnAuditSuccessCount} 个，失败 ${batchUnAuditFailCount} 个`"
            :type="batchUnAuditFailCount > 0 ? 'warning' : 'success'"
            :closable="false"
            show-icon
          />
          <div v-if="batchUnAuditFailList.length > 0" class="batch-recognize-fail-list">
            <div class="fail-list-title">失败文件：</div>
            <ul>
              <li v-for="(item, idx) in batchUnAuditFailList" :key="idx">
                {{ item.fileName }}：{{ item.message }}
              </li>
            </ul>
          </div>
        </template>
      </div>
      <template #footer>
        <el-button v-if="batchUnAuditDone" type="primary" @click="batchUnAuditDialogVisible = false">
          关闭
        </el-button>
      </template>
    </el-dialog>

    <!-- 上传对话框 -->
    <el-dialog
      v-model="uploadDialogVisible"
      title="上传原始文档"
      width="500px"
      class="upload-dialog"
    >
      <el-form
        ref="uploadFormRef"
        :model="uploadForm"
        :rules="uploadRules"
        label-width="120px"
        class="upload-form"
      >
        <el-form-item label="知识库" prop="documentCode">
          <el-select
            v-model="uploadForm.documentCode"
            placeholder="请选择知识库"
            filterable
            style="width: 100%"
            @change="handleDocumentCodeChange"
          >
            <el-option
              v-for="item in documentsTypeList"
              :key="item.code"
              :label="item.name"
              :value="item.code"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="标签">
          <el-select
            v-model="uploadForm.tag"
            placeholder="请选择标签"
            filterable
            clearable
            style="width: 100%"
          >
            <el-option
              v-for="tag in availableTagsForUpload"
              :key="tag"
              :label="tag"
              :value="tag"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="文件" prop="file">
          <el-upload
            ref="uploadRef"
            class="upload-inner"
            :auto-upload="false"
            :show-file-list="false"
            :on-change="handleFileChange"
            :on-remove="handleFileRemove"
            :on-exceed="handleUploadExceed"
            :limit="1"
            accept=".docx,.doc,.pdf,.txt,.zip,.md,.xlsx"
            drag
          >
            <template v-if="currentFile">
              <el-icon class="el-icon--upload"><Document /></el-icon>
              <div class="el-upload__text upload-selected-name">
                {{ currentFile.name }}
              </div>
              <div class="upload-replace-hint">再次拖放或点击可替换文件</div>
            </template>
            <template v-else>
              <el-icon class="el-icon--upload"><upload-filled /></el-icon>
              <div class="el-upload__text">
                将文件拖到此处，或<em>点击上传</em>
              </div>
            </template>
            <template #tip>
              <div class="el-upload__tip">
                支持 docx, doc, pdf, txt, zip, md, xlsx 文件。如果是 zip 文件会自动解压，如果是 xlsx 文件会自动转换为 md
              </div>
            </template>
          </el-upload>
          <!-- 上传进度条：上传中或已有进度时显示 -->
          <div v-if="uploadLoading || uploadProgress > 0" class="upload-progress-wrap">
            <div class="upload-progress-label">
              {{ uploadProgress === 100 ? '上传完成' : '上传中...' }} {{ uploadProgress }}%
            </div>
            <el-progress
              :percentage="uploadProgress"
              :status="uploadProgress === 100 ? 'success' : undefined"
              :stroke-width="12"
              style="margin-top: 6px"
            />
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="uploadDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitUpload" :loading="uploadLoading">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 内容比较对话框 -->
    <el-dialog
      v-model="diffDialogVisible"
      title="内容比较"
      width="95%"
      :close-on-click-modal="false"
      class="diff-dialog"
    >
      <div v-if="diffData && diffLines.length > 0" class="diff-container">
        <div class="diff-header">
          <div class="diff-label original">
            <span>原始内容 (content_org)</span>
          </div>
          <div class="diff-label current">
            <span>当前内容 (content)</span>
          </div>
        </div>
        <div class="diff-content-wrapper">
          <div class="diff-content">
            <table class="diff-table">
              <tbody>
                <tr
                  v-for="(line, index) in diffLines"
                  :key="index"
                  :class="getDiffLineClass(line)"
                >
                  <td class="diff-line-number original" :class="{ empty: !line.oldLine }">
                    {{ line.oldLine || '' }}
                  </td>
                  <td class="diff-line-content original">
                    <span v-if="line.oldContent">{{ line.oldContent }}</span>
                    <span v-else class="empty-line">&nbsp;</span>
                  </td>
                  <td class="diff-line-number current" :class="{ empty: !line.newLine }">
                    {{ line.newLine || '' }}
                  </td>
                  <td class="diff-line-content current">
                    <span v-if="line.newContent">{{ line.newContent }}</span>
                    <span v-else class="empty-line">&nbsp;</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
      <div v-else-if="diffData && diffLines.length === 0" class="diff-no-changes">
        <el-icon><Check /></el-icon>
        <span>内容相同，没有差异</span>
      </div>
      <div v-else class="diff-loading">
        <el-icon class="is-loading"><Loading /></el-icon>
        <span>加载中...</span>
      </div>
      <template #footer>
        <el-button @click="diffDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- LLM Markdown 规范化预览 -->
    <el-dialog
      v-model="normalizeDialogVisible"
      title="LLM Markdown 规范化"
      width="92%"
      :close-on-click-modal="false"
      class="normalize-md-dialog"
      destroy-on-close
    >
      <p class="normalize-hint">
        左侧为参与规范化的原文，右侧为模型输出。确认无误后可直接保存到知识库，或复制后去「编辑」继续人工调整。
      </p>
      <el-row :gutter="16" class="normalize-panels">
        <el-col :span="12">
          <div class="normalize-panel-title">原文</div>
          <el-input v-model="normalizeOriginal" type="textarea" :rows="22" readonly class="normalize-textarea" />
        </el-col>
        <el-col :span="12">
          <div class="normalize-panel-title">规范化后</div>
          <el-input v-model="normalizeFormatted" type="textarea" :rows="22" class="normalize-textarea" />
        </el-col>
      </el-row>
      <template #footer>
        <el-button @click="normalizeDialogVisible = false">关闭</el-button>
        <el-button type="success" :loading="normalizeSaveLoading" @click="handleMarkdownNormalizeSave">
          保存到知识库
        </el-button>
      </template>
    </el-dialog>

    <!-- Qdrant 分块查看 -->
    <el-dialog
      v-model="qdrantChunksDialogVisible"
      title="Qdrant 分块情况"
      width="900px"
      class="qdrant-chunks-dialog"
      destroy-on-close
      @closed="onQdrantChunksDialogClosed"
    >
      <div v-if="qdrantChunksData" class="qdrant-chunks-wrap">
        <el-descriptions :column="1" border size="small" class="qdrant-chunks-meta">
          <el-descriptions-item label="文件名">{{ qdrantChunksData.fileName }}</el-descriptions-item>
          <el-descriptions-item label="知识库">{{ getDocumentTypeName(qdrantChunksData.documentCode) }} ({{ qdrantChunksData.documentCode }})</el-descriptions-item>
          <el-descriptions-item label="主分块集合">{{ qdrantChunksData.mainCollection }} · 共 {{ qdrantChunksData.mainTotal }} 块</el-descriptions-item>
          <el-descriptions-item label="全文概要集合">{{ qdrantChunksData.summaryCollection }} · 共 {{ qdrantChunksData.summaryTotal }} 块</el-descriptions-item>
        </el-descriptions>
        <el-tabs v-model="qdrantChunksActiveTab" class="qdrant-chunks-tabs">
          <el-tab-pane label="正文分块" name="main">
            <el-table :data="qdrantChunksData.mainChunks || []" border stripe max-height="420" size="small">
              <el-table-column prop="pageIndex" label="块序号" width="88" />
              <el-table-column prop="totalPages" label="总块数" width="88" />
              <el-table-column prop="length" label="字数" width="88" />
              <el-table-column prop="contentPreview" label="预览" min-width="200" show-overflow-tooltip />
              <el-table-column label="操作" width="100" fixed="right">
                <template #default="{ row }">
                  <el-button type="primary" link size="small" @click="openChunkDetail(row)">全文</el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-tab-pane>
          <el-tab-pane label="全文概要" name="summary">
            <el-table :data="qdrantChunksData.summaryChunks || []" border stripe max-height="420" size="small">
              <el-table-column prop="pageIndex" label="块序号" width="88" />
              <el-table-column prop="totalPages" label="总块数" width="88" />
              <el-table-column prop="length" label="字数" width="88" />
              <el-table-column prop="contentPreview" label="预览" min-width="200" show-overflow-tooltip />
              <el-table-column label="操作" width="100" fixed="right">
                <template #default="{ row }">
                  <el-button type="primary" link size="small" @click="openChunkDetail(row)">全文</el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-tab-pane>
        </el-tabs>
      </div>
      <template #footer>
        <el-button type="primary" @click="qdrantChunksDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <el-drawer
      v-model="chunkDetailDrawerVisible"
      :title="chunkDetailTitle"
      size="60%"
      destroy-on-close
    >
      <pre class="chunk-detail-pre">{{ chunkDetailText }}</pre>
    </el-drawer>

    <el-dialog
      v-model="qaPairsDialogVisible"
      title="问答队列"
      width="72%"
      @closed="onQaPairsDialogClosed"
    >
      <div class="qa-pairs-header" v-if="qaPairsData">
        <el-tag type="info">文件：{{ qaPairsData.fileName || '-' }}</el-tag>
        <el-tag type="success">知识库：{{ getDocumentTypeName(qaPairsData.documentCode) || qaPairsData.documentCode || '-' }}</el-tag>
        <el-tag>总数：{{ qaPairsData.total || 0 }}</el-tag>
      </div>
      <el-table :data="qaPairsData?.list || []" border stripe max-height="460" size="small">
        <el-table-column prop="chunkIndex" label="分块" width="90" />
        <el-table-column prop="question" label="Q（content）" min-width="260" show-overflow-tooltip />
        <el-table-column prop="answer" label="A（辅助结果）" min-width="360" show-overflow-tooltip />
        <el-table-column prop="qdrantPointId" label="Qdrant点ID" min-width="220" show-overflow-tooltip />
      </el-table>
      <template #footer>
        <el-button type="primary" @click="qaPairsDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onBeforeUnmount, computed, watch, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Upload, Search, Refresh, UploadFilled, Edit, Document, Files, Check, Loading, MagicStick } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import {
  listRawDocuments,
  uploadRawDocuments,
  deleteRawDocuments,
  listDocumentsType,
  changeFileDocumentType,
  changeRawDocumentsTag,
  auditRawDocuments,
  cancelAuditRawDocuments,
  downloadRawDocumentsFile,
  updateRawDocumentsContent,
  getRawDocumentsContentDiff,
  getRawDocumentQdrantChunks,
  listRawDocumentQaPairs,
  previewRawDocumentsMarkdownNormalize
} from '@/api/knowdata'
import { diffLines as diffLinesFn } from 'diff'
import { convertDocumentToMD, convertDocumentToZIP, cancelConvertDocument, listFrTag } from '@/api/knowsource'
import { encodeListStateForRoute } from '@/utils/rawDocumentsListNavigation'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const hasPermission = (p) => (userStore.empPermissions || []).includes(p)

/** 仅当当前访问地址的域名或端口中含 6280 时显示「分块」按钮 */
const showQdrantChunksButton = computed(() => {
  if (typeof window === 'undefined') return false
  const { host, port, hostname } = window.location
  const h = String(host || '')
  const p = String(port || '')
  const hn = String(hostname || '')
  return p === '6280' || h.includes('6280') || hn.includes('6280')
})

const loading = ref(false)
const uploadLoading = ref(false)
const uploadDialogVisible = ref(false)
const uploadFormRef = ref(null)
const uploadRef = ref(null)
const tableData = ref([])
const selectedRows = ref([])
const currentFile = ref(null)
const documentsTypeList = ref([])
const changeTypeDialogVisible = ref(false)
const changeTypeLoading = ref(false)
const changeTypeFormRef = ref(null)
const changeTagDialogVisible = ref(false)
const changeTagLoading = ref(false)
const changeTagFormRef = ref(null)
const uploadProgress = ref(0)
const availableTags = ref([])
const availableTagsForUpload = ref([])
const convertingIds = ref([])

const mdContentDialogVisible = ref(false)
const mdContentData = reactive({
  id: 0,
  fileName: '',
  content: ''
})
const zipResultDialogVisible = ref(false)
const zipResultData = reactive({
  id: 0,
  fileName: '',
  zipFilePath: '',
  extractedDir: '',
  fileList: []
})

const compareLoadingId = ref(null)
const qdrantChunksLoadingId = ref(null)
const qdrantChunksDialogVisible = ref(false)
const qdrantChunksData = ref(null)
const qdrantChunksActiveTab = ref('main')
const chunkDetailDrawerVisible = ref(false)
const chunkDetailTitle = ref('')
const chunkDetailText = ref('')
const qaPairsLoadingId = ref(null)
const qaPairsDialogVisible = ref(false)
const qaPairsData = ref(null)
const diffDialogVisible = ref(false)
const diffData = ref(null)
const diffLines = ref([])

const normalizeDialogVisible = ref(false)
const normalizeLoadingId = ref(null)
const normalizeSaveLoading = ref(false)
const normalizeOriginal = ref('')
const normalizeFormatted = ref('')
const normalizeDocId = ref(0)

// 批量识别进度
const batchRecognizeDialogVisible = ref(false)
const batchRecognizeProgress = ref(0)
const batchRecognizeCurrent = ref(0)
const batchRecognizeTotal = ref(0)
const batchRecognizeFileName = ref('')
const batchRecognizeDone = ref(false)
const batchRecognizeSuccessCount = ref(0)
const batchRecognizeFailCount = ref(0)
const batchRecognizeFailList = ref([])

const batchRecognizeEligibleCount = computed(() =>
  selectedRows.value.filter((r) => r.isAudit !== 1 && r.isToMd !== 1).length
)

// 批量审核：未审核且已转MD的文档才可审核
const batchAuditDialogVisible = ref(false)
const batchAuditProgress = ref(0)
const batchAuditCurrent = ref(0)
const batchAuditTotal = ref(0)
const batchAuditFileName = ref('')
const batchAuditDone = ref(false)
const batchAuditSuccessCount = ref(0)
const batchAuditFailCount = ref(0)
const batchAuditFailList = ref([])
const batchAuditEligibleCount = computed(() =>
  selectedRows.value.filter((r) => r.isAudit !== 1 && r.isToMd === 1).length
)

// 批量取消审核：仅已审核文档可取消
const batchUnAuditDialogVisible = ref(false)
const batchUnAuditProgress = ref(0)
const batchUnAuditCurrent = ref(0)
const batchUnAuditTotal = ref(0)
const batchUnAuditFileName = ref('')
const batchUnAuditDone = ref(false)
const batchUnAuditSuccessCount = ref(0)
const batchUnAuditFailCount = ref(0)
const batchUnAuditFailList = ref([])
const batchUnAuditEligibleCount = computed(() =>
  selectedRows.value.filter((r) => r.isAudit === 1).length
)

/** 文档 status 含「正在」→ 后台处理中，需定时刷新列表 */
const rawDocStatusText = (row) => String(row?.status || '')
const rawDocStatusIsFailed = (row) => {
  const s = rawDocStatusText(row)
  return s === '识别失败' || s === '入库失败'
}
const rawDocStatusTagType = (row) => (rawDocStatusIsFailed(row) ? 'danger' : 'info')
const rawDocStatusIsBusy = (row) => !rawDocStatusIsFailed(row) && rawDocStatusText(row).includes('正在')
const rawDocStatusBusyInserting = (row) => rawDocStatusText(row).includes('正在入库')
const rawDocStatusBusyExtracting = (row) => {
  const s = rawDocStatusText(row)
  return s.includes('正在提取') || s.includes('正在转文字')
}
const rawDocStatusBusyRemoving = (row) => rawDocStatusText(row).includes('正在出库')
/** 入库/提取/出库进行中：仅对应「中断」可点，其余操作置灰 */
const rawDocRowOpsLocked = (row) =>
  rawDocStatusBusyInserting(row) || rawDocStatusBusyExtracting(row) || rawDocStatusBusyRemoving(row)
const rawDocRowSelectable = (row) => !rawDocRowOpsLocked(row)

const STATUS_POLL_MS = 5000
let statusPollTimerId = null

const tableHasBusyStatus = () => tableData.value.some((r) => rawDocStatusIsBusy(r))

const clearStatusPollTimer = () => {
  if (statusPollTimerId != null) {
    clearInterval(statusPollTimerId)
    statusPollTimerId = null
  }
}

const syncStatusPollTimer = () => {
  if (!tableHasBusyStatus()) {
    clearStatusPollTimer()
    return
  }
  if (statusPollTimerId != null) return
  statusPollTimerId = setInterval(() => {
    loadData({ silent: true })
  }, STATUS_POLL_MS)
}

const searchForm = reactive({
  documentCode: '',
  fileName: '',
  tag: '',
  isAudit: ''
})

const uploadForm = reactive({
  documentCode: '',
  file: null,
  tag: ''
})

const uploadRules = {
  documentCode: [
    { required: true, message: '请输入知识库编码', trigger: 'blur' }
  ],
  file: [
    { required: true, message: '请选择文件', trigger: 'change' }
  ]
}

const changeTypeForm = reactive({
  fileName: '',
  oldDocumentCode: '',
  newDocumentCode: ''
})

const changeTypeRules = {
  newDocumentCode: [
    { required: true, message: '请选择新知识库', trigger: 'change' }
  ]
}

const changeTagForm = reactive({
  id: 0,
  fileName: '',
  tag: ''
})

const changeTagRules = {
  tag: [
    { required: false, message: '请输入标签', trigger: 'blur' }
  ]
}


const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

/** 从路由 query 恢复分页与筛选（从编辑页返回时带 query） */
const applyListQueryFromRoute = (q) => {
  if (!q || typeof q !== 'object') return
  const p = parseInt(q.page, 10)
  const ps = parseInt(q.pageSize, 10)
  pagination.page = Number.isFinite(p) && p > 0 ? p : 1
  pagination.pageSize = Number.isFinite(ps) && ps > 0 ? ps : 10
  searchForm.documentCode = q.documentCode != null ? String(q.documentCode) : ''
  searchForm.fileName = q.fileName != null ? String(q.fileName) : ''
  searchForm.tag = q.tag != null ? String(q.tag) : ''
  searchForm.isAudit = q.isAudit != null && q.isAudit !== '' ? String(q.isAudit) : ''
}

/** 进入编辑页时携带，返回后可还原列表状态 */
const getListStateSnapshot = () => ({
  page: pagination.page,
  pageSize: pagination.pageSize,
  documentCode: searchForm.documentCode,
  fileName: searchForm.fileName,
  tag: searchForm.tag,
  isAudit: searchForm.isAudit
})

const formatTime = (timestamp) => {
  if (!timestamp) return '-'
  const date = new Date(timestamp * 1000)
  return date.toLocaleString('zh-CN')
}

const formatFileSize = (bytes) => {
  if (!bytes) return '-'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB'
  if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(2) + ' MB'
  return (bytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB'
}

const loadData = async (options = {}) => {
  const silent = options.silent === true
  if (!silent) {
    loading.value = true
  }
  try {
    const res = await listRawDocuments({
      page: pagination.page,
      pageSize: pagination.pageSize,
      ...searchForm
    })
    if (res.code === 200 && res.data) {
      tableData.value = res.data.list || []
      pagination.total = res.data.total || 0
    }
  } catch (error) {
    if (!silent) {
      ElMessage.error('加载数据失败')
    }
  } finally {
    if (!silent) {
      loading.value = false
    }
    syncStatusPollTimer()
  }
}

const handleSearch = () => {
  pagination.page = 1
  loadData()
}

const handleReset = () => {
  Object.assign(searchForm, {
    documentCode: '',
    fileName: '',
    tag: '',
    isAudit: ''
  })
  handleSearch()
}

const handleUpload = () => {
  Object.assign(uploadForm, {
    documentCode: '',
    file: null,
    tag: ''
  })
  currentFile.value = null
  uploadProgress.value = 0
  availableTagsForUpload.value = []
  if (uploadRef.value) {
    uploadRef.value.clearFiles()
  }
  uploadDialogVisible.value = true
}

const applySelectedUploadFile = (rawFile) => {
  currentFile.value = rawFile
  uploadForm.file = rawFile
}

const handleFileChange = (file) => {
  applySelectedUploadFile(file.raw)
}

const handleUploadExceed = (files) => {
  uploadRef.value?.clearFiles()
  const raw = files[0]
  if (!raw) return
  uploadRef.value?.handleStart(raw)
  applySelectedUploadFile(raw)
}

const handleFileRemove = () => {
  currentFile.value = null
  uploadForm.file = null
}

const handleSubmitUpload = async () => {
  if (!uploadFormRef.value) return
  
  await uploadFormRef.value.validate(async (valid) => {
    if (valid) {
      if (!currentFile.value) {
        ElMessage.warning('请选择文件')
        return
      }
      
      uploadLoading.value = true
      uploadProgress.value = 0
      try {
        const formData = new FormData()
        formData.append('file', currentFile.value)
        formData.append('fileName', currentFile.value.name)
        formData.append('fileType', currentFile.value.type || '')
        formData.append('documentCode', uploadForm.documentCode)
        if (uploadForm.tag) {
          formData.append('tag', uploadForm.tag)
        }
        
        const res = await uploadRawDocuments(formData, (progressEvent) => {
          if (progressEvent.total) {
            uploadProgress.value = Math.round((progressEvent.loaded * 100) / progressEvent.total)
          }
        })
        if (res.code === 200) {
          uploadProgress.value = 100
          ElMessage.success('上传成功')
          setTimeout(() => {
            uploadDialogVisible.value = false
            uploadProgress.value = 0
            loadData()
          }, 500)
        } else {
          ElMessage.error(res.message || res.msg || '上传失败')
          uploadProgress.value = 0
        }
      } catch (error) {
        const errorMessage = error?.response?.data?.message || error?.message || '上传失败，请稍后重试'
        ElMessage.error(errorMessage)
        uploadProgress.value = 0
      } finally {
        uploadLoading.value = false
      }
    }
  })
}

const handleDelete = (row) => {
  ElMessageBox.confirm('确定要删除这条记录吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      const res = await deleteRawDocuments({ ids: [row.id] })
      if (res.code === 200) {
        ElMessage.success('删除成功')
        loadData()
      } else {
        ElMessage.error(res.msg || '删除失败')
      }
    } catch (error) {
      ElMessage.error('删除失败，请稍后重试')
    }
  }).catch(() => {})
}

const handleBatchDelete = () => {
  if (selectedRows.value.length === 0) {
    ElMessage.warning('请选择要删除的记录')
    return
  }
  
  ElMessageBox.confirm(`确定要删除选中的 ${selectedRows.value.length} 条记录吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      const ids = selectedRows.value.map(row => row.id)
      const res = await deleteRawDocuments({ ids })
      if (res.code === 200) {
        ElMessage.success('删除成功')
        selectedRows.value = []
        loadData()
      } else {
        ElMessage.error(res.msg || '删除失败')
      }
    } catch (error) {
      ElMessage.error('删除失败，请稍后重试')
    }
  }).catch(() => {})
}

const handleSizeChange = () => {
  loadData()
}

const handlePageChange = () => {
  loadData()
}

const handleSelectionChange = (selection) => {
  selectedRows.value = selection
}

const handleFileNameClick = (row) => {
  // 点击文件名跳转到 md-preview 预览页（带 list 以便返回时还原分页）
  router.push({
    name: 'MdPreview',
    params: { id: row.id },
    query: {
      list: encodeListStateForRoute(getListStateSnapshot())
    }
  })
}

const handleDownload = async (row) => {
  try {
    const resp = await downloadRawDocumentsFile(row.id)
    const blob = resp.data

    // Try to read filename from Content-Disposition, fallback to row.fileName
    const cd = resp.headers?.['content-disposition'] || resp.headers?.['Content-Disposition']
    let filename = row.fileName || `rawdoc-${row.id}`
    if (cd) {
      const m = /filename\*\s*=\s*UTF-8''([^;]+)/i.exec(cd)
      if (m && m[1]) {
        try {
          filename = decodeURIComponent(m[1])
        } catch (e) {
          // ignore decode error
        }
      }
    }

    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    window.URL.revokeObjectURL(url)
  } catch (error) {
    // 兼容后端返回：纯文本 / JSON / blob(JSON)
    try {
      const resp = error?.response
      const data = resp?.data

      // axios + responseType=blob 时，错误也可能是 Blob
      if (data instanceof Blob) {
        const text = await data.text()
        try {
          const json = JSON.parse(text)
          ElMessage.error(json?.message || json?.msg || json?.info || text || '下载失败')
        } catch (e) {
          ElMessage.error(text || '下载失败')
        }
        return
      }

      if (typeof data === 'string') {
        ElMessage.error(data || '下载失败')
        return
      }

      if (data && typeof data === 'object') {
        ElMessage.error(data?.message || data?.msg || data?.info || '下载失败')
        return
      }
    } catch (e) {
      // ignore
    }

    ElMessage.error(error?.message || '下载失败，请稍后重试')
  }
}

const processDiff = (contentOrg, content) => {
  const org = contentOrg || ''
  const cur = content || ''
  const changes = diffLinesFn(org, cur)
  const lines = []
  let oldLineNum = 0
  let newLineNum = 0
  changes.forEach((change) => {
    const changeLines = change.value.split('\n')
    if (changeLines.length > 0 && changeLines[changeLines.length - 1] === '') {
      changeLines.pop()
    }
    if (change.added) {
      changeLines.forEach((line) => {
        newLineNum++
        lines.push({ oldLine: null, newLine: newLineNum, oldContent: null, newContent: line, type: 'added' })
      })
    } else if (change.removed) {
      changeLines.forEach((line) => {
        oldLineNum++
        lines.push({ oldLine: oldLineNum, newLine: null, oldContent: line, newContent: null, type: 'removed' })
      })
    } else {
      changeLines.forEach((line) => {
        oldLineNum++
        newLineNum++
        lines.push({ oldLine: oldLineNum, newLine: newLineNum, oldContent: line, newContent: line, type: 'normal' })
      })
    }
  })
  return lines
}

const getDiffLineClass = (line) => ({
  'diff-line-removed': line.type === 'removed',
  'diff-line-added': line.type === 'added',
  'diff-line-normal': line.type === 'normal'
})

const onQdrantChunksDialogClosed = () => {
  qdrantChunksData.value = null
}

const handleViewQdrantChunks = async (row) => {
  qdrantChunksLoadingId.value = row.id
  try {
    const res = await getRawDocumentQdrantChunks({ id: row.id })
    if (res.code === 200 && res.data) {
      qdrantChunksData.value = res.data
      qdrantChunksActiveTab.value = 'main'
      qdrantChunksDialogVisible.value = true
    } else {
      ElMessage.error(res.message || res.msg || res.info || '查询分块失败')
    }
  } catch (e) {
    ElMessage.error(e?.message || '查询分块失败')
  } finally {
    qdrantChunksLoadingId.value = null
  }
}

const openChunkDetail = (chunkRow) => {
  chunkDetailTitle.value = `Qdrant 点 ${chunkRow.qdrantId || '-'} · 块 ${chunkRow.pageIndex}/${chunkRow.totalPages || '-'}`
  chunkDetailText.value = chunkRow.content || chunkRow.contentPreview || '（无文本）'
  chunkDetailDrawerVisible.value = true
}

const onQaPairsDialogClosed = () => {
  qaPairsData.value = null
}

const handleViewQaPairs = async (row) => {
  qaPairsLoadingId.value = row.id
  try {
    const res = await listRawDocumentQaPairs({
      id: row.id,
      page: 1,
      pageSize: 200
    })
    if (res.code === 200) {
      qaPairsData.value = {
        list: res?.data?.list || [],
        total: res?.data?.total || 0,
        fileName: row.fileName,
        documentCode: row.documentCode
      }
      qaPairsDialogVisible.value = true
    } else {
      ElMessage.error(res.message || res.msg || res.info || '查询问答队列失败')
    }
  } catch (e) {
    ElMessage.error(e?.message || '查询问答队列失败')
  } finally {
    qaPairsLoadingId.value = null
  }
}

const handleCompare = async (row) => {
  diffDialogVisible.value = true
  diffData.value = null
  diffLines.value = []
  compareLoadingId.value = row.id
  try {
    const res = await getRawDocumentsContentDiff({ id: row.id })
    if (res.code === 200 && res.data) {
      diffData.value = res.data
      diffLines.value = processDiff(res.data.contentOrg, res.data.content)
    } else {
      ElMessage.error(res.message || '获取内容差异失败')
      diffDialogVisible.value = false
    }
  } catch (error) {
    ElMessage.error('获取内容差异失败，请稍后重试')
    diffDialogVisible.value = false
  } finally {
    compareLoadingId.value = null
  }
}

const handleMarkdownNormalizePreviewRow = async (row) => {
  if (!row) return
  normalizeLoadingId.value = row.id
  try {
    const res = await previewRawDocumentsMarkdownNormalize({ id: row.id })
    if (res.code === 200 && res.data) {
      normalizeDocId.value = row.id
      normalizeOriginal.value = res.data.originalContent ?? ''
      normalizeFormatted.value = res.data.formattedContent ?? ''
      normalizeDialogVisible.value = true
    } else {
      ElMessage.error(res.message || 'LLM 规范化失败')
    }
  } catch (e) {
    ElMessage.error('请求失败，请稍后重试')
  } finally {
    normalizeLoadingId.value = null
  }
}

const handleMarkdownNormalizeSave = async () => {
  if (!normalizeDocId.value) return
  normalizeSaveLoading.value = true
  try {
    const res = await updateRawDocumentsContent({
      id: normalizeDocId.value,
      content: normalizeFormatted.value
    })
    if (res.code === 200) {
      ElMessage.success('已保存规范化内容')
      normalizeDialogVisible.value = false
      await fetchTableData()
    } else {
      ElMessage.error(res.message || '保存失败')
    }
  } catch (e) {
    ElMessage.error('保存失败，请稍后重试')
  } finally {
    normalizeSaveLoading.value = false
  }
}

const handleEdit = (row) => {
  // 跳转到编辑页面（非审核且 isToMd=1 的文档可以编辑）
  if (row.isAudit !== 1 && row.isToMd === 1) {
    router.push({
      name: 'RawDocumentContent',
      params: { id: row.id },
      query: {
        edit: 'true',
        list: encodeListStateForRoute(getListStateSnapshot())
      }
    })
  }
}

const handleChangeDocumentType = (row) => {
  Object.assign(changeTypeForm, {
    fileName: row.fileName,
    oldDocumentCode: row.documentCode,
    newDocumentCode: ''
  })
  changeTypeDialogVisible.value = true
}

const handleChangeTag = (row) => {
  Object.assign(changeTagForm, {
    id: row.id,
    fileName: row.fileName,
    tag: row.tag || ''
  })
  changeTagDialogVisible.value = true
}

const handleAudit = async (row) => {
  const isAudit = row.isAudit === 1
  const action = isAudit ? '取消审核' : '审核'
  const newIsAudit = isAudit ? 0 : 1
  
  try {
    await ElMessageBox.confirm(
      `确定要${action}文档 "${row.fileName}" 吗？`,
      `${action}确认`,
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    try {
      const res = await auditRawDocuments({
        id: row.id,
        isAudit: newIsAudit
      })
      if (res.code === 200) {
        const successText = res.info ? `${action}操作成功：${res.info}` : `${action}操作成功`
        ElMessage.success(successText)
        await loadData()
      } else {
        ElMessage.error(res.message || res.msg || `${action}操作失败`)
      }
    } catch (error) {
      ElMessage.error(`${action}操作失败，请稍后重试`)
    }
  } catch (error) {
    // 用户取消操作，不做任何处理
  }
}

const handleSubmitChangeTag = async () => {
  if (!changeTagFormRef.value) return
  
  await changeTagFormRef.value.validate(async (valid) => {
    if (valid) {
      changeTagLoading.value = true
      try {
        const res = await changeRawDocumentsTag({
          id: changeTagForm.id,
          tag: changeTagForm.tag || ''
        })
        if (res.code === 200) {
          ElMessage.success('更改标签成功')
          changeTagDialogVisible.value = false
          loadData()
        } else {
          ElMessage.error(res.message || res.msg || '更改标签失败')
        }
      } catch (error) {
        ElMessage.error('更改标签失败，请稍后重试')
      } finally {
        changeTagLoading.value = false
      }
    }
  })
}

const getDocumentTypeName = (code) => {
  if (!code) return '-'
  const docType = documentsTypeList.value.find(item => item.code === code)
  return docType ? docType.name : code
}

const handleSubmitChangeType = async () => {
  if (!changeTypeFormRef.value) return
  
  await changeTypeFormRef.value.validate(async (valid) => {
    if (valid) {
      if (!changeTypeForm.newDocumentCode) {
        ElMessage.warning('请选择新知识库')
        return
      }
      
      if (changeTypeForm.oldDocumentCode === changeTypeForm.newDocumentCode) {
        ElMessage.warning('新知识库不能与原知识库相同')
        return
      }
      
      changeTypeLoading.value = true
      try {
        const res = await changeFileDocumentType({
          fileName: changeTypeForm.fileName,
          oldDocumentCode: changeTypeForm.oldDocumentCode,
          newDocumentCode: changeTypeForm.newDocumentCode
        })
        if (res.code === 200) {
          ElMessage.success('更改知识库成功')
          changeTypeDialogVisible.value = false
          loadData()
        } else {
          ElMessage.error(res.message || res.msg || '更改知识库失败')
        }
      } catch (error) {
        ElMessage.error('更改知识库失败，请稍后重试')
      } finally {
        changeTypeLoading.value = false
      }
    }
  })
}

const loadDocumentsTypeList = async () => {
  try {
    const res = await listDocumentsType({})
    if (res.code === 200 && res.data) {
      documentsTypeList.value = res.data.list || []
    }
  } catch (error) {
    console.error('加载知识库列表失败', error)
  }
}

const loadAvailableTags = async () => {
  try {
    const res = await listFrTag({
      page: 1,
      pageSize: 1000
    })
    if (res.code === 200 && res.data) {
      // 从标签列表中提取 tag 字段
      availableTags.value = (res.data.list || []).map(item => item.tag).filter(tag => tag)
    }
  } catch (error) {
    console.error('加载标签列表失败', error)
    availableTags.value = []
  }
}

const handleDocumentCodeChange = async () => {
  // 标签列表不依赖知识库，使用全局标签列表
  availableTagsForUpload.value = availableTags.value
}

// 转换文档为 MD
const handleConvertToMD = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定要将文档 "${row.fileName}" 转换为 Markdown 吗？`,
      '确认转换',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    convertingIds.value.push(row.id)
    const res = await convertDocumentToMD({ id: row.id })
    if (res.code === 200) {
      ElMessage.success(res.message || '已提交识别任务，后台处理中')
      loadData()
    } else {
      ElMessage.error(res.message || res.info || '转换失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '转换失败')
    }
  } finally {
    const index = convertingIds.value.indexOf(row.id)
    if (index > -1) {
      convertingIds.value.splice(index, 1)
    }
  }
}

const handleConvertToZIP = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定要将文档 "${row.fileName}" 转换为 MD 文件吗？`,
      '确认转换',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    convertingIds.value.push(row.id)
    const res = await convertDocumentToZIP({ id: row.id })
    if (res.code === 200) {
      ElMessage.success(res.message || '已提交识别任务，后台处理中')
      loadData()
    } else {
      ElMessage.error(res.message || res.info || '转换失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '转换失败')
    }
  } finally {
    const index = convertingIds.value.indexOf(row.id)
    if (index > -1) {
      convertingIds.value.splice(index, 1)
    }
  }
}

const handleCancelRecognize = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定要中断识别文档 "${row.fileName}" 吗？`,
      '中断识别确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
  } catch (e) {
    return
  }
  try {
    const res = await cancelConvertDocument({ id: row.id })
    if (res.info) {
      console.debug('[cancel-convert]', row.id, res.info)
    }
    if (res.code === 200) {
      ElMessage.success(res.message || '已提交取消请求')
      loadData()
    } else {
      const detail = res.info ? ` ${res.info}` : ''
      ElMessage.error((res.message || '取消失败') + detail)
    }
  } catch (e) {
    ElMessage.error(e?.message || '取消失败')
  }
}

const handleCancelAudit = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定要中断审核入库文档 "${row.fileName}" 吗？`,
      '中断审核入库确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
  } catch (e) {
    return
  }
  try {
    const res = await cancelAuditRawDocuments({ id: row.id })
    if (res.code === 200) {
      ElMessage.success(res.message || '已提交取消请求')
      loadData()
    } else {
      ElMessage.error(res.message || '取消失败')
    }
  } catch (e) {
    ElMessage.error(e?.message || '取消失败')
  }
}

const handleBatchRecognize = async () => {
  const eligible = selectedRows.value.filter((r) => r.isAudit !== 1 && r.isToMd !== 1)
  if (eligible.length === 0) {
    ElMessage.warning('请选择未审核且未识别（转MD）的文档进行识别')
    return
  }
  batchRecognizeDialogVisible.value = true
  batchRecognizeDone.value = false
  batchRecognizeTotal.value = eligible.length
  batchRecognizeCurrent.value = 0
  batchRecognizeProgress.value = 0
  batchRecognizeFileName.value = ''
  batchRecognizeSuccessCount.value = 0
  batchRecognizeFailCount.value = 0
  batchRecognizeFailList.value = []

  for (let i = 0; i < eligible.length; i++) {
    const row = eligible[i]
    batchRecognizeFileName.value = row.fileName
    batchRecognizeCurrent.value = i + 1
    batchRecognizeProgress.value = Math.round(((i + 1) / eligible.length) * 100)
    try {
      const res = await convertDocumentToZIP({ id: row.id })
      if (res.code === 200) {
        batchRecognizeSuccessCount.value += 1
      } else {
        batchRecognizeFailCount.value += 1
        batchRecognizeFailList.value.push({
          fileName: row.fileName,
          message: res.message || res.info || '转换失败'
        })
      }
    } catch (err) {
      batchRecognizeFailCount.value += 1
      batchRecognizeFailList.value.push({
        fileName: row.fileName,
        message: err.message || '请求异常'
      })
    }
  }

  batchRecognizeProgress.value = 100
  batchRecognizeDone.value = true
  batchRecognizeFileName.value = ''
  loadData()
}

const handleBatchAudit = async () => {
  const eligible = selectedRows.value.filter((r) => r.isAudit !== 1 && r.isToMd === 1)
  if (eligible.length === 0) {
    ElMessage.warning('请选择未审核且已转MD的文档进行审核')
    return
  }
  batchAuditDialogVisible.value = true
  batchAuditDone.value = false
  batchAuditTotal.value = eligible.length
  batchAuditCurrent.value = 0
  batchAuditProgress.value = 0
  batchAuditFileName.value = ''
  batchAuditSuccessCount.value = 0
  batchAuditFailCount.value = 0
  batchAuditFailList.value = []

  for (let i = 0; i < eligible.length; i++) {
    const row = eligible[i]
    batchAuditFileName.value = row.fileName
    batchAuditCurrent.value = i + 1
    batchAuditProgress.value = Math.round(((i + 1) / eligible.length) * 100)
    try {
      const res = await auditRawDocuments({
        id: row.id,
        isAudit: 1
      })
      if (res.code === 200) {
        batchAuditSuccessCount.value += 1
      } else {
        batchAuditFailCount.value += 1
        batchAuditFailList.value.push({
          fileName: row.fileName,
          message: res.message || res.msg || res.info || '审核失败'
        })
      }
    } catch (err) {
      batchAuditFailCount.value += 1
      batchAuditFailList.value.push({
        fileName: row.fileName,
        message: err.message || '请求异常'
      })
    }
  }

  batchAuditProgress.value = 100
  batchAuditDone.value = true
  batchAuditFileName.value = ''
  loadData()
}

const handleBatchUnAudit = async () => {
  const eligible = selectedRows.value.filter((r) => r.isAudit === 1)
  if (eligible.length === 0) {
    ElMessage.warning('请选择已审核的文档进行取消审核')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要取消审核选中的 ${eligible.length} 个文档吗？`,
      '取消审核确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
  } catch (e) {
    return
  }

  batchUnAuditDialogVisible.value = true
  batchUnAuditDone.value = false
  batchUnAuditTotal.value = eligible.length
  batchUnAuditCurrent.value = 0
  batchUnAuditProgress.value = 0
  batchUnAuditFileName.value = ''
  batchUnAuditSuccessCount.value = 0
  batchUnAuditFailCount.value = 0
  batchUnAuditFailList.value = []

  for (let i = 0; i < eligible.length; i++) {
    const row = eligible[i]
    batchUnAuditFileName.value = row.fileName
    batchUnAuditCurrent.value = i + 1
    batchUnAuditProgress.value = Math.round(((i + 1) / eligible.length) * 100)
    try {
      const res = await auditRawDocuments({
        id: row.id,
        isAudit: 0
      })
      if (res.code === 200) {
        batchUnAuditSuccessCount.value += 1
      } else {
        batchUnAuditFailCount.value += 1
        batchUnAuditFailList.value.push({
          fileName: row.fileName,
          message: res.message || res.msg || res.info || '取消审核失败'
        })
      }
    } catch (err) {
      batchUnAuditFailCount.value += 1
      batchUnAuditFailList.value.push({
        fileName: row.fileName,
        message: err.message || '请求异常'
      })
    }
  }

  batchUnAuditProgress.value = 100
  batchUnAuditDone.value = true
  batchUnAuditFileName.value = ''
  loadData()
}

// 确认 MD 内容（内容已经在转换时保存，这里只是关闭对话框）
const handleConfirmMDContent = () => {
  mdContentDialogVisible.value = false
  ElMessage.success('Markdown 内容已保存')
}

watch(
  () => route.query,
  () => {
    if (route.name !== 'RawDocuments') return
    applyListQueryFromRoute(route.query)
    loadData()
    if (route.query.openUpload === '1') {
      nextTick(() => {
        handleUpload()
        const q = { ...route.query }
        delete q.openUpload
        router.replace({ path: route.path, query: q })
      })
    }
  },
  { deep: true, immediate: true }
)

onMounted(() => {
  loadDocumentsTypeList()
  loadAvailableTags()
})

onBeforeUnmount(() => {
  clearStatusPollTimer()
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 18px;
  font-weight: 500;
}

.search-form {
  margin-bottom: 20px;
}

.batch-actions {
  margin-top: 20px;
  margin-bottom: 10px;
}

.batch-recognize-content {
  padding: 8px 0;
}
.batch-recognize-current {
  margin-bottom: 12px;
  color: var(--el-text-color-regular);
  font-size: 14px;
}
.batch-recognize-stats {
  margin-bottom: 8px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}
.batch-recognize-fail-list {
  margin-top: 12px;
  font-size: 13px;
  max-height: 200px;
  overflow-y: auto;
}
.batch-recognize-fail-list .fail-list-title {
  font-weight: 500;
  margin-bottom: 6px;
}
.batch-recognize-fail-list ul {
  margin: 0;
  padding-left: 20px;
  color: var(--el-text-color-secondary);
}
.batch-recognize-fail-list li {
  margin-bottom: 4px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.file-name-link {
  color: #409eff;
  cursor: pointer;
  text-decoration: none;
}

.file-name-cell {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.file-status-tag {
  max-width: 100%;
  white-space: normal;
  word-break: break-all;
}

.file-name-link:hover {
  color: #66b1ff;
  text-decoration: underline;
}

.op-rows {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.op-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

/* 让操作按钮宽度一致、更整齐（作用到 ElementPlus 按钮内部样式） */
.op-row :deep(.el-button) {
  min-width: 86px;
  justify-content: center;
}

.md-content-header {
  display: flex;
  gap: 10px;
  align-items: center;
  margin-bottom: 10px;
}

.md-content-container {
  max-height: 600px;
  overflow-y: auto;
}

.zip-result-header {
  display: flex;
  gap: 10px;
  align-items: center;
  margin-bottom: 10px;
  flex-wrap: wrap;
}

.zip-filename {
  word-break: break-word;
  flex: 1;
  min-width: 0;
}

.zip-file-item {
  display: flex;
  align-items: center;
  gap: 6px;
  word-break: break-word;
  margin-bottom: 8px;
}

.zip-file-item span {
  flex: 1;
  min-width: 0;
}

/* 转换结果弹框宽度自适应 */
.zip-result-dialog :deep(.el-dialog) {
  width: 90vw;
  max-width: 900px;
}

.zip-result-container {
  min-height: 200px;
}

.file-list {
  max-height: 300px;
  overflow-y: auto;
}

.no-files {
  color: #909399;
  font-style: italic;
}

/* 上传对话框：内容不突出弹框，文件名折行 */
.upload-dialog :deep(.el-dialog__body) {
  overflow-x: hidden;
  max-height: 70vh;
  overflow-y: auto;
}
.upload-dialog .upload-form {
  overflow: hidden;
}
.upload-dialog .upload-inner {
  width: 100%;
  max-width: 100%;
  box-sizing: border-box;
}
.upload-dialog :deep(.el-upload) {
  width: 100%;
  max-width: 100%;
  box-sizing: border-box;
}
.upload-dialog :deep(.el-upload-dragger) {
  width: 100%;
  max-width: 100%;
  box-sizing: border-box;
}
.upload-dialog :deep(.el-upload-dragger .el-upload__text),
.upload-dialog :deep(.upload-selected-name) {
  word-break: break-all;
  overflow-wrap: break-word;
  white-space: normal;
  text-align: center;
}
.upload-selected-name {
  font-weight: 600;
  color: var(--el-text-color-primary);
  padding: 0 12px;
}
.upload-replace-hint {
  margin-top: 8px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.upload-progress-wrap {
  margin-top: 12px;
  padding: 10px 0;
  width: 100%;
  min-width: 280px;
}
.upload-progress-wrap :deep(.el-progress) {
  width: 100%;
}
.upload-progress-wrap :deep(.el-progress-bar__outer) {
  width: 100%;
}
.upload-progress-label {
  font-size: 13px;
  color: var(--el-text-color-regular);
  margin-bottom: 2px;
}

/* 内容比较对话框 */
.diff-dialog :deep(.el-dialog__body) {
  padding: 0;
}
.diff-container {
  display: flex;
  flex-direction: column;
  height: 75vh;
  border: 1px solid #d0d7de;
  border-radius: 6px;
  overflow: hidden;
  background-color: #fff;
}
.diff-header {
  display: flex;
  background-color: #f6f8fa;
  border-bottom: 1px solid #d0d7de;
  padding: 8px 16px;
  font-size: 12px;
  font-weight: 600;
}
.diff-label { flex: 1; color: #656d76; }
.diff-label.original { border-right: 1px solid #d0d7de; padding-right: 16px; }
.diff-label.current { padding-left: 16px; }
.diff-content-wrapper { flex: 1; overflow: auto; background-color: #fff; }
.diff-content { width: 100%; }
.diff-table {
  width: 100%;
  border-collapse: collapse;
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Consolas, 'Liberation Mono', monospace;
  font-size: 12px;
  line-height: 1.45;
}
.diff-table tbody tr { border-top: 1px solid transparent; }
.diff-table tbody tr:hover { background-color: #f6f8fa; }
.diff-line-number {
  width: 1%;
  min-width: 50px;
  padding: 0 10px;
  text-align: right;
  color: #656d76;
  background-color: #f6f8fa;
  border-right: 1px solid #d0d7de;
  user-select: none;
  font-variant-numeric: tabular-nums;
}
.diff-line-number.empty { background-color: #f6f8fa; }
.diff-line-number.original { border-right: 1px solid #d0d7de; }
.diff-line-number.current { border-right: 1px solid #d0d7de; }

/* LLM Markdown 规范化对话框 */
.normalize-hint {
  margin: 0 0 12px;
  font-size: 13px;
  color: #606266;
  line-height: 1.5;
}
.normalize-panel-title {
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 8px;
  color: #303133;
}
.normalize-textarea :deep(textarea) {
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Consolas, 'Liberation Mono', monospace;
  font-size: 13px;
  line-height: 1.5;
}
.diff-line-content {
  padding: 0 10px;
  white-space: pre-wrap;
  word-wrap: break-word;
  word-break: break-word;
  overflow-x: hidden;
  color: #24292f;
}
.diff-line-content.original { border-right: 1px solid #d0d7de; }
.diff-line-removed { background-color: #fff1f2; }
.diff-line-removed .diff-line-number.original { background-color: #ffebe9; color: #82071e; }
.diff-line-removed .diff-line-content.original { background-color: #ffebe9; color: #82071e; position: relative; padding-left: 20px; }
.diff-line-removed .diff-line-content.original::before { content: '-'; position: absolute; left: 10px; color: #cf222e; }
.diff-line-added { background-color: #f0fff4; }
.diff-line-added .diff-line-number.current { background-color: #ccfdf4; color: #116329; }
.diff-line-added .diff-line-content.current { background-color: #ccfdf4; color: #116329; position: relative; padding-left: 20px; }
.diff-line-added .diff-line-content.current::before { content: '+'; position: absolute; left: 10px; color: #1a7f37; }
.diff-line-normal { background-color: #fff; }
.diff-line-normal .diff-line-content { color: #24292f; }
.empty-line { display: inline-block; width: 100%; }
.diff-loading { display: flex; align-items: center; justify-content: center; height: 200px; gap: 10px; color: #909399; }
.diff-no-changes { display: flex; align-items: center; justify-content: center; height: 200px; gap: 10px; color: #67c23a; font-size: 14px; }

.qdrant-chunks-meta {
  margin-bottom: 12px;
}
.qdrant-chunks-tabs {
  margin-top: 8px;
}
.qa-pairs-header {
  margin-bottom: 10px;
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.chunk-detail-pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 13px;
  line-height: 1.5;
  max-height: calc(100vh - 120px);
  overflow: auto;
}

</style>
