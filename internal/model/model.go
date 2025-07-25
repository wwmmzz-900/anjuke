package model

import "time"

// 常量定义
const (
	// 默认分页参数
	DefaultPage     = 1
	DefaultPageSize = 10
	MaxPageSize     = 100
	
	// 默认价格区间
	DefaultMinPrice  = 800.0
	DefaultMaxPrice  = 1500.0
	FallbackMaxPrice = 5000.0
	
	// 用户行为分析参数
	MaxRecentViewCount = 20
	
	// 默认用户名称
	DefaultUserName     = "用户"
	DefaultLandlordName = "房东"
)

// 房源状态枚举
type HouseStatus string

const (
	HouseStatusActive   HouseStatus = "active"   // 活跃
	HouseStatusInactive HouseStatus = "inactive" // 不活跃
	HouseStatusRented   HouseStatus = "rented"   // 已租用
)

// 预约状态枚举
type ReservationStatus string

const (
	ReservationStatusPending   ReservationStatus = "pending"   // 待确认
	ReservationStatusConfirmed ReservationStatus = "confirmed" // 已确认
	ReservationStatusCancelled ReservationStatus = "cancelled" // 已取消
	ReservationStatusCompleted ReservationStatus = "completed" // 已完成
)

// 消息类型枚举
type MessageType int

const (
	MessageTypeText     MessageType = 0 // 文本消息
	MessageTypeImage    MessageType = 1 // 图片消息
	MessageTypeVoice    MessageType = 2 // 语音消息
	MessageTypeLocation MessageType = 3 // 位置消息
	MessageTypeSystem   MessageType = 4 // 系统消息
)

// 聊天会话状态枚举
type ChatSessionStatus string

const (
	ChatSessionStatusActive ChatSessionStatus = "active" // 活跃
	ChatSessionStatusClosed ChatSessionStatus = "closed" // 关闭
)

// WebSocket消息类型枚举
type WSMessageType string

const (
	WSMessageTypeConnection  WSMessageType = "connection"
	WSMessageTypeChat        WSMessageType = "chat"
	WSMessageTypeSystem      WSMessageType = "system"
	WSMessageTypeError       WSMessageType = "error"
	WSMessageTypeEcho        WSMessageType = "echo"
	WSMessageTypeChatConfirm WSMessageType = "chat_confirm"
)

// 用户行为类型枚举
const (
	UserBehaviorView     = "view"     // 浏览
	UserBehaviorFavorite = "favorite" // 收藏
	UserBehaviorContact  = "contact"  // 联系
	UserBehaviorReserve  = "reserve"  // 预约
)

// 错误码定义
const (
	ErrCodeSuccess           = 0
	ErrCodeInvalidParams     = 400
	ErrCodeUnauthorized      = 401
	ErrCodeNotFound          = 404
	ErrCodeInternalError     = 500
	ErrCodeDuplicateReserve  = 1001
	ErrCodeChatSessionExists = 1002
)

// House 房源模型
type House struct {
	HouseId                 int64       `gorm:"column:house_id;type:bigint;comment:房源ID;primaryKey;not null;" json:"house_id"`
	Title                   string      `gorm:"column:title;type:varchar(100);comment:房源标题;not null;" json:"title"`
	Description             string      `gorm:"column:description;type:text;comment:房源描述;" json:"description"`
	LandlordId              int64       `gorm:"column:landlord_id;type:bigint;comment:发布人ID;not null;" json:"landlord_id"`
	Address                 string      `gorm:"column:address;type:varchar(255);comment:详细地址;not null;" json:"address"`
	RegionId                int64       `gorm:"column:region_id;type:bigint;comment:区域/小区ID;default:NULL;" json:"region_id"`
	CommunityId             int64       `gorm:"column:community_id;type:bigint;comment:小区ID;default:NULL;" json:"community_id"`
	Price                   float64     `gorm:"column:price;type:decimal(10, 2);comment:价格;not null;" json:"price"`
	Area                    float32     `gorm:"column:area;type:float;comment:面积;default:NULL;" json:"area"`
	Layout                  string      `gorm:"column:layout;type:varchar(50);comment:户型;default:NULL;" json:"layout"`
	Floor                   string      `gorm:"column:floor;type:varchar(20);comment:楼层;default:NULL;" json:"floor"`
	OwnershipCertificateUrl string      `gorm:"column:ownership_certificate_url;type:varchar(255);comment:产权证明图片;not null;" json:"ownership_certificate_url"`
	Orientation             string      `gorm:"column:orientation;type:varchar(20);comment:朝向;default:NULL;" json:"orientation"`
	Decoration              string      `gorm:"column:decoration;type:varchar(50);comment:装修;default:NULL;" json:"decoration"`
	Facilities              string      `gorm:"column:facilities;type:varchar(255);comment:配套设施（逗号分隔）;default:NULL;" json:"facilities"`
	Status                  HouseStatus `gorm:"column:status;type:enum('active', 'inactive', 'rented');comment:状态;not null;default:'active'" json:"status"`
	CreatedAt               time.Time   `gorm:"column:created_at;type:datetime;comment:发布时间;not null;default:CURRENT_TIMESTAMP;" json:"created_at"`
	UpdatedAt               time.Time   `gorm:"column:updated_at;type:datetime;comment:更新时间;not null;default:CURRENT_TIMESTAMP;" json:"updated_at"`
	DeletedAt               *time.Time  `gorm:"column:deleted_at;type:datetime;comment:删除时间;default:NULL;" json:"deleted_at"`
}

func (*House) TableName() string {
	return "house"
}

// HouseReservation 房源预约模型
type HouseReservation struct {
	ID          int64             `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	LandlordID  int64             `gorm:"column:landlord_id;not null;index" json:"landlord_id"`
	UserID      int64             `gorm:"column:user_id;not null;index" json:"user_id"`
	UserName    string            `gorm:"column:user_name;size:100;not null" json:"user_name"`
	HouseID     int64             `gorm:"column:house_id;not null;index" json:"house_id"`
	HouseTitle  string            `gorm:"column:house_title;size:200;not null" json:"house_title"`
	ReserveTime string            `gorm:"column:reserve_time;not null" json:"reserve_time"`
	Status      ReservationStatus `gorm:"column:status;size:20;not null;default:'pending'" json:"status"`
	CreatedAt   time.Time         `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time         `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName 指定表名
func (HouseReservation) TableName() string {
	return "house_reservations"
}

// IsValidStatus 检查预约状态是否有效
func (r *HouseReservation) IsValidStatus() bool {
	switch r.Status {
	case ReservationStatusPending, ReservationStatusConfirmed, 
		 ReservationStatusCancelled, ReservationStatusCompleted:
		return true
	default:
		return false
	}
}

// CanTransitionTo 检查是否可以转换到指定状态
func (r *HouseReservation) CanTransitionTo(newStatus ReservationStatus) bool {
	transitions := map[ReservationStatus][]ReservationStatus{
		ReservationStatusPending:   {ReservationStatusConfirmed, ReservationStatusCancelled},
		ReservationStatusConfirmed: {ReservationStatusCompleted, ReservationStatusCancelled},
		ReservationStatusCancelled: {},
		ReservationStatusCompleted: {},
	}
	
	allowedTransitions := transitions[r.Status]
	for _, allowed := range allowedTransitions {
		if allowed == newStatus {
			return true
		}
	}
	return false
}