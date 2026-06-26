<template>
  <div class="login-container">
    <TechBackground />
    <div class="login-box">
      <div class="login-header">
        <h2>知源智库 AI</h2>
        <p>欢迎登录</p>
      </div>
      <el-form
        ref="loginFormRef"
        :model="loginForm"
        :rules="loginRules"
        class="login-form"
      >
        <el-form-item prop="clientId">
          <el-input
            v-model="loginForm.clientId"
            placeholder="请输入企业账户名"
            size="large"
            prefix-icon="Key"
            @keyup.enter="focusEmpCode"
          />
        </el-form-item>
        <el-form-item prop="empCode">
          <el-input
            ref="empCodeInputRef"
            v-model="loginForm.empCode"
            placeholder="请输入员工编码"
            size="large"
            prefix-icon="User"
            @keyup.enter="focusPassword"
          />
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            ref="passwordInputRef"
            v-model="loginForm.password"
            type="password"
            placeholder="请输入密码"
            size="large"
            prefix-icon="Lock"
            show-password
            @keyup.enter="focusCaptcha"
          />
        </el-form-item>
        <el-form-item prop="captcha">
          <div class="captcha-row">
            <el-input
              ref="captchaInputRef"
              v-model="loginForm.captcha"
              placeholder="请输入验证码"
              size="large"
              prefix-icon="Key"
              style="flex: 1; margin-right: 10px"
              @keyup.enter="handleLogin"
            />
            <div class="captcha-img" @click="refreshCaptcha">
              <img v-if="captchaUrl" :src="captchaUrl" alt="验证码" />
              <span v-else>点击获取验证码</span>
            </div>
          </div>
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            @click="handleLogin"
            style="width: 100%"
          >
            登录
          </el-button>
        </el-form-item>
        <div class="login-links">
          <el-button link @click="goIndex">了解产品</el-button>
          <el-button link type="primary" @click="goEnterpriseRegister"
            >注册企业账户</el-button
          >
        </div>
      </el-form>
      <div class="" style="text-align: center; font-size: 12px; color: #909399; margin-top: 50px;margin-bottom: 50px;">
        演示的企业账户名称:demo 员工编码:demo 密码:Demo1234!
      </div>
      <div class="copyright">
     
        <SiteIcpLine theme="light" class="icp-footer" />
      </div>
      
    </div>
    
  </div>
 
</template>

<script setup>
import { ref, reactive, onMounted, nextTick } from "vue";
import { useRouter } from "vue-router";
import { useUserStore } from "@/stores/user";
import { ElMessage } from "element-plus";
import { getCaptcha } from "@/api/knowsource";
import TechBackground from "@/components/TechBackground.vue";
import SiteIcpLine from "@/components/SiteIcpLine.vue";

const router = useRouter();
const userStore = useUserStore();

const loginFormRef = ref(null);
const empCodeInputRef = ref(null);
const passwordInputRef = ref(null);
const captchaInputRef = ref(null);
const loading = ref(false);
const showCaptcha = ref(true);
const captchaUrl = ref("");

const loginForm = reactive({
  clientId: localStorage.getItem("clientId") || "",
  empCode: "",
  password: "",
  captcha: "",
  captchaId: "",
});

const loginRules = {
  clientId: [{ required: true, message: "请输入企业账户名", trigger: "blur" }],
  empCode: [{ required: true, message: "请输入员工编码", trigger: "blur" }],
  password: [{ required: true, message: "请输入密码", trigger: "blur" }],
  captcha: [{ required: true, message: "请输入验证码", trigger: "blur" }],
};

const focusPassword = async () => {
  await nextTick();
  passwordInputRef.value?.focus();
};

const focusEmpCode = async () => {
  await nextTick();
  empCodeInputRef.value?.focus();
};

const focusCaptcha = async () => {
  await nextTick();
  captchaInputRef.value?.focus();
};

const refreshCaptcha = async () => {
  try {
    const res = await getCaptcha();
    if (res.code === 200 && res.data) {
      loginForm.captchaId = res.data.captchaId;
      captchaUrl.value = `${res.data.imageBase64}`;
    } else {
      ElMessage.error("获取验证码失败");
    }
  } catch (error) {
    console.error("获取验证码失败:", error);
    ElMessage.error("获取验证码失败，请稍后重试");
  }
};

const handleLogin = async () => {
  if (!loginFormRef.value) return;

  await loginFormRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true;
      try {
        localStorage.setItem("clientId", loginForm.clientId);
        const result = await userStore.login(loginForm);
        if (result.success) {
          ElMessage.success("登录成功");
          router.push("/");
        } else {
          // 显示后端返回的错误消息
          ElMessage.error(result.message || "登录失败，请检查账号密码");
          // 登录失败后刷新验证码
          refreshCaptcha();
        }
      } catch (error) {
        // 捕获未处理的错误
        const errorMessage =
          error.response?.data?.message ||
          error.message ||
          "登录失败，请稍后重试";
        ElMessage.error(errorMessage);
        // 登录失败后刷新验证码
        refreshCaptcha();
      } finally {
        loading.value = false;
      }
    }
  });
};

const goEnterpriseRegister = () => {
  router.push("/enterprise/register");
};

const goIndex = () => {
  router.push("/index");
};

// 页面加载时获取验证码
onMounted(() => {
  refreshCaptcha();
});
</script>

<style scoped>
.login-container {
  width: 100%;
  height: 100vh;
  position: relative;
  display: flex;
  justify-content: center;
  align-items: center;
  overflow: hidden;
}

.login-box {
  width: 400px;
  padding: 40px;
  background: rgba(255, 255, 255, 0.92);
  border: 1px solid rgba(148, 163, 184, 0.35);
  backdrop-filter: blur(12px);
  border-radius: 10px;
  box-shadow: 0 30px 80px rgba(0, 0, 0, 0.45);
  position: relative;
  z-index: 1;
}

.login-header {
  text-align: center;
  margin-bottom: 30px;
}

.login-header h2 {
  margin: 0 0 10px 0;
  color: #0f172a;
  font-size: 24px;
}

.login-header p {
  margin: 0;
  color: #475569;
  font-size: 14px;
}

.login-form {
  margin-top: 20px;
}

.login-links {
  display: flex;
  justify-content: center;
  gap: 10px;
  margin-top: 4px;
}

.captcha-row {
  display: flex;
  align-items: center;
}

.captcha-img {
  width: 120px;
  height: 40px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: rgba(255, 255, 255, 0.85);
}

.captcha-img img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.captcha-img span {
  font-size: 12px;
  color: #64748b;
}

.copyright {
  margin-top: 20px;
  text-align: center;
  font-size: 12px;
  color: #64748b;
}

.copyright .icp-footer {
  margin-top: 10px;
}
</style>
