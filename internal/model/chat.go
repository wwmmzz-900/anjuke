package model

import "time"

// 聊天会话
type ChatSession struct {
	ChatID        string            `gorm:"column:chat_id;primaryKey;size:64" json:"chat_id"`              // 聊天ID
	ReservationID int64             `gorm:"column:reservation_id;not null;index" json:"reservation_id"`    // 预约ID
	UserID        int64             `gorm:"column:user_id;not null;index" json:"user_id"`                  // 用户ID
	LandlordID    int64             `gorm:"column:landlord_id;not null;index" json:"landlord_id"`          // 房东ID
	HouseID       int64             `gorm:"column:house_id;not null;index" json:"house_id"`                // 房源ID
	Status        ChatSessionStatus `gorm:"column:status;size:20;not null;default:'active'" json:"status"` // 状态：active/closed
	CreatedAt     time.Time         `gorm:"column:created_at;not null" json:"created_at"`                  // 创建时间
	UpdatedAt     time.Time         `gorm:"column:updated_at;not null" json:"updated_at"`                  // 更新时间
}

// 表名
func (ChatSession) TableName() string {
	return "chat_sessions"
}

// 聊天消息
type ChatMessage struct {
	ID           int64       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                // 消息ID
	ChatID       string      `gorm:"column:chat_id;size:64;not null;index" json:"chat_id"`        // 聊天ID
	SenderID     int64       `gorm:"column:sender_id;not null;index" json:"sender_id"`            // 发送者ID
	SenderName   string      `gorm:"column:sender_name;size:100;not null" json:"sender_name"`     // 发送者名称
	ReceiverID   int64       `gorm:"column:receiver_id;not null;index" json:"receiver_id"`        // 接收者ID
	ReceiverName string      `gorm:"column:receiver_name;size:100;not null" json:"receiver_name"` // 接收者名称
	Type         MessageType `gorm:"column:type;not null;default:0" json:"type"`                  // 消息类型：0-文本，1-图片，2-语音，3-位置，4-系统消息
	Content      string      `gorm:"column:content;type:text;not null" json:"content"`            // 消息内容
	Read         bool        `gorm:"column:read;not null;default:false" json:"read"`              // 是否已读
	CreatedAt    time.Time   `gorm:"column:created_at;not null" json:"created_at"`                // 创建时间
}

// 表名
func (ChatMessage) TableName() string {
	return "chat_messages"
}
