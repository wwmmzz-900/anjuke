import api from './index'

export const uploadApi = {
  // 智能上传
  async smartUpload(file, onProgress) {
    // 使用FileReader读取文件内容为二进制数据
    const fileData = await new Promise((resolve, reject) => {
      const reader = new FileReader()
      reader.onload = () => resolve(new Uint8Array(reader.result))
      reader.onerror = reject
      reader.readAsArrayBuffer(file)
    })

    // 构造与后端SimpleUploadRequest格式匹配的对象
    const uploadData = {
      filename: file.name,
      content_type: file.type,
      file_data: Array.from(fileData) // 转换为普通数组以便JSON序列化
    }

    return api.post('/api/upload/smart', uploadData, {
      headers: {
        'Content-Type': 'application/json'
      },
      timeout: 300000, // 5分钟超时
      onUploadProgress: onProgress
    })
  },

  // 获取文件列表
  async getFileList(params = {}) {
    try {
      const response = await api.get('/user/fileList', { params })
      // 确保返回正确的数据结构
      return {
        list: response?.list || response?.data || [],
        total: response?.total || response?.count || 0
      }
    } catch (error) {
      console.warn('文件列表接口调用失败:', error.message)
      // 返回空数据结构而不是抛出错误
      return { list: [], total: 0 }
    }
  },

  // 获取上传统计
  async getUploadStats() {
    try {
      const response = await api.get('/user/uploadStats')
      // 确保返回正确的数据结构
      return {
        totalUploads: response?.totalUploads || response?.total || 0,
        successUploads: response?.successUploads || response?.success || 0,
        totalSize: response?.totalSize || response?.size || 0,
        todayUploads: response?.todayUploads || response?.today || 0
      }
    } catch (error) {
      console.warn('统计接口调用失败:', error.message)
      // 返回空统计数据而不是抛出错误
      return { totalUploads: 0, successUploads: 0, totalSize: 0, todayUploads: 0 }
    }
  },

  // 删除文件
  deleteFile(objectName) {
    return api.delete(`/file/${objectName}`)
  },

  // 获取文件详情
  getFileDetail(fileId) {
    return api.get(`/file/${fileId}`)
  },

  // 直接上传文件
  uploadFile(file, onProgress) {
    const formData = new FormData()
    formData.append('file', file)

    return api.post('/upload/file', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      },
      onUploadProgress: onProgress
    })
  },

  // 清理上传
  cleanupUploads() {
    return api.post('/upload/cleanup')
  }
}