import api from './index'

export const userApi = {
  // 获取用户列表
  getUserList(params) {
    return api.get('/user/list', { params })
  },

  // 获取用户详情
  getUserDetail(userId) {
    return api.get(`/user/${userId}`)
  },

  // 创建用户
  createUser(data) {
    return api.post('/user', data)
  },

  // 更新用户信息
  updateUser(userId, data) {
    return api.put(`/user/${userId}`, data)
  },

  // 更新用户状态
  updateUserStatus(data) {
    return api.put('/user/status', data)
  },

  // 实名认证
  realName(data) {
    return api.post('/user/realname', data)
  },

  // 删除用户
  deleteUser(userId) {
    return api.delete(`/user/${userId}`)
  }
}