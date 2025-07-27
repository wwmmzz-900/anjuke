@echo off
setlocal enabledelayedexpansion

REM 测试运行脚本 (Windows版本)
REM 用法: run_tests.bat [选项]
REM 选项:
REM   -v: 详细输出
REM   -c: 生成覆盖率报告
REM   -u: 只运行单元测试（跳过集成测试）
REM   -i: 只运行集成测试
REM   -h: 显示帮助

set VERBOSE=false
set COVERAGE=false
set UNIT_ONLY=false
set INTEGRATION_ONLY=false

REM 解析命令行参数
:parse_args
if "%1"=="-v" (
    set VERBOSE=true
    shift
    goto parse_args
)
if "%1"=="-c" (
    set COVERAGE=true
    shift
    goto parse_args
)
if "%1"=="-u" (
    set UNIT_ONLY=true
    shift
    goto parse_args
)
if "%1"=="-i" (
    set INTEGRATION_ONLY=true
    shift
    goto parse_args
)
if "%1"=="-h" (
    echo 用法: %0 [选项]
    echo 选项:
    echo   -v: 详细输出
    echo   -c: 生成覆盖率报告
    echo   -u: 只运行单元测试（跳过集成测试）
    echo   -i: 只运行集成测试
    echo   -h: 显示帮助
    exit /b 0
)
if not "%1"=="" (
    echo 无效选项: %1
    exit /b 1
)

REM 进入项目根目录
cd /d "%~dp0\.."

echo 🚀 开始运行测试...

REM 构建测试参数
set TEST_ARGS=
if "%VERBOSE%"=="true" (
    set TEST_ARGS=!TEST_ARGS! -v
)
if "%COVERAGE%"=="true" (
    set TEST_ARGS=!TEST_ARGS! -coverprofile=coverage.out
)

REM 运行测试
if "%UNIT_ONLY%"=="true" (
    echo 📋 运行单元测试...
    go test !TEST_ARGS! -short ./internal/...
) else if "%INTEGRATION_ONLY%"=="true" (
    echo 🔗 运行集成测试...
    go test !TEST_ARGS! -run Integration ./internal/...
) else (
    echo 🧪 运行所有测试...
    go test !TEST_ARGS! ./internal/...
)

if errorlevel 1 (
    echo ❌ 测试失败!
    exit /b 1
)

REM 生成覆盖率报告
if "%COVERAGE%"=="true" (
    if exist coverage.out (
        echo 📊 生成覆盖率报告...
        go tool cover -html=coverage.out -o coverage.html
        echo ✅ 覆盖率报告已生成: coverage.html
        
        REM 显示覆盖率统计
        echo 📈 覆盖率统计:
        go tool cover -func=coverage.out | findstr "total:"
    )
)

echo ✅ 测试完成!