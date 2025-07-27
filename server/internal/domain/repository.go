// Package domain 定义了项目的核心业务领域模型和仓储接口。
// 这是整个应用最低层、最核心的包，它不依赖任何其他内部包。
// 所有业务实体（Entities）和数据操作规约（Repository Interfaces）都在这里定义。
package domain

import (
	"context"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

// Greeter is a Greeter model.
type Greeter struct {
	ID    int64
	Hello string
}

type UserBase struct {
	UserId     uint64         `gorm:"column:user_id;primaryKey;autoIncrement;comment:用户唯一ID" json:"user_id"`
	Name       string         `gorm:"type:varchar(50);not null;comment:用户昵称/姓名" json:"name"`
	RealName   string         `gorm:"type:varchar(30);comment:真实姓名" json:"real_name,omitempty"`
	Phone      string         `gorm:"type:char(11);not null;comment:手机号" json:"phone"`
	Password   string         `gorm:"type:char(32);not null;comment:密码（加密存储）" json:"-"`
	Avatar     string         `gorm:"type:text;comment:头像URL" json:"avatar,omitempty"`
	RoleID     uint64         `gorm:"not null;comment:角色id" json:"role_id"`
	Sex        Sex            `gorm:"type:enum('男','女');default:null;comment:用户性别" json:"sex,omitempty"`
	RealStatus RealStatus     `gorm:"type:tinyint;comment:用户实名状态(1:已实名2:未实名)" json:"real_status,omitempty"`
	Status     Status         `gorm:"type:tinyint;not null;default:1;comment:状态（0禁用1正常）" json:"status"`
	CreatedAt  time.Time      `gorm:"column:created_at;autoCreateTime;comment:注册时间" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间" json:"deleted_at,omitempty"`
}

func (UserBase) TableName() string {
	return "user_base"
}

type RealName struct {
	Id        uint64    `gorm:"column:id;type:bigint UNSIGNED;primaryKey;not null;" json:"id"`
	UserId    uint32    `gorm:"column:user_id;type:int UNSIGNED;comment:用户id;not null;default:0;" json:"user_id"`
	Name      string    `gorm:"column:name;type:varchar(30);comment:姓名;not null;" json:"name"`
	IdCard    string    `gorm:"column:id_card;type:char(18);comment:身份证号码;not null;" json:"id_card"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;comment:创建时间;not null;default:CURRENT_TIMESTAMP;" json:"created_at"`
}

func (RealName) TableName() string {
	return "real_name"
}

type (
	Sex        string
	RealStatus int8
	Status     int8
)

// Constants for enums
const (
	RealNameVerified   RealStatus = 1 // 已实名
	RealNameUnverified RealStatus = 2 // 未实名

	UserStatusDisabled Status = 0 // 禁用
	UserStatusNormal   Status = 1 // 正常
)

// --- 积分模块领域模型 ---

// UserPoints 用户积分信息
type UserPoints struct {
	UserID      uint64    `gorm:"column:user_id;primaryKey;comment:用户ID" json:"user_id"`
	TotalPoints int64     `gorm:"column:total_points;not null;default:0;comment:总积分" json:"total_points"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (UserPoints) TableName() string {
	return "user_points"
}

// PointsRecord 积分记录
type PointsRecord struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement;comment:记录ID" json:"id"`
	UserID      uint64    `gorm:"column:user_id;not null;index;comment:用户ID" json:"user_id"`
	Type        string    `gorm:"column:type;type:varchar(20);not null;comment:类型:checkin签到,consume消费,use使用" json:"type"`
	Points      int64     `gorm:"column:points;not null;comment:积分变动数量" json:"points"`
	Description string    `gorm:"column:description;type:varchar(200);comment:描述" json:"description"`
	OrderID     string    `gorm:"column:order_id;type:varchar(50);comment:关联订单ID" json:"order_id"`
	Amount      int64     `gorm:"column:amount;default:0;comment:关联金额(分)" json:"amount"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
}

func (PointsRecord) TableName() string {
	return "points_records"
}

// CheckInRecord 签到记录
type CheckInRecord struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement;comment:记录ID" json:"id"`
	UserID    uint64    `gorm:"column:user_id;not null;index;comment:用户ID" json:"user_id"`
	CheckDate string    `gorm:"column:check_date;type:date;not null;comment:签到日期" json:"check_date"`
	Points    int64     `gorm:"column:points;not null;comment:获得积分" json:"points"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
}

func (CheckInRecord) TableName() string {
	return "checkin_records"
}

// 积分操作结果结构体
type CheckInResult struct {
	PointsEarned    int64 `json:"points_earned"`
	TotalPoints     int64 `json:"total_points"`
	ConsecutiveDays int32 `json:"consecutive_days"`
}

type EarnResult struct {
	PointsEarned int64 `json:"points_earned"`
	TotalPoints  int64 `json:"total_points"`
}

type UseResult struct {
	PointsUsed     int64 `json:"points_used"`
	AmountDeducted int64 `json:"amount_deducted"`
	TotalPoints    int64 `json:"total_points"`
}

// 积分类型常量
const (
	PointsTypeCheckIn = "checkin" // 签到获得
	PointsTypeConsume = "consume" // 消费获得
	PointsTypeUse     = "use"     // 积分使用
)

// 积分规则常量
const (
	CheckInBasePoints  = 10 // 签到基础积分
	CheckInBonusPoints = 10 // 连续签到奖励积分（每7天）
	ConsumePointsRate  = 1  // 消费积分比例：1元=1积分
	UsePointsRate      = 10 // 使用积分比例：10积分=1元
)

// --- Repositories ---
// Repository 接口定义了数据持久化的规约，由 data 层实现。
// 这种方式将业务逻辑（biz）与数据存储（data）完全解耦。

// SmsRepo 定义了短信发送的仓储接口。
type SmsRepo interface {
	SendSms(ctx context.Context, phone, scene string) (string, error)       // 简化：只返回结果和错误
	VerifySms(ctx context.Context, phone, code, scene string) (bool, error) // 简化：不需要codeId
}

// GreeterRepo 定义了 Greeter 模块的仓储接口。
type GreeterRepo interface {
	Save(context.Context, *Greeter) (*Greeter, error)
	Update(context.Context, *Greeter) (*Greeter, error)
	FindByID(context.Context, int64) (*Greeter, error)
	ListByHello(context.Context, string) ([]*Greeter, error)
	ListAll(context.Context) ([]*Greeter, error)
}

// UserRepo 定义了用户模块的仓储接口，包含了用户实名、状态更新和短信风控等数据操作。
type UserRepo interface {
	RealName(ctx context.Context, user *RealName) (*RealName, error)
	UpdateUserStatus(ctx context.Context, status *UserBase) (*UserBase, error)
	SmsRiskControl(ctx context.Context, phone, deviceID, ip string) error
	// 检查手机号是否已存在
	CheckPhoneExists(ctx context.Context, phone string) (bool, error)
	// 创建新用户
	CreateUser(ctx context.Context, user *UserBase) (*UserBase, error)
	// 根据手机号和密码获取用户（密码登录）
	GetUserByPhoneAndPassword(ctx context.Context, phone, password string) (*UserBase, error)
	// 根据手机号获取用户（短信登录）
	GetUserByPhone(ctx context.Context, phone string) (*UserBase, error)
}

// PointsRepo 定义了积分模块的仓储接口
type PointsRepo interface {
	// 查询用户积分余额
	GetUserPoints(ctx context.Context, userID uint64) (*UserPoints, error)

	// 查询积分明细记录
	GetPointsHistory(ctx context.Context, userID uint64, page, pageSize int32, pointsType string) ([]*PointsRecord, int32, error)

	// 签到获取积分
	CheckIn(ctx context.Context, userID uint64) (*CheckInResult, error)

	// 消费获取积分
	EarnPointsByConsume(ctx context.Context, userID uint64, orderID string, amount int64) (*EarnResult, error)

	// 使用积分抵扣
	UsePoints(ctx context.Context, userID uint64, points int64, orderID, description string) (*UseResult, error)

	// 检查今日是否已签到
	HasCheckedInToday(ctx context.Context, userID uint64) (bool, error)

	// 获取连续签到天数
	GetConsecutiveCheckInDays(ctx context.Context, userID uint64) (int32, error)
}

// MinioRepo 定义了文件存储接口
// 只保留智能上传和常规文件操作方法
// 删除所有分片相关方法
type MinioRepo interface {
	SimpleUpload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error)
	SmartUpload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string, progressCb func(uploaded, total int64)) (string, error)
	Download(ctx context.Context, objectName string) (io.Reader, error)
	Delete(ctx context.Context, objectName string) error
	ListFiles(ctx context.Context, prefix string, maxKeys int) ([]FileInfo, error)
	GetFileInfo(ctx context.Context, objectName string) (*FileInfo, error)
	SearchFiles(ctx context.Context, keyword string, maxKeys int) ([]FileInfo, error)
	GetFileStats(ctx context.Context, prefix string) (map[string]any, error)
	CleanupIncompleteUploads(ctx context.Context, prefix string, olderThan time.Duration) error
	GetBucket() string
}

// MultipartUploadInfo 分片上传信息
type MultipartUploadInfo struct {
	UploadID    string               `json:"uploadId"`
	ObjectName  string               `json:"objectName"`
	Bucket      string               `json:"bucket"`
	Parts       []minio.CompletePart `json:"parts"`
	ContentType string               `json:"contentType"`
	TotalSize   int64                `json:"totalSize"`
	ChunkSize   int64                `json:"chunkSize"`
}

// UploadPartInfo 分片上传结果
type UploadPartInfo struct {
	PartNumber int    `json:"partNumber"`
	ETag       string `json:"etag"`
	Size       int64  `json:"size"`
}

// FileInfo 文件信息结构
// 放在domain包，供MinioRepo接口和data包共用
type FileInfo struct {
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"lastModified"`
	ContentType  string    `json:"contentType"`
	ETag         string    `json:"etag"`
	URL          string    `json:"url"`
}
