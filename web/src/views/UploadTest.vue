<template>
  <div class="upload-test">
    <el-card>
      <template #header>
        <div class="card-header">
          <h2>文件上传功能测试</h2>
          <p>测试后端文件上传API的各种功能</p>
        </div>
      </template>

      <!-- 智能上传测试 -->
      <el-card class="test-section">
        <template #header>
          <h3>1. 智能上传测试</h3>
        </template>
        <el-form :model="smartUploadForm" label-width="120px">
          <el-form-item label="选择文件">
            <el-upload
              ref="smartUploadRef"
              :auto-upload="false"
              :on-change="handleSmartFileChange"
              :show-file-list="true"
              :limit="1"
              accept="*/*"
            >
              <el-button type="primary">选择文件</el-button>
            </el-upload>
          </el-form-item>
          <el-form-item v-if="smartUploadForm.file">
            <el-progress 
              :percentage="smartUploadProgress" 
              :status="smartUploadProgress === 100 ? 'success' : ''"
              :stroke-width="20"
            />
          </el-form-item>
          <el-form-item>
            <el-button 
              type="warning" 
              @click="testSmartUpload"
              :loading="smartUploadLoading"
              :disabled="!smartUploadForm.file"
            >
              测试智能上传
            </el-button>
          </el-form-item>
          <el-form-item v-if="smartUploadResult">
            <el-alert
              :title="smartUploadResult.success ? '上传成功' : '上传失败'"
              :type="smartUploadResult.success ? 'success' : 'error'"
              :description="smartUploadResult.message"
              show-icon
            />
          </el-form-item>
        </el-form>
      </el-card>

      <!-- 测试结果展示 -->
      <el-card class="test-section">
        <template #header>
          <h3>2. 测试结果日志</h3>
        </template>
        <div class="log-container">
          <div 
            v-for="(log, index) in testLogs" 
            :key="index" 
            :class="['log-item', log.type]"
          >
            <span class="log-time">{{ log.time }}</span>
            <span class="log-message">{{ log.message }}</span>
          </div>
        </div>
        <div class="log-actions">
          <el-button @click="clearLogs" size="small">清空日志</el-button>
          <el-button @click="testWebSocketConnection" size="small" type="info">测试WebSocket连接</el-button>
        </div>
      </el-card>
    </el-card>
  </div>
</template>

<script>
import { ref, reactive } from 'vue'
import { uploadApi } from '@/api/upload'

export default {
  name: 'UploadTest',
  setup() {
    // 表单数据
    const smartUploadForm = reactive({
      file: null
    })

    // 加载状态
    const smartUploadLoading = ref(false)

    // 进度
    const smartUploadProgress = ref(0)

    // 结果
    const smartUploadResult = ref(null)

    // 日志
    const testLogs = ref([])

    // 上传组件引用
    const smartUploadRef = ref()

    // 添加日志
    const addLog = (message, type = 'info') => {
      const time = new Date().toLocaleTimeString()
      testLogs.value.unshift({
        time,
        message,
        type
      })
    }

    // 清空日志
    const clearLogs = () => {
      testLogs.value = []
    }

    // 文件选择处理
    const handleSmartFileChange = (file) => {
      smartUploadForm.file = file.raw
      smartUploadProgress.value = 0
      addLog(`选择了文件: ${file.name}`, 'info')
    }

    // 生成uploadID的函数
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

    // WebSocket连接函数
    const connectWebSocket = (uploadID, onProgress) => {
      const protocol = window.location.protocol === 'https:' ? 'wss://' : 'ws://'
      const wsUrl = `${protocol}${window.location.host}/api/upload/progress?uploadID=${uploadID}`
      
      console.log('WebSocket连接URL:', wsUrl)
      const ws = new WebSocket(wsUrl)
      
      // 添加连接超时处理
      const connectionTimeout = setTimeout(() => {
        if (ws.readyState === WebSocket.CONNECTING) {
          addLog('WebSocket连接超时，尝试重连...', 'warning')
          ws.close()
        }
      }, 5000) // 5秒超时
      
      ws.onopen = () => {
        clearTimeout(connectionTimeout)
        addLog('WebSocket连接已建立', 'info')
      }
      
      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          console.log('WebSocket进度更新:', data)
          
          if (data.uploadID === uploadID) {
            onProgress(data.progress, data.status)
            
            // 如果处理完成，主动关闭连接
            if (data.status === '处理完成') {
              addLog('上传处理完成，关闭WebSocket连接', 'info')
              setTimeout(() => {
                if (ws.readyState === WebSocket.OPEN) {
                  ws.close(1000, '上传完成')
                }
              }, 500)
            }
            // 如果上传失败，也关闭连接
            else if (data.status === '上传失败') {
              setTimeout(() => {
                if (ws.readyState === WebSocket.OPEN) {
                  ws.close(1000, '上传失败')
                }
              }, 1000)
            }
          }
        } catch (error) {
          console.error('解析WebSocket消息失败:', error)
          addLog(`解析WebSocket消息失败: ${error.message}`, 'error')
        }
      }
      
      ws.onerror = (error) => {
        console.error('WebSocket错误:', error)
        addLog(`WebSocket连接错误: ${error.type || '未知错误'}`, 'error')
      }
      
      ws.onclose = (event) => {
        clearTimeout(connectionTimeout)
        if (event.code === 1000) {
          addLog('WebSocket连接正常关闭', 'info')
        } else if (event.code === 1005) {
          addLog('WebSocket连接已关闭（上传完成）', 'info')
        } else {
          addLog(`WebSocket连接异常关闭: 代码=${event.code}, 原因=${event.reason}`, 'warning')
        }
      }
      
      return ws
    }

    // 智能上传测试
    const testSmartUpload = async () => {
      if (!smartUploadForm.file) {
        addLog('请先选择文件', 'error')
        return
      }

      smartUploadLoading.value = true
      smartUploadProgress.value = 0
      addLog('开始智能上传测试...', 'info')

      let websocket = null

      try {
        // 先生成uploadID
        const uploadID = generateUploadID()
        addLog(`生成uploadID: ${uploadID}`, 'info')
        
        // 先连接WebSocket，确保能接收到进度更新
        websocket = connectWebSocket(uploadID, (progress, status) => {
          smartUploadProgress.value = progress
          addLog(`进度更新: ${progress}% - ${status}`, 'info')
          
          // 当进度达到100%时，记录完成状态
          if (progress === 100 && status === '上传完成') {
            addLog('文件传输完成，等待服务器处理...', 'info')
          }
        })
        
        // 等待WebSocket连接建立
        await new Promise(resolve => setTimeout(resolve, 500))
        
        // 使用智能上传接口，传入uploadID
        const response = await uploadApi.uploadSmart(smartUploadForm.file, uploadID)
        console.log('智能上传响应:', response)

        // 等待一段时间确保WebSocket消息已处理
        await new Promise(resolve => setTimeout(resolve, 1000))
        
        // 检查响应
        if (response && response.data && response.data.url) {
          smartUploadResult.value = {
            success: true,
            message: `智能上传成功！文件URL: ${response.data.url}`
          }
          addLog('智能上传测试成功', 'success')
        } else if (response && response.url) {
          // 兼容不同的响应格式
          smartUploadResult.value = {
            success: true,
            message: `智能上传成功！文件URL: ${response.url}`
          }
          addLog('智能上传测试成功', 'success')
        } else {
          throw new Error(response?.msg || '上传失败：服务器返回无效响应')
        }
      } catch (error) {
        smartUploadResult.value = {
          success: false,
          message: `上传失败: ${error.message}`
        }
        addLog(`智能上传测试失败: ${error.message}`, 'error')
        
        // 如果进度已经100%但HTTP请求失败，给出特殊提示
        if (smartUploadProgress.value === 100) {
          addLog('注意：文件传输已完成，但服务器处理失败，请检查网络连接或联系管理员', 'warning')
        }
      } finally {
        smartUploadLoading.value = false
        // 关闭WebSocket连接
        if (websocket) {
          websocket.close()
        }
      }
    }

    // 测试WebSocket连接
    const testWebSocketConnection = async () => {
      const uploadID = generateUploadID()
      addLog(`测试WebSocket连接: uploadID=${uploadID}`, 'info')
      
      const ws = connectWebSocket(uploadID, (progress, status) => {
        addLog(`收到进度更新: ${progress}% - ${status}`, 'success')
      })
      
      // 等待连接建立
      await new Promise(resolve => setTimeout(resolve, 2000))
      
      if (ws.readyState === WebSocket.OPEN) {
        addLog('WebSocket连接测试成功', 'success')
      } else {
        addLog(`WebSocket连接测试失败，状态: ${ws.readyState}`, 'error')
      }
      
      // 关闭连接
      setTimeout(() => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.close()
        }
      }, 1000)
    }

    return {
      // 表单数据
      smartUploadForm,
      
      // 加载状态
      smartUploadLoading,
      
      // 进度
      smartUploadProgress,
      
      // 结果
      smartUploadResult,
      
      // 日志
      testLogs,
      
      // 组件引用
      smartUploadRef,
      
      // 方法
      handleSmartFileChange,
      testSmartUpload,
      clearLogs,
      addLog,
      testWebSocketConnection
    }
  }
}
</script>

<style scoped>
.upload-test {
  padding: 20px;
}

.card-header {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.card-header h2 {
  margin: 0;
  color: #303133;
}

.card-header p {
  margin: 0;
  color: #909399;
  font-size: 14px;
}

.test-section {
  margin-bottom: 20px;
}

.test-section h3 {
  margin: 0;
  color: #303133;
}

.log-container {
  max-height: 300px;
  overflow-y: auto;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  padding: 10px;
  background-color: #f5f7fa;
  margin-bottom: 10px;
}

.log-item {
  display: flex;
  gap: 10px;
  margin-bottom: 5px;
  font-family: 'Courier New', monospace;
  font-size: 12px;
}

.log-time {
  color: #909399;
  min-width: 80px;
}

.log-message {
  flex: 1;
}

.log-item.info .log-message {
  color: #409eff;
}

.log-item.success .log-message {
  color: #67c23a;
}

.log-item.warning .log-message {
  color: #e6a23c;
}

.log-item.error .log-message {
  color: #f56c6c;
}

.log-actions {
  display: flex;
  justify-content: space-between;
  margin-top: 10px;
}
</style>