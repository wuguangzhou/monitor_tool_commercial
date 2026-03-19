<template>
  <div class="alert-record-container">
    <!-- 搜索栏 -->
    <div class="search-bar mb-4 flex justify-between items-center">
      <el-row :gutter="30">
        <el-col :span="18">
          <el-input
            v-model="searchKeyword"
            placeholder="请输入监控项名称/告警内容搜索"
            prefix-icon="el-icon-search"
            @keyup.enter="loadAlertList"
          ></el-input>
        </el-col>
        <el-col :span="11">
          <el-select
            v-model="alertType"
            placeholder="请选择告警类型"
            @change="loadAlertList"
          >
            <el-option label="全部" value=""></el-option>
            <el-option label="宕机告警" value="1"></el-option>
            <el-option label="恢复通知" value="2"></el-option>
          </el-select>
        </el-col>
        <el-col :span="11">
          <el-select
            v-model="alertStatus"
            placeholder="请选择发送状态"
            @change="loadAlertList"
          >
            <el-option label="全部" value=""></el-option>
            <el-option label="已发送" value="1"></el-option>
            <el-option label="未发送" value="0"></el-option>
            <el-option label="发送失败" value="2"></el-option>
          </el-select>
        </el-col>
      </el-row>
    </div>

    <!-- 告警记录列表 -->
    <el-card class="shadow-sm">
      <el-table
        :data="alertList"
        border
        stripe
        hover
        v-loading="loading"
        empty-text="暂无告警记录"
      >
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="monitorId" label="监控项ID" width="100" />
        <el-table-column prop="monitorName" label="监控项名称" min-width="150" />
        <el-table-column prop="alertSubType" label="告警类型" width="100">
          <template #default="scope">
            <el-tag v-if="scope.row.alertSubType === 1" type="danger">宕机告警</el-tag>
            <el-tag v-if="scope.row.alertSubType === 2" type="success">恢复通知</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="alertType" label="发送方式" width="100">
          <template #default="scope">
            <el-tag type="info">{{ scope.row.alertType === 1 ? '邮箱' : '钉钉' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="content" label="告警内容" min-width="300" show-overflow-tooltip />
        <el-table-column prop="status" label="发送状态" width="100">
          <template #default="scope">
            <el-tag v-if="scope.row.status === 1" type="success">已发送</el-tag>
            <el-tag v-if="scope.row.status === 0" type="warning">未发送</el-tag>
            <el-tag v-if="scope.row.status === 2" type="danger">发送失败</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="180" />
        <el-table-column prop="sendTime" label="发送时间" width="180" />
      </el-table>

      <!-- 分页 -->
      <el-pagination
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
        :current-page="page"
        :page-sizes="[10, 20, 50]"
        :page-size="size"
        layout="total, sizes, prev, pager, next, jumper"
        :total="total"
        class="mt-4 flex justify-end"
      >
      </el-pagination>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getAlertList } from '@/api/alert'

// 列表数据
const loading = ref(false)
const alertList = ref([])
const page = ref(1)
const size = ref(10)
const total = ref(0)

// 搜索条件
const searchKeyword = ref('')
const alertType = ref('')
const alertStatus = ref('')

// 加载告警记录列表
const loadAlertList = async () => {
  try {
    loading.value = true
    const params = {
      page: page.value,
      size: size.value,
      keyword: searchKeyword.value,
      alert_sub_type: alertType.value,
      status: alertStatus.value,
    }
    const res = await getAlertList(params)
    alertList.value = res.data.list || []
    total.value = res.data.total || 0
  } catch (error) {
    ElMessage.error('获取告警记录失败：' + error.msg)
    console.error('加载告警记录错误：', error)
  } finally {
    loading.value = false
  }
}

// 分页处理
const handleSizeChange = (val) => {
  size.value = val
  loadAlertList()
}
const handleCurrentChange = (val) => {
  page.value = val
  loadAlertList()
}

onMounted(() => {
  loadAlertList()
})
</script>

<style scoped>
.alert-record-container {
  padding: 0;
}

.search-bar {
  margin-bottom: 1rem;
}

.mb-4 {
  margin-bottom: 1rem;
}

.flex {
  display: flex;
}

.justify-between {
  justify-content: space-between;
}

.justify-end {
  justify-content: flex-end;
}

.items-center {
  align-items: center;
}

.mt-4 {
  margin-top: 1rem;
}
</style>