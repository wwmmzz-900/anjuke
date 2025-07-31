package biz

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"anjuke/server/internal/domain"
	"anjuke/server/internal/service/dto"

	"github.com/go-kratos/kratos/v2/log"
)

// AppointmentUsecase 预约业务用例实现
type AppointmentUsecase struct {
	appointmentRepo domain.AppointmentRepo
	storeRepo       domain.StoreRepo
	realtorRepo     domain.RealtorRepo
	log             *log.Helper
}

// NewAppointmentUsecase 创建预约业务用例
func NewAppointmentUsecase(
	appointmentRepo domain.AppointmentRepo,
	storeRepo domain.StoreRepo,
	realtorRepo domain.RealtorRepo,
	logger log.Logger,
) *AppointmentUsecase {
	return &AppointmentUsecase{
		appointmentRepo: appointmentRepo,
		storeRepo:       storeRepo,
		realtorRepo:     realtorRepo,
		log:             log.NewHelper(logger),
	}
}

// CreateAppointment 创建预约
func (uc *AppointmentUsecase) CreateAppointment(ctx context.Context, req *dto.CreateAppointmentRequest) (*dto.CreateAppointmentResponse, error) {
	uc.log.Infof("开始创建预约，用户ID: %d, 门店ID: %s", req.UserID, req.StoreID)

	// 基础验证
	if err := uc.validateAppointmentRequest(ctx, req); err != nil {
		return nil, err
	}

	// 解析时间数据
	storeID, appointmentDateTime, endDateTime, err := uc.parseAppointmentTime(req)
	if err != nil {
		return nil, err
	}

	// 业务规则验证
	if err := uc.validateBusinessRules(ctx, req.UserID, storeID, appointmentDateTime, endDateTime); err != nil {
		return nil, err
	}

	// 创建预约
	appointment, needQueue, err := uc.assignRealtorAndCreateAppointment(ctx, req, storeID, appointmentDateTime, endDateTime)
	if err != nil {
		return nil, err
	}

	return &dto.CreateAppointmentResponse{
		Appointment: appointment,
		NeedQueue:   needQueue,
		Message:     "预约创建成功",
	}, nil
}

// GetAppointmentByCode 根据预约码获取预约信息
func (uc *AppointmentUsecase) GetAppointmentByCode(ctx context.Context, code string) (*dto.GetAppointmentResponse, error) {
	if code == "" {
		return nil, fmt.Errorf("预约码不能为空")
	}

	appointment, err := uc.appointmentRepo.GetAppointmentByCode(ctx, code)
	if err != nil {
		uc.log.Warnf("根据预约码查询失败，预约码: %s, 错误: %v", code, err)
		return nil, err
	}

	uc.log.Debugf("成功获取预约信息，预约码: %s, ID: %d", code, appointment.ID)
	return &dto.GetAppointmentResponse{
		Appointment: appointment,
	}, nil
}

// CancelAppointment 取消预约
func (uc *AppointmentUsecase) CancelAppointment(ctx context.Context, req *dto.CancelAppointmentRequest) (*dto.CancelAppointmentResponse, error) {
	// 获取预约信息
	appointment, err := uc.appointmentRepo.GetAppointmentByID(ctx, req.AppointmentID)
	if err != nil {
		return &dto.CancelAppointmentResponse{
			Success: false,
			Message: "预约不存在",
		}, err
	}

	// 检查是否可以取消
	if !appointment.CanBeCancelled() {
		return &dto.CancelAppointmentResponse{
			Success: false,
			Message: "当前状态不允许取消预约",
		}, nil
	}

	// 更新预约状态
	appointment.Status = domain.AppointmentStatusCancelled
	err = uc.appointmentRepo.UpdateAppointment(ctx, appointment)
	if err != nil {
		return &dto.CancelAppointmentResponse{
			Success: false,
			Message: "取消预约失败",
		}, err
	}

	return &dto.CancelAppointmentResponse{
		Success: true,
		Message: "预约已取消",
	}, nil
}

// GetAvailableTimeSlots 获取可预约时段
func (uc *AppointmentUsecase) GetAvailableTimeSlots(ctx context.Context, req *dto.ListAvailableSlotsRequest) (*dto.ListAvailableSlotsResponse, error) {
	storeID, err := strconv.ParseUint(req.StoreID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("无效的门店ID")
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("无效的日期格式")
	}

	slots, err := uc.appointmentRepo.GetAvailableTimeSlots(ctx, storeID, startDate, int(req.Days))
	if err != nil {
		return nil, err
	}

	return &dto.ListAvailableSlotsResponse{
		Slots: slots,
	}, nil
}

// AcceptAppointment 经纪人接单
func (uc *AppointmentUsecase) AcceptAppointment(ctx context.Context, appointmentID, realtorID uint64) error {
	appointment, err := uc.appointmentRepo.GetAppointmentByID(ctx, appointmentID)
	if err != nil {
		return fmt.Errorf("预约不存在")
	}

	if appointment.Status != domain.AppointmentStatusPending {
		return fmt.Errorf("预约状态不允许接单")
	}

	appointment.RealtorID = &realtorID
	appointment.Status = domain.AppointmentStatusConfirmed

	return uc.appointmentRepo.UpdateAppointment(ctx, appointment)
}

// ConfirmAppointment 经纪人确认服务
func (uc *AppointmentUsecase) ConfirmAppointment(ctx context.Context, appointmentID, realtorID uint64) error {
	appointment, err := uc.appointmentRepo.GetAppointmentByID(ctx, appointmentID)
	if err != nil {
		return fmt.Errorf("预约不存在")
	}

	if appointment.RealtorID == nil || *appointment.RealtorID != realtorID {
		return fmt.Errorf("无权限操作此预约")
	}

	if appointment.Status != domain.AppointmentStatusConfirmed {
		return fmt.Errorf("预约状态不允许确认")
	}

	appointment.Status = domain.AppointmentStatusInProgress
	now := time.Now()
	appointment.ConfirmedAt = &now

	return uc.appointmentRepo.UpdateAppointment(ctx, appointment)
}

// CompleteAppointment 完成服务
func (uc *AppointmentUsecase) CompleteAppointment(ctx context.Context, appointmentID, realtorID uint64, serviceNotes string) error {
	appointment, err := uc.appointmentRepo.GetAppointmentByID(ctx, appointmentID)
	if err != nil {
		return fmt.Errorf("预约不存在")
	}

	if appointment.RealtorID == nil || *appointment.RealtorID != realtorID {
		return fmt.Errorf("无权限操作此预约")
	}

	if appointment.Status != domain.AppointmentStatusInProgress {
		return fmt.Errorf("预约状态不允许完成")
	}

	appointment.Status = domain.AppointmentStatusCompleted
	now := time.Now()
	appointment.CompletedAt = &now

	return uc.appointmentRepo.UpdateAppointment(ctx, appointment)
}

// UpdateRealtorStatus 更新经纪人状态
func (uc *AppointmentUsecase) UpdateRealtorStatus(ctx context.Context, req *dto.UpdateRealtorStatusRequest) (*dto.UpdateRealtorStatusResponse, error) {
	// 这里应该有经纪人状态仓储的实现，暂时返回成功
	return &dto.UpdateRealtorStatusResponse{
		Success: true,
		Message: "状态更新成功",
	}, nil
}

// 辅助方法
func (uc *AppointmentUsecase) validateAppointmentRequest(ctx context.Context, req *dto.CreateAppointmentRequest) error {
	// 验证门店是否存在且激活
	// 将string ID转换为uint64
	storeID, err := strconv.ParseUint(req.StoreID, 10, 64)
	if err != nil {
		return fmt.Errorf("无效的门店ID格式: %s", req.StoreID)
	}
	store, err := uc.storeRepo.GetStoreByID(ctx, storeID)
	if err != nil {
		return fmt.Errorf("门店不存在或已停用")
	}

	// 检查门店是否激活
	if !store.IsActive {
		return fmt.Errorf("门店暂停服务")
	}

	return nil
}

func (uc *AppointmentUsecase) parseAppointmentTime(req *dto.CreateAppointmentRequest) (uint64, time.Time, time.Time, error) {
	// 解析门店ID
	storeID, err := strconv.ParseUint(req.StoreID, 10, 64)
	if err != nil {
		return 0, time.Time{}, time.Time{}, fmt.Errorf("无效的门店ID格式: %s", req.StoreID)
	}

	// 解析预约日期
	appointmentDate, err := time.Parse("2006-01-02", req.AppointmentDate)
	if err != nil {
		return 0, time.Time{}, time.Time{}, fmt.Errorf("无效的预约日期格式: %v", err)
	}

	// 解析预约时间
	startTime, err := time.Parse("15:04", req.StartTime)
	if err != nil {
		return 0, time.Time{}, time.Time{}, fmt.Errorf("无效的预约时间格式: %v", err)
	}

	// 组合日期和时间
	appointmentDateTime := time.Date(
		appointmentDate.Year(), appointmentDate.Month(), appointmentDate.Day(),
		startTime.Hour(), startTime.Minute(), 0, 0, appointmentDate.Location(),
	)

	endDateTime := appointmentDateTime.Add(time.Duration(req.DurationMinutes) * time.Minute)

	return storeID, appointmentDateTime, endDateTime, nil
}

func (uc *AppointmentUsecase) validateBusinessRules(ctx context.Context, userID int64, storeID uint64, startTime, endTime time.Time) error {
	// 检查用户时间冲突
	hasConflict, err := uc.appointmentRepo.CheckUserTimeConflict(ctx, userID, startTime, endTime)
	if err != nil {
		uc.log.Errorf("检查用户时间冲突失败: %v", err)
		return fmt.Errorf("系统繁忙，请稍后重试")
	}
	if hasConflict {
		return fmt.Errorf("该时间段您已有其他预约")
	}

	return nil
}

func (uc *AppointmentUsecase) assignRealtorAndCreateAppointment(
	ctx context.Context,
	req *dto.CreateAppointmentRequest,
	storeID uint64,
	startTime, endTime time.Time,
) (*domain.AppointmentInfo, bool, error) {
	// 创建预约信息
	appointment := &domain.AppointmentInfo{
		UserID:          req.UserID,
		StoreID:         storeID,
		CustomerName:    req.CustomerName,
		CustomerPhone:   req.CustomerPhone,
		AppointmentDate: startTime.Truncate(24 * time.Hour),
		StartTime:       startTime,
		EndTime:         endTime,
		DurationMinutes: req.DurationMinutes,
		Requirements:    req.Requirements,
		Status:          domain.AppointmentStatusPending,
	}

	// 保存预约
	savedAppointment, err := uc.appointmentRepo.CreateAppointment(ctx, appointment)
	if err != nil {
		return nil, false, fmt.Errorf("创建预约失败: %v", err)
	}

	return savedAppointment, false, nil
}
