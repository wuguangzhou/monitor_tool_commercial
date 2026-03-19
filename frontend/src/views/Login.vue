<template>
  <div class="login-container">
    <el-card class="login-card shadow-lg">
      <div class="login-header text-center mb-6">
        <h2 class="text-primary font-bold text-2xl">监控工具管理系统</h2>
        <p class="text-gray-500 mt-2">高效监控，实时告警</p>
      </div>
      <el-form
        :model="loginForm"
        :rules="loginRules"
        ref="loginFormRef"
        label-width="80px"
        class="login-form"
      >
        <el-form-item label="手机号" prop="phone">
          <el-input
            v-model="loginForm.phone"
            placeholder="请输入手机号"
            prefix-icon="el-icon-user"
            size="large"
          ></el-input>
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="请输入密码"
            prefix-icon="el-icon-lock"
            size="large"
            show-password
          ></el-input>
        </el-form-item>
        <el-form-item class="mt-4">
          <el-button
            type="primary"
            @click="handleLogin"
            class="w-full"
            size="large"
            :loading="loading"
          >
            登录
          </el-button>
        </el-form-item>
        <el-form-item class="text-center">
          <span class="text-gray-500">还没有账号？</span>
          <el-button type="text" @click="goToRegister" class="text-primary">
            立即注册
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { login } from '@/api/user'

const router = useRouter()
const loading = ref(false)
const loginFormRef = ref()

// 登录表单
const loginForm = reactive({
  phone: '',
  password: '',
})

// 表单校验规则
const loginRules = reactive({
  phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { len: 11, message: '手机号必须为11位', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' },
  ],
})

// 登录处理
const handleLogin = async () => {
  try {
    await loginFormRef.value.validate()
    loading.value = true
    const res = await login(loginForm)
    // 存储token到localStorage
    localStorage.setItem('token', res.data.token)
    ElMessage.success('登录成功')
    router.push('/dashboard')
  } catch (error) {
    ElMessage.error('登录失败：' + (error.msg || '账号或密码错误'))
    console.error('登录错误：', error)
  } finally {
    loading.value = false
  }
}

// 跳转到注册页
const goToRegister = () => {
  router.push('/register')
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background: linear-gradient(135deg, #e8f4f8 0%, #f0f8fb 100%);
  padding: 20px;
}

.login-card {
  width: 100%;
  max-width: 450px;
  border-radius: 12px;
  border: none;
  padding: 30px;
}

.shadow-lg {
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.1);
}

.login-form {
  margin-top: 20px;
}

.mb-6 {
  margin-bottom: 1.5rem;
}

.mt-2 {
  margin-top: 0.5rem;
}

.mt-4 {
  margin-top: 1rem;
}
</style>