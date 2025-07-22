#!/bin/bash

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 默认配置
HOST="localhost"
PORT="8080"
USER_ID="1001"
TOKEN=""
CAPTCHA=""

# 显示标题
echo -e "${BLUE}=======================================${NC}"
echo -e "${BLUE}       个性化推荐接口调试工具         ${NC}"
echo -e "${BLUE}=======================================${NC}"

# 显示帮助信息
show_help() {
  echo -e "${YELLOW}用法:${NC}"
  echo -e "  ./test_recommendation.sh [选项]"
  echo
  echo -e "${YELLOW}选项:${NC}"
  echo -e "  -h, --host HOST      指定主机名 (默认: localhost)"
  echo -e "  -p, --port PORT      指定端口号 (默认: 8080)"
  echo -e "  -u, --user ID        指定用户ID (默认: 1001)"
  echo -e "  -t, --token TOKEN    指定身份验证令牌"
  echo -e "  -c, --captcha CODE   指定验证码"
  echo -e "  --help               显示此帮助信息"
  echo
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
  case $1 in
    -h|--host)
      HOST="$2"
      shift 2
      ;;
    -p|--port)
      PORT="$2"
      shift 2
      ;;
    -u|--user)
      USER_ID="$2"
      shift 2
      ;;
    -t|--token)
      TOKEN="$2"
      shift 2
      ;;
    -c|--captcha)
      CAPTCHA="$2"
      shift 2
      ;;
    --help)
      show_help
      exit 0
      ;;
    *)
      echo -e "${RED}错误: 未知选项 $1${NC}"
      show_help
      exit 1
      ;;
  esac
done

# 检查是否安装了jq
check_jq() {
  if ! command -v jq &> /dev/null; then
    echo -e "${YELLOW}警告: 未安装jq工具，JSON输出将不会格式化${NC}"
    return 1
  fi
  return 0
}

# 获取验证码
get_captcha() {
  echo -e "${GREEN}获取验证码...${NC}"
  
  if check_jq; then
    RESPONSE=$(curl -s "http://${HOST}:${PORT}/api/captcha/get" | jq .)
    echo "$RESPONSE"
    
    # 提取验证码ID
    CAPTCHA_ID=$(echo "$RESPONSE" | jq -r '.data.captcha_id')
    echo -e "${YELLOW}验证码ID: ${CAPTCHA_ID}${NC}"
    echo -e "${YELLOW}请在图片中查看验证码并输入:${NC}"
    read -r CAPTCHA
  else
    curl -s "http://${HOST}:${PORT}/api/captcha/get"
    echo -e "${YELLOW}请输入验证码ID:${NC}"
    read -r CAPTCHA_ID
    echo -e "${YELLOW}请输入验证码:${NC}"
    read -r CAPTCHA
  fi
}

# 获取身份验证令牌
get_token() {
  echo -e "${GREEN}获取身份验证令牌...${NC}"
  
  if [ -z "$CAPTCHA" ] || [ -z "$CAPTCHA_ID" ]; then
    get_captcha
  fi
  
  if check_jq; then
    RESPONSE=$(curl -s -X POST "http://${HOST}:${PORT}/api/user/login" \
      -H "Content-Type: application/json" \
      -d "{
        \"user_id\": \"${USER_ID}\",
        \"captcha_id\": \"${CAPTCHA_ID}\",
        \"captcha_code\": \"${CAPTCHA}\"
      }" | jq .)
    echo "$RESPONSE"
    
    # 提取令牌
    TOKEN=$(echo "$RESPONSE" | jq -r '.data.token')
    if [ "$TOKEN" != "null" ]; then
      echo -e "${YELLOW}令牌: ${TOKEN}${NC}"
    else
      echo -e "${RED}获取令牌失败${NC}"
      exit 1
    fi
  else
    curl -s -X POST "http://${HOST}:${PORT}/api/user/login" \
      -H "Content-Type: application/json" \
      -d "{
        \"user_id\": \"${USER_ID}\",
        \"captcha_id\": \"${CAPTCHA_ID}\",
        \"captcha_code\": \"${CAPTCHA}\"
      }"
    echo -e "${YELLOW}请输入获取到的令牌:${NC}"
    read -r TOKEN
  fi
}

# 获取个性化推荐
get_recommendations() {
  echo -e "${GREEN}获取个性化推荐...${NC}"
  
  if [ -z "$TOKEN" ]; then
    echo -e "${YELLOW}未提供令牌，尝试获取...${NC}"
    get_token
  fi
  
  # 构建请求参数
  PARAMS="user_id=${USER_ID}"
  
  # 添加可选参数
  echo -e "${YELLOW}请输入城市ID (可选，直接回车跳过):${NC}"
  read -r CITY_ID
  if [ -n "$CITY_ID" ]; then
    PARAMS="${PARAMS}&city_id=${CITY_ID}"
  fi
  
  echo -e "${YELLOW}请输入区域ID (可选，直接回车跳过):${NC}"
  read -r DISTRICT_ID
  if [ -n "$DISTRICT_ID" ]; then
    PARAMS="${PARAMS}&district_id=${DISTRICT_ID}"
  fi
  
  echo -e "${YELLOW}请输入价格范围，格式为'最低-最高' (可选，直接回车跳过):${NC}"
  read -r PRICE_RANGE
  if [ -n "$PRICE_RANGE" ]; then
    PARAMS="${PARAMS}&price_range=${PRICE_RANGE}"
  fi
  
  echo -e "${YELLOW}请输入房型，如'2室1厅' (可选，直接回车跳过):${NC}"
  read -r HOUSE_TYPE
  if [ -n "$HOUSE_TYPE" ]; then
    PARAMS="${PARAMS}&house_type=${HOUSE_TYPE}"
  fi
  
  echo -e "${YELLOW}请输入页码 (默认: 1):${NC}"
  read -r PAGE
  PAGE=${PAGE:-1}
  PARAMS="${PARAMS}&page=${PAGE}"
  
  echo -e "${YELLOW}请输入每页数量 (默认: 10):${NC}"
  read -r PAGE_SIZE
  PAGE_SIZE=${PAGE_SIZE:-10}
  PARAMS="${PARAMS}&page_size=${PAGE_SIZE}"
  
  # 发送请求
  echo -e "${GREEN}发送请求: http://${HOST}:${PORT}/api/house/recommend?${PARAMS}${NC}"
  
  if check_jq; then
    curl -s "http://${HOST}:${PORT}/api/house/recommend?${PARAMS}" \
      -H "Authorization: Bearer ${TOKEN}" | jq .
  else
    curl -s "http://${HOST}:${PORT}/api/house/recommend?${PARAMS}" \
      -H "Authorization: Bearer ${TOKEN}"
  fi
}

# 验证验证码
verify_captcha() {
  echo -e "${GREEN}验证验证码...${NC}"
  
  if [ -z "$CAPTCHA" ] || [ -z "$CAPTCHA_ID" ]; then
    get_captcha
  fi
  
  if check_jq; then
    curl -s -X POST "http://${HOST}:${PORT}/api/captcha/verify" \
      -H "Content-Type: application/json" \
      -d "{
        \"captcha_id\": \"${CAPTCHA_ID}\",
        \"captcha_code\": \"${CAPTCHA}\"
      }" | jq .
  else
    curl -s -X POST "http://${HOST}:${PORT}/api/captcha/verify" \
      -H "Content-Type: application/json" \
      -d "{
        \"captcha_id\": \"${CAPTCHA_ID}\",
        \"captcha_code\": \"${CAPTCHA}\"
      }"
  fi
}

# 主菜单
show_menu() {
  echo
  echo -e "${YELLOW}当前配置:${NC}"
  echo -e "  主机: ${HOST}"
  echo -e "  端口: ${PORT}"
  echo -e "  用户ID: ${USER_ID}"
  echo -e "  令牌: ${TOKEN:-未设置}"
  echo -e "  验证码: ${CAPTCHA:-未设置}"
  echo
  echo -e "${YELLOW}请选择操作:${NC}"
  echo -e "  1) 获取验证码"
  echo -e "  2) 验证验证码"
  echo -e "  3) 获取身份验证令牌"
  echo -e "  4) 获取个性化推荐"
  echo -e "  5) 设置/修改参数"
  echo -e "  h) 显示帮助信息"
  echo -e "  q) 退出"
  echo
  echo -n "请输入选项: "
}

# 设置参数
set_params() {
  echo -e "${YELLOW}设置参数 (直接回车保持当前值)${NC}"
  
  echo -e "主机 (当前: ${HOST}):"
  read -r input
  if [ -n "$input" ]; then
    HOST="$input"
  fi
  
  echo -e "端口 (当前: ${PORT}):"
  read -r input
  if [ -n "$input" ]; then
    PORT="$input"
  fi
  
  echo -e "用户ID (当前: ${USER_ID}):"
  read -r input
  if [ -n "$input" ]; then
    USER_ID="$input"
  fi
  
  echo -e "令牌 (当前: ${TOKEN:-未设置}):"
  read -r input
  if [ -n "$input" ]; then
    TOKEN="$input"
  fi
  
  echo -e "验证码ID (当前: ${CAPTCHA_ID:-未设置}):"
  read -r input
  if [ -n "$input" ]; then
    CAPTCHA_ID="$input"
  fi
  
  echo -e "验证码 (当前: ${CAPTCHA:-未设置}):"
  read -r input
  if [ -n "$input" ]; then
    CAPTCHA="$input"
  fi
}

# 主循环
while true; do
  show_menu
  read -r option
  
  case $option in
    1)
      get_captcha
      ;;
    2)
      verify_captcha
      ;;
    3)
      get_token
      ;;
    4)
      get_recommendations
      ;;
    5)
      set_params
      ;;
    h)
      show_help
      ;;
    q)
      echo -e "${BLUE}感谢使用个性化推荐接口调试工具!${NC}"
      exit 0
      ;;
    *)
      echo -e "${RED}无效选项，请重试${NC}"
      ;;
  esac
done