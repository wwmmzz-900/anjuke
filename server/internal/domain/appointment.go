package domain

import (
	"context"
	"time"
)

// 预约状态枚举
type AppointmentStatus string

const (
	AppointmentStatusPending    AppointmentStatus = "pending"     // 待确认
	AppointmentStatusConfirmed  AppointmentStatus = "confirmed"   // 已确认
	AppointmentStatusInProgress AppointmentStatus = "in_progress" // 进行中
	AppointmentStatusCompleted  AppointmentStatus = "completed"   // 已完成
	AppointmentStatusCancelled  AppointmentStatus = "cancelled"   // 已取消
)

// 经纪人状态枚举
type RealtorStatus string

const (
	RealtorStatusOnline  RealtorStatus = "online"  // 在线
	RealtorStatusOffline RealtorStatus = "offline" // 离线
	RealtorStatusBusy    RealtorStatus = "busy"    // 忙碌
)

// AppointmentInfo 预约信息业务实体
type AppointmentInfo struct {
	ID                   uint64            `json:"id"`
	AppointmentCode      string            `json:"appointment_code"`
	UserID               int64             `json:"user_id"`
	StoreID              uint64            `json:"store_id"`
	RealtorID            *uint64           `json:"realtor_id,omitempty"`
	CustomerName         string            `json:"customer_name"`
	CustomerPhone        string            `json:"customer_phone"`
	AppointmentDate      time.Time         `json:"appointment_date"`
	StartTime            time.Time         `json:"start_time"`
	EndTime              time.Time         `json:"end_time"`
	DurationMinutes      int32             `json:"duration_minutes"`
	Requirements         string            `json:"requirements"`
	Status               AppointmentStatus `json:"status"`
	QueuePosition        int32             `json:"queue_position"`
	EstimatedWaitMinutes int32             `json:"estimated_wait_minutes"`
	CreatedAt            time.Time         `json:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at"`
	ConfirmedAt          *time.Time        `json:"confirmed_at,omitempty"`
	CompletedAt          *time.Time        `json:"completed_at,omitempty"`
	CancelledAt          *time.Time        `json:"cancelled_at,omitempty"`

	// 关联信息
	StoreInfo   *StoreBasicInfo   `json:"store_info,omitempty"`
	RealtorInfo *RealtorBasicInfo `json:"realtor_info,omitempty"`
}

// IsValid 验证预约信息是否有效
func (a *AppointmentInfo) IsValid() bool {
	return a.UserID != 0 &&
		a.StoreID != 0 &&
		a.CustomerName != "" &&
		a.CustomerPhone != "" &&
		!a.StartTime.IsZero() &&
		a.DurationMinutes > 0
}

// CanBeCancelled 判断预约是否可以取消
func (a *AppointmentInfo) CanBeCancelled() bool {
	return a.Status == AppointmentStatusPending || a.Status == AppointmentStatusConfirmed
}

// CanBeConfirmed 判断预约是否可以确认
func (a *AppointmentInfo) CanBeConfirmed() bool {
	return a.Status == AppointmentStatusPending
}

// IsInProgress 判断预约是否进行中
func (a *AppointmentInfo) IsInProgress() bool {
	return a.Status == AppointmentStatusInProgress
}

// IsCompleted 判断预约是否已完成
func (a *AppointmentInfo) IsCompleted() bool {
	return a.Status == AppointmentStatusCompleted
}

// StoreBasicInfo 门店基础信息
type StoreBasicInfo struct {
	ID            uint64  `json:"id"`
	Name          string  `json:"name"`
	Address       string  `json:"address"`
	Phone         string  `json:"phone"`
	BusinessHours string  `json:"business_hours"`
	Rating        float64 `json:"rating"`
	ReviewCount   int32   `json:"review_count"`
}

// RealtorBasicInfo 经纪人基础信息
type RealtorBasicInfo struct {
	ID       uint64 `json:"id"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
}

// RealtorStatusInfo 经纪人状态信息
type RealtorStatusInfo struct {
	RealtorID    uint64        `json:"realtor_id"`
	Status       RealtorStatus `json:"status"`
	CurrentLoad  int32         `json:"current_load"`
	MaxLoad      int32         `json:"max_load"`
	LastActiveAt time.Time     `json:"last_active_at"`
}

// TimeSlot 时间段信息
type TimeSlot struct {
	Date              string `json:"date"`
	StartTime         string `json:"start_time"`
	EndTime           string `json:"end_time"`
	Available         bool   `json:"available"`
	AvailableRealtors int32  `json:"available_realtors"`
	TotalCapacity     int32  `json:"total_capacity"`
	BookedCount       int32  `json:"booked_count"`
	QueueCount        int32  `json:"queue_count"`
}

// AppointmentLog 预约操作日志
type AppointmentLog struct {
	ID            uint64    `json:"id"`
	AppointmentID uint64    `json:"appointment_id"`
	Action        string    `json:"action"`
	OperatorType  string    `json:"operator_type"`
	OperatorID    *uint64   `json:"operator_id,omitempty"`
	OldStatus     *string   `json:"old_status,omitempty"`
	NewStatus     *string   `json:"new_status,omitempty"`
	Remark        string    `json:"remark"`
	CreatedAt     time.Time `json:"created_at"`
}

// StoreWorkingHours 门店工作时间
type StoreWorkingHours struct {
	ID        uint64    `json:"id"`
	StoreID   uint64    `json:"store_id"`
	DayOfWeek int32     `json:"day_of_week"`
	StartTime string    `json:"start_time"`
	EndTime   string    `json:"end_time"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RealtorWorkingHours 经纪人工作时间
type RealtorWorkingHours struct {
	ID        uint64    `json:"id"`
	RealtorID uint64    `json:"realtor_id"`
	StoreID   uint64    `json:"store_id"`
	DayOfWeek int32     `json:"day_of_week"`
	StartTime string    `json:"start_time"`
	EndTime   string    `json:"end_time"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AppointmentReview 预约评价
type AppointmentReview struct {
	ID                 uint64    `json:"id"`
	AppointmentID      uint64    `json:"appointment_id"`
	UserID             int64     `json:"user_id"`
	RealtorID          uint64    `json:"realtor_id"`
	StoreID            uint64    `json:"store_id"`
	ServiceRating      int32     `json:"service_rating"`
	ProfessionalRating int32     `json:"professional_rating"`
	ResponseRating     int32     `json:"response_rating"`
	OverallRating      float64   `json:"overall_rating"`
	ReviewContent      string    `json:"review_content"`
	CreatedAt          time.Time `json:"created_at"`
}

// AppointmentCode 预约码值对象
type AppointmentCode struct {
	Value string
}

// IsValid 验证预约码是否有效
func (ac AppointmentCode) IsValid() bool {
	return len(ac.Value) == 6
}

// UserID 用户ID值对象
type UserID struct {
	Value int64
}

// AppointmentSearchCriteria 预约搜索条件
type AppointmentSearchCriteria struct {
	UserID    *int64             `json:"user_id,omitempty"`
	StoreID   *uint64            `json:"store_id,omitempty"`
	RealtorID *uint64            `json:"realtor_id,omitempty"`
	Status    *AppointmentStatus `json:"status,omitempty"`
	Date      *time.Time         `json:"date,omitempty"`
	Page      int32              `json:"page"`
	PageSize  int32              `json:"page_size"`
}

// 仓储接口定义

// AppointmentRepo 预约仓储接口
type AppointmentRepo interface {
	// 预约管理
	CreateAppointment(ctx context.Context, appointment *AppointmentInfo) (*AppointmentInfo, error)
	GetAppointmentByID(ctx context.Context, id uint64) (*AppointmentInfo, error)
	GetAppointmentByCode(ctx context.Context, code string) (*AppointmentInfo, error)
	UpdateAppointment(ctx context.Context, appointment *AppointmentInfo) error
	DeleteAppointment(ctx context.Context, id uint64) error

	// 查询方法
	GetAppointmentsByUser(ctx context.Context, userID int64, page, pageSize int32) ([]*AppointmentInfo, int64, error)
	GetAppointmentsByRealtor(ctx context.Context, realtorID uint64, date time.Time) ([]*AppointmentInfo, error)
	GetAppointmentsByStore(ctx context.Context, storeID uint64, date time.Time) ([]*AppointmentInfo, error)
	GetUserRecentAppointment(ctx context.Context, userID int64, storeID uint64, minutes int) (*AppointmentInfo, error)

	// 冲突检查
	CheckRealtorTimeConflict(ctx context.Context, realtorID uint64, startTime, endTime time.Time) (bool, error)
	CheckUserTimeConflict(ctx context.Context, userID int64, startTime, endTime time.Time) (bool, error)

	// 排队管理
	GetQueueCount(ctx context.Context, storeID uint64, date, startTime time.Time) (int, error)
	GetQueuedAppointments(ctx context.Context, storeID uint64, date time.Time) ([]*AppointmentInfo, error)
	UpdateQueuePositions(ctx context.Context, storeID uint64, date time.Time) error

	// 时间段管理
	GetAvailableTimeSlots(ctx context.Context, storeID uint64, startDate time.Time, days int) ([]*TimeSlot, error)
}

// RealtorStatusRepo 经纪人状态仓储接口
type RealtorStatusRepo interface {
	FindByID(ctx context.Context, realtorID uint64) (*RealtorStatusInfo, error)
	FindOnlineByStore(ctx context.Context, storeID uint64) ([]*RealtorStatusInfo, error)
	Update(ctx context.Context, status *RealtorStatusInfo) error
	SetOnline(ctx context.Context, realtorID uint64) error
	SetOffline(ctx context.Context, realtorID uint64) error
}

// WorkingHoursRepo 工作时间仓储接口
type WorkingHoursRepo interface {
	FindStoreWorkingHours(ctx context.Context, storeID uint64) ([]*StoreWorkingHours, error)
	SaveStoreWorkingHours(ctx context.Context, hours *StoreWorkingHours) error
	UpdateStoreWorkingHours(ctx context.Context, hours *StoreWorkingHours) error
	DeleteStoreWorkingHours(ctx context.Context, id uint64) error

	FindRealtorWorkingHours(ctx context.Context, realtorID uint64) ([]*RealtorWorkingHours, error)
	SaveRealtorWorkingHours(ctx context.Context, hours *RealtorWorkingHours) error
	UpdateRealtorWorkingHours(ctx context.Context, hours *RealtorWorkingHours) error
	DeleteRealtorWorkingHours(ctx context.Context, id uint64) error
}

// AppointmentLogRepo 预约日志仓储接口
type AppointmentLogRepo interface {
	Save(ctx context.Context, appointmentID uint64, log *AppointmentLog) error
	FindByAppointmentID(ctx context.Context, appointmentID uint64) ([]*AppointmentLog, error)
}

// 领域服务接口

// AppointmentDomainService 预约领域服务接口
type AppointmentDomainService interface {
	// 业务规则验证
	ValidateAppointmentTime(ctx context.Context, storeID uint64, startTime, endTime time.Time) error
	CalculateQueuePosition(ctx context.Context, storeID uint64, appointmentDate time.Time) (int32, error)
	EstimateWaitTime(ctx context.Context, queuePosition int32) int32

	// 预约分配
	AssignRealtor(ctx context.Context, appointment *AppointmentInfo) (*uint64, error)
	ProcessQueue(ctx context.Context, storeID uint64, date time.Time) error
}
