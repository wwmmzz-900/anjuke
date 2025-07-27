import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    redirect: '/upload'
  },
  // 文件上传
  {
    path: '/upload',
    name: 'FileUpload',
    component: () => import('../views/upload/FileUploadPage.vue')
  },
  {
    path: '/upload/test',
    name: 'UploadTest',
    component: () => import('../views/upload/UploadTest.vue')
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router