import api from './index'

// 文件上传API
export const uploadApi = {
  // 智能上传接口
  uploadSmart(file, uploadID = null) {
    const formData = new FormData()
    formData.append('file', file)
    
    // 如果提供了uploadID，添加到请求中
    if (uploadID) {
      formData.append('uploadID', uploadID)
    }
    
    return api.post('/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      },
      timeout: 600000 // 10分钟超时，适合大文件上传
    })
  },

  // 获取上传统计
  getUploadStats() {
    return api.get('/user/uploadStats')
  },

  // 获取文件列表
  getFileList(params = {}) {
    return api.get('/user/getFileList', { params })
  },

  // 删除文件
  deleteFile(objectName) {
    return api.post('/user/deleteFile', { objectName })
  },

  // 清理上传
  cleanupUploads() {
    return api.post('/user/cleanupUploads')
  }
}

export default uploadApi 