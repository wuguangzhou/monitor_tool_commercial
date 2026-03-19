import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'

// 创建Axios实例
const service = axios.create({
    baseURL: '/api',
    timeout: 10000, // 请求超时时间
    headers: {
        'Content-Type': 'application/json;charset=utf-8',
    },
})

// 请求拦截器：添加token
service.interceptors.request.use(
    (config) => {
        const token = localStorage.getItem('token')
        if (token) {
            config.headers.Authorization = `Bearer ${token}`
        }
        return config
    },
    (error) => {
        console.error('请求错误：', error)
        return Promise.reject(error)
    }
)

// 响应拦截器：统一处理响应
service.interceptors.response.use(
    (response) => {
        const res = response.data
        // 业务错误处理
        if (res.code !== 200) {
            ElMessage.error(res.msg || '请求失败')
            // token过期/未登录
            if (res.code === 401) {
                ElMessageBox.confirm('登录状态已过期，请重新登录', '提示', {
                    confirmButtonText: '重新登录',
                    cancelButtonText: '取消',
                    type: 'warning',
                }).then(() => {
                    localStorage.removeItem('token')
                    window.location.href = '/login'
                })
            }
            return Promise.reject(res)
        }
        return res
    },
    (error) => {
        console.error('响应错误：', error)
        ElMessage.error(error.message || '服务器错误')
        return Promise.reject(error)
    }
)

export default service