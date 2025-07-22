#!/bin/bash

# 生成 protobuf Go 代码
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       --go-http_out=. --go-http_opt=paths=source_relative \
       api/house/v3/house.proto

echo "Proto 文件生成完成"