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
          <el-menu-item index="/home">
            <el-icon><House /></el-icon>
            <span>首页</span>
          </el-menu-item>
          <el-menu-item index="/upload-test">
            <el-icon><Upload /></el-icon>
            <span>上传测试</span>
          </el-menu-item>

        </el-menu>
      </el-aside>
      <el-container>
        <el-header>
          <div style="display: flex; justify-content: space-between; align-items: center;">
            <h2>安居客管理系统</h2>
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
import { House, Upload } from '@element-plus/icons-vue'

export default {
  name: 'App',
  components: {
    House,
    Upload
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