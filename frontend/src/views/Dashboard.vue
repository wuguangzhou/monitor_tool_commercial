<template>
  <div class="dashboard-container">
    <el-row :gutter="20" class="mb-6">
      <!-- 统计卡片 -->
      <el-col :span="6">
        <el-card class="stat-card shadow-sm">
          <div class="stat-content">
            <p class="stat-label text-gray-500">总监控项</p>
            <h3 class="stat-value text-3xl font-bold text-primary mt-2">{{ totalMonitor }}</h3>
            <p class="stat-desc text-green-500 mt-2">
              <el-icon><TrendCharts /></el-icon> 正常：{{ normalMonitor }} 个
            </p>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card shadow-sm">
          <div class="stat-content">
            <p class="stat-label text-gray-500">宕机监控项</p>
            <h3 class="stat-value text-3xl font-bold text-danger mt-2">{{ downMonitor }}</h3>
            <p class="stat-desc text-orange-500 mt-2">
              <el-icon><Warning /></el-icon> 需及时处理
            </p>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card shadow-sm">
          <div class="stat-content">
            <p class="stat-label text-gray-500">今日告警数</p>
            <h3 class="stat-value text-3xl font-bold text-warning mt-2">{{ todayAlert }}</h3>
            <p class="stat-desc text-gray-500 mt-2">
              <el-icon><Clock /></el-icon> 恢复：{{ todayRecovery }} 个
            </p>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card shadow-sm">
          <div class="stat-content">
            <p class="stat-label text-gray-500">告警配置率</p>
            <h3 class="stat-value text-3xl font-bold text-success mt-2">{{ alertConfigRate }}%</h3>
            <p class="stat-desc text-gray-500 mt-2">
              <el-icon><Setting /></el-icon> 已配置：{{ configCount }} 人
            </p>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 最近告警记录 -->
    <el-card class="shadow-sm">
      <template #header>
        <div class="flex justify-between items-center">
          <span class="font-bold">最近告警记录</span>
          <el-button type="text" @click="goToAlertRecord">查看全部</el-button>
        </div>
      </template>
      <el-table :data="recentAlertList" border stripe hover>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="monitorName" label="监控项名称" min-width="150" />
        <el-table-column prop="alertSubType" label="类型" width="100">
          <template #default="scope">
            <el-tag v-if="scope.row.alertSubType === 1" type="danger">宕机告警</el-tag>
            <el-tag v-if="scope.row.alertSubType === 2" type="success">恢复通知</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="content" label="内容" min-width="300" show-overflow-tooltip />
        <el-table-column prop="createdAt" label="时间" width="180" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag v-if="scope.row.status === 1" type="success">已发送</el-tag>
            <el-tag v-if="scope.row.status === 0" type="warning">未发送</el-tag>
            <el-tag v-if="scope.row.status === 2" type="danger">发送失败</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import {
  TrendCharts,
  Warning,
  Clock,
  Setting,
} from '@element-plus/icons-vue'
import { getMonitorList } from '@/api/monitor'
import { getAlertList } from '@/api/alert'

const router = useRouter()

// 统计数据
const totalMonitor = ref(0)
const normalMonitor = ref(0)
const downMonitor = ref(0)
const todayAlert = ref(0)
const todayRecovery = ref(0)
const alertConfigRate = ref(85)
const configCount = ref(12)

// 最近告警记录
const recentAlertList = ref([])

// 获取仪表盘数据
const getDashboardData = async () => {
  try {
    // 获取监控项统计
    const monitorRes = await getMonitorList({ page: 1, size: 1000 })
    const monitorList = monitorRes.data.list || []
    totalMonitor.value = monitorList.length
    normalMonitor.value = monitorList.filter(item => item.status === 1).length
    downMonitor.value = monitorList.filter(item => item.status === 2).length

    // 获取最近告警记录
    const alertRes = await getAlertList({ page: 1, size: 10 })
    recentAlertList.value = alertRes.data.list || []
    todayAlert.value = recentAlertList.value.filter(item =>
      item.alertSubType === 1 && item.createdAt.includes(new Date().toLocaleDateString().replace(/\//g, '-'))
    ).length
    todayRecovery.value = recentAlertList.value.filter(item =>
      item.alertSubType === 2 && item.createdAt.includes(new Date().toLocaleDateString().replace(/\//g, '-'))
    ).length
  } catch (error) {
    console.error('获取仪表盘数据失败：', error)
  }
}

// 跳转到告警记录页
const goToAlertRecord = () => {
  router.push('/alert/record')
}

onMounted(() => {
  getDashboardData()
})
</script>

<style scoped>
.dashboard-container {
  padding: 0;
}

.mb-6 {
  margin-bottom: 1.5rem;
}

.stat-card {
  height: 100%;
  border-radius: 8px;
}

.stat-content {
  padding: 10px 0;
  text-align: center;
}

.stat-label {
  font-size: 14px;
}

.stat-value {
  margin: 8px 0;
}

.stat-desc {
  font-size: 12px;
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
</style>