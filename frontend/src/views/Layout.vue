<template>
  <el-container class="layout-container h-screen">
    <!-- 侧边栏，可收缩 -->
    <el-aside :width="isCollapse ? '64px' : '200px'" class="aside-container bg-white shadow-sm">
      <div class="logo-container text-center py-4 border-b">
        <h3 class="logo-title text-primary font-bold">智云监控中心</h3>
      </div>
      <el-menu
        default-active="1"
        class="el-menu-vertical-demo"
        router
        active-text-color="#409eff"
        background-color="#ffffff"
        text-color="#333333"
        :collapse="isCollapse"
      >
        <el-menu-item index="/dashboard">
          <el-icon><House /></el-icon>
          <template #title>仪表盘</template>
        </el-menu-item>
        <el-menu-item index="/monitor/list">
          <el-icon><Monitor /></el-icon>
          <template #title>监控项管理</template>
        </el-menu-item>
        <el-sub-menu index="2">
          <template #title>
            <el-icon><Bell /></el-icon>
            <span>告警管理</span>
          </template>
          <el-menu-item index="/alert/config">告警配置</el-menu-item>
          <el-menu-item index="/alert/record">告警记录</el-menu-item>
        </el-sub-menu>
        <el-menu-item @click="handleLogout">
          <el-icon><SwitchButton /></el-icon>
          <template #title>退出登录</template>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <!-- 主内容区 -->
    <el-container>
      <!-- 顶部导航 -->
      <el-header class="header-container shadow-sm bg-white">
        <div class="flex justify-between items-center h-full">
          <el-icon
            class="cursor-pointer text-xl collapse-toggle"
            @click="toggleCollapse"
          >
            <Fold />
          </el-icon>
          <div class="user-info">
            <el-dropdown @visible-change="handleDropdownVisible">
              <span class="flex items-center cursor-pointer">
                <el-avatar :src="userAvatar" icon="el-icon-user" class="mr-2"></el-avatar>
                <span>{{ userName }}</span>
                <el-icon :class="['ml-1', 'user-arrow', dropdownVisible ? 'user-arrow-open' : '']">
                  <ArrowDown />
                </el-icon>
              </span>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item @click="openProfileDialog">个人信息</el-dropdown-item>
                  <el-dropdown-item @click="triggerAvatarUpload">上传头像</el-dropdown-item>
                  <el-dropdown-item divided @click="handleLogout">退出登录</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
            <!-- 隐藏文件选择，用于上传头像 -->
            <input
              ref="avatarInputRef"
              type="file"
              accept="image/*"
              style="display: none"
              @change="handleAvatarChange"
            />
          </div>
        </div>
      </el-header>

      <!-- 个人信息弹窗 -->
      <el-dialog
        v-model="profileDialogVisible"
        title="个人信息"
        width="420px"
        :close-on-click-modal="false"
      >
        <el-form
          :model="profileForm"
          :rules="profileRules"
          ref="profileFormRef"
          label-width="90px"
        >
          <el-form-item label="昵称" prop="username">
            <el-input v-model="profileForm.username" placeholder="请输入昵称"></el-input>
          </el-form-item>
          <el-form-item label="手机号">
            <el-input v-model="profileForm.phone" disabled></el-input>
          </el-form-item>
          <el-form-item label="会员等级">
            <el-tag type="warning">Lv{{ profileForm.memberLevel || 1 }}</el-tag>
          </el-form-item>
          <el-form-item label="注册时间" v-if="profileForm.createdAt">
            <span>{{ profileForm.createdAt }}</span>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="profileDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="submitProfile">保存</el-button>
        </template>
      </el-dialog>

      <!-- 内容区域 -->
      <el-main class="main-container bg-gray-50 p-6">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  House,
  Monitor,
  Bell,
  SwitchButton,
  Fold,
  ArrowDown,
} from '@element-plus/icons-vue'
import { getUserInfo, uploadAvatar, updateProfile } from '@/api/user'

const router = useRouter()
const isCollapse = ref(false)
const userName = ref('管理员')
const userAvatar = ref('')
const avatarInputRef = ref(null)
const dropdownVisible = ref(false)

// 个人信息弹窗
const profileDialogVisible = ref(false)
const profileFormRef = ref(null)
const profileForm = reactive({
  username: '',
  phone: '',
  memberLevel: '',
  createdAt: '',
})
const profileRules = reactive({
  username: [
    { required: true, message: '请输入昵称', trigger: 'blur' },
    { min: 3, max: 20, message: '昵称长度需在3-20个字符之间', trigger: 'blur' },
  ],
})

// 获取用户信息
const getUserData = async () => {
  try {
    const res = await getUserInfo()
    const data = res.data || {}
    userName.value = data.username || '管理员'
    userAvatar.value = data.avatar || ''
    profileForm.username = data.username || '管理员'
    profileForm.phone = data.phone || ''
    profileForm.memberLevel = data.member_level || 1
    profileForm.createdAt = data.created_at || ''
  } catch (error) {
    console.error('获取用户信息失败：', error)
  }
}

// 侧边栏收缩
const toggleCollapse = () => {
  isCollapse.value = !isCollapse.value
}

// 下拉箭头动画
const handleDropdownVisible = (visible) => {
  dropdownVisible.value = visible
}

// 触发头像上传
const triggerAvatarUpload = () => {
  avatarInputRef.value && avatarInputRef.value.click()
}

// 处理头像文件选择并上传
const handleAvatarChange = async (event) => {
  const file = event.target.files && event.target.files[0]
  if (!file) return
  const formData = new FormData()
  formData.append('file', file)
  try {
    const res = await uploadAvatar(formData)
    userAvatar.value = res.data.avatar || ''
    ElMessage.success('头像上传成功')
  } catch (error) {
    ElMessage.error('头像上传失败：' + (error.msg || '请稍后重试'))
    console.error('头像上传错误：', error)
  } finally {
    // 清空文件选择，方便下次重新选择同一文件
    event.target.value = ''
  }
}

// 打开个人信息弹窗
const openProfileDialog = () => {
  profileDialogVisible.value = true
}

// 提交个人信息修改
const submitProfile = async () => {
  try {
    await profileFormRef.value?.validate()
    await updateProfile({ username: profileForm.username })
    userName.value = profileForm.username || userName.value
    ElMessage.success('个人信息更新成功')
    profileDialogVisible.value = false
  } catch (error) {
    // 表单校验错误已由组件提示，这里仅处理接口错误
    if (error && error.msg) {
      ElMessage.error(error.msg)
    } else if (error?.response?.data?.msg) {
      ElMessage.error('个人信息更新失败：' + error.response.data.msg)
    } else if (error && error.message) {
      console.error('个人信息更新错误：', error)
    }
  }
}

// 退出登录
const handleLogout = () => {
  ElMessageBox.confirm('确定要退出登录吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  }).then(async () => {
    localStorage.removeItem('token')
    ElMessage.success('退出成功')
    router.push('/login')
  })
}

onMounted(() => {
  getUserData()
})
</script>

<style scoped>
.layout-container {
  --el-header-height: 60px;
}

/* 侧边栏随容器布局，不再遮挡内容 */
.aside-container {
  height: 100vh;
  transition: width 0.2s ease;
  overflow: hidden;
}

.header-container {
  padding: 0 20px;
}

.main-container {
  min-height: calc(100vh - 60px);
}

.logo-title {
  font-size: 22px;
  letter-spacing: 2px;
}

.flex {
  display: flex;
}

.justify-between {
  justify-content: space-between;
}

.items-center {
  align-items: center;
}

.cursor-pointer {
  cursor: pointer;
}

.mr-2 {
  margin-right: 8px;
}

.ml-1 {
  margin-left: 4px;
}

.text-xl {
  font-size: 20px;
}

.collapse-toggle {
  transition: transform 0.2s ease;
}

.user-arrow {
  transition: transform 0.2s ease;
}

.user-arrow-open {
  transform: rotate(180deg);
}
</style>