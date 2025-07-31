const { defineConfig } = require('@vue/cli-service')

module.exports = defineConfig({
  transpileDependencies: true,
  
  // 生产环境配置
  publicPath: process.env.NODE_ENV === 'production' ? '/' : '/',
  
  // 输出目录
  outputDir: 'dist',
  
  // 静态资源目录
  assetsDir: 'static',
  
  // 生产环境不生成source map
  productionSourceMap: false,
  
  // 开发服务器配置
  devServer: {
    port: 8080,
    host: '0.0.0.0',
    proxy: {
      '/api': {
        target: 'http://localhost:8001',
        changeOrigin: true,
        ws: true, // 支持WebSocket
      },
      '/user': {
        target: 'http://localhost:8001',
        changeOrigin: true,
      }
    }
  },
  
  // 构建优化
  configureWebpack: {
    optimization: {
      splitChunks: {
        chunks: 'all',
        cacheGroups: {
          vendor: {
            name: 'chunk-vendors',
            test: /[\\/]node_modules[\\/]/,
            priority: 10,
            chunks: 'initial'
          },
          elementPlus: {
            name: 'chunk-element-plus',
            test: /[\\/]node_modules[\\/]element-plus[\\/]/,
            priority: 20
          }
        }
      }
    }
  },
  
  // CSS配置
  css: {
    extract: process.env.NODE_ENV === 'production',
    sourceMap: false
  },
  
  // PWA配置（可选）
  pwa: {
    name: '安居客管理系统',
    themeColor: '#409EFF',
    msTileColor: '#000000',
    appleMobileWebAppCapable: 'yes',
    appleMobileWebAppStatusBarStyle: 'black',
    workboxOptions: {
      skipWaiting: true
    }
  }
})