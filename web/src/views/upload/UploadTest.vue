<template>
  <div class="upload-test">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>文件上传测试</span>
        </div>
      </template>

      <el-alert
        title="测试说明"
        type="info"
        :closable="false"
        style="margin-bottom: 20px;"
      >
        <p>此页面用于测试文件上传功能，确保与后端 MinIO 服务正常连接。</p>
        <p>后端接口: <code>/user/uploadFile</code> (单文件) 和 <code>/user/uploadFiles</code> (多文件)</p>
      </el-alert>

      <!-- 简单上传测试 -->
      <el-row :gutter="20">
        <el-col :span="12">
          <el-card>
            <template #header>
              <span>简单上传测试</span>
            </template>
            
            <el-upload
              ref="simpleUploadRef"
              action="/api/upload/smart"
              :headers="uploadHeaders"
              :on-success="handleSimpleSuccess"
              :on-error="handleSimpleError"
              :on-progress="handleSimpleProgress"
              :before-upload="beforeSimpleUpload"
              :show-file-list="true"
            >
              <el-button type="primary">
                <el-icon><Upload /></el-icon>
                选择文件
              </el-button>
              <template #tip>
                <div class="el-upload__tip">
                  支持任意格式文件，最大100MB
                </div>
              </template>
            </el-upload>

            <!-- 上传结果 -->
            <div v-if="simpleUploadResult" class="upload-result">
              <h4>上传结果:</h4>
              <el-descriptions :column="1" border>
                <el-descriptions-item label="状态">
                  <el-tag :type="simpleUploadResult.success ? 'success' : 'danger'">
                    {{ simpleUploadResult.success ? '成功' : '失败' }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="文件名">
                  {{ simpleUploadResult.fileName }}
                </el-descriptions-item>
                <el-descriptions-item label="文件大小">
                  {{ formatFileSize(simpleUploadResult.fileSize) }}
                </el-descriptions-item>
                <el-descriptions-item label="上传时间">
                  {{ simpleUploadResult.uploadTime }}
                </el-descriptions-item>
                <el-descriptions-item label="文件URL" v-if="simpleUploadResult.url">
                  <el-link :href="simpleUploadResult.url" target="_blank">
                    {{ simpleUploadResult.url }}
                  </el-link>
                </el-descriptions-item>
                <el-descriptions-item label="对象名" v-if="simpleUploadResult.objectName">
                  {{ simpleUploadResult.objectName }}
                </el-descriptions-item>
                <el-descriptions-item label="错误信息" v-if="simpleUploadResult.error">
                  <el-text type="danger">{{ simpleUploadResult.error }}</el-text>
                </el-descriptions-item>
              </el-descriptions>
            </div>
          </el-card>
        </el-col>

        <el-col :span="12">
          <el-card>
            <template #header>
              <span>拖拽上传测试</span>
            </template>
            
            <el-upload
              ref="dragUploadRef"
              action="/api/upload/smart"
              :headers="uploadHeaders"
              :on-success="handleDragSuccess"
              :on-error="handleDragError"
              :on-progress="handleDragProgress"
              :before-upload="beforeDragUpload"
              drag
              multiple
            >
              <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
              <div class="el-upload__text">
                将文件拖到此处，或<em>点击上传</em>
              </div>
              <template #tip>
                <div class="el-upload__tip">
                  支持多文件拖拽上传
                </div>
              </template>
            </el-upload>

            <!-- 拖拽上传结果 -->
            <div v-if="dragUploadResults.length > 0" class="upload-results">
              <h4>上传结果:</h4>
              <el-table :data="dragUploadResults" style="width: 100%">
                <el-table-column prop="fileName" label="文件名" />
                <el-table-column prop="status" label="状态" width="80">
                  <template #default="scope">
                    <el-tag :type="scope.row.success ? 'success' : 'danger'" size="small">
                      {{ scope.row.success ? '成功' : '失败' }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="fileSize" label="大小" width="100">
                  <template #default="scope">
                    {{ formatFileSize(scope.row.fileSize) }}
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <!-- API 测试 -->
      <el-card style="margin-top: 20px;">
        <template #header>
          <span>API 直接测试</span>
        </template>

        <el-row :gutter="20">
          <el-col :span="8">
            <div class="api-test-section">
              <h4>选择测试文件:</h4>
              <input
                ref="fileInputRef"
                type="file"
                @change="handleFileSelect"
                style="margin-bottom: 10px;"
              />
              <div v-if="selectedFile">
                <p><strong>文件名:</strong> {{ selectedFile.name }}</p>
                <p><strong>大小:</strong> {{ formatFileSize(selectedFile.size) }}</p>
                <p><strong>类型:</strong> {{ selectedFile.type }}</p>
              </div>
            </div>
          </el-col>

          <el-col :span="8">
            <div class="api-test-section">
              <h4>测试操作:</h4>
              <el-button 
                type="primary" 
                @click="testDirectUpload" 
                :loading="directUploading"
                :disabled="!selectedFile"
              >
                直接调用上传API
              </el-button>
              <el-button @click="testCleanup" :loading="cleanupLoading">
                测试清理接口
              </el-button>
              <el-button @click="testStats" :loading="statsLoading">
                测试统计接口
              </el-button>
              <el-button @click="testFileList" :loading="fileListLoading">
                测试文件列表
              </el-button>
            </div>
          </el-col>

          <el-col :span="8">
            <div class="api-test-section">
              <h4>上传进度:</h4>
              <div v-if="directUploadProgress.show">
                <el-progress 
                  :percentage="directUploadProgress.percent" 
                  :status="directUploadProgress.status"
                />
                <p>{{ directUploadProgress.text }}</p>
              </div>
            </div>
          </el-col>
        </el-row>

        <!-- API 测试结果 -->
        <div v-if="apiTestResult" class="api-test-result">
          <h4>API 测试结果:</h4>
          <el-card>
            <pre>{{ JSON.stringify(apiTestResult, null, 2) }}</pre>
          </el-card>
        </div>
      </el-card>
    </el-card>
  </div>
</template>

<script>
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { uploadApi } from '@/api/upload'

export default {
  name: 'UploadTest',
  setup() {
    const simpleUploadRef = ref()
    const dragUploadRef = ref()
    const fileInputRef = ref()
    
    const simpleUploadResult = ref(null)
    const dragUploadResults = ref([])
    const selectedFile = ref(null)
    const apiTestResult = ref(null)
    
    const directUploading = ref(false)
    const cleanupLoading = ref(false)
    const statsLoading = ref(false)
    const fileListLoading = ref(false)
    
    const directUploadProgress = ref({
      show: false,
      percent: 0,
      status: 'active',
      text: ''
    })

    // 计算属性
    const uploadHeaders = computed(() => {
      const token = localStorage.getItem('token')
      return token ? { Authorization: `Bearer ${token}` } : {}
    })

    // 简单上传处理
    const beforeSimpleUpload = (file) => {
      console.log('准备上传文件:', file.name)
      return true
    }

    const handleSimpleSuccess = (response, file) => {
      console.log('简单上传成功:', response, file)
      simpleUploadResult.value = {
        success: true,
        fileName: file?.name || '未知文件',
        fileSize: file?.size || 0,
        uploadTime: new Date().toLocaleString(),
        url: response?.data?.url || response?.url || '#',
        objectName: response?.data?.objectName || response?.objectName || ''
      }
      ElMessage.success('文件上传成功')
    }

    const handleSimpleError = (error, file) => {
      console.error('简单上传失败:', error, file)
      simpleUploadResult.value = {
        success: false,
        fileName: file.name,
        fileSize: file.size,
        uploadTime: new Date().toLocaleString(),
        error: error.message || '上传失败'
      }
      ElMessage.error('文件上传失败')
    }

    const handleSimpleProgress = (event, file) => {
      console.log('简单上传进度:', event.percent, file.name)
    }

    // 拖拽上传处理
    const beforeDragUpload = (file) => {
      console.log('准备拖拽上传文件:', file.name)
      return true
    }

    const handleDragSuccess = (response, file) => {
      console.log('拖拽上传成功:', response, file)
      dragUploadResults.value.push({
        success: true,
        fileName: file.name,
        fileSize: file.size,
        uploadTime: new Date().toLocaleString(),
        url: response.data?.url || response.url,
        objectName: response.data?.objectName || response.objectName
      })
      ElMessage.success(`${file.name} 上传成功`)
    }

    const handleDragError = (error, file) => {
      console.error('拖拽上传失败:', error, file)
      dragUploadResults.value.push({
        success: false,
        fileName: file.name,
        fileSize: file.size,
        uploadTime: new Date().toLocaleString(),
        error: error.message || '上传失败'
      })
      ElMessage.error(`${file.name} 上传失败`)
    }

    const handleDragProgress = (event, file) => {
      console.log('拖拽上传进度:', event.percent, file.name)
    }

    // 文件选择
    const handleFileSelect = (event) => {
      const file = event.target.files[0]
      if (file) {
        selectedFile.value = file
        console.log('选择文件:', file)
      }
    }

    // 直接API测试
    const testDirectUpload = async () => {
      if (!selectedFile.value) {
        ElMessage.warning('请先选择文件')
        return
      }

      try {
        directUploading.value = true
        directUploadProgress.value = {
          show: true,
          percent: 0,
          status: 'active',
          text: '开始上传...'
        }

        const result = await uploadApi.uploadFile(selectedFile.value, (progress) => {
          directUploadProgress.value.percent = progress.percent
          directUploadProgress.value.text = `上传中... ${progress.percent}%`
        })

        directUploadProgress.value.status = 'success'
        directUploadProgress.value.text = '上传完成'

        apiTestResult.value = {
          type: 'upload',
          success: true,
          data: result,
          timestamp: new Date().toISOString()
        }

        ElMessage.success('API 上传测试成功')

      } catch (error) {
        directUploadProgress.value.status = 'exception'
        directUploadProgress.value.text = '上传失败'

        apiTestResult.value = {
          type: 'upload',
          success: false,
          error: error.message,
          timestamp: new Date().toISOString()
        }

        ElMessage.error('API 上传测试失败: ' + error.message)
      } finally {
        directUploading.value = false
      }
    }

    // 测试清理接口
    const testCleanup = async () => {
      try {
        cleanupLoading.value = true
        const result = await uploadApi.cleanupUploads()
        
        apiTestResult.value = {
          type: 'cleanup',
          success: true,
          data: result,
          timestamp: new Date().toISOString()
        }

        ElMessage.success('清理接口测试成功')
      } catch (error) {
        apiTestResult.value = {
          type: 'cleanup',
          success: false,
          error: error.message,
          timestamp: new Date().toISOString()
        }

        ElMessage.error('清理接口测试失败: ' + error.message)
      } finally {
        cleanupLoading.value = false
      }
    }

    // 测试统计接口
    const testStats = async () => {
      try {
        statsLoading.value = true
        const result = await uploadApi.getUploadStats()
        
        apiTestResult.value = {
          type: 'stats',
          success: true,
          data: result || {},
          timestamp: new Date().toISOString()
        }

        ElMessage.success('统计接口测试成功')
      } catch (error) {
        apiTestResult.value = {
          type: 'stats',
          success: false,
          error: error.message || '未知错误',
          timestamp: new Date().toISOString()
        }

        ElMessage.error('统计接口测试失败: ' + (error.message || '未知错误'))
      } finally {
        statsLoading.value = false
      }
    }

    // 测试文件列表接口
    const testFileList = async () => {
      try {
        fileListLoading.value = true
        const result = await uploadApi.getFileList({
          page: 1,
          pageSize: 10,
          keyword: ''
        })
        
        apiTestResult.value = {
          type: 'fileList',
          success: true,
          data: result || {},
          timestamp: new Date().toISOString()
        }

        ElMessage.success('文件列表接口测试成功')
      } catch (error) {
        apiTestResult.value = {
          type: 'fileList',
          success: false,
          error: error.message || '未知错误',
          timestamp: new Date().toISOString()
        }

        ElMessage.error('文件列表接口测试失败: ' + (error.message || '未知错误'))
      } finally {
        fileListLoading.value = false
      }
    }

    // 格式化文件大小
    const formatFileSize = (size) => {
      if (!size) return '0 B'
      const units = ['B', 'KB', 'MB', 'GB']
      let index = 0
      while (size >= 1024 && index < units.length - 1) {
        size /= 1024
        index++
      }
      return `${size.toFixed(2)} ${units[index]}`
    }

    return {
      simpleUploadRef,
      dragUploadRef,
      fileInputRef,
      simpleUploadResult,
      dragUploadResults,
      selectedFile,
      apiTestResult,
      directUploading,
      cleanupLoading,
      statsLoading,
      fileListLoading,
      directUploadProgress,
      uploadHeaders,
      beforeSimpleUpload,
      handleSimpleSuccess,
      handleSimpleError,
      handleSimpleProgress,
      beforeDragUpload,
      handleDragSuccess,
      handleDragError,
      handleDragProgress,
      handleFileSelect,
      testDirectUpload,
      testCleanup,
      testStats,
      testFileList,
      formatFileSize
    }
  }
}
</script>

<style scoped>
.upload-test {
  padding: 20px;
}

.card-header {
  font-weight: bold;
}

.upload-result,
.upload-results {
  margin-top: 20px;
  padding: 15px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.upload-result h4,
.upload-results h4 {
  margin: 0 0 15px 0;
  color: #303133;
}

.api-test-section {
  padding: 15px;
  background-color: #f9f9f9;
  border-radius: 4px;
  height: 200px;
}

.api-test-section h4 {
  margin: 0 0 15px 0;
  color: #303133;
}

.api-test-result {
  margin-top: 20px;
}

.api-test-result h4 {
  margin: 0 0 15px 0;
  color: #303133;
}

.api-test-result pre {
  background-color: #f5f5f5;
  padding: 15px;
  border-radius: 4px;
  overflow-x: auto;
  font-size: 12px;
  line-height: 1.4;
}
</style>