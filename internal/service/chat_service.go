package service

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/wwmmzz-900/anjuke/api/house/v3"
	"github.com/wwmmzz-900/anjuke/internal/data"
	"github.com/wwmmzz-900/anjuke/internal/model"
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
	log.Printf("发起聊天: 预约ID=%d, 用户ID=%d, 房东ID=%d",
		req.ReservationId, req.UserId, req.LandlordId)

	// 检查参数有效性
	if req.ReservationId <= 0 || req.UserId <= 0 || req.LandlordId <= 0 {
		return &pb.StartChatReply{
			Code: 400,
			Msg:  "无效的参数",
			Data: &pb.StartChatData{
				Success: false,
			},
		}, nil
	}

	// 检查是否已存在聊天会话
	exists, err := s.chatRepo.ChatSessionExists(ctx, req.ReservationId)
	if err != nil {
		log.Printf("检查聊天会话失败: %v", err)
	}

	var chatID string
	
	if exists {
		// 获取已存在的聊天会话
		session, err := s.chatRepo.GetChatSessionByReservationID(ctx, req.ReservationId)
		if err != nil {
			log.Printf("获取聊天会话失败: %v", err)
			// 创建新会话作为备选方案
			chatID = fmt.Sprintf("chat_%d_%d_%d", req.ReservationId, req.UserId, time.Now().Unix())
		} else {
			chatID = session.ChatID
		}
	} else {
		// 创建新的聊天会话
		// 在实际项目中，应该从预约信息中获取房源ID
		houseID := int64(0) // 这里应该从预约信息中获取
		chatID, err = s.chatRepo.CreateChatSession(ctx, req.ReservationId, req.UserId, req.LandlordId, houseID)
		if err != nil {
			log.Printf("创建聊天会话失败: %v", err)
			// 创建临时会话ID作为备选方案
			chatID = fmt.Sprintf("chat_%d_%d_%d", req.ReservationId, req.UserId, time.Now().Unix())
		}
	}

	// 如果有初始消息，保存并发送
	if req.InitialMessage != "" {
		// 保存消息到数据库
		message := &model.ChatMessage{
			ChatID:       chatID,
			SenderID:     req.UserId,
			SenderName:   "用户", // 实际项目中应该从用户信息中获取
			ReceiverID:   req.LandlordId,
			ReceiverName: "房东", // 实际项目中应该从房东信息中获取
			Type:         0,    // 文本消息
			Content:      req.InitialMessage,
			Read:         false,
			CreatedAt:    time.Now(),
		}
		
		if err := s.chatRepo.AddChatMessage(ctx, message); err != nil {
			log.Printf("保存聊天消息失败: %v", err)
		}
		
		// 通过WebSocket发送消息给房东
		chatMessage := map[string]interface{}{
			"type":           "chat",
			"chat_id":        chatID,
			"from":           req.UserId,
			"to":             req.LandlordId,
			"content":        req.InitialMessage,
			"message":        req.InitialMessage, // 兼容旧版本
			"reservation_id": req.ReservationId,
			"timestamp":      time.Now().Unix(),
			"sequence":       GlobalSequenceManager.GetNextSequence(req.UserId),
		}
		
		// 使用全局WebSocket管理器发送消息
		if err := GlobalWebSocketManager.SendMessageToUser(req.LandlordId, chatMessage); err != nil {
			log.Printf("发送WebSocket消息失败: %v", err)
		}
		
		// 发送确认消息给用户
		confirmMessage := map[string]interface{}{
			"type":           "chat_confirm",
			"chat_id":        chatID,
			"message":        "消息已发送",
			"reservation_id": req.ReservationId,
			"timestamp":      time.Now().Unix(),
		}
		
		if err := GlobalWebSocketManager.SendMessageToUser(req.UserId, confirmMessage); err != nil {
			log.Printf("发送确认消息失败: %v", err)
		}
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

// SendChatMessage 发送聊天消息
func (s *ChatService) SendChatMessage(ctx context.Context, senderID, receiverID int64, chatID, content string, msgType int) error {
	// 参数验证
	if senderID <= 0 || receiverID <= 0 || chatID == "" || content == "" {
		return fmt.Errorf("无效的参数")
	}
	
	// 获取发送者和接收者信息（实际项目中应该从用户服务获取）
	senderName := fmt.Sprintf("用户%d", senderID)
	receiverName := fmt.Sprintf("用户%d", receiverID)
	
	// 保存消息到数据库
	message := &model.ChatMessage{
		ChatID:       chatID,
		SenderID:     senderID,
		SenderName:   senderName,
		ReceiverID:   receiverID,
		ReceiverName: receiverName,
		Type:         msgType,
		Content:      content,
		Read:         false,
		CreatedAt:    time.Now(),
	}
	
	if err := s.chatRepo.AddChatMessage(ctx, message); err != nil {
		return fmt.Errorf("保存聊天消息失败: %w", err)
	}
	
	// 通过WebSocket发送消息给接收者
	chatMessage := map[string]interface{}{
		"type":        "chat",
		"chat_id":     chatID,
		"from":        senderID,
		"from_name":   senderName,
		"to":          receiverID,
		"to_name":     receiverName,
		"content":     content,
		"message":     content, // 兼容旧版本
		"message_type": msgType,
		"timestamp":   time.Now().Unix(),
		"sequence":    GlobalSequenceManager.GetNextSequence(senderID),
	}
	
	// 使用全局WebSocket管理器发送消息
	if err := GlobalWebSocketManager.SendMessageToUser(receiverID, chatMessage); err != nil {
		log.Printf("发送WebSocket消息失败: %v", err)
	}
	
	// 发送确认消息给发送者
	confirmMessage := map[string]interface{}{
		"type":      "chat_confirm",
		"chat_id":   chatID,
		"message":   "消息已发送",
		"timestamp": time.Now().Unix(),
	}
	
	if err := GlobalWebSocketManager.SendMessageToUser(senderID, confirmMessage); err != nil {
		log.Printf("发送确认消息失败: %v", err)
	}
	
	return nil
}

// GetChatMessages 获取聊天消息列表
func (s *ChatService) GetChatMessages(ctx context.Context, chatID string, page, pageSize int) ([]*model.ChatMessage, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	
	return s.chatRepo.GetChatMessages(ctx, chatID, page, pageSize)
}

// MarkMessagesAsRead 标记消息为已读
func (s *ChatService) MarkMessagesAsRead(ctx context.Context, chatID string, receiverID int64) error {
	return s.chatRepo.MarkMessagesAsRead(ctx, chatID, receiverID)
}

// GetUnreadMessageCount 获取未读消息数量
func (s *ChatService) GetUnreadMessageCount(ctx context.Context, receiverID int64) (int64, error) {
	return s.chatRepo.GetUnreadMessageCount(ctx, receiverID)
}
