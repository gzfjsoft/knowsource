<template>
  <div class="setup-init">
    <div class="header">
      <h1>系统配置向导</h1>
      <p class="sub">
        当 MySQL 或 Redis 不可用时会进入本页。请按步骤填写
        <code>knowsource.yaml</code>
        与可选的
        <code>fca-emails.yaml</code>
        ，保存后
        <strong>必须重启后端进程</strong>
        才能正常使用系统。
      </p>
    </div>

    <el-alert
      v-if="statusText"
      :title="statusText"
      :type="statusAlertType"
      show-icon
      class="mb"
      closable
      @close="statusText = ''"
    />

    <el-card class="mb">
      <template #header>
        <span>连接状态（自磁盘上的当前配置检测）</span>
        <el-button size="small" class="ml" @click="refreshStatus">刷新</el-button>
      </template>
      <el-descriptions :column="2" border size="small">
        <el-descriptions-item label="MySQL">
          <el-tag :type="status.mysqlOk ? 'success' : 'danger'">{{ status.mysqlOk ? '正常' : '失败' }}</el-tag>
          {{ status.mysqlMsg || '—' }}
        </el-descriptions-item>
        <el-descriptions-item label="Redis">
          <el-tag :type="status.redisOk ? 'success' : 'danger'">{{ status.redisOk ? '正常' : '失败' }}</el-tag>
          {{ status.redisMsg || '—' }}
        </el-descriptions-item>
        <el-descriptions-item label="初始化模式">
          {{ status.initMode ? '是（仅开放配置接口，需重启）' : '否' }}
        </el-descriptions-item>
        <el-descriptions-item label="应用就绪">
          {{ status.appReady ? '是，可前往登录' : '否' }}
        </el-descriptions-item>
        <el-descriptions-item v-if="mainPath" label="主配置文件" :span="2">{{ mainPath }}</el-descriptions-item>
      </el-descriptions>
    </el-card>

    <el-steps :active="step" finish-status="success" align-center class="mb">
      <el-step title="说明" />
      <el-step title="数据库 / Redis" />
      <el-step title="HTTP / 日志" />
      <el-step title="认证与 Boot" />
      <el-step title="路径与文档" />
      <el-step title="AI / RAG" />
      <el-step title="邮件与其它" />
      <el-step title="邮件模板 YAML" />
      <el-step title="保存" />
    </el-steps>

    <el-card v-loading="loading">
      <!-- 0 说明 -->
      <div v-show="step === 0" class="step-panel">
        <el-alert type="info" show-icon :closable="false" title="分页说明" />
        <ul class="desc-list">
          <li><strong>数据库 / Redis：</strong>DSN 与 Redis 地址密码；错误的配置会导致无法登录与缓存。</li>
          <li><strong>HTTP / 日志：</strong>服务监听 Host、Port 及日志目录（与 go-zero RestConf 一致）。</li>
          <li><strong>认证与 Boot：</strong>JWT Secret、过期时间、Salt、Boot/Fca 引导地址等。</li>
          <li><strong>路径与文档：</strong>Knowdata 知识路径、本地上传目录、FilesRoot 等。</li>
          <li><strong>AI / RAG：</strong>Qdrant、Embeddings、Completion、Reranker 等 URL。</li>
          <li><strong>邮件与其它：</strong>SMTP、短信、对象存储等扩展项。</li>
          <li><strong>fca-emails.yaml：</strong>可选，与主配置同目录下的邮件文案。</li>
        </ul>
      </div>

      <!-- 1 MySQL Redis -->
      <div v-show="step === 1" class="step-panel">
        <el-form label-width="140px">
          <el-form-item label="Mysql.DataSource">
            <el-input v-model="mainConfig.MySQL.DataSource" type="textarea" :rows="4" placeholder="user:pass@tcp(host:3306)/db?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai" />
            <div class="hint">完整 DSN，需含 charset、parseTime、loc。</div>
          </el-form-item>
          <el-divider>Redis（第一项为主连接，验证码与 JWT 等使用）</el-divider>
          <el-form-item label="Host">
            <el-input v-model="mainConfig.CacheRedis[0].Host" placeholder="127.0.0.1:6379" />
          </el-form-item>
          <el-form-item label="Type">
            <el-input v-model="mainConfig.CacheRedis[0].Type" placeholder="node" />
          </el-form-item>
          <el-form-item label="Pass">
            <el-input v-model="mainConfig.CacheRedis[0].Pass" type="password" show-password placeholder="无密码可留空" />
          </el-form-item>
          <el-divider>第二段 Redis（可选，与第一段相同时可保持不变）</el-divider>
          <el-form-item label="Host 2">
            <el-input v-model="mainConfig.CacheRedis[1].Host" placeholder="可选" />
          </el-form-item>
          <el-form-item label="Pass 2">
            <el-input v-model="mainConfig.CacheRedis[1].Pass" type="password" show-password />
          </el-form-item>
        </el-form>
      </div>

      <!-- 2 HTTP Log -->
      <div v-show="step === 2" class="step-panel">
        <el-form label-width="140px">
          <el-form-item label="Name">
            <el-input v-model="mainConfig.Name" />
          </el-form-item>
          <el-form-item label="Host">
            <el-input v-model="mainConfig.Host" placeholder="0.0.0.0" />
          </el-form-item>
          <el-form-item label="Port">
            <el-input-number v-model="mainConfig.Port" :min="1" :max="65535" controls-position="right" />
          </el-form-item>
          <el-form-item label="MaxBytes">
            <el-input-number v-model="mainConfig.MaxBytes" :min="1024" controls-position="right" />
          </el-form-item>
          <el-divider>日志 Log（go-zero logx）</el-divider>
          <el-form-item label="Log.Mode">
            <el-input v-model="mainConfig.Log.Mode" placeholder="file" />
          </el-form-item>
          <el-form-item label="Log.Path">
            <el-input v-model="mainConfig.Log.Path" placeholder="./logs" />
          </el-form-item>
          <el-form-item label="Log.Level">
            <el-input v-model="mainConfig.Log.Level" placeholder="info" />
          </el-form-item>
        </el-form>
      </div>

      <!-- 3 Auth -->
      <div v-show="step === 3" class="step-panel">
        <el-form label-width="160px">
          <el-form-item label="Auth.AccessSecret">
            <el-input v-model="mainConfig.Auth.AccessSecret" type="password" show-password />
          </el-form-item>
          <el-form-item label="Auth.AccessExpire(秒)">
            <el-input-number v-model="mainConfig.Auth.AccessExpire" :min="60" controls-position="right" />
          </el-form-item>
          <el-form-item label="Salt">
            <el-input v-model="mainConfig.Salt" />
          </el-form-item>
          <el-form-item label="Boot.Url">
            <el-input v-model="mainConfig.Boot.Url" />
          </el-form-item>
          <el-form-item label="Fca.Url">
            <el-input v-model="mainConfig.Fca.Url" />
          </el-form-item>
        </el-form>
      </div>

      <!-- 4 Paths -->
      <div v-show="step === 4" class="step-panel">
        <el-form label-width="200px">
          <el-form-item label="FilesRoot">
            <el-input v-model="mainConfig.FilesRoot" />
          </el-form-item>
          <el-form-item label="UploadPath">
            <el-input v-model="mainConfig.UploadPath" />
          </el-form-item>
          <el-form-item label="BucketPath">
            <el-input v-model="mainConfig.BucketPath" />
          </el-form-item>
          <el-form-item label="Knowdata.DocumentPath">
            <el-input v-model="mainConfig.Knowdata.DocumentPath" type="textarea" :rows="2" />
          </el-form-item>
          <el-form-item label="Knowdata.KnowledgeFilePath">
            <el-input v-model="mainConfig.Knowdata.KnowledgeFilePath" />
          </el-form-item>
          <el-form-item label="Knowdata.MarkdownPath">
            <el-input v-model="mainConfig.Knowdata.MarkdownPath" />
          </el-form-item>
          <el-form-item label="Knowdata.TempFilePath">
            <el-input v-model="mainConfig.Knowdata.TempFilePath" />
          </el-form-item>
        </el-form>
      </div>

      <!-- 5 AI -->
      <div v-show="step === 5" class="step-panel">
        <el-form label-width="180px">
          <el-form-item label="Qdrant.Host">
            <el-input v-model="mainConfig.Qdrant.Host" />
          </el-form-item>
          <el-form-item label="Qdrant.Port">
            <el-input-number v-model="mainConfig.Qdrant.Port" :min="0" controls-position="right" />
          </el-form-item>
          <el-form-item label="Rag.EmbeddingsUrl">
            <el-input v-model="mainConfig.Rag.EmbeddingsUrl" />
          </el-form-item>
          <el-form-item label="Rag.RerankerUrl">
            <el-input v-model="mainConfig.Rag.RerankerUrl" />
          </el-form-item>
          <el-form-item label="Llm.CompletionUrl">
            <el-input v-model="mainConfig.Llm.CompletionUrl" />
          </el-form-item>
          <el-form-item label="RAGURL">
            <el-input v-model="mainConfig.RAGURL" />
          </el-form-item>
          <el-form-item label="MinerU.URL">
            <el-input v-model="mainConfig.MinerU.URL" />
          </el-form-item>
        </el-form>
      </div>

      <!-- 6 Mail misc -->
      <div v-show="step === 6" class="step-panel">
        <el-form label-width="160px">
          <el-form-item label="Mail.MailHost">
            <el-input v-model="mainConfig.Mail.MailHost" />
          </el-form-item>
          <el-form-item label="Mail.MailPort">
            <el-input-number v-model="mainConfig.Mail.MailPort" :min="0" controls-position="right" />
          </el-form-item>
          <el-form-item label="Mail.MailAccount">
            <el-input v-model="mainConfig.Mail.MailAccount" />
          </el-form-item>
          <el-form-item label="Mail.MailPass">
            <el-input v-model="mainConfig.Mail.MailPass" type="password" show-password />
          </el-form-item>
          <el-form-item label="AdminJWT">
            <el-input v-model="mainConfig.AdminJWT" />
          </el-form-item>
          <el-form-item label="SensitiveWordsFile">
            <el-input v-model="mainConfig.SensitiveWordsFile" />
          </el-form-item>
          <el-form-item label="DefaultUserRoleId">
            <el-input-number v-model="mainConfig.DefaultUserRoleId" controls-position="right" />
          </el-form-item>
        </el-form>
      </div>

      <!-- 7 email yaml -->
      <div v-show="step === 7" class="step-panel">
        <el-checkbox v-model="saveEmailDraft">同时保存到 {{ emailPathShort }}</el-checkbox>
        <el-form label-width="200px" class="mt">
          <el-form-item label="VerifyMailTitle">
            <el-input v-model="emailConfig.VerifyMailTitle" />
          </el-form-item>
          <el-form-item label="LoginMailTitle">
            <el-input v-model="emailConfig.LoginMailTitle" />
          </el-form-item>
          <el-form-item label="VerifyMailContent">
            <el-input v-model="emailConfig.VerifyMailContent" type="textarea" :rows="3" />
          </el-form-item>
          <el-form-item label="LoginMailContent">
            <el-input v-model="emailConfig.LoginMailContent" type="textarea" :rows="3" />
          </el-form-item>
          <el-form-item label="ForgetPasswordMailTitle">
            <el-input v-model="emailConfig.ForgetPasswordMailTitle" />
          </el-form-item>
          <el-form-item label="ForgetPasswordMailContent">
            <el-input v-model="emailConfig.ForgetPasswordMailContent" type="textarea" :rows="3" />
          </el-form-item>
        </el-form>
      </div>

      <!-- 8 save -->
      <div v-show="step === 8" class="step-panel">
        <el-alert type="warning" show-icon :closable="false" title="保存将直接覆盖磁盘上的 YAML 文件；请先确认 MySQL/Redis 已正确。" />
        <div class="mt">
          <el-button type="primary" :loading="saving" @click="onSave">写入配置</el-button>
          <el-button @click="refreshStatus">仅检测连接</el-button>
        </div>
      </div>

      <div class="nav-btns">
        <el-button :disabled="step <= 0" @click="step--">上一步</el-button>
        <el-button :disabled="step >= 8" type="primary" @click="step++">下一步</el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getBootstrapConfig, getBootstrapStatus, saveBootstrapConfig } from '@/api/bootstrap'
import { clearBootstrapGateCache } from '@/utils/bootstrapGate'

const step = ref(0)
const loading = ref(false)
const saving = ref(false)
const statusText = ref('')
const statusAlertType = ref('info')
const mainPath = ref('')
const emailPath = ref('')
const saveEmailDraft = ref(false)

const status = reactive({
  mysqlOk: false,
  redisOk: false,
  coreOk: false,
  initMode: false,
  appReady: false,
  mysqlMsg: '',
  redisMsg: ''
})

function emptyMain() {
  return {
    Name: '',
    Host: '0.0.0.0',
    Port: 8070,
    MaxBytes: 10485760,
    Log: { Mode: 'file', Path: './logs', Level: 'info' },
    MySQL: { DataSource: '' },
    CacheRedis: [
      { Host: '', Type: 'node', Pass: '' },
      { Host: '', Type: 'node', Pass: '' }
    ],
    Boot: { Url: '' },
    Fca: { Url: '' },
    Auth: { AccessSecret: '', AccessExpire: 259200 },
    Salt: '',
    FilesRoot: '',
    UploadPath: './upload/',
    BucketPath: './bucket/',
    Knowdata: {
      Host: '',
      KnowledgeFilePath: '',
      MarkdownPath: '',
      TempFilePath: '',
      DocumentPath: ''
    },
    Qdrant: { Host: '', Port: 6333 },
    Rag: { EmbeddingsUrl: '', RerankerUrl: '', EmbeddingsType: '', RerankerType: '' },
    Llm: { CompletionUrl: '', CompletionType: '' },
    RAGURL: '',
    MinerU: { URL: '' },
    Mail: { MailHost: '', MailPort: 25, MailAccount: '', MailPass: '' },
    AdminJWT: '',
    SensitiveWordsFile: '',
    DefaultUserRoleId: 0
  }
}

const mainConfig = reactive(emptyMain())

const emailConfig = reactive({
  VerifyMailTitle: '',
  VerifyMailContent: '',
  LoginMailTitle: '',
  LoginMailContent: '',
  ForgetPasswordMailTitle: '',
  ForgetPasswordMailContent: '',
  StopInstanceMailTitle: '',
  StopInstanceMailContent: ''
})

const emailPathShort = computed(() => emailPath.value || 'fca-emails.yaml')

function ensureRedisSlots() {
  const slot = { Host: '', Type: 'node', Pass: '' }
  if (!Array.isArray(mainConfig.CacheRedis)) mainConfig.CacheRedis = []
  while (mainConfig.CacheRedis.length < 2) {
    mainConfig.CacheRedis.push({ ...slot })
  }
}

async function refreshStatus() {
  try {
    const res = await getBootstrapStatus()
    if (res.code !== 200) {
      statusText.value = res.message || '状态获取失败'
      statusAlertType.value = 'error'
      return
    }
    const d = res.data || {}
    Object.assign(status, d)
    if (d.mainPath) mainPath.value = d.mainPath
  } catch (e) {
    statusText.value = e.message || '网络错误'
    statusAlertType.value = 'error'
  }
}

async function loadConfig() {
  loading.value = true
  try {
    const res = await getBootstrapConfig()
    if (res.code !== 200) {
      ElMessage.error(res.message || '加载配置失败')
      return
    }
    const d = res.data || {}
    if (d.mainPath) mainPath.value = d.mainPath
    if (d.emailPath) emailPath.value = d.emailPath
    if (d.mainConfig) {
      const base = emptyMain()
      const merged = deepMerge(base, d.mainConfig)
      Object.assign(mainConfig, merged)
      ensureRedisSlots()
      if (!mainConfig.Knowdata) mainConfig.Knowdata = { ...base.Knowdata }
      if (!mainConfig.MySQL) mainConfig.MySQL = { DataSource: '' }
      if (!mainConfig.Log) mainConfig.Log = { ...base.Log }
      if (!mainConfig.Boot) mainConfig.Boot = { Url: '' }
      if (!mainConfig.Fca) mainConfig.Fca = { Url: '' }
      if (!mainConfig.Auth) mainConfig.Auth = { AccessSecret: '', AccessExpire: 259200 }
      if (!mainConfig.Qdrant) mainConfig.Qdrant = { Host: '', Port: 6333 }
      if (!mainConfig.Rag) mainConfig.Rag = { EmbeddingsUrl: '', RerankerUrl: '' }
      if (!mainConfig.Llm) mainConfig.Llm = { CompletionUrl: '' }
      if (!mainConfig.MinerU) mainConfig.MinerU = { URL: '' }
      if (!mainConfig.Mail) mainConfig.Mail = { MailHost: '', MailPort: 25, MailAccount: '', MailPass: '' }
    }
    if (d.emailConfig) {
      Object.assign(emailConfig, d.emailConfig)
    }
  } catch (e) {
    ElMessage.error(e.message || '加载失败')
  } finally {
    loading.value = false
  }
}

function deepMerge(target, src) {
  if (!src || typeof src !== 'object') return target
  const out = { ...target }
  for (const k of Object.keys(src)) {
    const sv = src[k]
    const tv = out[k]
    if (sv && typeof sv === 'object' && !Array.isArray(sv) && tv && typeof tv === 'object' && !Array.isArray(tv)) {
      out[k] = deepMerge(tv, sv)
    } else {
      out[k] = sv
    }
  }
  return out
}

async function onSave() {
  saving.value = true
  try {
    const body = {
      mainConfig: structuredClone(mainConfig),
      saveEmail: saveEmailDraft.value
    }
    if (saveEmailDraft.value) {
      body.emailConfig = { ...emailConfig }
    }
    const res = await saveBootstrapConfig(body)
    if (res.code !== 200) {
      ElMessage.error(res.message || res.info || '保存失败')
      return
    }
    ElMessage.success(res.message || '已保存')
    statusText.value = res.message || ''
    statusAlertType.value = 'success'
    clearBootstrapGateCache()
    await refreshStatus()
  } catch (e) {
    ElMessage.error(e.message || '保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  ensureRedisSlots()
  await loadConfig()
  ensureRedisSlots()
  await refreshStatus()
})
</script>

<style scoped>
.setup-init {
  max-width: 960px;
  margin: 24px auto;
  padding: 0 16px;
}
.header h1 {
  margin: 0 0 8px;
  font-size: 22px;
}
.sub {
  color: #606266;
  line-height: 1.6;
  margin: 0 0 16px;
}
.mb {
  margin-bottom: 16px;
}
.ml {
  margin-left: 12px;
}
.mt {
  margin-top: 12px;
}
.hint {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}
.desc-list {
  margin: 12px 0 0 20px;
  line-height: 1.7;
  color: #606266;
}
.step-panel {
  min-height:200px;
}
.nav-btns {
  margin-top: 24px;
  display: flex;
  justify-content: space-between;
}
</style>
