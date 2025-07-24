# 修复说明

## 问题1：房源推荐列表无法获取房源图片地址

### 原因
在查询房源数据时，没有包含图片URL字段，并且没有正确关联房源图片表。

### 解决方案
1. 创建房源图片表（如果不存在）
2. 添加房源图片数据
3. 修改数据层代码，正确查询和关联房源图片

### 执行步骤
1. 执行SQL脚本创建房源图片表：
   ```bash
   mysql -u root -p < scripts/create_house_image_table.sql
   ```

2. 执行SQL脚本添加房源图片数据：
   ```bash
   mysql -u root -p < scripts/add_house_images.sql
   ```

3. 重新编译并启动服务：
   ```bash
   go build -o anjuke ./cmd/anjuke
   ./anjuke
   ```

## 问题2：WebSocket消息为空

### 原因
在WebSocket消息处理中，有两个不同的字段名用于消息内容：`message` 和 `content`。

### 解决方案
修改WebSocket消息处理代码，统一使用`message`字段，同时保留`content`字段以兼容。

### 执行步骤
1. 重新编译并启动服务：
   ```bash
   go build -o anjuke ./cmd/anjuke
   ./anjuke
   ```

## 测试方法

### 测试房源推荐列表
```bash
curl -s "http://localhost:8000/house/recommend?page=1&page_size=10" | python -m json.tool
```

### 测试个性化推荐列表
```bash
curl -s "http://localhost:8000/house/personal-recommend?user_id=1001&page=1&page_size=10" | python -m json.tool
```

### 测试WebSocket消息
使用`test_websocket.sh`脚本或WebSocket客户端测试消息发送和接收。