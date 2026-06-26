<template>
  <div class="register-container">
    <TechBackground />
    <div class="register-box">
      <div class="register-header">
        <h2>企业账户注册</h2>
        <p>创建企业账户后，会自动创建 superadmin</p>
      </div>

      <el-steps :active="step" align-center finish-status="success" class="mb">
        <el-step title="填写信息" />
        <el-step title="邮箱验证" />
        <el-step title="完成" />
      </el-steps>

      <el-form
        v-if="step === 0"
        ref="formRef"
        :model="form"
        :rules="rules"
        class="register-form"
        label-position="top"
      >
        <el-form-item label="企业账户名" prop="clientId">
          <el-input
            v-model="form.clientId"
            placeholder="长度 > 6，仅字母/数字/下划线/横线"
          />
        </el-form-item>
        <el-form-item label="用户名">
          <el-input :model-value="SUPERADMIN_USERNAME" disabled />
        </el-form-item>
        <el-form-item label="企业名称" prop="companyName">
          <el-input v-model="form.companyName" placeholder="用于界面展示" />
        </el-form-item>
        <el-form-item label="邮箱" prop="ownerEmail">
          <el-input
            v-model="form.ownerEmail"
            placeholder="用于验证与找回密码"
          />
        </el-form-item>
        <el-form-item label="手机号" prop="ownerMobile">
          <el-input
            v-model="form.ownerMobile"
            placeholder="用于找回密码/绑定"
          />
        </el-form-item>
        <el-form-item label="初始密码（员工编码：superadmin）" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            show-password
            placeholder=">=8 位，含大小写/数字/符号"
          />
        </el-form-item>
        <el-form-item label="确认密码" prop="password2">
          <el-input
            v-model="form.password2"
            type="password"
            show-password
            placeholder="再次输入密码"
          />
        </el-form-item>
        <el-form-item label="备注（可选）" prop="desp">
          <el-input v-model="form.desp" type="textarea" :rows="3" />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            :loading="loading"
            style="width: 100%"
            @click="handleSubmit"
          >
            注册企业账户
          </el-button>
        </el-form-item>

        <div class="footer-actions">
          <el-button link type="primary" @click="goLogin">返回登录</el-button>
        </div>
      </el-form>

      <div v-else-if="step === 1" class="verify-panel">
        <el-alert
          type="success"
          show-icon
          :closable="false"
          title="注册成功：验证码已发送到邮箱"
          class="mb"
        />
        <div class="account-info-card">
          <p class="account-info-title">请妥善保存以下账户信息</p>
          <p><span class="label">企业账户名：</span>{{ form.clientId }}</p>
          <p><span class="label">用户名：</span>{{ registerInfo.username }}</p>
          <p><span class="label">企业名称：</span>{{ form.companyName }}</p>
        </div>
        <div class="verify-hint">
          请到
          <strong>{{ form.ownerEmail }}</strong>
          查收验证码并完成验证。邮件中亦包含企业账户名与用户名。验证通过后即可登录。
        </div>
        <el-form
          ref="verifyFormRef"
          :model="verifyForm"
          :rules="verifyRules"
          label-position="top"
          class="register-form"
        >
          <el-form-item label="邮箱验证码" prop="code">
            <el-input v-model="verifyForm.code" placeholder="请输入验证码" />
          </el-form-item>
          <el-form-item>
            <el-button
              type="primary"
              :loading="verifyLoading"
              style="width: 100%"
              @click="handleVerify"
            >
              验证邮箱并启用企业账户
            </el-button>
          </el-form-item>
          <div class="footer-actions">
            <el-button link type="primary" @click="goLogin">返回登录</el-button>
          </div>
        </el-form>
      </div>

      <div v-else class="done-panel">
        <el-result
          icon="success"
          title="注册完成"
          sub-title="企业账户已启用，请使用以下信息登录"
        >
          <template #extra>
            <div class="account-info-card done">
              <p><span class="label">企业账户名：</span>{{ form.clientId }}</p>
              <p><span class="label">用户名：</span>{{ registerInfo.username }}</p>
              <p><span class="label">企业名称：</span>{{ form.companyName }}</p>
              <p><span class="label">邮箱：</span>{{ form.ownerEmail }}</p>
            </div>
            <el-button type="primary" @click="goLogin">去登录</el-button>
          </template>
        </el-result>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from "vue";
import { useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import { tenantRegister, tenantVerifyEmail } from "@/api/knowsource";
import TechBackground from "@/components/TechBackground.vue";

const SUPERADMIN_USERNAME = "superadmin";

const router = useRouter();
const step = ref(0);
const loading = ref(false);
const verifyLoading = ref(false);
const formRef = ref(null);
const verifyFormRef = ref(null);

const registerInfo = reactive({
  username: SUPERADMIN_USERNAME,
});

const form = reactive({
  clientId: "",
  companyName: "",
  ownerEmail: "",
  ownerMobile: "",
  password: "",
  password2: "",
  desp: "",
});

const verifyForm = reactive({
  code: "",
});

const accountPattern = /^[a-zA-Z0-9_-]+$/;
const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const passwordStrong =
  /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[^a-zA-Z0-9]).{8,}$/;

const rules = {
  clientId: [
    { required: true, message: "请输入企业账户名", trigger: "blur" },
    { min: 7, message: "企业账户名长度必须大于 6 位", trigger: "blur" },
    {
      validator: (_, v, cb) =>
        accountPattern.test(String(v || ""))
          ? cb()
          : cb(new Error("仅支持字母、数字、下划线和横线")),
      trigger: "blur",
    },
  ],
  companyName: [{ required: true, message: "请输入企业名称", trigger: "blur" }],
  ownerEmail: [
    { required: true, message: "请输入邮箱", trigger: "blur" },
    {
      validator: (_, v, cb) =>
        emailPattern.test(String(v || "").trim())
          ? cb()
          : cb(new Error("邮箱格式不正确")),
      trigger: "blur",
    },
  ],
  ownerMobile: [{ required: true, message: "请输入手机号", trigger: "blur" }],
  password: [
    { required: true, message: "请输入密码", trigger: "blur" },
    {
      validator: (_, v, cb) =>
        passwordStrong.test(String(v || ""))
          ? cb()
          : cb(new Error(">=8 位，且包含大小写字母、数字和符号")),
      trigger: "blur",
    },
  ],
  password2: [
    { required: true, message: "请确认密码", trigger: "blur" },
    {
      validator: (_, v, cb) =>
        String(v || "") === String(form.password || "")
          ? cb()
          : cb(new Error("两次输入密码不一致")),
      trigger: "blur",
    },
  ],
};

const verifyRules = {
  code: [{ required: true, message: "请输入验证码", trigger: "blur" }],
};

const goLogin = () => router.push("/login");

const handleSubmit = async () => {
  if (!formRef.value) return;
  await formRef.value.validate(async (ok) => {
    if (!ok) return;
    loading.value = true;
    try {
      const res = await tenantRegister({
        clientId: String(form.clientId || "").trim(),
        companyName: String(form.companyName || "").trim(),
        ownerEmail: String(form.ownerEmail || "").trim(),
        ownerMobile: String(form.ownerMobile || "").trim(),
        password: form.password,
        desp: String(form.desp || "").trim(),
      });
      if (res.code === 200) {
        registerInfo.username =
          res.data?.username || SUPERADMIN_USERNAME;
        ElMessage.success("注册成功，请完成邮箱验证");
        step.value = 1;
      } else {
        ElMessage.error(res.message || "注册失败");
      }
    } catch (e) {
      ElMessage.error(e?.message || "注册失败，请稍后重试");
    } finally {
      loading.value = false;
    }
  });
};

const handleVerify = async () => {
  if (!verifyFormRef.value) return;
  await verifyFormRef.value.validate(async (ok) => {
    if (!ok) return;
    verifyLoading.value = true;
    try {
      const res = await tenantVerifyEmail({
        clientId: String(form.clientId || "").trim(),
        email: String(form.ownerEmail || "").trim(),
        code: String(verifyForm.code || "").trim(),
      });
      if (res.code === 200) {
        ElMessage.success("验证成功");
        step.value = 2;
      } else {
        ElMessage.error(res.message || "验证失败");
      }
    } catch (e) {
      ElMessage.error(e?.message || "验证失败，请稍后重试");
    } finally {
      verifyLoading.value = false;
    }
  });
};
</script>

<style scoped>
.register-container {
  width: 100%;
  min-height: 100vh;
  position: relative;
  display: flex;
  justify-content: center;
  align-items: center;
  overflow: hidden;
  padding: 24px 12px;
  box-sizing: border-box;
}
.register-box {
  width: 520px;
  padding: 32px 32px 18px;
  background: rgba(255, 255, 255, 0.92);
  border: 1px solid rgba(148, 163, 184, 0.35);
  backdrop-filter: blur(12px);
  border-radius: 10px;
  box-shadow: 0 30px 80px rgba(0, 0, 0, 0.45);
  position: relative;
  z-index: 1;
}
.register-header {
  text-align: center;
  margin-bottom: 18px;
}
.register-header h2 {
  margin: 0 0 8px 0;
  color: #0f172a;
  font-size: 22px;
}
.register-header p {
  margin: 0;
  color: #475569;
  font-size: 13px;
}
.mb {
  margin-bottom: 14px;
}
.verify-panel .verify-hint {
  color: #606266;
  font-size: 13px;
  line-height: 1.5;
  margin: 8px 0 10px;
}
.account-info-card {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 12px 14px;
  margin: 10px 0 12px;
  font-size: 13px;
  color: #334155;
  line-height: 1.8;
}
.account-info-card.done {
  text-align: left;
  margin-bottom: 16px;
}
.account-info-title {
  margin: 0 0 6px;
  font-weight: 600;
  color: #0f172a;
}
.account-info-card .label {
  color: #64748b;
}
.footer-actions {
  display: flex;
  justify-content: center;
  margin-top: 6px;
}
</style>
