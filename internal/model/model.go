package model

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"strings"
	"time"
)

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

// 房源排序类型枚举
type HouseSortType string

// 分页查询参数
type PaginationParams struct {
	Page     int    `json:"page"`      // 页码
	PageSize int    `json:"page_size"` // 每页大小
	SortBy   string `json:"sort_by"`   // 排序字段
	Order    string `json:"order"`     // 排序方向 asc/desc
}

// 分页查询结果
type PaginationResult struct {
	Total    int64 `json:"total"`     // 总记录数
	Page     int   `json:"page"`      // 当前页码
	PageSize int   `json:"page_size"` // 每页大小
	Pages    int   `json:"pages"`     // 总页数
}

// 房源查询过滤参数
type HouseFilterParams struct {
	Status   string  `json:"status"`    // 房源状态
	MinPrice float64 `json:"min_price"` // 最低价格
	MaxPrice float64 `json:"max_price"` // 最高价格
	MinArea  float32 `json:"min_area"`  // 最小面积
	MaxArea  float32 `json:"max_area"`  // 最大面积
	Layout   string  `json:"layout"`    // 户型
	Keyword  string  `json:"keyword"`   // 关键词搜索
	RegionId int64   `json:"region_id"` // 区域ID
}

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

// 用户状态枚举
type UserStatus int8

const (
	UserStatusInactive UserStatus = 0 // 未激活
	UserStatusActive   UserStatus = 1 // 活跃
	UserStatusDisabled UserStatus = 2 // 禁用
	UserStatusDeleted  UserStatus = 3 // 已删除
)

// 用户实名认证状态枚举
type UserRealStatus int8

const (
	UserRealStatusUnverified UserRealStatus = 0 // 未认证
	UserRealStatusVerified   UserRealStatus = 1 // 已认证
	UserRealStatusRejected   UserRealStatus = 2 // 认证被拒绝
)

// 用户性别枚举
type UserSex string

const (
	UserSexMale    UserSex = "male"    // 男性
	UserSexFemale  UserSex = "female"  // 女性
	UserSexUnknown UserSex = "unknown" // 未知
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
	ErrCodeUserNotFound      = 2001 // 用户不存在
	ErrCodeUserDisabled      = 2002 // 用户已禁用
	ErrCodeInvalidUserId     = 2003 // 无效的用户ID
	ErrCodeStatsQueryFailed  = 2004 // 统计数据查询失败
	ErrCodeProfileNotFound   = 2005 // 博主主页不存在
	ErrCodeProfileDisabled   = 2006 // 博主主页已禁用
	ErrCodeHouseQueryFailed  = 2007 // 房源查询失败
	ErrCodeInvalidPageParam  = 2008 // 无效的分页参数
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

// Favorite 房源收藏模型
type Favorite struct {
	Id        int64          `gorm:"column:id;type:bigint;comment:收藏ID;primaryKey;not null;" json:"id"`
	UserId    int64          `gorm:"column:user_id;type:bigint;comment:用户ID;not null;" json:"user_id"`
	HouseId   int64          `gorm:"column:house_id;type:bigint;comment:房源ID;not null;" json:"house_id"`
	CreatedAt time.Time      `gorm:"column:created_at;type:datetime;comment:收藏时间;not null;default:CURRENT_TIMESTAMP;" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;type:datetime;comment:删除时间;" json:"deleted_at"`
}

// TableName 指定表名
func (Favorite) TableName() string {
	return "favorite"
}

// IsValidFavorite 检查收藏记录是否有效
func (f *Favorite) IsValidFavorite() bool {
	return f.UserId > 0 && f.HouseId > 0
}

// UserBase 用户基础信息模型
type UserBase struct {
	UserId     int64          `gorm:"column:user_id;type:bigint;comment:用户ID;primaryKey;autoIncrement;not null;" json:"user_id"`
	Name       string         `gorm:"column:name;type:varchar(50);comment:用户昵称;not null;" json:"name"`
	RealName   string         `gorm:"column:real_name;type:varchar(30);comment:真实姓名;" json:"real_name"`
	Phone      string         `gorm:"column:phone;type:varchar(11);comment:手机号;not null;" json:"phone"`
	Email      string         `gorm:"column:email;type:varchar(32);comment:邮箱;" json:"email"`
	Password   string         `gorm:"column:password;type:varchar(32);comment:密码;not null;" json:"-"`
	Avatar     string         `gorm:"column:avatar;type:text;comment:头像URL;" json:"avatar"`
	RoleId     int64          `gorm:"column:role_id;type:bigint;comment:角色ID;not null;" json:"role_id"`
	Sex        UserSex        `gorm:"column:sex;type:enum('male','female','unknown');comment:性别;default:'unknown'" json:"sex"`
	RealStatus UserRealStatus `gorm:"column:real_status;type:tinyint;comment:实名认证状态;default:0" json:"real_status"`
	Status     UserStatus     `gorm:"column:status;type:tinyint;comment:用户状态;default:1" json:"status"`
	CreatedAt  time.Time      `gorm:"column:created_at;type:datetime;comment:创建时间;not null;default:CURRENT_TIMESTAMP;" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"column:updated_at;type:datetime;comment:更新时间;not null;default:CURRENT_TIMESTAMP;" json:"updated_at"`
	DeletedAt  *time.Time     `gorm:"column:deleted_at;type:datetime;comment:删除时间;" json:"deleted_at"`
}

// TableName 指定表名
func (UserBase) TableName() string {
	return "user_base"
}

// IsActive 检查用户是否为活跃状态
func (u *UserBase) IsActive() bool {
	return u.Status == UserStatusActive && u.DeletedAt == nil
}

// IsRealNameVerified 检查用户是否已实名认证
func (u *UserBase) IsRealNameVerified() bool {
	return u.RealStatus == UserRealStatusVerified
}

// GetMaskedPhone 获取脱敏后的手机号
func (u *UserBase) GetMaskedPhone() string {
	if len(u.Phone) != 11 {
		return u.Phone
	}
	return u.Phone[:3] + "****" + u.Phone[7:]
}

// GetDefaultAvatar 获取默认头像URL
func (u *UserBase) GetDefaultAvatar() string {
	if u.Avatar != "" {
		return u.Avatar
	}
	return "/static/images/default_avatar.png"
}

// GetDisplayName 获取显示名称
func (u *UserBase) GetDisplayName() string {
	if u.Name != "" {
		return u.Name
	}
	if u.RealName != "" {
		return u.RealName
	}
	return DefaultUserName
}

// IsDisabled 检查用户是否被禁用
func (u *UserBase) IsDisabled() bool {
	return u.Status == UserStatusDisabled
}

// IsDeleted 检查用户是否被删除
func (u *UserBase) IsDeleted() bool {
	return u.DeletedAt != nil || u.Status == UserStatusDeleted
}

// GetStatusText 获取用户状态文本
func (u *UserBase) GetStatusText() string {
	switch u.Status {
	case UserStatusInactive:
		return "未激活"
	case UserStatusActive:
		return "活跃"
	case UserStatusDisabled:
		return "已禁用"
	case UserStatusDeleted:
		return "已删除"
	default:
		return "未知状态"
	}
}

// GetRealStatusText 获取实名认证状态文本
func (u *UserBase) GetRealStatusText() string {
	switch u.RealStatus {
	case UserRealStatusUnverified:
		return "未认证"
	case UserRealStatusVerified:
		return "已认证"
	case UserRealStatusRejected:
		return "认证被拒绝"
	default:
		return "未知状态"
	}
}

// HouseStatistics 房源统计信息
type HouseStatistics struct {
	TotalCount  int64 `json:"total_count"`  // 总房源数
	ActiveCount int64 `json:"active_count"` // 活跃房源数
	RentedCount int64 `json:"rented_count"` // 已租房源数
}

// IsEmpty 检查统计是否为空
func (h *HouseStatistics) IsEmpty() bool {
	return h.TotalCount == 0 && h.ActiveCount == 0 && h.RentedCount == 0
}

// GetInactiveCount 获取非活跃房源数
func (h *HouseStatistics) GetInactiveCount() int64 {
	return h.TotalCount - h.ActiveCount - h.RentedCount
}

// GetActiveRate 获取活跃房源比例
func (h *HouseStatistics) GetActiveRate() float64 {
	if h.TotalCount == 0 {
		return 0.0
	}
	return float64(h.ActiveCount) / float64(h.TotalCount)
}

// FormatTotalCount 格式化总房源数显示
func (h *HouseStatistics) FormatTotalCount() string {
	return FormatStatNumber(h.TotalCount)
}

// FormatActiveCount 格式化活跃房源数显示
func (h *HouseStatistics) FormatActiveCount() string {
	return FormatStatNumber(h.ActiveCount)
}

// FormatRentedCount 格式化已租房源数显示
func (h *HouseStatistics) FormatRentedCount() string {
	return FormatStatNumber(h.RentedCount)
}

// FormatActiveRate 格式化活跃房源比例显示
func (h *HouseStatistics) FormatActiveRate() string {
	return FormatPercentage(h.GetActiveRate())
}

// InteractStatistics 互动统计信息
type InteractStatistics struct {
	TotalViews        int64   `json:"total_views"`        // 总浏览量
	TotalFavorites    int64   `json:"total_favorites"`    // 总收藏量
	TotalReservations int64   `json:"total_reservations"` // 总预约量
	ResponseRate      float64 `json:"response_rate"`      // 响应率
}

// IsEmpty 检查统计是否为空
func (i *InteractStatistics) IsEmpty() bool {
	return i.TotalViews == 0 && i.TotalFavorites == 0 && i.TotalReservations == 0
}

// FormatResponseRate 格式化响应率显示
func (i *InteractStatistics) FormatResponseRate() string {
	return FormatPercentage(i.ResponseRate)
}

// FormatViews 格式化浏览量显示（超过1万显示为1.2万格式）
func (i *InteractStatistics) FormatViews() string {
	return FormatStatNumber(i.TotalViews)
}

// FormatFavorites 格式化收藏量显示
func (i *InteractStatistics) FormatFavorites() string {
	return FormatStatNumber(i.TotalFavorites)
}

// FormatReservations 格式化预约量显示
func (i *InteractStatistics) FormatReservations() string {
	return FormatStatNumber(i.TotalReservations)
}

// GetEngagementRate 获取互动率（收藏+预约/浏览）
func (i *InteractStatistics) GetEngagementRate() float64 {
	if i.TotalViews == 0 {
		return 0.0
	}
	return float64(i.TotalFavorites+i.TotalReservations) / float64(i.TotalViews)
}

// BloggerProfile 博主主页信息聚合模型
type BloggerProfile struct {
	UserInfo      *UserBase           `json:"user_info"`      // 用户基础信息
	HouseStats    *HouseStatistics    `json:"house_stats"`    // 房源统计
	InteractStats *InteractStatistics `json:"interact_stats"` // 互动统计
	RecentHouses  []*House            `json:"recent_houses"`  // 最近房源
}

// IsValid 检查博主主页信息是否有效
func (b *BloggerProfile) IsValid() bool {
	return b.UserInfo != nil && b.UserInfo.IsActive()
}

// GetHouseCount 获取房源总数
func (b *BloggerProfile) GetHouseCount() int64 {
	if b.HouseStats == nil {
		return 0
	}
	return b.HouseStats.TotalCount
}

// GetActiveHouseCount 获取活跃房源数
func (b *BloggerProfile) GetActiveHouseCount() int64 {
	if b.HouseStats == nil {
		return 0
	}
	return b.HouseStats.ActiveCount
}

// GetTotalViews 获取总浏览量
func (b *BloggerProfile) GetTotalViews() int64 {
	if b.InteractStats == nil {
		return 0
	}
	return b.InteractStats.TotalViews
}

// formatLargeNumber 格式化大数字显示
func formatLargeNumber(num int64) string {
	if num >= 10000 {
		return fmt.Sprintf("%.1f万", float64(num)/10000)
	}
	return fmt.Sprintf("%d", num)
}

// ValidatePaginationParams 验证分页参数
func (p *PaginationParams) ValidatePaginationParams() {
	if p.Page <= 0 {
		p.Page = DefaultPage
	}
	if p.PageSize <= 0 {
		p.PageSize = DefaultPageSize
	}
	if p.PageSize > MaxPageSize {
		p.PageSize = MaxPageSize
	}
}

// GetOffset 获取偏移量
func (p *PaginationParams) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

// GetOrderClause 获取排序子句
func (p *PaginationParams) GetOrderClause() string {
	if p.SortBy == "" {
		return "created_at DESC" // 默认按创建时间降序
	}

	order := "DESC"
	if p.Order == "asc" {
		order = "ASC"
	}

	// 验证排序字段，防止SQL注入
	validSortFields := map[string]bool{
		"created_at": true,
		"updated_at": true,
		"price":      true,
		"area":       true,
		"title":      true,
	}

	if !validSortFields[p.SortBy] {
		return "created_at DESC"
	}

	return fmt.Sprintf("%s %s", p.SortBy, order)
}

// CalculatePages 计算总页数
func (r *PaginationResult) CalculatePages() {
	if r.PageSize <= 0 {
		r.Pages = 0
		return
	}
	r.Pages = int((r.Total + int64(r.PageSize) - 1) / int64(r.PageSize))
}

// HasNextPage 是否有下一页
func (r *PaginationResult) HasNextPage() bool {
	return r.Page < r.Pages
}

// HasPrevPage 是否有上一页
func (r *PaginationResult) HasPrevPage() bool {
	return r.Page > 1
}

// ValidateHouseFilterParams 验证房源过滤参数
func (f *HouseFilterParams) ValidateHouseFilterParams() {
	// 价格范围验证
	if f.MinPrice < 0 {
		f.MinPrice = 0
	}
	if f.MaxPrice < 0 {
		f.MaxPrice = 0
	}
	if f.MaxPrice > 0 && f.MinPrice > f.MaxPrice {
		f.MinPrice, f.MaxPrice = f.MaxPrice, f.MinPrice
	}

	// 面积范围验证
	if f.MinArea < 0 {
		f.MinArea = 0
	}
	if f.MaxArea < 0 {
		f.MaxArea = 0
	}
	if f.MaxArea > 0 && f.MinArea > f.MaxArea {
		f.MinArea, f.MaxArea = f.MaxArea, f.MinArea
	}
}

// IsEmpty 检查过滤参数是否为空
func (f *HouseFilterParams) IsEmpty() bool {
	return f.Status == "" && f.MinPrice == 0 && f.MaxPrice == 0 &&
		f.MinArea == 0 && f.MaxArea == 0 && f.Layout == "" &&
		f.Keyword == "" && f.RegionId == 0
}

// FormatStatNumber 格式化统计数字显示，0时显示"0"
func FormatStatNumber(num int64) string {
	if num == 0 {
		return "0"
	}
	return formatLargeNumber(num)
}

// FormatStatFloat 格式化统计浮点数显示
func FormatStatFloat(num float64, precision int) string {
	if num == 0 {
		return "0"
	}
	format := fmt.Sprintf("%%.%df", precision)
	return fmt.Sprintf(format, num)
}

// FormatPercentage 格式化百分比显示
func FormatPercentage(rate float64) string {
	if rate == 0 {
		return "0%"
	}
	return fmt.Sprintf("%.1f%%", rate*100)
}

// BloggerProfileAccessLog 博主主页访问日志模型
type BloggerProfileAccessLog struct {
	ID           int64     `gorm:"column:id;type:bigint;comment:日志ID;primaryKey;autoIncrement;not null;" json:"id"`
	BloggerID    int64     `gorm:"column:blogger_id;type:bigint;comment:被访问的博主ID;not null;index" json:"blogger_id"`
	VisitorID    int64     `gorm:"column:visitor_id;type:bigint;comment:访问者ID;default:0;index" json:"visitor_id"`
	VisitorIP    string    `gorm:"column:visitor_ip;type:varchar(45);comment:访问者IP地址;not null" json:"visitor_ip"`
	UserAgent    string    `gorm:"column:user_agent;type:varchar(500);comment:用户代理信息" json:"user_agent"`
	Referer      string    `gorm:"column:referer;type:varchar(500);comment:来源页面" json:"referer"`
	RequestPath  string    `gorm:"column:request_path;type:varchar(255);comment:请求路径;not null" json:"request_path"`
	RequestMethod string   `gorm:"column:request_method;type:varchar(10);comment:请求方法;not null" json:"request_method"`
	ResponseTime int64     `gorm:"column:response_time;type:bigint;comment:响应时间(毫秒);default:0" json:"response_time"`
	StatusCode   int       `gorm:"column:status_code;type:int;comment:HTTP状态码;not null" json:"status_code"`
	DeviceType   string    `gorm:"column:device_type;type:varchar(20);comment:设备类型(mobile/desktop/tablet);default:'unknown'" json:"device_type"`
	Platform     string    `gorm:"column:platform;type:varchar(50);comment:操作系统平台" json:"platform"`
	Browser      string    `gorm:"column:browser;type:varchar(50);comment:浏览器信息" json:"browser"`
	SessionID    string    `gorm:"column:session_id;type:varchar(100);comment:会话ID" json:"session_id"`
	CreatedAt    time.Time `gorm:"column:created_at;type:datetime;comment:访问时间;not null;default:CURRENT_TIMESTAMP;index" json:"created_at"`
}

// TableName 指定表名
func (BloggerProfileAccessLog) TableName() string {
	return "blogger_profile_access_logs"
}

// IsValidLog 检查访问日志是否有效
func (l *BloggerProfileAccessLog) IsValidLog() bool {
	return l.BloggerID > 0 && l.VisitorIP != "" && l.RequestPath != ""
}

// IsSuccessfulAccess 检查是否为成功访问
func (l *BloggerProfileAccessLog) IsSuccessfulAccess() bool {
	return l.StatusCode >= 200 && l.StatusCode < 300
}

// GetDeviceTypeFromUserAgent 从User-Agent解析设备类型
func (l *BloggerProfileAccessLog) GetDeviceTypeFromUserAgent() string {
	if l.UserAgent == "" {
		return "unknown"
	}
	
	userAgent := strings.ToLower(l.UserAgent)
	
	// 检查移动设备
	mobileKeywords := []string{"mobile", "android", "iphone", "ipad", "ipod", "blackberry", "windows phone"}
	for _, keyword := range mobileKeywords {
		if strings.Contains(userAgent, keyword) {
			if strings.Contains(userAgent, "ipad") || strings.Contains(userAgent, "tablet") {
				return "tablet"
			}
			return "mobile"
		}
	}
	
	return "desktop"
}

// GetBrowserFromUserAgent 从User-Agent解析浏览器信息
func (l *BloggerProfileAccessLog) GetBrowserFromUserAgent() string {
	if l.UserAgent == "" {
		return "unknown"
	}
	
	userAgent := strings.ToLower(l.UserAgent)
	
	browsers := map[string]string{
		"chrome":  "Chrome",
		"firefox": "Firefox",
		"safari":  "Safari",
		"edge":    "Edge",
		"opera":   "Opera",
		"ie":      "Internet Explorer",
	}
	
	for keyword, browser := range browsers {
		if strings.Contains(userAgent, keyword) {
			return browser
		}
	}
	
	return "unknown"
}

// GetPlatformFromUserAgent 从User-Agent解析操作系统平台
func (l *BloggerProfileAccessLog) GetPlatformFromUserAgent() string {
	if l.UserAgent == "" {
		return "unknown"
	}
	
	userAgent := strings.ToLower(l.UserAgent)
	
	platforms := map[string]string{
		"windows":   "Windows",
		"mac":       "macOS",
		"linux":     "Linux",
		"android":   "Android",
		"ios":       "iOS",
		"iphone":    "iOS",
		"ipad":      "iOS",
	}
	
	for keyword, platform := range platforms {
		if strings.Contains(userAgent, keyword) {
			return platform
		}
	}
	
	return "unknown"
}

// BloggerProfileAccessStats 博主主页访问统计模型
type BloggerProfileAccessStats struct {
	BloggerID      int64     `gorm:"column:blogger_id;type:bigint;comment:博主ID;primaryKey;not null;" json:"blogger_id"`
	TotalViews     int64     `gorm:"column:total_views;type:bigint;comment:总访问量;default:0" json:"total_views"`
	TodayViews     int64     `gorm:"column:today_views;type:bigint;comment:今日访问量;default:0" json:"today_views"`
	WeekViews      int64     `gorm:"column:week_views;type:bigint;comment:本周访问量;default:0" json:"week_views"`
	MonthViews     int64     `gorm:"column:month_views;type:bigint;comment:本月访问量;default:0" json:"month_views"`
	UniqueVisitors int64     `gorm:"column:unique_visitors;type:bigint;comment:独立访客数;default:0" json:"unique_visitors"`
	AvgResponseTime float64  `gorm:"column:avg_response_time;type:decimal(10,2);comment:平均响应时间(毫秒);default:0" json:"avg_response_time"`
	LastAccessAt   *time.Time `gorm:"column:last_access_at;type:datetime;comment:最后访问时间" json:"last_access_at"`
	CreatedAt      time.Time `gorm:"column:created_at;type:datetime;comment:创建时间;not null;default:CURRENT_TIMESTAMP;" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;type:datetime;comment:更新时间;not null;default:CURRENT_TIMESTAMP;" json:"updated_at"`
}

// TableName 指定表名
func (BloggerProfileAccessStats) TableName() string {
	return "blogger_profile_access_stats"
}

// IsEmpty 检查统计是否为空
func (s *BloggerProfileAccessStats) IsEmpty() bool {
	return s.TotalViews == 0 && s.UniqueVisitors == 0
}

// FormatTotalViews 格式化总访问量显示
func (s *BloggerProfileAccessStats) FormatTotalViews() string {
	return FormatStatNumber(s.TotalViews)
}

// FormatUniqueVisitors 格式化独立访客数显示
func (s *BloggerProfileAccessStats) FormatUniqueVisitors() string {
	return FormatStatNumber(s.UniqueVisitors)
}

// GetViewsGrowthRate 计算访问量增长率（今日相比昨日）
func (s *BloggerProfileAccessStats) GetViewsGrowthRate(yesterdayViews int64) float64 {
	if yesterdayViews == 0 {
		if s.TodayViews > 0 {
			return 1.0 // 100%增长
		}
		return 0.0
	}
	return float64(s.TodayViews-yesterdayViews) / float64(yesterdayViews)
}

// AccessLogRequest 访问日志请求参数
type AccessLogRequest struct {
	BloggerID     int64  `json:"blogger_id"`
	VisitorID     int64  `json:"visitor_id"`
	VisitorIP     string `json:"visitor_ip"`
	UserAgent     string `json:"user_agent"`
	Referer       string `json:"referer"`
	RequestPath   string `json:"request_path"`
	RequestMethod string `json:"request_method"`
	SessionID     string `json:"session_id"`
}

// Validate 验证访问日志请求参数
func (r *AccessLogRequest) Validate() error {
	if r.BloggerID <= 0 {
		return errors.New("博主ID不能为空或小于等于0")
	}
	if r.VisitorIP == "" {
		return errors.New("访问者IP不能为空")
	}
	if r.RequestPath == "" {
		return errors.New("请求路径不能为空")
	}
	if r.RequestMethod == "" {
		r.RequestMethod = "GET" // 默认为GET请求
	}
	return nil
}

// ToAccessLog 转换为访问日志模型
func (r *AccessLogRequest) ToAccessLog() *BloggerProfileAccessLog {
	log := &BloggerProfileAccessLog{
		BloggerID:     r.BloggerID,
		VisitorID:     r.VisitorID,
		VisitorIP:     r.VisitorIP,
		UserAgent:     r.UserAgent,
		Referer:       r.Referer,
		RequestPath:   r.RequestPath,
		RequestMethod: r.RequestMethod,
		SessionID:     r.SessionID,
		StatusCode:    200, // 默认成功状态码
		CreatedAt:     time.Now(),
	}
	
	// 自动解析设备信息
	log.DeviceType = log.GetDeviceTypeFromUserAgent()
	log.Browser = log.GetBrowserFromUserAgent()
	log.Platform = log.GetPlatformFromUserAgent()
	
	return log
}
