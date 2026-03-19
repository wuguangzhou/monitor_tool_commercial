import request from '@/utils/request'

// 获取监控项列表（分页）
export function getMonitorList(params) {
    return request({
        url: '/monitor/list',
        method: 'get',
        params,
    })
}

// 创建监控项
export function createMonitor(data) {
    return request({
        url: '/monitor/create',
        method: 'post',
        data,
    })
}

// 更新监控项
export function updateMonitor(id, data) {
    return request({
        url: `/monitor/update/${id}`,
        method: 'put',
        data,
    })
}

// 删除监控项
export function deleteMonitor(id) {
    return request({
        url: `/monitor/delete/${id}`,
        method: 'delete',
    })
}

// 手动执行监控检测
export function runMonitor(id) {
    return request({
        url: `/monitor/run/${id}`,
        method: 'post',
    })
}

// 暂停监控项
export function pauseMonitor(id) {
    return request({
        url: `/monitor/pause/${id}`,
        method: 'post',
    })
}

// 恢复监控项
export function resumeMonitor(id) {
    return request({
        url: `/monitor/resume/${id}`,
        method: 'post',
    })
}