#!/bin/bash

echo "生成权限模块 protobuf 代码..."

# 生成 Go 代码
kratos proto client api/permission/v1/permission.proto

# 如果上面命令失败，使用 protoc 直接生成
if [ $? -ne 0 ]; then
    echo "使用 protoc 生成代码..."
    protoc --proto_path=. \
           --proto_path=./third_party \
           --go_out=paths=source_relative:. \
           --go-http_out=paths=source_relative:. \
           --go-grpc_out=paths=source_relative:. \
           api/permission/v1/permission.proto
fi

echo "权限模块 protobuf 代码生成完成！"