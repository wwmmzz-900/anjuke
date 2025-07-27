<template>
  <div id="app">
    <el-container>
      <el-aside width="200px">
        <el-menu
          :default-active="$route.path"
          router
          background-color="#545c64"
          text-color="#fff"
          active-text-color="#ffd04b"
        >
          <el-menu-item index="/upload">
            <el-icon><Upload /></el-icon>
            <span>文件管理</span>
          </el-menu-item>

          <el-menu-item index="/upload/test">
            <el-icon><Setting /></el-icon>
            <span>上传测试</span>
          </el-menu-item>

        </el-menu>
      </el-aside>
      <el-container>
        <el-header>
          <div style="display: flex; justify-content: space-between; align-items: center;">
            <h2>文件上传管理系统</h2>
          </div>
        </el-header>
        <el-main>
          <router-view />
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script>
import { Upload, Setting } from '@element-plus/icons-vue'

export default {
  name: 'App',
  components: {
    Upload,
    Setting
  }
}
</script>

<style>
#app {
  height: 100vh;
}

.el-container {
  height: 100%;
}

.el-header {
  background-color: #B3C0D1;
  color: #333;
  line-height: 60px;
}

.el-aside {
  background-color: #545c64;
}
</style>