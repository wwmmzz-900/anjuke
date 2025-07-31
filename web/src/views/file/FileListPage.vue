<template>
  <div class="file-list-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>文件管理</span>
          <div>
            <el-input v-model="searchKeyword" placeholder="搜索文件名" style="width: 200px; margin-right: 10px;" clearable
              @keyup.enter="handleSearch" @clear="handleSearch">
              <template #prefix>
                <el-icon>
                  <Search />
                </el-icon>
              </template>
            </el-input>
            <el-button @click="handleSearch" style="margin-right: 10px;">
              <el-icon>
                <Search />
              </el-icon>
              搜索
            </el-button>
            <el-button @click="refreshFileList">
              <el-icon>
                <Refresh />
              </el-icon>
              刷新
            </el-button>
          </div>
        </div>
      </template>

      <!-- 统计信息 -->
      <el-row :gutter="20" style="margin-bottom: 20px;">
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-item">
              <div class="stat-icon upload-icon">
                <el-icon>
                  <Upload />
                </el-icon>
              </div>
              <div class="stat-content">
                <div class="stat-number">{{ stats.totalUploads }}</div>
                <div class="stat-label">总文件数</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-item">
              <div class="stat-icon success-icon">
                <el-icon>
                  <Check />
                </el-icon>
              </div>
              <div class="stat-content">
                <div class="stat-number">{{ stats.successUploads }}</div>
                <div class="stat-label">成功上传</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-item">
              <div class="stat-icon size-icon">
                <el-icon>
                  <Folder />
                </el-icon>
              </div>
              <div class="stat-content">
                <div class="stat-number">{{ formatFileSize(stats.totalSize) }}</div>
                <div class="stat-label">总文件大小</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-item">
              <div class="stat-icon today-icon">
                <el-icon>
                  <Calendar />
                </el-icon>
              </div>
              <div class="stat-content">
                <div class="stat-number">{{ stats.todayUploads }}</div>
                <div class="stat-label">今日上传</div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <!-- 文件表格 -->
      <el-table :data="fileList" style="width: 100%" v-loading="loading">
        <el-table-column prop="name" label="文件名" min-width="200" show-overflow-tooltip>
          <template #default="scope">
            <div class="file-info">
              <el-icon class="file-icon">
                <Picture v-if="isImage(scope.row)" />
                <Document v-else />
              </el-icon>
              <span>{{ scope.row.name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="size" label="文件大小" width="120">
          <template #default="scope">
            {{ formatFileSize(scope.row.size) }}
          </template>
        </el-table-column>
        <el-table-column prop="type" label="文件类型" width="120">
          <template #default="scope">
            <el-tag size="small">{{ getFileType(scope.row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="uploadTime" label="上传时间" width="180" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.status === 'success' ? 'success' : 'danger'" size="small">
              {{ scope.row.status === 'success' ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="scope">
            <el-button size="small" @click="previewFile(scope.row)">
              <el-icon>
                <View />
              </el-icon>
              预览
            </el-button>
            <el-button size="small" @click="downloadFile(scope.row)">
              <el-icon>
                <Download />
              </el-icon>
              下载
            </el-button>
            <el-button size="small" type="danger" @click="deleteFile(scope.row)">
              <el-icon>
                <Delete />
              </el-icon>
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize"
        :page-sizes="[10, 20, 50, 100]" :total="pagination.total" layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange" @current-change="handleCurrentChange"
        style="margin-top: 20px; text-align: right;" />
    </el-card>

    <!-- 文件预览对话框 -->
    <el-dialog v-model="previewVisible" title="文件预览" width="80%" top="5vh">
      <div class="preview-container">
        <img v-if="isImage(previewFileData)" :src="previewFileData.url" class="preview-image" />
        <iframe v-else-if="isPdf(previewFileData)" :src="previewFileData.url" class="preview-iframe"></iframe>
        <div v-else class="preview-placeholder">
          <el-icon class="large-icon">
            <Document />
          </el-icon>
          <p>{{ previewFileData.name }}</p>
          <p>此文件类型不支持在线预览</p>
          <el-button type="primary" @click="downloadFile(previewFileData)">
            <el-icon>
              <Download />
            </el-icon>
            下载文件
          </el-button>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { uploadApi } from '@/api/upload'

export default {
  name: 'FileListPage',
  setup() {
    const fileList = ref([])
    const searchKeyword = ref('')
    const loading = ref(false)
    const previewVisible = ref(false)
    const previewFileData = ref({})

    const stats = ref({
      totalUploads: 0,
      successUploads: 0,
      totalSize: 0,
      todayUploads: 0
    })

    const pagination = ref({
      page: 1,
      pageSize: 10,
      total: 0
    })

    // 初始化数据
    onMounted(() => {
      loadFileList()
      loadUploadStats()
    })

    // 加载文件列表
    const loadFileList = async () => {
      try {
        loading.value = true
        const params = {
          page: pagination.value.page,
          pageSize: pagination.value.pageSize,
          keyword: searchKeyword.value
        }

        const response = await uploadApi.getFileList(params)

        if (response && typeof response === 'object') {
          const list = response.list || []
          fileList.value = Array.isArray(list) ? list.map(file => ({
            id: file?.id || Date.now() + Math.random(),
            name: file?.name || file?.fileName || '未知文件',
            size: file?.size || file?.fileSize || 0,
            type: file?.type || file?.mimeType || file?.contentType || 'application/octet-stream',
            url: file?.url || file?.downloadUrl || '#',
            uploadTime: file?.uploadTime || file?.createTime || new Date().toLocaleString(),
            status: file?.status || 'success',
            objectName: file?.objectName || file?.name || ''
          })) : []
          pagination.value.total = response.total || 0
        } else {
          fileList.value = []
          pagination.value.total = 0
        }
      } catch (error) {
        console.error('加载文件列表失败:', error)
        fileList.value = []
        pagination.value.total = 0
      } finally {
        loading.value = false
      }
    }

    // 加载上传统计
    const loadUploadStats = async () => {
      try {
        const response = await uploadApi.getUploadStats()

        if (response && typeof response === 'object') {
          stats.value = {
            totalUploads: response.totalUploads || 0,
            successUploads: response.successUploads || 0,
            totalSize: response.totalSize || 0,
            todayUploads: response.todayUploads || 0
          }
        } else {
          stats.value = {
            totalUploads: 0,
            successUploads: 0,
            totalSize: 0,
            todayUploads: 0
          }
        }
      } catch (error) {
        console.error('加载统计数据失败:', error)
        stats.value = {
          totalUploads: 0,
          successUploads: 0,
          totalSize: 0,
          todayUploads: 0
        }
      }
    }

    // 预览文件
    const previewFile = (file) => {
      previewFileData.value = file
      previewVisible.value = true
    }

    // 下载文件
    const downloadFile = (file) => {
      if (file.url === '#') {
        ElMessage.warning('文件下载链接不可用')
        return
      }

      const link = document.createElement('a')
      link.href = file.url
      link.download = file.name
      link.target = '_blank'
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
    }

    // 删除文件
    const deleteFile = (file) => {
      ElMessageBox.confirm('确定要删除这个文件吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        try {
          const objectName = file.objectName || file.name
          if (!objectName) {
            ElMessage.error('无法获取文件标识')
            return
          }

          await uploadApi.deleteFile(objectName)
          ElMessage.success('文件删除成功')

          // 重新加载文件列表和统计
          loadFileList()
          loadUploadStats()
        } catch (error) {
          ElMessage.error('文件删除失败: ' + (error.message || '未知错误'))
        }
      }).catch(() => {
        ElMessage.info('已取消删除')
      })
    }

    // 刷新文件列表
    const refreshFileList = () => {
      loadFileList()
      loadUploadStats()
      ElMessage.success('文件列表已刷新')
    }

    // 判断是否为图片
    const isImage = (file) => {
      return file.type && file.type.startsWith('image/')
    }

    // 判断是否为PDF
    const isPdf = (file) => {
      return file.type === 'application/pdf'
    }

    // 获取文件类型显示名称
    const getFileType = (mimeType) => {
      const typeMap = {
        'image/jpeg': 'JPEG',
        'image/png': 'PNG',
        'image/gif': 'GIF',
        'application/pdf': 'PDF',
        'application/msword': 'DOC',
        'application/vnd.openxmlformats-officedocument.wordprocessingml.document': 'DOCX',
        'text/plain': 'TXT'
      }
      return typeMap[mimeType] || '其他'
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

    // 分页处理
    const handleSizeChange = (val) => {
      pagination.value.pageSize = val
      pagination.value.page = 1
      loadFileList()
    }

    const handleCurrentChange = (val) => {
      pagination.value.page = val
      loadFileList()
    }

    // 搜索处理
    const handleSearch = () => {
      pagination.value.page = 1
      loadFileList()
    }

    return {
      fileList,
      searchKeyword,
      loading,
      previewVisible,
      previewFileData,
      stats,
      pagination,
      loadFileList,
      loadUploadStats,
      previewFile,
      downloadFile,
      deleteFile,
      refreshFileList,
      handleSearch,
      isImage,
      isPdf,
      getFileType,
      formatFileSize,
      handleSizeChange,
      handleCurrentChange
    }
  }
}
</script>

<style scoped>
.file-list-page {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: bold;
}

.stat-card {
  margin-bottom: 20px;
}

.stat-item {
  display: flex;
  align-items: center;
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 15px;
  font-size: 24px;
  color: white;
}

.upload-icon {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.success-icon {
  background: linear-gradient(135deg, #67C23A 0%, #85ce61 100%);
}

.size-icon {
  background: linear-gradient(135deg, #E6A23C 0%, #ebb563 100%);
}

.today-icon {
  background: linear-gradient(135deg, #409EFF 0%, #66b1ff 100%);
}

.stat-content {
  flex: 1;
}

.stat-number {
  font-size: 32px;
  font-weight: bold;
  color: #303133;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 5px;
}

.file-info {
  display: flex;
  align-items: center;
}

.file-icon {
  margin-right: 8px;
  color: #909399;
}

.preview-container {
  text-align: center;
}

.preview-image {
  max-width: 100%;
  max-height: 70vh;
  object-fit: contain;
}

.preview-iframe {
  width: 100%;
  height: 70vh;
  border: none;
}

.preview-placeholder {
  padding: 60px 20px;
}

.large-icon {
  font-size: 64px;
  color: #c0c4cc;
  margin-bottom: 20px;
}

.preview-placeholder p {
  margin: 10px 0;
  color: #606266;
}
</style>