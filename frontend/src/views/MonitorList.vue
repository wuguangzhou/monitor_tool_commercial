<template>
  <div class="monitor-list-container">
    <!-- 搜索和操作栏 -->
    <div class="search-bar mb-4 flex justify-between items-center">
      <el-input
        v-model="searchKeyword"
        placeholder="请输入监控项名称/URL搜索"
        prefix-icon="Search"
        style="width: 300px"
        @keyup.enter="loadMonitorList"
      ></el-input>
      <el-button type="primary" @click="openAddDialog" icon="Plus">
        添加监控项
      </el-button>
    </div>

    <!-- 监控项列表 -->
    <el-card class="shadow-sm">
      <el-table
        :data="monitorList"
        border
        stripe
        hover
        v-loading="loading"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="监控项名称" min-width="150" />
        <el-table-column prop="url" label="监控URL" min-width="200" show-overflow-tooltip />
        <el-table-column prop="monitorType" label="类型" width="150">
          <template #default="scope">
            <el-tag type="info">{{ scope.row.monitorType === 1 ? 'HTTP/HTTPS' : '未知' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="frequency" label="频率(秒)" width="100" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag v-if="Number(scope.row.status) === 1" type="success">正常</el-tag>
            <el-tag v-if="Number(scope.row.status) === 2" type="danger">宕机</el-tag>
            <el-tag v-if="Number(scope.row.status) === 3" type="warning">暂停</el-tag>
            <el-tag v-if="Number(scope.row.status) === 0" type="info">初始化</el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="createAt" label="创建时间" width="200" />
        <el-table-column prop="updateAt" label="更新时间" width="200" />

        <el-table-column label="操作" width="350">
          <template #default="scope">
            <el-button
              size="small"
              type="info"
              icon="Edit"
              @click="openEditDialog(scope.row)"
            >
              编辑
            </el-button>
            <el-button
              size="small"
              type="primary"
              icon="Refresh"
              @click="handleRunMonitor(scope.row)"
            >
              手动检测
            </el-button>

            <!-- 暂停按钮：增加NaN兜底 + 严格判断有效状态 -->
            <el-button
              size="small"
              type="warning"
              :icon="VideoPause"
              @click="handlePauseMonitor(scope.row)"
              v-if="!isNaN(Number(scope.row.status)) && [0,1,2].includes(Number(scope.row.status))"
            >
              暂停
            </el-button>

            <!-- 恢复按钮：增加NaN兜底 + 严格等于3 -->
            <el-button
              size="small"
              type="success"
              icon="CaretRight"
              @click="handleResumeMonitor(scope.row)"
              v-if="!isNaN(Number(scope.row.status)) && Number(scope.row.status) === 3"
            >
              恢复
            </el-button>

            <el-button
              size="small"
              type="danger"
              icon="Delete"
              @click="handleDeleteMonitor(scope.row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
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

    <!-- 添加/编辑监控项弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑监控项' : '添加监控项'"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form
        :model="monitorForm"
        :rules="monitorRules"
        ref="monitorFormRef"
        label-width="100px"
      >
        <el-form-item label="监控项名称" prop="name">
          <el-input
            v-model="monitorForm.name"
            placeholder="请输入监控项名称（如：百度首页）"
          ></el-input>
        </el-form-item>
        <el-form-item label="监控URL" prop="url">
          <el-input
            v-model="monitorForm.url"
            placeholder="请输入HTTP/HTTPS地址（如：https://www.baidu.com）"
          ></el-input>
        </el-form-item>
        <el-form-item label="监控类型" prop="monitorType">
          <el-select v-model="monitorForm.monitorType" placeholder="请选择监控类型">
            <el-option label="HTTP/HTTPS" :value="1"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="监控频率(秒)" prop="frequency">
          <el-input
            v-model.number="monitorForm.frequency"
            type="number"
            placeholder="请输入监控频率（最小10秒）"
            min="10"
          ></el-input>
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input
            v-model="monitorForm.remark"
            type="textarea"
            placeholder="请输入备注（可选）"
            rows="3"
          ></el-input>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          @click="submitMonitorForm"
          :loading="submitLoading"
        >
          确认
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { VideoPause } from '@element-plus/icons-vue'
import {
  getMonitorList,
  createMonitor,
  updateMonitor,
  deleteMonitor as apiDeleteMonitor,
  runMonitor as apiRunMonitor,
  pauseMonitor as apiPauseMonitor,
  resumeMonitor as apiResumeMonitor,
} from '@/api/monitor'

// 列表数据
const loading = ref(false)
const monitorList = ref([])
const page = ref(1)
const size = ref(10)
const total = ref(0)
const searchKeyword = ref('')

// 弹窗相关
const dialogVisible = ref(false)
const isEdit = ref(false)
const monitorFormRef = ref()
const submitLoading = ref(false)

// 表单数据
const monitorForm = reactive({
  id: '',
  name: '',
  url: '',
  monitorType: 1,
  frequency: 60,
  remark: '',
})

// 表单校验规则
const monitorRules = reactive({
  name: [
    { required: true, message: '请输入监控项名称', trigger: 'blur' },
    { min: 2, max: 50, message: '名称长度在2-50个字符之间', trigger: 'blur' },
  ],
  url: [
    { required: true, message: '请输入监控URL', trigger: 'blur' },
    {
      pattern: /^https?:\/\/.+/,
      message: '请输入有效的HTTP/HTTPS地址',
      trigger: 'blur',
    },
  ],
  frequency: [
    { required: true, message: '请输入监控频率', trigger: 'blur' },
    { type: 'number', min: 10, message: '监控频率最小为10秒', trigger: 'blur' },
  ],
})

// 加载监控项列表
const loadMonitorList = async () => {
  try {
    loading.value = true
    const params = {
      page: page.value,
      size: size.value,
      keyword: searchKeyword.value,
    }
    const res = await getMonitorList(params)
    console.log('监控项列表数据:', res.data.list)
    monitorList.value = res.data.list || []
    total.value = res.data.total || 0
  } catch (error) {
    ElMessage.error('获取监控项列表失败：' + (error.msg || '服务器错误'))
    console.error('加载列表错误：', error)
  } finally {
    loading.value = false
  }
}

// 分页处理
const handleSizeChange = (val) => {
  size.value = val
  loadMonitorList()
}
const handleCurrentChange = (val) => {
  page.value = val
  loadMonitorList()
}

// 打开添加弹窗
const openAddDialog = () => {
  resetForm()
  isEdit.value = false
  dialogVisible.value = true
}

// 打开编辑弹窗
const openEditDialog = (row) => {
  monitorForm.id = row.id
  monitorForm.name = row.name
  monitorForm.url = row.url
  monitorForm.monitorType = row.monitorType
  monitorForm.frequency = Number(row.frequency)
  monitorForm.remark = row.remark || ''
  isEdit.value = true
  dialogVisible.value = true
}

// 重置表单
const resetForm = () => {
  monitorForm.id = ''
  monitorForm.name = ''
  monitorForm.url = ''
  monitorForm.monitorType = 1
  monitorForm.frequency = 60
  monitorForm.remark = ''
  monitorFormRef.value?.resetFields()
}

// 提交表单（添加/编辑）
const submitMonitorForm = async () => {
  try {
    await monitorFormRef.value.validate()
    submitLoading.value = true

    const submitData = {
      ...monitorForm,
      frequency: Number(monitorForm.frequency)
    }

    if (isEdit.value) {
      await updateMonitor(submitData.id, submitData)
      ElMessage.success('监控项编辑成功')
    } else {
      await createMonitor(submitData)
      ElMessage.success('监控项添加成功')
    }
    dialogVisible.value = false
    loadMonitorList()
  } catch (error) {
    ElMessage.error('操作失败：' + (error.msg || error.message || '未知错误'))
    console.error('提交表单错误：', error)
  } finally {
    submitLoading.value = false
  }
}

// 手动检测
const handleRunMonitor = async (row) => {
  try {
    await ElMessageBox.confirm(`确定要手动检测「${row.name}」吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
    })
    await apiRunMonitor(row.id)
    ElMessage.success('手动检测已触发，结果将实时更新')
    setTimeout(() => {
      loadMonitorList()
    }, 1000)
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('手动检测失败：' + (error.msg || '未知错误'))
    }
  }
}

// 暂停监控
const handlePauseMonitor = async (row) => {
  const monitorId = Number(row.id)
  if (!monitorId) {
    ElMessage.error('监控项ID无效')
    return
  }
  try {
    await ElMessageBox.confirm(`确定要暂停「${row.name}」监控吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await apiPauseMonitor(monitorId)
    ElMessage.success('监控项已暂停')
    loadMonitorList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('暂停失败：' + (error.msg || '未知错误'))
    }
  }
}

// 恢复监控
const handleResumeMonitor = async (row) => {
  const monitorId = Number(row.id)
  if (!monitorId) {
    ElMessage.error('监控项ID无效')
    return
  }
  try {
    await ElMessageBox.confirm(`确定要恢复「${row.name}」监控吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await apiResumeMonitor(monitorId)
    ElMessage.success('监控项已恢复')
    loadMonitorList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('恢复失败：' + (error.msg || '未知错误'))
    }
  }
}

// 删除监控项
const handleDeleteMonitor = async (row) => {
  try {
    await ElMessageBox.confirm(`确定要删除「${row.name}」吗？此操作不可恢复！`, '警告', {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'danger',
    })
    await apiDeleteMonitor(row.id)
    ElMessage.success('监控项已删除')
    loadMonitorList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败：' + (error.msg || '未知错误'))
    }
  }
}

// 多选框选择
const handleSelectionChange = (val) => {
  console.log('选中的监控项：', val)
}

onMounted(() => {
  loadMonitorList()
})
</script>

<style scoped>
.monitor-list-container {
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