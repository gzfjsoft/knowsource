<template>
  <div class="email-verification-container">
    <div class="verification-box">
      <div class="verification-header">
        <h2>邮箱验证</h2>
        <p>请验证您的邮箱地址以激活账户</p>
      </div>
      
      <div v-if="!showVerificationForm" class="email-info">
        <p>我们已向您的邮箱 <span class="email-address">{{ email }}</span> 发送了验证码</p>
        <p class="tip">请检查您的邮箱（包括垃圾邮件文件夹）获取验证码</p>
        <el-button type="primary" @click="showVerificationForm = true">
          已收到验证码，现在验证
        </el-button>
        <el-button @click="resendVerificationCode">
          重新发送验证码
        </el-button>
      </div>
      
      <el-form
        v-else
        ref="verificationFormRef"
        :model="verificationForm"
        :rules="verificationRules"
        class="verification-form"
      >
        <el-form-item label="验证码" prop="code">
          <div class="code-row">
            <el-input
              v-model="verificationForm.code"
              placeholder="请输入邮箱验证码"
              size="large"
              prefix-icon="Key"
              style="flex: 1; margin-right: 10px"
              @keyup.enter="handleVerify"
            />
            <el-button
              type="primary"
              :loading="resending"
              :disabled="countdown > 0"
              @click="resendVerificationCode"
            >
              {{ countdown > 0 ? `${countdown}秒后重发` : '重新发送' }}
            </el-button>
          </div>
        </el-form-item>
        
        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="verifying"
            @click="handleVerify"
            style="width: 100%"
          >
            验证
          </el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from "vue";
import { useRouter, useRoute } from "vue-router";
import { ElMessage } from "element-plus";
import { sendVerificationCode, tenantVerifyEmail } from "@/api/knowsource";

const router = useRouter();
const route = useRoute();

const showVerificationForm = ref(false);
const resending = ref(false);
const verifying = ref(false);
const countdown = ref(0);
const verificationFormRef = ref(null);

const clientId = ref(localStorage.getItem("clientId") || "");
const email = ref("");

const verificationForm = reactive({
  code: ""
});

const verificationRules = {
  code: [{ required: true, message: "请输入验证码", trigger: "blur" }]
};

const startCountdown = () => {
  countdown.value = 60;
  const timer = setInterval(() => {
    if (countdown.value > 0) {
      countdown.value--;
    } else {
      clearInterval(timer);
    }
  }, 1000);
};

const resendVerificationCode = async () => {
  if (countdown.value > 0) return;
  
  resending.value = true;
  try {
    const res = await sendVerificationCode({ clientId: clientId.value });
    if (res.code === 200) {
      ElMessage.success("验证码已重新发送，请查收邮箱");
      startCountdown();
    } else {
      ElMessage.error(res.message || "发送验证码失败");
    }
  } catch (error) {
    console.error("发送验证码失败:", error);
    ElMessage.error("发送验证码失败，请稍后重试");
  } finally {
    resending.value = false;
  }
};

const handleVerify = async () => {
  if (!verificationFormRef.value) return;
  
  await verificationFormRef.value.validate(async (valid) => {
    if (valid) {
      verifying.value = true;
      try {
        const res = await tenantVerifyEmail({
          clientId: clientId.value,
          code: verificationForm.code
        });
        if (res.code === 200) {
          ElMessage.success("邮箱验证成功，即将跳转到登录页面");
          setTimeout(() => {
            router.push("/login");
          }, 2000);
        } else {
          ElMessage.error(res.message || "验证失败");
        }
      } catch (error) {
        console.error("验证失败:", error);
        ElMessage.error("验证失败，请稍后重试");
      } finally {
        verifying.value = false;
      }
    }
  });
};

// 页面加载时发送验证码
onMounted(async () => {
  if (!clientId.value) {
    ElMessage.error("缺少企业账户信息");
    setTimeout(() => {
      router.push("/login");
    }, 1500);
    return;
  }
  
  // 发送验证码
  resendVerificationCode();
});
</script>

<style scoped>
.email-verification-container {
  width: 100%;
  height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  justify-content: center;
  align-items: center;
}

.verification-box {
  width: 400px;
  padding: 40px;
  background: #fff;
  border-radius: 10px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.1);
}

.verification-header {
  text-align: center;
  margin-bottom: 30px;
}

.verification-header h2 {
  margin: 0 0 10px 0;
  color: #303133;
  font-size: 24px;
}

.verification-header p {
  margin: 0;
  color: #909399;
  font-size: 14px;
}

.email-info {
  text-align: center;
  margin-bottom: 30px;
}

.email-info p {
  margin: 10px 0;
  color: #606266;
}

.email-address {
  font-weight: bold;
  color: #409eff;
}

.tip {
  font-size: 12px;
  color: #909399;
}

.verification-form {
  margin-top: 20px;
}

.code-row {
  display: flex;
  align-items: center;
}

.code-row .el-button {
  white-space: nowrap;
}
</style>