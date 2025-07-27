#!/bin/bash

# 测试运行脚本
# 用法: ./scripts/run_tests.sh [选项]
# 选项:
#   -v: 详细输出
#   -c: 生成覆盖率报告
#   -u: 只运行单元测试（跳过集成测试）
#   -i: 只运行集成测试
#   -h: 显示帮助

set -e

# 默认参数
VERBOSE=false
COVERAGE=false
UNIT_ONLY=false
INTEGRATION_ONLY=false

# 解析命令行参数
while getopts "vcuih" opt; do
  case $opt in
    v)
      VERBOSE=true
      ;;
    c)
      COVERAGE=true
      ;;
    u)
      UNIT_ONLY=true
      ;;
    i)
      INTEGRATION_ONLY=true
      ;;
    h)
      echo "用法: $0 [选项]"
      echo "选项:"
      echo "  -v: 详细输出"
      echo "  -c: 生成覆盖率报告"
      echo "  -u: 只运行单元测试（跳过集成测试）"
      echo "  -i: 只运行集成测试"
      echo "  -h: 显示帮助"
      exit 0
      ;;
    \?)
      echo "无效选项: -$OPTARG" >&2
      exit 1
      ;;
  esac
done

# 进入项目根目录
cd "$(dirname "$0")/.."

echo "🚀 开始运行测试..."

# 构建测试参数
TEST_ARGS=""
if [ "$VERBOSE" = true ]; then
    TEST_ARGS="$TEST_ARGS -v"
fi

if [ "$COVERAGE" = true ]; then
    TEST_ARGS="$TEST_ARGS -coverprofile=coverage.out"
fi

# 运行测试
if [ "$UNIT_ONLY" = true ]; then
    echo "📋 运行单元测试..."
    go test $TEST_ARGS -short ./internal/...
elif [ "$INTEGRATION_ONLY" = true ]; then
    echo "🔗 运行集成测试..."
    go test $TEST_ARGS -run Integration ./internal/...
else
    echo "🧪 运行所有测试..."
    go test $TEST_ARGS ./internal/...
fi

# 生成覆盖率报告
if [ "$COVERAGE" = true ] && [ -f coverage.out ]; then
    echo "📊 生成覆盖率报告..."
    go tool cover -html=coverage.out -o coverage.html
    echo "✅ 覆盖率报告已生成: coverage.html"
    
    # 显示覆盖率统计
    echo "📈 覆盖率统计:"
    go tool cover -func=coverage.out | tail -1
fi

echo "✅ 测试完成!"