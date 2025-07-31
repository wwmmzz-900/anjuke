package data

import (
	"context"
	"fmt"
	"time"

	"anjuke/server/internal/domain"

	"gorm.io/gorm"
)

// AppointmentModel 预约表模型 - 与领域对象匹配
type AppointmentModel struct {
	ID                   uint64     `gorm:"primaryKey;autoIncrement;comment:预约ID，主键，自增" json:"id"`
	AppointmentCode      string     `gorm:"type:varchar(6);uniqueIndex;not null;comment:预约码，6位唯一编码，不能为空" json:"appointment_code"`
	UserID               int64      `gorm:"not null;index;comment:用户ID，不能为空，索引" json:"user_id"`
	StoreID              uint64     `gorm:"not null;index;comment:门店ID，不能为空，索引" json:"store_id"`
	RealtorID            *uint64    `gorm:"index;comment:经纪人ID，可为空，索引" json:"realtor_id"`
	CustomerName         string     `gorm:"type:varchar(50);not null;comment:客户姓名，不能为空" json:"customer_name"`
	CustomerPhone        string     `gorm:"type:varchar(20);not null;comment:客户电话，不能为空" json:"customer_phone"`
	AppointmentDate      time.Time  `gorm:"not null;index;comment:预约日期，不能为空，索引" json:"appointment_date"`
	StartTime            time.Time  `gorm:"not null;comment:开始时间，不能为空" json:"start_time"`
	EndTime              time.Time  `gorm:"not null;comment:结束时间，不能为空" json:"end_time"`
	DurationMinutes      int32      `gorm:"not null;comment:预约时长（分钟），不能为空" json:"duration_minutes"`
	Requirements         string     `gorm:"type:text;comment:预约需求描述" json:"requirements"`
	Status               string     `gorm:"type:varchar(20);not null;default:'pending';comment:预约状态，不能为空，默认为pending" json:"status"`
	QueuePosition        int32      `gorm:"default:0;comment:排队位置，默认为0" json:"queue_position"`
	EstimatedWaitMinutes int32      `gorm:"default:0;comment:预计等待时间（分钟），默认为0" json:"estimated_wait_minutes"`
	CreatedAt            time.Time  `gorm:"autoCreateTime;comment:创建时间，自动生成" json:"created_at"`
	UpdatedAt            time.Time  `gorm:"autoUpdateTime;comment:更新时间，自动更新" json:"updated_at"`
	ConfirmedAt          *time.Time `gorm:"comment:确认时间，可为空" json:"confirmed_at"`
	CompletedAt          *time.Time `gorm:"comment:完成时间，可为空" json:"completed_at"`
	CancelledAt          *time.Time `gorm:"comment:取消时间，可为空" json:"cancelled_at"`
}

func (AppointmentModel) TableName() string {
	return "appointments" // 预约表
}

// ToAppointmentInfo 转换为领域对象
func (a *AppointmentModel) ToAppointmentInfo() *domain.AppointmentInfo {
	return &domain.AppointmentInfo{
		ID:                   a.ID,
		AppointmentCode:      a.AppointmentCode,
		UserID:               a.UserID,
		StoreID:              a.StoreID,
		RealtorID:            a.RealtorID,
		CustomerName:         a.CustomerName,
		CustomerPhone:        a.CustomerPhone,
		AppointmentDate:      a.AppointmentDate,
		StartTime:            a.StartTime,
		EndTime:              a.EndTime,
		DurationMinutes:      a.DurationMinutes,
		Requirements:         a.Requirements,
		Status:               domain.AppointmentStatus(a.Status),
		QueuePosition:        a.QueuePosition,
		EstimatedWaitMinutes: a.EstimatedWaitMinutes,
		CreatedAt:            a.CreatedAt,
		UpdatedAt:            a.UpdatedAt,
		ConfirmedAt:          a.ConfirmedAt,
		CompletedAt:          a.CompletedAt,
		CancelledAt:          a.CancelledAt,
	}
}

// AppointmentLogModel 预约日志表模型
type AppointmentLogModel struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement;comment:日志ID，主键，自增" json:"id"`
	AppointmentID uint64    `gorm:"not null;index;comment:预约ID，不能为空，索引" json:"appointment_id"`
	Action        string    `gorm:"type:varchar(50);not null;comment:操作类型，不能为空" json:"action"`
	OperatorType  string    `gorm:"type:varchar(20);not null;comment:操作者类型，不能为空" json:"operator_type"`
	OperatorID    *uint64   `gorm:"comment:操作者ID，可为空" json:"operator_id"`
	OldStatus     *string   `gorm:"type:varchar(20);comment:旧状态，可为空" json:"old_status"`
	NewStatus     *string   `gorm:"type:varchar(20);comment:新状态，可为空" json:"new_status"`
	Remark        string    `gorm:"type:text;comment:备注信息" json:"remark"`
	CreatedAt     time.Time `gorm:"autoCreateTime;comment:创建时间，自动生成" json:"created_at"`
}

func (AppointmentLogModel) TableName() string {
	return "appointment_logs" // 预约日志表
}

// ToAppointmentLog 转换为领域对象
func (a *AppointmentLogModel) ToAppointmentLog() *domain.AppointmentLog {
	return &domain.AppointmentLog{
		ID:            a.ID,
		AppointmentID: a.AppointmentID,
		Action:        a.Action,
		OperatorType:  a.OperatorType,
		OperatorID:    a.OperatorID,
		OldStatus:     a.OldStatus,
		NewStatus:     a.NewStatus,
		Remark:        a.Remark,
		CreatedAt:     a.CreatedAt,
	}
}

// StoreWorkingHoursModel 门店工作时间表模型
type StoreWorkingHoursModel struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement;comment:工作时间ID，主键，自增" json:"id"`
	StoreID   uint64    `gorm:"not null;index;comment:门店ID，不能为空，索引" json:"store_id"`
	DayOfWeek int32     `gorm:"not null;comment:星期几(1-7)" json:"day_of_week"`
	StartTime string    `gorm:"type:time;not null;comment:开始时间" json:"start_time"`
	EndTime   string    `gorm:"type:time;not null;comment:结束时间" json:"end_time"`
	IsActive  bool      `gorm:"default:true;comment:是否激活，默认为true" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime;comment:创建时间，自动生成" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;comment:更新时间，自动更新" json:"updated_at"`
}

func (StoreWorkingHoursModel) TableName() string {
	return "store_working_hours" // 门店工作时间表
}

// ToStoreWorkingHours 转换为领域对象
func (s *StoreWorkingHoursModel) ToStoreWorkingHours() *domain.StoreWorkingHours {
	return &domain.StoreWorkingHours{
		ID:        s.ID,
		StoreID:   s.StoreID,
		DayOfWeek: s.DayOfWeek,
		StartTime: s.StartTime,
		EndTime:   s.EndTime,
		IsActive:  s.IsActive,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

// RealtorWorkingHoursModel 经纪人工作时间表模型
type RealtorWorkingHoursModel struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement;comment:工作时间ID，主键，自增" json:"id"`
	RealtorID uint64    `gorm:"not null;index;comment:经纪人ID，不能为空，索引" json:"realtor_id"`
	StoreID   uint64    `gorm:"not null;index;comment:门店ID，不能为空，索引" json:"store_id"`
	DayOfWeek int32     `gorm:"not null;comment:星期几(1-7)" json:"day_of_week"`
	StartTime string    `gorm:"type:time;not null;comment:开始时间" json:"start_time"`
	EndTime   string    `gorm:"type:time;not null;comment:结束时间" json:"end_time"`
	IsActive  bool      `gorm:"default:true;comment:是否激活，默认为true" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime;comment:创建时间，自动生成" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;comment:更新时间，自动更新" json:"updated_at"`
}

func (RealtorWorkingHoursModel) TableName() string {
	return "realtor_working_hours" // 经纪人工作时间表
}

// ToRealtorWorkingHours 转换为领域对象
func (r *RealtorWorkingHoursModel) ToRealtorWorkingHours() *domain.RealtorWorkingHours {
	return &domain.RealtorWorkingHours{
		ID:        r.ID,
		RealtorID: r.RealtorID,
		StoreID:   r.StoreID,
		DayOfWeek: r.DayOfWeek,
		StartTime: r.StartTime,
		EndTime:   r.EndTime,
		IsActive:  r.IsActive,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

// RealtorStatusModel 经纪人状态表模型
type RealtorStatusModel struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement;comment:状态ID，主键，自增" json:"id"`
	RealtorID    uint64    `gorm:"not null;uniqueIndex;comment:经纪人ID，不能为空，唯一索引" json:"realtor_id"`
	Status       string    `gorm:"type:varchar(20);default:'offline';comment:状态，默认为offline" json:"status"`
	CurrentLoad  int32     `gorm:"default:0;comment:当前负载数量，默认为0" json:"current_load"`
	MaxLoad      int32     `gorm:"default:10;comment:最大负载数量，默认为10" json:"max_load"`
	LastActiveAt time.Time `gorm:"autoUpdateTime;comment:最后活跃时间，自动更新" json:"last_active_at"`
}

func (RealtorStatusModel) TableName() string {
	return "realtor_status" // 经纪人状态表
}

// ToRealtorStatusInfo 转换为领域对象
func (r *RealtorStatusModel) ToRealtorStatusInfo() *domain.RealtorStatusInfo {
	return &domain.RealtorStatusInfo{
		RealtorID:    r.RealtorID,
		Status:       domain.RealtorStatus(r.Status),
		CurrentLoad:  r.CurrentLoad,
		MaxLoad:      r.MaxLoad,
		LastActiveAt: r.LastActiveAt,
	}
}

// AppointmentReviewModel 预约评价表模型
type AppointmentReviewModel struct {
	ID                 uint64    `gorm:"primaryKey;autoIncrement;comment:评价ID，主键，自增" json:"id"`
	AppointmentID      uint64    `gorm:"not null;uniqueIndex;comment:预约ID，不能为空，唯一索引" json:"appointment_id"`
	UserID             int64     `gorm:"not null;index;comment:用户ID，不能为空，索引" json:"user_id"`
	RealtorID          uint64    `gorm:"not null;index;comment:经纪人ID，不能为空，索引" json:"realtor_id"`
	StoreID            uint64    `gorm:"not null;index;comment:门店ID，不能为空，索引" json:"store_id"`
	ServiceRating      int32     `gorm:"not null;comment:服务评分，不能为空" json:"service_rating"`
	ProfessionalRating int32     `gorm:"not null;comment:专业度评分，不能为空" json:"professional_rating"`
	ResponseRating     int32     `gorm:"not null;comment:响应速度评分，不能为空" json:"response_rating"`
	OverallRating      float64   `gorm:"not null;comment:总体评分，不能为空" json:"overall_rating"`
	ReviewContent      string    `gorm:"type:text;comment:评价内容" json:"review_content"`
	CreatedAt          time.Time `gorm:"autoCreateTime;comment:创建时间，自动生成" json:"created_at"`
}

func (AppointmentReviewModel) TableName() string {
	return "appointment_reviews" // 预约评价表
}

// ToAppointmentReview 转换为领域对象
func (a *AppointmentReviewModel) ToAppointmentReview() *domain.AppointmentReview {
	return &domain.AppointmentReview{
		ID:                 a.ID,
		AppointmentID:      a.AppointmentID,
		UserID:             a.UserID,
		RealtorID:          a.RealtorID,
		StoreID:            a.StoreID,
		ServiceRating:      a.ServiceRating,
		ProfessionalRating: a.ProfessionalRating,
		ResponseRating:     a.ResponseRating,
		OverallRating:      a.OverallRating,
		ReviewContent:      a.ReviewContent,
		CreatedAt:          a.CreatedAt,
	}
}

// AppointmentDBRepo 预约数据库仓储实现
type AppointmentDBRepo struct {
	data *Data
}

// NewAppointmentDBRepo 创建预约数据库仓储
func NewAppointmentDBRepo(data *Data) domain.AppointmentRepo {
	return &AppointmentDBRepo{
		data: data,
	}
}

// CreateAppointment 创建预约
func (r *AppointmentDBRepo) CreateAppointment(ctx context.Context, appointment *domain.AppointmentInfo) (*domain.AppointmentInfo, error) {
	model := &AppointmentModel{
		AppointmentCode:      appointment.AppointmentCode,
		UserID:               appointment.UserID,
		StoreID:              appointment.StoreID,
		RealtorID:            appointment.RealtorID,
		CustomerName:         appointment.CustomerName,
		CustomerPhone:        appointment.CustomerPhone,
		AppointmentDate:      appointment.AppointmentDate,
		StartTime:            appointment.StartTime,
		EndTime:              appointment.EndTime,
		DurationMinutes:      appointment.DurationMinutes,
		Requirements:         appointment.Requirements,
		Status:               string(appointment.Status),
		QueuePosition:        appointment.QueuePosition,
		EstimatedWaitMinutes: appointment.EstimatedWaitMinutes,
	}

	if err := r.data.db.WithContext(ctx).Create(model).Error; err != nil {
		return nil, err
	}

	return model.ToAppointmentInfo(), nil
}

// GetAppointmentByID 根据ID获取预约
func (r *AppointmentDBRepo) GetAppointmentByID(ctx context.Context, id uint64) (*domain.AppointmentInfo, error) {
	var model AppointmentModel
	if err := r.data.db.WithContext(ctx).First(&model, id).Error; err != nil {
		return nil, err
	}
	return model.ToAppointmentInfo(), nil
}

// GetAppointmentByCode 根据预约码获取预约
func (r *AppointmentDBRepo) GetAppointmentByCode(ctx context.Context, code string) (*domain.AppointmentInfo, error) {
	var model AppointmentModel
	if err := r.data.db.WithContext(ctx).Where("appointment_code = ?", code).First(&model).Error; err != nil {
		return nil, err
	}
	return model.ToAppointmentInfo(), nil
}

// UpdateAppointment 更新预约
func (r *AppointmentDBRepo) UpdateAppointment(ctx context.Context, appointment *domain.AppointmentInfo) error {
	updates := map[string]interface{}{
		"realtor_id":             appointment.RealtorID,
		"appointment_date":       appointment.AppointmentDate,
		"start_time":             appointment.StartTime,
		"end_time":               appointment.EndTime,
		"duration_minutes":       appointment.DurationMinutes,
		"requirements":           appointment.Requirements,
		"status":                 string(appointment.Status),
		"queue_position":         appointment.QueuePosition,
		"estimated_wait_minutes": appointment.EstimatedWaitMinutes,
		"updated_at":             time.Now(),
	}

	// 根据状态设置时间戳
	switch appointment.Status {
	case domain.AppointmentStatusConfirmed:
		if appointment.ConfirmedAt != nil {
			updates["confirmed_at"] = appointment.ConfirmedAt
		} else {
			now := time.Now()
			updates["confirmed_at"] = &now
		}
	case domain.AppointmentStatusCompleted:
		if appointment.CompletedAt != nil {
			updates["completed_at"] = appointment.CompletedAt
		} else {
			now := time.Now()
			updates["completed_at"] = &now
		}
	case domain.AppointmentStatusCancelled:
		if appointment.CancelledAt != nil {
			updates["cancelled_at"] = appointment.CancelledAt
		} else {
			now := time.Now()
			updates["cancelled_at"] = &now
		}
	}

	return r.data.db.WithContext(ctx).Model(&AppointmentModel{}).
		Where("id = ?", appointment.ID).
		Updates(updates).Error
}

// DeleteAppointment 删除预约
func (r *AppointmentDBRepo) DeleteAppointment(ctx context.Context, id uint64) error {
	return r.data.db.WithContext(ctx).Delete(&AppointmentModel{}, id).Error
}

// GetAppointmentsByUser 获取用户的预约列表
func (r *AppointmentDBRepo) GetAppointmentsByUser(ctx context.Context, userID int64, page, pageSize int32) ([]*domain.AppointmentInfo, int64, error) {
	var models []AppointmentModel
	var total int64

	// 计算总数
	if err := r.data.db.WithContext(ctx).Model(&AppointmentModel{}).
		Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := r.data.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(int(offset)).Limit(int(pageSize)).
		Find(&models).Error; err != nil {
		return nil, 0, err
	}

	// 转换为领域对象
	appointments := make([]*domain.AppointmentInfo, len(models))
	for i, model := range models {
		appointments[i] = model.ToAppointmentInfo()
	}

	return appointments, total, nil
}

// GetAppointmentsByRealtor 获取经纪人指定日期的预约列表
func (r *AppointmentDBRepo) GetAppointmentsByRealtor(ctx context.Context, realtorID uint64, date time.Time) ([]*domain.AppointmentInfo, error) {
	var models []AppointmentModel
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	if err := r.data.db.WithContext(ctx).
		Where("realtor_id = ? AND appointment_date >= ? AND appointment_date < ?", realtorID, startOfDay, endOfDay).
		Order("start_time ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	appointments := make([]*domain.AppointmentInfo, len(models))
	for i, model := range models {
		appointments[i] = model.ToAppointmentInfo()
	}

	return appointments, nil
}

// GetAppointmentsByStore 获取门店指定日期的预约列表
func (r *AppointmentDBRepo) GetAppointmentsByStore(ctx context.Context, storeID uint64, date time.Time) ([]*domain.AppointmentInfo, error) {
	var models []AppointmentModel
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	if err := r.data.db.WithContext(ctx).
		Where("store_id = ? AND appointment_date >= ? AND appointment_date < ?", storeID, startOfDay, endOfDay).
		Order("start_time ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	appointments := make([]*domain.AppointmentInfo, len(models))
	for i, model := range models {
		appointments[i] = model.ToAppointmentInfo()
	}

	return appointments, nil
}

// GetUserRecentAppointment 获取用户最近的预约
func (r *AppointmentDBRepo) GetUserRecentAppointment(ctx context.Context, userID int64, storeID uint64, minutes int) (*domain.AppointmentInfo, error) {
	var model AppointmentModel
	cutoffTime := time.Now().Add(-time.Duration(minutes) * time.Minute)

	if err := r.data.db.WithContext(ctx).
		Where("user_id = ? AND store_id = ? AND created_at >= ?", userID, storeID, cutoffTime).
		Order("created_at DESC").
		First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return model.ToAppointmentInfo(), nil
}

// CheckRealtorTimeConflict 检查经纪人时间冲突
func (r *AppointmentDBRepo) CheckRealtorTimeConflict(ctx context.Context, realtorID uint64, startTime, endTime time.Time) (bool, error) {
	var count int64
	if err := r.data.db.WithContext(ctx).Model(&AppointmentModel{}).
		Where("realtor_id = ? AND status NOT IN (?, ?) AND ((start_time < ? AND end_time > ?) OR (start_time < ? AND end_time > ?) OR (start_time >= ? AND end_time <= ?))",
			realtorID, string(domain.AppointmentStatusCancelled), string(domain.AppointmentStatusCompleted),
			startTime, startTime, endTime, endTime, startTime, endTime).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// CheckUserTimeConflict 检查用户时间冲突
func (r *AppointmentDBRepo) CheckUserTimeConflict(ctx context.Context, userID int64, startTime, endTime time.Time) (bool, error) {
	var count int64
	if err := r.data.db.WithContext(ctx).Model(&AppointmentModel{}).
		Where("user_id = ? AND status NOT IN (?, ?) AND ((start_time < ? AND end_time > ?) OR (start_time < ? AND end_time > ?) OR (start_time >= ? AND end_time <= ?))",
			userID, string(domain.AppointmentStatusCancelled), string(domain.AppointmentStatusCompleted),
			startTime, startTime, endTime, endTime, startTime, endTime).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetQueueCount 获取排队数量
func (r *AppointmentDBRepo) GetQueueCount(ctx context.Context, storeID uint64, date, startTime time.Time) (int, error) {
	var count int64
	if err := r.data.db.WithContext(ctx).Model(&AppointmentModel{}).
		Where("store_id = ? AND appointment_date = ? AND start_time = ? AND status = ?",
			storeID, date.Format("2006-01-02"), startTime.Format("15:04:05"), string(domain.AppointmentStatusPending)).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

// GetQueuedAppointments 获取排队中的预约
func (r *AppointmentDBRepo) GetQueuedAppointments(ctx context.Context, storeID uint64, date time.Time) ([]*domain.AppointmentInfo, error) {
	var models []AppointmentModel
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	if err := r.data.db.WithContext(ctx).
		Where("store_id = ? AND appointment_date >= ? AND appointment_date < ? AND status = ?",
			storeID, startOfDay, endOfDay, string(domain.AppointmentStatusPending)).
		Order("queue_position ASC, created_at ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	appointments := make([]*domain.AppointmentInfo, len(models))
	for i, model := range models {
		appointments[i] = model.ToAppointmentInfo()
	}

	return appointments, nil
}

// UpdateQueuePositions 更新排队位置
func (r *AppointmentDBRepo) UpdateQueuePositions(ctx context.Context, storeID uint64, date time.Time) error {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	// 获取所有排队中的预约，按创建时间排序
	var models []AppointmentModel
	if err := r.data.db.WithContext(ctx).
		Where("store_id = ? AND appointment_date >= ? AND appointment_date < ? AND status = ?",
			storeID, startOfDay, endOfDay, string(domain.AppointmentStatusPending)).
		Order("created_at ASC").
		Find(&models).Error; err != nil {
		return err
	}

	// 更新排队位置
	for i, model := range models {
		if err := r.data.db.WithContext(ctx).Model(&AppointmentModel{}).
			Where("id = ?", model.ID).
			Update("queue_position", i+1).Error; err != nil {
			return err
		}
	}

	return nil
}

// GetAvailableTimeSlots 获取可用时间段
func (r *AppointmentDBRepo) GetAvailableTimeSlots(ctx context.Context, storeID uint64, startDate time.Time, days int) ([]*domain.TimeSlot, error) {
	// 这里需要根据门店工作时间和已有预约来计算可用时间段
	// 简化实现，实际应该更复杂
	var timeSlots []*domain.TimeSlot

	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i)
		dateStr := date.Format("2006-01-02")

		// 获取当天的预约数量
		var bookedCount int64
		r.data.db.WithContext(ctx).Model(&AppointmentModel{}).
			Where("store_id = ? AND appointment_date = ?", storeID, dateStr).
			Count(&bookedCount)

		// 简化的时间段生成（实际应该根据工作时间）
		for hour := 9; hour <= 17; hour++ {
			startTime := fmt.Sprintf("%02d:00", hour)
			endTime := fmt.Sprintf("%02d:00", hour+1)

			timeSlot := &domain.TimeSlot{
				Date:              dateStr,
				StartTime:         startTime,
				EndTime:           endTime,
				Available:         bookedCount < 10, // 简化逻辑
				AvailableRealtors: 3,                // 简化逻辑
				TotalCapacity:     10,
				BookedCount:       int32(bookedCount),
				QueueCount:        0,
			}
			timeSlots = append(timeSlots, timeSlot)
		}
	}

	return timeSlots, nil
}
