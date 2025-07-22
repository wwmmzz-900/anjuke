package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// 颜色代码
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
)

// 消息类型
const (
	TypeSystem = "system"
	TypeChat   = "chat"
	TypeEcho   = "echo"
	TypeError  = "error"
)

func main() {
	// 解析命令行参数
	houseID, userID, host, port := parseArgs()

	// 构建 WebSocket URL
	u := url.URL{
		Scheme:   "ws",
		Host:     fmt.Sprintf("%s:%s", host, port),
		Path:     "/ws/house",
		RawQuery: fmt.Sprintf("house_id=%s&user_id=%s", houseID, userID),
	}

	fmt.Printf("%s连接到 WebSocket: %s%s\n", colorYellow, u.String(), colorReset)

	// 连接 WebSocket
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("WebSocket 连接失败:", err)
	}
	defer c.Close()

	fmt.Printf("%sWebSocket 连接成功！%s\n", colorGreen, colorReset)
	printHelp()

	// 处理中断信号
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// 接收消息的 goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					log.Printf("%s读取消息错误: %v%s", colorRed, err, colorReset)
				}
				return
			}
			
			// 解析并格式化消息
			formatAndPrintMessage(message)
		}
	}()

	// 发送心跳消息的 ticker
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// 读取用户输入的 goroutine
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("> ")
			input, err := reader.ReadString('\n')
			if err != nil {
				log.Printf("%s读取输入错误: %v%s", colorRed, err, colorReset)
				continue
			}

			// 去除换行符
			input = strings.TrimSpace(input)

			// 处理特殊命令
			if input == "/help" {
				printHelp()
				continue
			} else if input == "/quit" {
				fmt.Println("正在关闭连接...")
				c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				time.Sleep(time.Second)
				os.Exit(0)
			} else if strings.HasPrefix(input, "/to ") {
				// 格式: /to <user_id> <message>
				parts := strings.SplitN(input[4:], " ", 2)
				if len(parts) != 2 {
					fmt.Printf("%s无效的命令格式。使用: /to <user_id> <message>%s\n", colorRed, colorReset)
					continue
				}
				
				// 构造私聊消息
				msg := map[string]interface{}{
					"action":  "message",
					"to":      parts[0],
					"content": parts[1],
				}
				
				jsonMsg, err := json.Marshal(msg)
				if err != nil {
					fmt.Printf("%s消息格式化失败: %v%s\n", colorRed, err, colorReset)
					continue
				}
				
				if err := c.WriteMessage(websocket.TextMessage, jsonMsg); err != nil {
					fmt.Printf("%s发送消息失败: %v%s\n", colorRed, err, colorReset)
				}
				continue
			} else if strings.HasPrefix(input, "/json ") {
				// 直接发送JSON
				jsonStr := input[6:]
				if err := c.WriteMessage(websocket.TextMessage, []byte(jsonStr)); err != nil {
					fmt.Printf("%s发送JSON失败: %v%s\n", colorRed, err, colorReset)
				}
				continue
			}

			// 发送普通消息
			if err := c.WriteMessage(websocket.TextMessage, []byte(input)); err != nil {
				fmt.Printf("%s发送消息失败: %v%s\n", colorRed, err, colorReset)
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			// 发送心跳消息
			err := c.WriteMessage(websocket.TextMessage, []byte("{\"action\":\"ping\"}"))
			if err != nil {
				log.Println("发送心跳失败:", err)
				return
			}
		case <-interrupt:
			log.Println("收到中断信号，正在关闭连接...")
			
			// 发送关闭消息
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("发送关闭消息失败:", err)
				return
			}
			
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

// 解析命令行参数
func parseArgs() (string, string, string, string) {
	houseID := "101"
	userID := "1001"
	host := "localhost"
	port := "8080"

	if len(os.Args) >= 3 {
		houseID = os.Args[1]
		userID = os.Args[2]
	}

	if len(os.Args) >= 4 {
		host = os.Args[3]
	}

	if len(os.Args) >= 5 {
		port = os.Args[4]
	}

	if len(os.Args) < 3 {
		fmt.Println("用法: go run websocket_client.go <house_id> <user_id> [host] [port]")
		fmt.Println("示例: go run websocket_client.go 101 1001 localhost 8080")
		fmt.Println("使用默认值:", houseID, userID, host, port)
	}

	return houseID, userID, host, port
}

// 打印帮助信息
func printHelp() {
	fmt.Printf("\n%s可用命令:%s\n", colorCyan, colorReset)
	fmt.Printf("  %s/help%s - 显示此帮助信息\n", colorYellow, colorReset)
	fmt.Printf("  %s/quit%s - 关闭连接并退出\n", colorYellow, colorReset)
	fmt.Printf("  %s/to <user_id> <message>%s - 发送私聊消息\n", colorYellow, colorReset)
	fmt.Printf("  %s/json <json_string>%s - 发送原始JSON消息\n", colorYellow, colorReset)
	fmt.Printf("  %s其他任何输入将作为普通消息发送%s\n\n", colorGreen, colorReset)
}

// 格式化并打印消息
func formatAndPrintMessage(message []byte) {
	// 尝试解析为JSON
	var msgMap map[string]interface{}
	if err := json.Unmarshal(message, &msgMap); err != nil {
		// 不是JSON，直接打印
		fmt.Printf("%s[%s] 收到消息: %s%s\n", 
			colorPurple, time.Now().Format("15:04:05"), string(message), colorReset)
		return
	}

	// 格式化JSON消息
	msgType, _ := msgMap["type"].(string)
	timestamp := time.Now().Format("15:04:05")

	switch msgType {
	case TypeSystem:
		fmt.Printf("%s[%s] 系统消息: %v%s\n", 
			colorBlue, timestamp, msgMap["message"], colorReset)
	
	case TypeChat:
		from, _ := msgMap["from"].(float64)
		content, hasContent := msgMap["content"].(string)
		message, hasMessage := msgMap["message"].(string)
		
		var msgContent string
		if hasContent {
			msgContent = content
		} else if hasMessage {
			msgContent = message
		} else {
			msgContent = "<空消息>"
		}
		
		fmt.Printf("%s[%s] 用户 %.0f 说: %s%s\n", 
			colorGreen, timestamp, from, msgContent, colorReset)
	
	case TypeError:
		fmt.Printf("%s[%s] 错误: %v%s\n", 
			colorRed, timestamp, msgMap["message"], colorReset)
	
	default:
		// 美化输出JSON
		prettyJSON, err := json.MarshalIndent(msgMap, "", "  ")
		if err != nil {
			fmt.Printf("%s[%s] %s%s\n", 
				colorPurple, timestamp, string(message), colorReset)
		} else {
			fmt.Printf("%s[%s] 收到JSON:%s\n%s%s\n", 
				colorPurple, timestamp, colorReset, string(prettyJSON), colorReset)
		}
	}
}