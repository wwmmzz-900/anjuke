package service

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/wwmmzz-900/anjuke/api/house/v3"
	"github.com/wwmmzz-900/anjuke/internal/data"
)

// 聊天服务
type ChatService struct {
	chatRepo *data.ChatRepo
}

// 创建聊天服务
func NewChatService(chatRepo *data.ChatRepo) *ChatService {
	return &ChatService{
		chatRepo: chatRepo,
	}
}

// 发起聊天
func (s *ChatService) StartChat(ctx context.Context, req *pb.StartChatRequest) (*pb.StartChatReply, error) {
	// 简化实现：不验证预约信息
	log.Printf("发起聊天: 预约ID=%d, 用户ID=%d, 房东ID=%d", 
		req.ReservationId, req.UserId, req.LandlordId)

	// 简化实现：直接创建聊天会话
	chatID := fmt.Sprintf("chat_%d_%d_%d", req.ReservationId, req.UserId, time.Now().Unix())

	// 4. 如果有初始消息，记录日志
	if req.InitialMessage != "" {
		log.Printf("聊天初始消息: 聊天ID=%s, 内容=%s", chatID, req.InitialMessage)
	}

	return &pb.StartChatReply{
		Code: 0,
		Msg:  "发起聊天成功",
		Data: &pb.StartChatData{
			ChatId:  chatID,
			Success: true,
		},
	}, nil
}

// 注意：此处省略了消息通知功能，实际项目中应该实现