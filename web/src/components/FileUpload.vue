<template>
  <div class="file-upload">
    <el-upload
      ref="uploadRef"
      :action="action"
      :auto-upload="false"
      :multiple="multiple"
      :limit="limit"
      :accept="accept"
      :file-list="fileList"
      :on-change="handleFileChange"
      :on-remove="handleFileRemove"
      :on-exceed="handleExceed"
      :before-upload="beforeUpload"
      :drag="drag"
      :show-file-list="true"
      :list-type="listType"
      :disabled="disabled"
    >
      <template #trigger>
        <el-button type="primary" :disabled="disabled">
          <el-icon><Upload /></el-icon>
          选择文件
        </el-button>
      </template>
      
      <template #tip>
        <div class="el-upload__tip">
          {{ tip }}
        </div>
      </template>
    </el-upload>

    <!-- 上传进度 -->
    <div v-if="uploadProgress > 0 && uploadProgress < 100" class="upload-progress">
      <el-progress 
        :percentage="uploadProgress" 
        :status="uploadStatus"
        :stroke-width="8"
      />
      <div class="progress-text">{{ progressText }}</div>
    </div>

    <!-- 上传按钮 -->
    <div v-if="fileList.length > 0" class="upload-actions">
      <el-button 
        type="success" 
        @click="startUpload"
        :loading="uploading"
        :disabled="disabled || fileList.length === 0"
      >
        <el-icon><Upload /></el-icon>
        开始上传
      </el-button>
      <el-button 
        @click="clearFiles"
        :disabled="uploading"
      >
        <el-icon><Delete /></el-icon>
        清空文件
      </el-button>
    </div>
  </div>
</template>

<script>
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { uploadApi } from '@/api/upload'

export default {
  name: 'FileUpload',
  props: {
    // 文件列表
    modelValue: {
      type: Array,
      default: () => []
    },
    // 是否多文件
    multiple: {
      type: Boolean,
      default: false
    },
    // 文件数量限制
    limit: {
      type: Number,
      default: 10
    },
    // 接受的文件类型
    accept: {
      type: String,
      default: '*'
    },
    // 最大文件大小（MB）
    maxSize: {
      type: Number,
      default: 100
    },
    // 是否拖拽上传
    drag: {
      type: Boolean,
      default: false
    },
    // 上传地址
    action: {
      type: String,
      default: ''
    },
    // 提示文字
    tip: {
      type: String,
      default: '支持的文件类型和大小限制'
    },
    // 列表类型
    listType: {
      type: String,
      default: 'text'
    },
    // 是否禁用
    disabled: {
      type: Boolean,
      default: false
    }
  },
  emits: ['update:modelValue', 'success', 'error', 'change'],
  setup(props, { emit }) {
    const uploadRef = ref()
    const fileList = ref([])
    const uploading = ref(false)
    const uploadProgress = ref(0)
    const uploadStatus = ref('')
    const progressText = ref('')

    // 监听modelValue变化
    watch(() => props.modelValue, (newVal) => {
      fileList.value = [...newVal]
    }, { immediate: true })

    // 监听fileList变化，更新modelValue
    watch(fileList, (newVal) => {
      emit('update:modelValue', newVal)
      emit('change', newVal)
    }, { deep: true })

    // 文件变化处理
    const handleFileChange = (file, fileList) => {
      // 验证文件大小
      if (file.size > props.maxSize * 1024 * 1024) {
        ElMessage.error(`文件大小不能超过 ${props.maxSize}MB`)
        return false
      }
      
      // 验证文件类型
      if (props.accept !== '*' && props.accept !== '') {
        const acceptTypes = props.accept.split(',').map(type => type.trim())
        const fileType = file.type || ''
        const fileName = file.name || ''
        
        const isValidType = acceptTypes.some(type => {
          if (type.startsWith('.')) {
            return fileName.toLowerCase().endsWith(type.toLowerCase())
          } else if (type.includes('*')) {
            const pattern = type.replace('*', '.*')
            return new RegExp(pattern).test(fileType)
          } else {
            return fileType === type
          }
        })
        
        if (!isValidType) {
          ElMessage.error(`不支持的文件类型: ${fileType}`)
          return false
        }
      }
      
      return true
    }

    // 文件移除处理
    const handleFileRemove = (file, fileList) => {
      // 可以在这里添加额外的清理逻辑
    }

    // 文件超出限制处理
    const handleExceed = (files, fileList) => {
      ElMessage.warning(`最多只能上传 ${props.limit} 个文件`)
    }

    // 上传前验证
    const beforeUpload = (file) => {
      return handleFileChange(file, fileList.value)
    }

    // 开始上传
    const startUpload = async () => {
      if (fileList.value.length === 0) {
        ElMessage.warning('请先选择文件')
        return
      }

      uploading.value = true
      uploadProgress.value = 0
      uploadStatus.value = ''
      progressText.value = '准备上传...'

      try {
        for (let i = 0; i < fileList.value.length; i++) {
          const file = fileList.value[i]
          
          if (!file.raw) {
            ElMessage.error(`文件 ${file.name} 无效`)
            continue
          }

          progressText.value = `正在上传: ${file.name}`
          
          // 根据文件大小选择上传方式
          let response
          if (file.size > 5 * 1024 * 1024) { // 5MB以上使用智能上传
            // 生成uploadID
            const uploadID = generateUploadID()
            
            // 连接WebSocket监听进度
            const ws = connectWebSocket(uploadID, (progress, status) => {
              uploadProgress.value = progress
              progressText.value = `${status}: ${progress}%`
            })
            
            // 等待WebSocket连接
            await new Promise(resolve => setTimeout(resolve, 500))
            
            // 执行上传
            response = await uploadApi.uploadSmart(file.raw, uploadID)
            
            // 关闭WebSocket
            if (ws) {
              ws.close()
            }
          } else {
            // 小文件使用普通上传
            response = await uploadApi.uploadFile(file.raw)
            uploadProgress.value = 100
            progressText.value = '上传完成'
          }

          // 处理响应
          if (response && (response.data?.url || response.url)) {
            const url = response.data?.url || response.url
            file.url = url
            file.status = 'success'
            
            ElMessage.success(`文件 ${file.name} 上传成功`)
            emit('success', response, file)
          } else {
            file.status = 'error'
            throw new Error(response?.msg || '上传失败')
          }
        }

        uploadProgress.value = 100
        uploadStatus.value = 'success'
        progressText.value = '所有文件上传完成'
        
      } catch (error) {
        uploadProgress.value = 0
        uploadStatus.value = 'exception'
        progressText.value = `上传失败: ${error.message}`
        
        ElMessage.error(`上传失败: ${error.message}`)
        emit('error', error)
      } finally {
        uploading.value = false
      }
    }

    // 清空文件
    const clearFiles = () => {
      fileList.value = []
      uploadProgress.value = 0
      uploadStatus.value = ''
      progressText.value = ''
    }

    // 生成uploadID
    const generateUploadID = () => {
      const now = new Date()
      const timestamp = now.getFullYear().toString() +
        (now.getMonth() + 1).toString().padStart(2, '0') +
        now.getDate().toString().padStart(2, '0') +
        now.getHours().toString().padStart(2, '0') +
        now.getMinutes().toString().padStart(2, '0') +
        now.getSeconds().toString().padStart(2, '0')
      const random = Math.random().toString(36).substring(2, 10)
      return `${timestamp}_${random}`
    }

    // WebSocket连接
    const connectWebSocket = (uploadID, onProgress) => {
      const protocol = window.location.protocol === 'https:' ? 'wss://' : 'ws://'
      const wsUrl = `${protocol}${window.location.host}/api/upload/progress?uploadID=${uploadID}`
      
      const ws = new WebSocket(wsUrl)
      
      ws.onopen = () => {
        console.log('WebSocket连接已建立')
      }
      
      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          if (data.uploadID === uploadID) {
            onProgress(data.progress, data.status)
          }
        } catch (error) {
          console.error('解析WebSocket消息失败:', error)
        }
      }
      
      ws.onerror = (error) => {
        console.error('WebSocket错误:', error)
      }
      
      ws.onclose = () => {
        console.log('WebSocket连接已关闭')
      }
      
      return ws
    }

    return {
      uploadRef,
      fileList,
      uploading,
      uploadProgress,
      uploadStatus,
      progressText,
      handleFileChange,
      handleFileRemove,
      handleExceed,
      beforeUpload,
      startUpload,
      clearFiles
    }
  }
}
</script>

<style scoped>
.file-upload {
  width: 100%;
}

.upload-progress {
  margin-top: 15px;
}

.progress-text {
  margin-top: 5px;
  font-size: 12px;
  color: #909399;
  text-align: center;
}

.upload-actions {
  margin-top: 15px;
  display: flex;
  gap: 10px;
}

.el-upload__tip {
  color: #909399;
  font-size: 12px;
  margin-top: 5px;
}
</style> 