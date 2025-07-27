# 文件上传管理系统

基于 Vue 3 + Element Plus 开发的文件上传管理系统前端项目。

## 功能特性

- 📁 **文件上传** - 支持单文件和多文件上传
- 📋 **文件管理** - 文件列表查看、搜索、删除
- 📊 **统计信息** - 上传统计和文件分析
- 🔧 **测试工具** - 上传功能测试和调试

## 技术栈

- Vue 3
- Vue Router 4
- Vuex 4
- Element Plus
- Axios

## 项目结构

```
web/
├── public/
│   └── index.html
├── src/
│   ├── api/           # API 接口
│   ├── components/    # 公共组件
│   ├── router/        # 路由配置
│   ├── store/         # 状态管理
│   ├── views/         # 页面组件
│   │   ├── user/      # 用户管理
│   │   ├── customer/  # 客户管理
│   │   ├── house/     # 房屋管理
│   │   ├── points/    # 积分管理
│   │   └── transaction/ # 交易管理
│   ├── App.vue
│   └── main.js
├── package.json
├── vue.config.js
└── README.md
```

## 安装和运行

### 1. 安装依赖

```bash
cd web
npm install
```

### 2. 启动开发服务器

```bash
npm run serve
```

项目将在 http://localhost:8080 启动

### 3. 构建生产版本

```bash
npm run build
```

## API 接口

项目通过代理配置连接到后端 API（默认 http://localhost:8000）。

主要接口模块：
- `/user/*` - 用户相关接口
- `/customer/*` - 客户相关接口
- `/house/*` - 房屋相关接口
- `/points/*` - 积分相关接口
- `/transaction/*` - 交易相关接口

## 页面说明

### 登录页面
- 手机号 + 验证码登录
- 集成短信验证码发送

### 仪表盘
- 数据统计卡片
- 最近用户注册
- 最近交易记录

### 用户管理
- 用户列表查看和搜索
- 用户创建（手机号注册）
- 实名认证功能
- 用户状态管理

### 客户管理
- 客户信息管理
- 客户列表和搜索

### 房屋管理
- 房屋信息管理
- 房屋列表和搜索
- 房屋状态管理

### 积分管理
- 积分余额查询
- 积分明细记录
- 签到管理
- 积分获取和使用

### 交易管理
- 交易记录查看
- 交易创建和处理
- 交易状态管理

## 开发说明

1. 所有 API 调用都通过 `src/api/` 目录下的模块进行
2. 使用 Element Plus 组件库进行 UI 开发
3. 路由配置在 `src/router/index.js`
4. 状态管理使用 Vuex，配置在 `src/store/index.js`
5. 开发时后端 API 通过 vue.config.js 中的代理配置访问

## 注意事项

- 确保后端服务在 http://localhost:8000 运行
- 登录功能需要后端短信验证码接口支持
- 部分功能使用模拟数据，实际使用时需要连接真实 API