#!/bin/bash

# 权限管理功能测试脚本
BASE_URL="http://localhost:8000"

echo "=== 权限管理功能测试 ==="

# 1. 更新用户权限
echo "1. 测试更新用户权限..."
curl -X PUT "${BASE_URL}/permission/user/1001" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1001,
    "permissions": ["READ", "WRITE", "PUBLISH_HOUSE"],
    "role": "ROLE_LANDLORD",
    "reason": "升级为房东用户",
    "operator_id": 1
  }' | jq .

echo -e "\n"

# 2. 获取用户权限
echo "2. 测试获取用户权限..."
curl -X GET "${BASE_URL}/permission/user/1001" | jq .

echo -e "\n"

# 3. 批量更新用户权限
echo "3. 测试批量更新用户权限..."
curl -X PUT "${BASE_URL}/permission/batch" \
  -H "Content-Type: application/json" \
  -d '{
    "updates": [
      {
        "user_id": 1002,
        "permissions": ["READ", "WRITE"],
        "role": "ROLE_NORMAL_USER",
        "reason": "普通用户权限",
        "operator_id": 1
      },
      {
        "user_id": 1003,
        "permissions": ["READ", "WRITE", "PUBLISH_HOUSE", "MANAGE_USER"],
        "role": "ROLE_AGENT",
        "reason": "升级为经纪人",
        "operator_id": 1
      }
    ]
  }' | jq .

echo -e "\n"

# 4. 获取权限列表
echo "4. 测试获取权限列表..."
curl -X GET "${BASE_URL}/permission/list?page=1&page_size=10" | jq .

echo -e "\n"

# 5. 获取角色权限
echo "5. 测试获取角色权限..."
curl -X GET "${BASE_URL}/permission/role/ROLE_AGENT" | jq .

echo -e "\n"

# 6. 测试权限验证失败的情况
echo "6. 测试无效权限（游客用户设置管理员权限）..."
curl -X PUT "${BASE_URL}/permission/user/1004" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1004,
    "permissions": ["ADMIN"],
    "role": "ROLE_GUEST",
    "reason": "测试无效权限",
    "operator_id": 1
  }' | jq .

echo -e "\n"

# 7. 测试不存在的用户
echo "7. 测试不存在的用户..."
curl -X PUT "${BASE_URL}/permission/user/99999" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 99999,
    "permissions": ["READ"],
    "role": "NORMAL_USER",
    "reason": "测试不存在用户",
    "operator_id": 1
  }' | jq .

echo -e "\n=== 测试完成 ==="