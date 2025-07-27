const { defineConfig } = require('@vue/cli-service')

module.exports = defineConfig({
  transpileDependencies: true,
  devServer: {
    port: 8081,
    proxy: {
      '/api': {
        target: 'http://localhost:8001',
        changeOrigin: true,
        // 不重写路径，保留/api前缀
        pathRewrite: null
      }
    }
  }
})