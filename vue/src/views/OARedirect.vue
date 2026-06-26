<template>
  <div class="oa-redirect-container">
    <div class="loading-box">
      <el-icon class="loading-icon" :size="40">
        <Loading />
      </el-icon>
      <p class="loading-text">{{ loadingText }}</p>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useUserStore } from "@/stores/user";
import { ElMessage } from "element-plus";
import { Loading } from "@element-plus/icons-vue";
import { oaLogin } from "@/api/knowsource";

const route = useRoute();
const router = useRouter();
const userStore = useUserStore();

const loadingText = ref("正在验证登录信息...");

// 从 URL 获取参数并调用登录接口
const handleOALogin = async () => {
  // 获取 URL 参数
  const userId = route.query.userid || route.query.userId;
  const code = route.query.code;
  const clientId = route.query.clientId || route.query.client_id;

  // 检查参数是否存在
  if (!clientId || !userId || !code) {
    ElMessage.error("缺少必要的登录参数（clientId / userid / code）");
    loadingText.value = "登录失败：缺少参数";
    setTimeout(() => {
      router.push("/login");
    }, 2000);
    return;
  }

  try {
    loadingText.value = "正在验证...";

    // 调用 OA 登录接口
    const res = await oaLogin({
      clientId: clientId,
      userId: userId,
      code: code,
    });

    if (res.code === 200 && res.data) {
      // 登录成功，保存用户信息
      userStore.token = res.data.token;
      userStore.userInfo = res.data.userInfo;
      localStorage.setItem("token", res.data.token);
      localStorage.setItem("userInfo", JSON.stringify(res.data.userInfo));
      localStorage.setItem("clientId", String(clientId).trim());
      // 触发自定义事件，通知 App.vue 启动 token 刷新定时器
      window.dispatchEvent(new CustomEvent("user-logged-in"));

      loadingText.value = "登录成功，正在跳转...";
      ElMessage.success("登录成功");

      // 延迟一下再跳转，让用户看到成功提示
      setTimeout(() => {
        // 跳转到首页，路由守卫会根据 role 自动重定向
        router.push("/");
      }, 500);
    } else {
      // 登录失败
      const errorMsg = res.message || "登录失败，请检查参数";
      ElMessage.error(errorMsg);
      loadingText.value = `登录失败：${errorMsg}`;

      setTimeout(() => {
        router.push("/login");
      }, 2000);
    }
  } catch (error) {
    console.error("OA Login error:", error);
    const errorMessage =
      error.response?.data?.message || error.message || "登录失败，请稍后重试";
    ElMessage.error(errorMessage);
    loadingText.value = `登录失败：${errorMessage}`;

    setTimeout(() => {
      router.push("/login");
    }, 2000);
  }
};

// 页面加载时执行登录
onMounted(() => {
  handleOALogin();
});
</script>

<style scoped>
.oa-redirect-container {
  width: 100%;
  height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  justify-content: center;
  align-items: center;
}

.loading-box {
  text-align: center;
  color: #fff;
}

.loading-icon {
  animation: rotate 1s linear infinite;
  margin-bottom: 20px;
}

@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.loading-text {
  font-size: 16px;
  margin: 0;
  color: #fff;
}
</style>
