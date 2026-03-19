<template>
  <div class="register-container">
    <el-card class="register-card shadow-lg">
      <div class="register-header text-center mb-6">
        <h2 class="text-primary font-bold text-2xl">监控工具管理系统</h2>
        <p class="text-gray-500 mt-2">注册新账号</p>
      </div>
      <el-form
        :model="registerForm"
        :rules="registerRules"
        ref="registerFormRef"
        label-width="80px"
        class="register-form"
      >
        <el-form-item label="用户名" prop="username">
          <el-input
            v-model="registerForm.username"
            placeholder="请输入用户名（3-20位）"
            prefix-icon="el-icon-user"
            size="large"
          ></el-input>
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input
            v-model="registerForm.phone"
            placeholder="请输入手机号（11位）"
            prefix-icon="el-icon-mobile"
            size="large"
            maxlength="11"
          ></el-input>
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="registerForm.password"
            type="password"
            placeholder="请输入密码（8-20位）"
            prefix-icon="el-icon-lock"
            size="large"
            show-password
          ></el-input>
        </el-form-item>
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input
            v-model="registerForm.confirmPassword"
            type="password"
            placeholder="请再次输入密码"
            prefix-icon="el-icon-lock"
            size="large"
            show-password
          ></el-input>
        </el-form-item>
        <el-form-item class="mt-4">
          <el-button
            type="primary"
            @click="handleRegister"
            class="w-full"
            size="large"
            :loading="loading"
          >
            注册
          </el-button>
        </el-form-item>
        <el-form-item class="text-center">
          <span class="text-gray-500">已有账号？</span>
          <el-button type="text" @click="goToLogin" class="text-primary">
            立即登录
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
import { register } from '@/api/user'

const router = useRouter()
const loading = ref(false)
const registerFormRef = ref()

// 注册表单
const registerForm = reactive({
  username: '',
  phone: '',
  password: '',
  confirmPassword: '',
})

// 确认密码校验
const validateConfirmPassword = (rule, value, callback) => {
  if (value === '') {
    callback(new Error('请再次输入密码'))
  } else if (value !== registerForm.password) {
    callback(new Error('两次输入密码不一致'))
  } else {
    callback()
  }
}

// 表单校验规则
const registerRules = reactive({
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度在3-20个字符之间', trigger: 'blur' },
  ],
  phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { len: 11, message: '手机号必须为11位', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: '请输入有效的手机号', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 8, max: 20, message: '密码长度在8-20个字符之间', trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: '请再次输入密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' },
  ],
})

// 注册处理
const handleRegister = async () => {
  try {
    await registerFormRef.value.validate()
    loading.value = true
    const data = {
      username: registerForm.username,
      phone: registerForm.phone,
      password: registerForm.password,
    }
    await register(data)
    ElMessage.success('注册成功，请登录')
    router.push('/login')
  } catch (error) {
    ElMessage.error('注册失败：' + (error.msg || '请检查输入信息'))
    console.error('注册错误：', error)
  } finally {
    loading.value = false
  }
}

// 跳转到登录页
const goToLogin = () => {
  router.push('/login')
}
</script>

<style scoped>
.register-container {
  min-height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background: linear-gradient(135deg, #e8f4f8 0%, #f0f8fb 100%);
  padding: 20px;
}

.register-card {
  width: 100%;
  max-width: 450px;
  border-radius: 12px;
  border: none;
  padding: 30px;
}

.shadow-lg {
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.1);
}

.register-form {
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
