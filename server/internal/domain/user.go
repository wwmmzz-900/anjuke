package domain

import (
	"context"
	"io"
	"time"
)

// 用户状态常量
const (
	UserStatusNormal = 1 // 正常状态
	UserStatusBanned = 2 // 禁用状态
)

// 实名认证状态常量
const (
	RealNameUnverified = 0 // 未认证
	RealNameVerified   = 1 // 已认证
	RealNameRejected   = 2 // 认证失败
)

// UserBase 用户基础信息
type UserBase struct {
	ID         uint64    `json:"id"`
	UserId     uint64    `json:"user_id"` // 兼容字段
	Phone      string    `json:"phone"`
	Name       string    `json:"name"`
	Password   string    `json:"password"`
	Nickname   string    `json:"nickname"`
	Avatar     string    `json:"avatar"`
	Gender     int       `json:"gender"`
	Birthday   time.Time `json:"birthday"`
	Province   string    `json:"province"`
	City       string    `json:"city"`
	District   string    `json:"district"`
	Address    string    `json:"address"`
	IsRealName bool      `json:"is_real_name"`
	RoleID     int       `json:"role_id"`
	RealStatus int       `json:"real_status"`
	Status     int       `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// RealName 实名认证信息
type RealName struct {
	ID       uint64 `json:"id"`
	UserID   uint64 `json:"user_id"`
	UserId   uint64 `json:"user_id"` // 兼容字段
	RealName string `json:"real_name"`
	Name     string `json:"name"` // 兼容字段
	IDCard   string `json:"id_card"`
	IdCard   string `json:"id_card"` // 兼容字段
	Status   int    `json:"status"`
}

// FileInfo 文件信息
type FileInfo struct {
	ID           uint64    `json:"id"`
	FileName     string    `json:"file_name"`
	Name         string    `json:"name"` // 兼容字段
	FileSize     int64     `json:"file_size"`
	Size         int64     `json:"size"` // 兼容字段
	FileType     string    `json:"file_type"`
	ContentType  string    `json:"content_type"` // 兼容字段
	FileURL      string    `json:"file_url"`
	URL          string    `json:"url"`           // 兼容字段
	ETag         string    `json:"etag"`          // 兼容字段
	LastModified time.Time `json:"last_modified"` // 兼容字段
}

// SmsRepo 短信仓储接口
type SmsRepo interface {
	SendSms(ctx context.Context, phone, code string) error
	VerifySms(ctx context.Context, phone, code string) (bool, error)
}

// UserRepo 用户仓储接口
type UserRepo interface {
	CreateUser(ctx context.Context, user *UserBase) (*UserBase, error)
	GetUserByID(ctx context.Context, id uint64) (*UserBase, error)
	GetUserByPhone(ctx context.Context, phone string) (*UserBase, error)
	GetUserByPhoneAndPassword(ctx context.Context, phone, password string) (*UserBase, error)
	UpdateUser(ctx context.Context, user *UserBase) (*UserBase, error)
	DeleteUser(ctx context.Context, id uint64) error
	SmsRiskControl(ctx context.Context, phone string) error
	RealName(ctx context.Context, realName *RealName) (*RealName, error)
	UpdateUserStatus(ctx context.Context, user *UserBase) (*UserBase, error)
	CheckPhoneExists(ctx context.Context, phone string) (bool, error)
}

// MinioRepo 文件存储仓储接口
type MinioRepo interface {
	UploadFile(ctx context.Context, fileName string, reader io.Reader, size int64) (*FileInfo, error)
	DownloadFile(ctx context.Context, fileName string) (io.Reader, error)
	DeleteFile(ctx context.Context, fileName string) error
	SimpleUpload(ctx context.Context, fileName string, reader io.Reader) (*FileInfo, error)
	Delete(ctx context.Context, fileName string) error
	SearchFiles(ctx context.Context, keyword string, page, pageSize int32) ([]FileInfo, int32, error)
	ListFiles(ctx context.Context, prefix string, maxKeys int) ([]FileInfo, error)
	GetFileStats(ctx context.Context) (map[string]interface{}, error)
}
