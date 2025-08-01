package data

import (
	"context"
	"fmt"
	"time"

	"github.com/wwmmzz-900/anjuke/internal/model"
)

// 聊天数据仓库
type ChatRepo struct {
	data *Data
}

// 创建聊天数据仓库
func NewChatRepo(data *Data) *ChatRepo {
	return &ChatRepo{data: data}
}

// 创建聊天会话
func (r *ChatRepo) CreateChatSession(ctx context.Context, reservationID, userID, landlordID, houseID int64) (string, error) {
	// 参数验证
	if reservationID <= 0 || userID <= 0 || landlordID <= 0 {
		return "", fmt.Errorf("无效的参数")
	}

	// 生成聊天ID（简化版本）
	chatID := fmt.Sprintf("chat_%d_%d_%d", reservationID, userID, time.Now().Unix())

	// 创建聊天会话
	session := &model.ChatSession{
		ChatID:        chatID,
		ReservationID: reservationID,
		UserID:        userID,
		LandlordID:    landlordID,
		HouseID:       houseID,
		Status:        model.ChatSessionStatusActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 保存到数据库
	if err := r.data.db.WithContext(ctx).Create(session).Error; err != nil {
		return "", fmt.Errorf("创建聊天会话失败: %w", err)
	}

	return chatID, nil
}

// 获取聊天会话
func (r *ChatRepo) GetChatSession(ctx context.Context, chatID string) (*model.ChatSession, error) {
	var session model.ChatSession
	if err := r.data.db.WithContext(ctx).Where("chat_id = ?", chatID).First(&session).Error; err != nil {
		return nil, fmt.Errorf("获取聊天会话失败: %w", err)
	}
	return &session, nil
}

// 获取用户的聊天会话列表
func (r *ChatRepo) GetUserChatSessions(ctx context.Context, userID int64) ([]*model.ChatSession, error) {
	var sessions []*model.ChatSession
	if err := r.data.db.WithContext(ctx).Where("user_id = ? OR landlord_id = ?", userID, userID).
		Order("updated_at DESC").Find(&sessions).Error; err != nil {
		return nil, fmt.Errorf("获取用户聊天会话列表失败: %w", err)
	}
	return sessions, nil
}

// 添加聊天消息
func (r *ChatRepo) AddChatMessage(ctx context.Context, message *model.ChatMessage) error {
	// 设置创建时间
	message.CreatedAt = time.Now()

	// 保存到数据库
	if err := r.data.db.WithContext(ctx).Create(message).Error; err != nil {
		return fmt.Errorf("添加聊天消息失败: %w", err)
	}

	// 更新聊天会话的更新时间
	if err := r.data.db.WithContext(ctx).Model(&model.ChatSession{}).
		Where("chat_id = ?", message.ChatID).
		Update("updated_at", time.Now()).Error; err != nil {
		return fmt.Errorf("更新聊天会话时间失败: %w", err)
	}

	return nil
}

// 获取聊天消息列表
func (r *ChatRepo) GetChatMessages(ctx context.Context, chatID string, page, pageSize int) ([]*model.ChatMessage, int64, error) {
	var messages []*model.ChatMessage
	var total int64

	// 统计总数
	if err := r.data.db.WithContext(ctx).Model(&model.ChatMessage{}).
		Where("chat_id = ?", chatID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计聊天消息数量失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := r.data.db.WithContext(ctx).Where("chat_id = ?", chatID).
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&messages).Error; err != nil {
		return nil, 0, fmt.Errorf("获取聊天消息列表失败: %w", err)
	}

	return messages, total, nil
}

// 标记消息为已读
func (r *ChatRepo) MarkMessagesAsRead(ctx context.Context, chatID string, receiverID int64) error {
	if err := r.data.db.WithContext(ctx).Model(&model.ChatMessage{}).
		Where("chat_id = ? AND receiver_id = ? AND read = ?", chatID, receiverID, false).
		Update("read", true).Error; err != nil {
		return fmt.Errorf("标记消息为已读失败: %w", err)
	}
	return nil
}

// 获取未读消息数量
func (r *ChatRepo) GetUnreadMessageCount(ctx context.Context, receiverID int64) (int64, error) {
	var count int64
	if err := r.data.db.WithContext(ctx).Model(&model.ChatMessage{}).
		Where("receiver_id = ? AND read = ?", receiverID, false).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("获取未读消息数量失败: %w", err)
	}
	return count, nil
}

// 根据预约ID获取聊天会话
func (r *ChatRepo) GetChatSessionByReservationID(ctx context.Context, reservationID int64) (*model.ChatSession, error) {
	var session model.ChatSession
	if err := r.data.db.WithContext(ctx).Where("reservation_id = ?", reservationID).First(&session).Error; err != nil {
		return nil, fmt.Errorf("根据预约ID获取聊天会话失败: %w", err)
	}
	return &session, nil
}

// 检查聊天会话是否存在
func (r *ChatRepo) ChatSessionExists(ctx context.Context, reservationID int64) (bool, error) {
	var count int64
	if err := r.data.db.WithContext(ctx).Model(&model.ChatSession{}).
		Where("reservation_id = ?", reservationID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("检查聊天会话是否存在失败: %w", err)
	}
	return count > 0, nil
}
