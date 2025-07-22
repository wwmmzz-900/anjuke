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
HOUSE_ID="101"
USER_ID="1001"
LANDLORD_ID="2001"

# 显示标题
echo -e "${BLUE}=======================================${NC}"
echo -e "${BLUE}       WebSocket 接口调试工具         ${NC}"
echo -e "${BLUE}=======================================${NC}"

# 显示帮助信息
show_help() {
  echo -e "${YELLOW}用法:${NC}"
  echo -e "  ./test_websocket.sh [选项]"
  echo
  echo -e "${YELLOW}选项:${NC}"
  echo -e "  -h, --host HOST      指定主机名 (默认: localhost)"
  echo -e "  -p, --port PORT      指定端口号 (默认: 8080)"
  echo -e "  --house ID           指定房源ID (默认: 101)"
  echo -e "  --user ID            指定用户ID (默认: 1001)"
  echo -e "  --landlord ID        指定房东ID (默认: 2001)"
  echo -e "  --help               显示此帮助信息"
  echo
  echo -e "${YELLOW}可用命令:${NC}"
  echo -e "  1) 检查WebSocket连接状态"
  echo -e "  2) 连接WebSocket (需要安装websocat)"
  echo -e "  3) 发送预约请求 (触发WebSocket消息)"
  echo -e "  4) 发送聊天消息"
  echo -e "  5) 运行Go WebSocket客户端"
  echo -e "  q) 退出"
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
    --house)
      HOUSE_ID="$2"
      shift 2
      ;;
    --user)
      USER_ID="$2"
      shift 2
      ;;
    --landlord)
      LANDLORD_ID="$2"
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

# 检查是否安装了websocat
check_websocat() {
  if ! command -v websocat &> /dev/null; then
    echo -e "${RED}错误: 未安装websocat工具${NC}"
    echo -e "请安装websocat: https://github.com/vi/websocat/releases"
    return 1
  fi
  return 0
}

# 检查WebSocket连接状态
check_ws_status() {
  echo -e "${GREEN}检查WebSocket连接状态...${NC}"
  
  if check_jq; then
    curl -s "http://${HOST}:${PORT}/api/websocket/stats" | jq .
  else
    curl -s "http://${HOST}:${PORT}/api/websocket/stats"
  fi
  
  echo -e "${GREEN}完成${NC}"
}

# 连接WebSocket
connect_ws() {
  if ! check_websocat; then
    return 1
  fi
  
  echo -e "${GREEN}连接到WebSocket...${NC}"
  echo -e "${YELLOW}按Ctrl+C退出连接${NC}"
  echo
  
  websocat "ws://${HOST}:${PORT}/ws/house?house_id=${HOUSE_ID}&user_id=${USER_ID}"
}

# 发送预约请求
send_reservation() {
  echo -e "${GREEN}发送预约请求...${NC}"
  
  CURRENT_DATE=$(date +"%Y-%m-%d")
  
  if check_jq; then
    curl -X POST "http://${HOST}:${PORT}/house/reserve" \
      -H "Content-Type: application/json" \
      -d "{
        \"landlord_id\": ${LANDLORD_ID},
        \"user_id\": ${USER_ID},
        \"user_name\": \"张三\",
        \"house_id\": ${HOUSE_ID},
        \"house_title\": \"精装修两室一厅\",
        \"reserve_time\": \"${CURRENT_DATE} 14:00:00\"
      }" | jq .
  else
    curl -X POST "http://${HOST}:${PORT}/house/reserve" \
      -H "Content-Type: application/json" \
      -d "{
        \"landlord_id\": ${LANDLORD_ID},
        \"user_id\": ${USER_ID},
        \"user_name\": \"张三\",
        \"house_id\": ${HOUSE_ID},
        \"house_title\": \"精装修两室一厅\",
        \"reserve_time\": \"${CURRENT_DATE} 14:00:00\"
      }"
  fi
  
  echo -e "${GREEN}完成${NC}"
}

# 发送聊天消息
send_chat_message() {
  echo -e "${GREEN}发送聊天消息...${NC}"
  echo -e "请输入接收者ID (默认: ${LANDLORD_ID}):"
  read -r RECEIVER_ID
  RECEIVER_ID=${RECEIVER_ID:-$LANDLORD_ID}
  
  echo -e "请输入消息内容:"
  read -r MESSAGE
  
  if check_jq; then
    curl -X POST "http://${HOST}:${PORT}/api/chat/send" \
      -H "Content-Type: application/json" \
      -d "{
        \"sender_id\": ${USER_ID},
        \"receiver_id\": ${RECEIVER_ID},
        \"content\": \"${MESSAGE}\",
        \"type\": 1
      }" | jq .
  else
    curl -X POST "http://${HOST}:${PORT}/api/chat/send" \
      -H "Content-Type: application/json" \
      -d "{
        \"sender_id\": ${USER_ID},
        \"receiver_id\": ${RECEIVER_ID},
        \"content\": \"${MESSAGE}\",
        \"type\": 1
      }"
  fi
  
  echo -e "${GREEN}完成${NC}"
}

# 运行Go WebSocket客户端
run_go_client() {
  echo -e "${GREEN}运行Go WebSocket客户端...${NC}"
  go run websocket_client.go "${HOUSE_ID}" "${USER_ID}"
}

# 主菜单
show_menu() {
  echo
  echo -e "${YELLOW}当前配置:${NC}"
  echo -e "  主机: ${HOST}"
  echo -e "  端口: ${PORT}"
  echo -e "  房源ID: ${HOUSE_ID}"
  echo -e "  用户ID: ${USER_ID}"
  echo -e "  房东ID: ${LANDLORD_ID}"
  echo
  echo -e "${YELLOW}请选择操作:${NC}"
  echo -e "  1) 检查WebSocket连接状态"
  echo -e "  2) 连接WebSocket (需要安装websocat)"
  echo -e "  3) 发送预约请求 (触发WebSocket消息)"
  echo -e "  4) 发送聊天消息"
  echo -e "  5) 运行Go WebSocket客户端"
  echo -e "  h) 显示帮助信息"
  echo -e "  q) 退出"
  echo
  echo -n "请输入选项: "
}

# 主循环
while true; do
  show_menu
  read -r option
  
  case $option in
    1)
      check_ws_status
      ;;
    2)
      connect_ws
      ;;
    3)
      send_reservation
      ;;
    4)
      send_chat_message
      ;;
    5)
      run_go_client
      ;;
    h)
      show_help
      ;;
    q)
      echo -e "${BLUE}感谢使用WebSocket接口调试工具!${NC}"
      exit 0
      ;;
    *)
      echo -e "${RED}无效选项，请重试${NC}"
      ;;
  esac
done