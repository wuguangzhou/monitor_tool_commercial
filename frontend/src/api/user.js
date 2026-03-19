import request from '@/utils/request'

// 登录
export function login(data) {
    return request({
        url: '/user/login',
        method: 'post',
        data,
    })
}

// 注册
export function register(data) {
    return request({
        url: '/user/register',
        method: 'post',
        data,
    })
}

// 获取用户信息
export function getUserInfo() {
    return request({
        url: '/user/info',
        method: 'get',
    })
}

// 上传头像
export function uploadAvatar(formData) {
    return request({
        url: '/user/avatar',
        method: 'post',
        data: formData,
        headers: {
            'Content-Type': 'multipart/form-data',
        },
    })
}

// 更新个人信息（昵称）
export function updateProfile(data) {
    return request({
        url: '/user/profile',
        method: 'post',
        data,
    })
}