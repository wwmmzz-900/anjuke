<template>
  <div class="file-upload">
    <div class="upload-container">
      <div class="upload-header">
        <h3>文件上传</h3>
        <button @click="selectFile">选择文件</button>
        <input ref="fileInput" type="file" style="display: none" @change="handleFileChange" />
      </div>
      <div v-if="currentFile" class="file-info">
        <div class="file-details">
          <span>{{ currentFile.name }}</span>
          <span>{{ formatSize(currentFile.size) }}</span>
        </div>
        <div class="progress-container">
          <progress :value="uploadProgress" max="100"></progress>
          <span>{{ uploadProgress }}%</span>
          <div class="upload-actions">
            <button v-if="!uploading && !uploadComplete" @click="startUpload">开始上传</button>
            <button v-if="uploading && !uploadComplete" @click="cancelUpload">取消</button>
          </div>
        </div>
      </div>
      <div class="upload-status">
        <p v-if="uploadMessage">{{ uploadMessage }}</p>
        <p v-if="uploadComplete">
          上传完成！文件URL: <a :href="fileUrl" target="_blank">{{ fileUrl }}</a>
        </p>
      </div>
    </div>
  </div>
</template>

<script>
import { ref } from 'vue'
import { uploadApi } from '@/api/upload'

export default {
  name: 'FileUpload',
  setup() {
    const fileInput = ref(null);
    const currentFile = ref(null);
    const uploading = ref(false);
    const uploadComplete = ref(false);
    const uploadProgress = ref(0);
    const uploadMessage = ref('');
    const fileUrl = ref('');

    const selectFile = () => {
      fileInput.value.click();
    };
    const handleFileChange = (event) => {
      const file = event.target.files[0];
      if (!file) return;
      currentFile.value = file;
      uploadProgress.value = 0;
      uploadComplete.value = false;
      uploading.value = false;
      uploadMessage.value = '准备上传...';
      fileUrl.value = '';
    };
    const startUpload = async () => {
      if (!currentFile.value) return;
      try {
        uploading.value = true;
        uploadMessage.value = '上传中...';
        const response = await uploadApi.smartUpload(currentFile.value, (progressEvent) => {
          const percentCompleted = Math.round((progressEvent.loaded * 100) / progressEvent.total);
          uploadProgress.value = percentCompleted;
        });
        fileUrl.value = response.data.url;
        uploadComplete.value = true;
        uploadMessage.value = '上传完成！';
      } catch (error) {
        uploadMessage.value = `上传失败: ${error.message}`;
      } finally {
        uploading.value = false;
      }
    };
    const cancelUpload = () => {
      // 这里只能重置状态，无法中断已发出的请求
      currentFile.value = null;
      uploading.value = false;
      uploadProgress.value = 0;
      uploadMessage.value = '';
      fileUrl.value = '';
      uploadComplete.value = false;
    };
    const formatSize = (bytes) => {
      if (bytes === 0) return '0 B';
      const k = 1024;
      const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
      const i = Math.floor(Math.log(bytes) / Math.log(k));
      return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i];
    };
    return {
      fileInput,
      currentFile,
      uploading,
      uploadComplete,
      uploadProgress,
      uploadMessage,
      fileUrl,
      selectFile,
      handleFileChange,
      startUpload,
      cancelUpload,
      formatSize
    };
  }
};
</script>

<style scoped>
.file-upload {
  width: 100%;
  max-width: 600px;
  margin: 0 auto;
  padding: 20px;
}
.upload-container {
  border: 1px solid #ebeef5;
  border-radius: 4px;
  padding: 20px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}
.upload-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.file-info {
  margin-bottom: 20px;
}
.file-details {
  display: flex;
  justify-content: space-between;
  margin-bottom: 10px;
}
.progress-container {
  margin-bottom: 20px;
}
.upload-actions {
  margin-top: 10px;
  display: flex;
  gap: 10px;
}
.upload-status {
  margin-top: 20px;
  word-break: break-all;
}
</style>