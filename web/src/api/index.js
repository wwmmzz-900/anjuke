import axios from 'axios'
import { ElMessage } from 'element-plus'

// 创建 axios 实例
const api = axios.create({
  baseURL: '/api',
  timeout: 30000 // 增加到30秒，上传接口会单独设置更长的超时
})

// 请求拦截器
api.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  response => {
    // 文件上传等特殊接口可能直接返回数据
    if (response?.config?.url && response.config.url.includes('upload')) {
      return response.data || {}
    }
    
    // 安全地处理响应数据
    if (response?.data && typeof response.data === 'object') {
      const { code, msg, data } = response.data
      if (code === 0 || code === undefined) {
        return data !== undefined ? data : response.data
      } else {
        ElMessage.error(msg || '请求失败')
        return Promise.reject(new Error(msg || '请求失败'))
      }
    }
    
    // 如果响应数据格式不符合预期，直接返回
    return response?.data || {}
  },
  error => {
    // 更详细的错误处理
    let errorMessage = '网络错误'
    if (error.response) {
      // 服务器响应了错误状态码
      errorMessage = error.response.data?.msg || error.response.statusText || `HTTP ${error.response.status}`
    } else if (error.request) {
      // 请求已发出但没有收到响应
      errorMessage = '服务器无响应'
    } else {
      // 其他错误
      errorMessage = error.message || '未知错误'
    }
    
    ElMessage.error(errorMessage)
    return Promise.reject(error)
  }
)

export default api