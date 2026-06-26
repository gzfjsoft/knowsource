<template>
  <div class="llm-setting-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>LLM 配置</span>
          <div>
            <el-button type="primary" @click="handleLoad" :loading="loading">
              <el-icon><Refresh /></el-icon>
              刷新配置
            </el-button>
          </div>
        </div>
      </template>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="140px"
        style="max-width: 800px"
      >
        <el-form-item label="最大大小 (MaxTokens)" prop="maxSize">
          <el-input-number
            v-model="form.maxSize"
            :min="1"
            :max="32768"
            :step="512"
            placeholder="请输入最大大小"
            style="width: 100%"
          />
          <div class="form-item-tip">控制模型处理的最大 token 数量，建议范围：1024-16384</div>
        </el-form-item>

        <el-form-item label="对话模型 (Model)" prop="model">
          <el-select
            v-model="form.model"
            placeholder="请选择或输入模型"
            filterable
            allow-create
            default-first-option
            style="width: 100%"
          >
            <el-option
              v-for="id in chatModelIds"
              :key="id"
              :label="id"
              :value="id"
            />
          </el-select>
          <div class="form-item-tip">从对话服务拉取的模型列表（排除名称含 embedding 的模型），选择后保存到配置文件</div>
        </el-form-item>

        <el-divider content-position="left" style="margin-top: 20px;">服务端点覆盖（租户级）</el-divider>

        <el-form-item label="快捷预设">
          <el-button type="success" plain @click="handleApplyDeepSeek">
            一键配置 DeepSeek
          </el-button>
          <div class="form-item-tip">
            填入 Chat URL <code>https://api.deepseek.com</code>、模型 <code>deepseek-v4-pro</code>、Chat Type <code>vllm</code>（OpenAI 兼容，与 vLLM 相同调用方式）。后端请求 <code>/v1/chat/completions</code>。请自行填写 Chat API Key 后保存。
          </div>
        </el-form-item>

        <el-form-item label="Chat URL" prop="completionUrl">
          <el-input v-model="form.completionUrl" placeholder="OpenAI 兼容 base URL，如 http://127.0.0.1:8000" />
          <div class="form-item-tip">仅当前租户生效。配置后优先覆盖系统 Llm.CompletionUrl</div>
        </el-form-item>
        <el-form-item label="Chat API Key" prop="completionApiKey">
          <el-input v-model="form.completionApiKey" show-password placeholder="可选，Bearer Token" />
        </el-form-item>
        <el-form-item label="Chat Type" prop="completionType">
          <el-select v-model="form.completionType" placeholder="自动/手动指定" clearable style="width: 100%">
            <el-option label="vllm" value="vllm" />
            <el-option label="ollama" value="ollama" />
            <el-option label="llamacpp" value="llamacpp" />
          </el-select>
        </el-form-item>

        <el-form-item label="Embedding URL" prop="embeddingsUrl">
          <el-input v-model="form.embeddingsUrl" placeholder="OpenAI 兼容 embeddings base URL" />
          <div class="form-item-tip">仅当前租户生效。配置后优先覆盖系统 Rag.EmbeddingsUrl</div>
        </el-form-item>
        <el-form-item label="Embedding API Key" prop="embeddingsApiKey">
          <el-input v-model="form.embeddingsApiKey" show-password placeholder="可选，Bearer Token" />
        </el-form-item>
        <el-form-item label="Embedding Type" prop="embeddingsType">
          <el-select v-model="form.embeddingsType" placeholder="自动/手动指定" clearable style="width: 100%">
            <el-option label="vllm" value="vllm" />
            <el-option label="ollama" value="ollama" />
          </el-select>
        </el-form-item>

        <el-form-item label="Rerank URL" prop="rerankerUrl">
          <el-input v-model="form.rerankerUrl" placeholder="OpenAI 兼容 rerank base URL" />
          <div class="form-item-tip">仅当前租户生效。配置后优先覆盖系统 Rag.RerankerUrl</div>
        </el-form-item>
        <el-form-item label="Rerank API Key" prop="rerankerApiKey">
          <el-input v-model="form.rerankerApiKey" show-password placeholder="可选，Bearer Token" />
        </el-form-item>
        <el-form-item label="Rerank Type" prop="rerankerType">
          <el-select v-model="form.rerankerType" placeholder="自动/手动指定" clearable style="width: 100%">
            <el-option label="vllm" value="vllm" />
            <el-option label="llama.cpp" value="llama.cpp" />
          </el-select>
        </el-form-item>
        <el-form-item label="Rerank 模型" prop="rerankerModel">
          <el-input v-model="form.rerankerModel" placeholder="例如 Qwen3-Reranker-0.6B" />
        </el-form-item>

        <el-form-item label="温度参数 (Temperature)" prop="temperature">
          <el-slider
            v-model="form.temperature"
            :min="0"
            :max="2"
            :step="0.1"
            show-input
            :show-input-controls="false"
            style="width: 100%"
          />
          <div class="form-item-tip">控制输出的随机性，范围：0.0-2.0。值越大输出越随机，值越小输出越确定</div>
        </el-form-item>

        <el-form-item label="TopK 参数" prop="topK">
          <el-input-number
            v-model="form.topK"
            :min="1"
            :max="200"
            placeholder="请输入 TopK 值"
            style="width: 100%"
          />
          <div class="form-item-tip">限制候选词数量，范围：1-200。值越小输出越集中</div>
        </el-form-item>

        <el-form-item label="TopP 参数" prop="topP">
          <el-slider
            v-model="form.topP"
            :min="0"
            :max="1"
            :step="0.01"
            show-input
            :show-input-controls="false"
            style="width: 100%"
          />
          <div class="form-item-tip">核采样参数，范围：0.0-1.0。控制输出的多样性，值越大输出越多样</div>
        </el-form-item>

        <el-divider content-position="left" style="margin-top: 20px;">RAG 检索与重排</el-divider>

        <el-form-item label="Embedding 模型" prop="embeddingModel">
          <el-select
            v-model="form.embeddingModel"
            placeholder="请选择或输入 Embedding 模型"
            filterable
            allow-create
            default-first-option
            clearable
            style="width: 100%"
          >
            <el-option
              v-for="id in embeddingModelIds"
              :key="id"
              :label="id"
              :value="id"
            />
          </el-select>
          <div class="form-item-tip">从向量化服务（Rag.EmbeddingsUrl）拉取，仅显示名称含 embedding 的模型，用于 RAG 向量化</div>
        </el-form-item>


        <el-form-item label="Embedding TopK" prop="ragEmbeddingTopK">
          <el-input-number
            v-model="form.ragEmbeddingTopK"
            :min="1"
            :max="200"
            placeholder="向量检索返回条数"
            style="width: 100%"
          />
          <div class="form-item-tip">RAG 向量检索时从 Qdrant 取出的最相似条数，建议 5–50</div>
        </el-form-item>
        <el-form-item label="相似度底线" prop="ragSimilarityThreshold">
          <el-input-number
            v-model="form.ragSimilarityThreshold"
            :min="0"
            :max="1"
            :step="0.05"
            :precision="2"
            placeholder="0 表示不过滤"
            style="width: 100%"
          />
          <div class="form-item-tip">向量相似度低于此值的 chunk 将被过滤，0 表示不设底线</div>
        </el-form-item>
        <el-form-item label="概要检索相似度底线" prop="ragSummarySimilarityThreshold">
          <el-input-number
            v-model="form.ragSummarySimilarityThreshold"
            :min="0"
            :max="1"
            :step="0.05"
            :precision="2"
            placeholder="默认沿用“相似度底线”"
            style="width: 100%"
          />
          <div class="form-item-tip">概要向量检索（summary collection）专用相似度阈值；0 表示不过滤；未配置时会沿用“相似度底线”</div>
        </el-form-item>
        <el-form-item label="Rerank TopK" prop="ragRerankTopK">
          <el-input-number
            v-model="form.ragRerankTopK"
            :min="1"
            :max="50"
            placeholder="重排后保留条数"
            style="width: 100%"
          />
          <div class="form-item-tip">重排后保留的条数，用于 AI 上下文的参考资料数量</div>
        </el-form-item>
        <el-form-item label="重排分值底线" prop="ragRerankScoreThreshold">
          <el-input-number
            v-model="form.ragRerankScoreThreshold"
            :min="0"
            :max="1"
            :step="0.05"
            :precision="2"
            placeholder="0 表示不过滤"
            style="width: 100%"
          />
          <div class="form-item-tip">重排得分低于此值的片段将被过滤，0 表示不设底线</div>
        </el-form-item>
        <el-form-item label="概要检索重排底线" prop="ragSummaryRerankScoreThreshold">
          <el-input-number
            v-model="form.ragSummaryRerankScoreThreshold"
            :min="0"
            :max="1"
            :step="0.05"
            :precision="2"
            placeholder="默认沿用“重排分值底线”"
            style="width: 100%"
          />
          <div class="form-item-tip">概要检索（summary collection）专用重排阈值；0 表示不过滤；未配置时会沿用“重排分值底线”</div>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleSubmit" :loading="submitLoading">
            <el-icon><Check /></el-icon>
            保存配置
          </el-button>
          <el-button @click="handleReset" :loading="resetLoading">
            <el-icon><RefreshLeft /></el-icon>
            恢复系统默认
          </el-button>
        </el-form-item>

        <el-form-item label="连通性测试">
          <div class="test-actions">
            <el-button @click="runServiceCheck" :loading="testLoading.service">检测服务可达性</el-button>
            <el-button @click="runChatSmokeTest" :loading="testLoading.chat">测试 Chat 调用</el-button>
            <el-button @click="runEmbeddingRerankTest" :loading="testLoading.embedRerank">测试 Embedding + Rerank</el-button>
          </div>
          <div class="test-result" v-if="testResult.service">
            服务检查：{{ testResult.service }}
          </div>
          <div class="test-result" v-if="testResult.chat">
            Chat 调用：{{ testResult.chat }}
          </div>
          <div class="test-result" v-if="testResult.embedRerank">
            Embedding+Rerank：{{ testResult.embedRerank }}
          </div>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Check, RefreshLeft } from '@element-plus/icons-vue'
import { loadLLMSetting, saveLLMSetting, loadLLMSettingDefaults, getLLMChatModels, getLLMEmbeddingModels, callLLMOneShot, testLLMServices } from '@/api/knowdata'
import { sysCheck } from '@/api/knowsource'

const formRef = ref(null)
const loading = ref(false)
const submitLoading = ref(false)
const resetLoading = ref(false)
const testLoading = reactive({
  service: false,
  chat: false,
  embedRerank: false
})
const testResult = reactive({
  service: '',
  chat: '',
  embedRerank: ''
})
const chatModelIds = ref([])
const embeddingModelIds = ref([])

const DEEPSEEK_PRESET = {
  completionUrl: 'https://api.deepseek.com',
  completionType: 'vllm',
  model: 'deepseek-v4-pro'
}

const form = reactive({
  maxSize: 16384,
  model: 'qwen3:latest',
  completionUrl: '',
  completionApiKey: '',
  completionType: '',
  embeddingModel: '',
  embeddingsUrl: '',
  embeddingsApiKey: '',
  embeddingsType: '',
  rerankerUrl: '',
  rerankerApiKey: '',
  rerankerType: '',
  rerankerModel: '',
  temperature: 0.7,
  topK: 40,
  topP: 0.9,
  ragEmbeddingTopK: 10,
  ragSimilarityThreshold: 0,
  ragSummarySimilarityThreshold: 0,
  ragRerankTopK: 5,
  ragRerankScoreThreshold: 0,
  ragSummaryRerankScoreThreshold: 0
})

const applyFormFromSetting = (data) => {
  if (!data) return
  Object.assign(form, {
    maxSize: data.maxSize ?? 16384,
    model: data.model || '',
    completionUrl: data.completionUrl ?? '',
    completionApiKey: data.completionApiKey ?? '',
    completionType: data.completionType ?? '',
    embeddingModel: data.embeddingModel ?? '',
    embeddingsUrl: data.embeddingsUrl ?? '',
    embeddingsApiKey: data.embeddingsApiKey ?? '',
    embeddingsType: data.embeddingsType ?? '',
    rerankerUrl: data.rerankerUrl ?? '',
    rerankerApiKey: data.rerankerApiKey ?? '',
    rerankerType: data.rerankerType ?? '',
    rerankerModel: data.rerankerModel ?? '',
    temperature: data.temperature ?? 0.7,
    topK: data.topK ?? 40,
    topP: data.topP ?? 0.9,
    ragEmbeddingTopK: data.ragEmbeddingTopK ?? 10,
    ragSimilarityThreshold: data.ragSimilarityThreshold ?? 0,
    ragSummarySimilarityThreshold: data.ragSummarySimilarityThreshold ?? 0,
    ragRerankTopK: data.ragRerankTopK ?? 5,
    ragRerankScoreThreshold: data.ragRerankScoreThreshold ?? 0,
    ragSummaryRerankScoreThreshold: data.ragSummaryRerankScoreThreshold ?? 0
  })
  if (data.model && !chatModelIds.value.includes(data.model)) {
    chatModelIds.value = [data.model, ...chatModelIds.value]
  }
  if (data.embeddingModel && !embeddingModelIds.value.includes(data.embeddingModel)) {
    embeddingModelIds.value = [data.embeddingModel, ...embeddingModelIds.value]
  }
}

const rules = {
  maxSize: [
    { required: true, message: '请输入最大大小', trigger: 'blur' },
    { type: 'number', min: 1, max: 32768, message: '最大大小范围：1-32768', trigger: 'blur' }
  ],
  model: [
    { required: true, message: '请输入模型名称', trigger: 'blur' },
    { validator: (rule, value, cb) => {
      if (value && value.toLowerCase().includes('embedding')) {
        cb(new Error('对话模型名称不能包含 embedding'))
      } else {
        cb()
      }
    }, trigger: 'blur' }
  ],
  embeddingModel: [
    { validator: (rule, value, cb) => {
      if (value && !value.toLowerCase().includes('embedding')) {
        cb(new Error('Embedding 模型名称必须包含 embedding'))
      } else {
        cb()
      }
    }, trigger: 'blur' }
  ],
  temperature: [
    { required: true, message: '请输入温度参数', trigger: 'blur' },
    { type: 'number', min: 0, max: 2, message: '温度参数范围：0.0-2.0', trigger: 'blur' }
  ],
  topK: [
    { required: true, message: '请输入 TopK 值', trigger: 'blur' },
    { type: 'number', min: 1, max: 200, message: 'TopK 范围：1-200', trigger: 'blur' }
  ],
  topP: [
    { required: true, message: '请输入 TopP 值', trigger: 'blur' },
    { type: 'number', min: 0, max: 1, message: 'TopP 范围：0.0-1.0', trigger: 'blur' }
  ],
  ragEmbeddingTopK: [
    { type: 'number', min: 1, max: 200, message: '范围：1-200', trigger: 'blur' }
  ],
  ragSimilarityThreshold: [
    { type: 'number', min: 0, max: 1, message: '范围：0-1', trigger: 'blur' }
  ],
  ragSummarySimilarityThreshold: [
    { type: 'number', min: 0, max: 1, message: '范围：0-1', trigger: 'blur' }
  ],
  ragRerankTopK: [
    { type: 'number', min: 1, max: 50, message: '范围：1-50', trigger: 'blur' }
  ],
  ragRerankScoreThreshold: [
    { type: 'number', min: 0, max: 1, message: '范围：0-1', trigger: 'blur' }
  ],
  ragSummaryRerankScoreThreshold: [
    { type: 'number', min: 0, max: 1, message: '范围：0-1', trigger: 'blur' }
  ]
}

// 加载配置
const handleLoad = async () => {
  loading.value = true
  try {
    const [res, modelsRes, embeddingRes] = await Promise.all([
      loadLLMSetting(),
      getLLMChatModels(),
      getLLMEmbeddingModels()
    ])
    if (modelsRes.code === 200 && modelsRes.data && Array.isArray(modelsRes.data.modelIds)) {
      chatModelIds.value = modelsRes.data.modelIds
    } else {
      chatModelIds.value = []
    }
    if (embeddingRes.code === 200 && embeddingRes.data && Array.isArray(embeddingRes.data.modelIds)) {
      embeddingModelIds.value = embeddingRes.data.modelIds
    } else {
      embeddingModelIds.value = []
    }
    if (res.code === 200 && res.data) {
      applyFormFromSetting(res.data)
      ElMessage.success('配置加载成功')
    } else {
      ElMessage.error(res.message || '加载配置失败')
    }
  } catch (error) {
    ElMessage.error('加载配置失败，请稍后重试')
    console.error('Load LLM setting error:', error)
  } finally {
    loading.value = false
  }
}

// 保存配置
const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitLoading.value = true
      try {
        const res = await saveLLMSetting({
          maxSize: form.maxSize,
          model: form.model,
          completionUrl: form.completionUrl,
          completionApiKey: form.completionApiKey,
          completionType: form.completionType,
          embeddingModel: form.embeddingModel,
          embeddingsUrl: form.embeddingsUrl,
          embeddingsApiKey: form.embeddingsApiKey,
          embeddingsType: form.embeddingsType,
          rerankerUrl: form.rerankerUrl,
          rerankerApiKey: form.rerankerApiKey,
          rerankerType: form.rerankerType,
          rerankerModel: form.rerankerModel,
          temperature: form.temperature,
          topK: form.topK,
          topP: form.topP,
          ragEmbeddingTopK: form.ragEmbeddingTopK,
          ragSimilarityThreshold: form.ragSimilarityThreshold,
          ragSummarySimilarityThreshold: form.ragSummarySimilarityThreshold,
          ragRerankTopK: form.ragRerankTopK,
          ragRerankScoreThreshold: form.ragRerankScoreThreshold,
          ragSummaryRerankScoreThreshold: form.ragSummaryRerankScoreThreshold
        })
        if (res.code === 200) {
          ElMessage.success('配置保存成功')
        } else {
          ElMessage.error(res.message || '保存配置失败')
        }
      } catch (error) {
        ElMessage.error('保存配置失败，请稍后重试')
        console.error('Save LLM setting error:', error)
      } finally {
        submitLoading.value = false
      }
    }
  })
}

const handleApplyDeepSeek = () => {
  const preservedKey = form.completionApiKey
  form.completionUrl = DEEPSEEK_PRESET.completionUrl
  form.completionType = DEEPSEEK_PRESET.completionType
  form.model = DEEPSEEK_PRESET.model
  form.completionApiKey = preservedKey
  if (!chatModelIds.value.includes(DEEPSEEK_PRESET.model)) {
    chatModelIds.value = [DEEPSEEK_PRESET.model, ...chatModelIds.value]
  }
  ElMessage.success('已填入 DeepSeek 预设，请填写 Chat API Key 后点击「保存配置」')
}

// 恢复为 knowsource.yaml 系统默认（仅更新界面，不保存）
const handleReset = () => {
  ElMessageBox.confirm(
    '将表单恢复为 knowsource.yaml 中的系统默认配置。仅更新页面，不会自动保存；确认后请再点击「保存配置」才会写入租户配置。',
    '恢复系统默认',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(async () => {
    resetLoading.value = true
    try {
      const res = await loadLLMSettingDefaults()
      if (res.code === 200 && res.data) {
        applyFormFromSetting(res.data)
        formRef.value?.clearValidate()
        ElMessage.success('已恢复为系统默认配置（未保存）')
      } else {
        ElMessage.error(res.message || '获取系统默认配置失败')
      }
    } catch (error) {
      ElMessage.error('获取系统默认配置失败，请稍后重试')
      console.error('Load LLM setting defaults error:', error)
    } finally {
      resetLoading.value = false
    }
  }).catch(() => {})
}

const runServiceCheck = async () => {
  testLoading.service = true
  testResult.service = ''
  try {
    const res = await sysCheck()
    if (res.code !== 200 || !res.data) {
      testResult.service = res.message || '检查失败'
      ElMessage.error(testResult.service)
      return
    }
    const d = res.data
    const msg = [
      `chat:${d.vllmchat?.ok ? 'OK' : 'FAIL'}`,
      `embedding:${d.vllmembedding?.ok ? 'OK' : 'FAIL'}`,
      `rerank:${d.vllmreranker?.ok ? 'OK' : 'FAIL'}`
    ].join(' | ')
    testResult.service = msg
    if (d.vllmchat?.ok && d.vllmembedding?.ok && d.vllmreranker?.ok) {
      ElMessage.success('服务可达性检查通过')
    } else {
      ElMessage.warning('部分服务不可达，请检查配置')
    }
  } catch (error) {
    testResult.service = error?.message || '请求失败'
    ElMessage.error(testResult.service)
  } finally {
    testLoading.service = false
  }
}

const runChatSmokeTest = async () => {
  testLoading.chat = true
  testResult.chat = ''
  try {
    const res = await callLLMOneShot({
      prompt: '请回复：ok',
      maxTokens: 32,
      temperature: 0
    })
    if (res.code === 200 && res.data?.content) {
      const preview = String(res.data.content).slice(0, 80)
      testResult.chat = `成功（返回: ${preview}）`
      ElMessage.success('Chat 调用测试成功')
    } else {
      testResult.chat = res.message || '调用失败'
      ElMessage.error(testResult.chat)
    }
  } catch (error) {
    testResult.chat = error?.message || '请求失败'
    ElMessage.error(testResult.chat)
  } finally {
    testLoading.chat = false
  }
}

const runEmbeddingRerankTest = async () => {
  testLoading.embedRerank = true
  testResult.embedRerank = ''
  try {
    const res = await testLLMServices({
      query: '什么是知识库问答',
      doc1: '知识库问答可以结合企业私有文档，提高回答准确率。',
      doc2: '普通闲聊通常不依赖私有文档检索。'
    })
    if (res.code === 200 && res.data) {
      const d = res.data
      const scoreText = Array.isArray(d.rerankScores)
        ? d.rerankScores.map(v => Number(v).toFixed(4)).join(', ')
        : '-'
      testResult.embedRerank = `embedding=${d.embeddingOk ? 'OK' : 'FAIL'} dim=${d.embeddingDimension || 0}; rerank=${d.rerankOk ? 'OK' : 'FAIL'} top=${Number(d.rerankTopScore || 0).toFixed(4)} scores=[${scoreText}]`
      if (d.embeddingOk && d.rerankOk) {
        ElMessage.success('Embedding + Rerank 测试成功')
      } else {
        ElMessage.warning(`部分失败：embeddingErr=${d.embeddingError || '-'} rerankErr=${d.rerankError || '-'}`)
      }
    } else {
      testResult.embedRerank = res.message || '测试失败'
      ElMessage.error(testResult.embedRerank)
    }
  } catch (error) {
    testResult.embedRerank = error?.message || '请求失败'
    ElMessage.error(testResult.embedRerank)
  } finally {
    testLoading.embedRerank = false
  }
}

// 页面加载时自动获取配置
onMounted(() => {
  handleLoad()
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

.form-item-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
  line-height: 1.4;
}

.test-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.test-result {
  margin-top: 8px;
  font-size: 12px;
  color: #606266;
}

:deep(.el-form-item) {
  margin-bottom: 24px;
}

:deep(.el-slider) {
  margin-top: 8px;
}
</style>

