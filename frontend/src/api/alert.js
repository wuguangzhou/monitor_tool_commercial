import request from '@/utils/request'

// 获取告警配置
export function getAlertConfig() {
    return request({
        url: '/alert/config',
        method: 'get',
    })
}

// 更新告警配置
export function updateAlertConfig(data) {
    return request({
        url: '/alert/config/update',
        method: 'post',
        data,
    })
}

// 获取告警记录列表（分页）
export function getAlertList(params) {
    return request({
        url: '/alert/list',
        method: 'get',
        params,
    })
}