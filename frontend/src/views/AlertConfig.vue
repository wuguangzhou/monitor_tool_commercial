<template>
  <div class="alert-config-container">
    <el-card class="shadow-sm">
      <template #header>
        <span class="font-bold">告警配置</span>
      </template>
      <el-form
        :model="alertConfigForm"
        :rules="alertConfigRules"
        ref="alertConfigFormRef"
        label-width="120px"
        class="mt-4"
      >
        <el-form-item label="告警邮箱" prop="email">
          <el-input
            v-model="alertConfigForm.email"
            placeholder="请输入接收告警的邮箱"
          ></el-input>
          <div class="text-gray-500 text-sm mt-1">
            用于接收邮件告警/恢复通知
          </div>
        </el-form-item>
        <el-form-item label="默认告警方式" prop="alertType">
          <el-radio-group v-model="alertConfigForm.alertType">
            <el-radio label="1">邮箱告警</el-radio>
            <el-radio label="2">钉钉告警</el-radio>
          </el-radio-group>
        </el-form-item>

        <!-- 钉钉配置：仅当选择钉钉告警时展示 -->
        <template v-if="Number(alertConfigForm.alertType) === 2">
          <el-form-item label="钉钉Webhook" prop="dingtalkWebhook">
            <el-input
              v-model="alertConfigForm.dingtalkWebhook"
              placeholder="https://oapi.dingtalk.com/robot/send?access_token=xxx"
            />
            <div class="text-gray-500 text-sm mt-1">
              钉钉群机器人 Webhook 地址
            </div>
          </el-form-item>
          <el-form-item label="钉钉加签Secret" prop="dingtalkSecret">
            <el-input
              v-model="alertConfigForm.dingtalkSecret"
              placeholder="可选：开启加签时填写（SEC...）"
              show-password
            />
          </el-form-item>
          <el-form-item label="钉钉关键词" prop="dingtalkKeyword">
            <el-input
              v-model="alertConfigForm.dingtalkKeyword"
              placeholder="可选：群安全设置要求的关键词"
            />
          </el-form-item>
        </template>
        <el-form-item label="告警开关" prop="isEnabled">
          <el-switch
            v-model="alertConfigForm.isEnabled"
            active-value="1"
            inactive-value="0"
            active-text="开启"
            inactive-text="关闭"
          ></el-switch>
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            @click="submitAlertConfig"
            :loading="submitLoading"
          >
            保存配置
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getAlertConfig, updateAlertConfig } from '@/api/alert'

// 表单相关
const alertConfigFormRef = ref()
const submitLoading = ref(false)

// 表单数据
const alertConfigForm = reactive({
  email: '',
  alertType: '1',
  isEnabled: '1',
  dingtalkWebhook: '',
  dingtalkSecret: '',
  dingtalkKeyword: '',
})

// 表单校验规则
const alertConfigRules = reactive({
  email: [
    { required: true, message: '请输入告警邮箱', trigger: 'blur' },
    {
      pattern: /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/,
      message: '请输入有效的邮箱地址',
      trigger: 'blur',
    },
  ],
  alertType: [
    { required: true, message: '请选择告警方式', trigger: 'change' },
  ],
  isEnabled: [
    { required: true, message: '请选择告警开关状态', trigger: 'change' },
  ],
  dingtalkWebhook: [
    {
      validator: (rule, value, callback) => {
        if (Number(alertConfigForm.alertType) !== 2) return callback()
        if (!value || !value.trim()) return callback(new Error('请输入钉钉Webhook'))
        if (!/^https?:\/\/.+/.test(value.trim())) return callback(new Error('请输入有效的Webhook URL'))
        callback()
      },
      trigger: 'blur',
    },
  ],
})

// 获取告警配置
const getAlertConfigData = async () => {
  try {
    const res = await getAlertConfig()
    if (res.data) {
      alertConfigForm.email = res.data.email || ''
      // 后端返回字段多为 snake_case（alert_type/is_enabled），旧版本前端可能用 camelCase（alertType/isEnabled）
      const alertTypeVal = res.data.alert_type ?? res.data.alertType ?? 1
      const isEnabledVal = res.data.is_enabled ?? res.data.isEnabled ?? 1

      alertConfigForm.alertType = String(alertTypeVal || 1)
      alertConfigForm.isEnabled = String(isEnabledVal || 1)

      // 后端可能因字段命名/历史版本导致返回 key 存在差异，这里做兼容兜底：
      // 优先使用期望 key（dingtalk_*），如果不存在再尝试其它常见变体。
      const dtWebhook =
        res.data.dingtalk_webhook ??
        res.data.ding_talk_webhook ??
        res.data.dingtalkWebhook ??
        ''

      const dtSecret =
        res.data.dingtalk_secret ??
        res.data.ding_talk_secret ??
        res.data.dingtalkSecret ??
        ''

      const dtKeyword =
        res.data.dingtalk_keyword ??
        res.data.ding_talk_keyword ??
        res.data.dingtalkKeyword ??
        ''

      alertConfigForm.dingtalkWebhook = dtWebhook || ''
      alertConfigForm.dingtalkSecret = dtSecret || ''
      alertConfigForm.dingtalkKeyword = dtKeyword || ''
    }
  } catch (error) {
    // 无配置时不报错
    console.log('用户暂无告警配置：', error)
  }
}

// 提交告警配置
const submitAlertConfig = async () => {
  try {
    await alertConfigFormRef.value.validate()
    submitLoading.value = true
    // 转换为数字类型
    const submitData = {
      email: alertConfigForm.email,
      alert_type: parseInt(alertConfigForm.alertType),
      is_enabled: parseInt(alertConfigForm.isEnabled),
      dingtalk_webhook: alertConfigForm.dingtalkWebhook,
      dingtalk_secret: alertConfigForm.dingtalkSecret,
      dingtalk_keyword: alertConfigForm.dingtalkKeyword,
    }
    await updateAlertConfig(submitData)
    ElMessage.success('告警配置保存成功')
  } catch (error) {
    ElMessage.error('保存失败：' + error.msg)
    console.error('提交配置错误：', error)
  } finally {
    submitLoading.value = false
  }
}

onMounted(() => {
  getAlertConfigData()
})
</script>

<style scoped>
.alert-config-container {
  padding: 0;
}

.mt-4 {
  margin-top: 1rem;
}

.text-sm {
  font-size: 12px;
}

.text-gray-500 {
  color: #909399;
}
</style>