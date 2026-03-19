import { createRouter, createWebHistory } from 'vue-router'
import { ElMessage } from 'element-plus'

// 导入页面组件
import Login from '@/views/Login.vue'
import Register from '@/views/Register.vue'
import Layout from '@/views/Layout.vue'
import Dashboard from '@/views/Dashboard.vue'
import MonitorList from '@/views/MonitorList.vue'
import AlertConfig from '@/views/AlertConfig.vue'
import AlertRecord from '@/views/AlertRecord.vue'

// 路由守卫：验证登录状态
const requireAuth = (to, from, next) => {
    const token = localStorage.getItem('token')
    if (token) {
        next()
    } else {
        ElMessage.warning('请先登录')
        next('/login')
    }
}

const routes = [
    {
        path: '/',
        redirect: '/dashboard',
    },
    {
        path: '/login',
        name: 'Login',
        component: Login,
        meta: { title: '登录 - 监控工具' },
    },
    {
        path: '/register',
        name: 'Register',
        component: Register,
        meta: { title: '注册 - 监控工具' },
    },
    {
        path: '/',
        name: 'Layout',
        component: Layout,
        meta: { requiresAuth: true },
        beforeEnter: requireAuth,
        children: [
            {
                path: 'dashboard',
                name: 'Dashboard',
                component: Dashboard,
                meta: { title: '仪表盘 - 监控工具' },
            },
            {
                path: 'monitor/list',
                name: 'MonitorList',
                component: MonitorList,
                meta: { title: '监控项管理 - 监控工具' },
            },
            {
                path: 'alert/config',
                name: 'AlertConfig',
                component: AlertConfig,
                meta: { title: '告警配置 - 监控工具' },
            },
            {
                path: 'alert/record',
                name: 'AlertRecord',
                component: AlertRecord,
                meta: { title: '告警记录 - 监控工具' },
            },
        ],
    },
    {
        path: '/:pathMatch(.*)*',
        redirect: '/dashboard',
    },
]

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes,
    scrollBehavior() {
        return { top: 0 }
    },
})

// 路由导航：设置页面标题
router.beforeEach((to, from, next) => {
    if (to.meta.title) {
        document.title = to.meta.title
    }
    next()
})

export default router