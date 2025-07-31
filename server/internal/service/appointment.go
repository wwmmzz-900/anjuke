package service

import (
	"context"
	"strconv"

	pb "anjuke/server/api/appointment/v1"
	"anjuke/server/internal/biz"
	"anjuke/server/internal/domain"
	"anjuke/server/internal/service/dto"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AppointmentService 预约服务实现
type AppointmentService struct {
	pb.UnimplementedAppointmentServiceServer

	appointmentUC *biz.AppointmentUsecase
	log           *log.Helper
}

// NewAppointmentService 创建预约服务实例
func NewAppointmentService(appointmentUC *biz.AppointmentUsecase, logger log.Logger) *AppointmentService {
	return &AppointmentService{
		appointmentUC: appointmentUC,
		log:           log.NewHelper(logger),
	}
}

// CreateAppointment 创建预约
func (s *AppointmentService) CreateAppointment(ctx context.Context, req *pb.CreateAppointmentRequest) (*pb.CreateAppointmentResponse, error) {
	s.log.Infof("收到创建预约请求，门店ID: %s, 客户: %s", req.StoreId, req.CustomerName)

	// 转换请求对象
	domainReq := &dto.CreateAppointmentRequest{
		UserID:          1, // 这里应该从上下文中获取用户ID
		StoreID:         req.StoreId,
		CustomerName:    req.CustomerName,
		CustomerPhone:   req.CustomerPhone,
		AppointmentDate: req.AppointmentDate,
		StartTime:       req.StartTime,
		DurationMinutes: req.DurationMinutes,
		Requirements:    req.Requirements,
	}

	// 调用业务逻辑
	resp, err := s.appointmentUC.CreateAppointment(ctx, domainReq)
	if err != nil {
		s.log.Errorf("创建预约失败: %v", err)
		return nil, err
	}

	// 转换响应对象
	return &pb.CreateAppointmentResponse{
		Appointment: s.convertAppointmentToPB(resp.Appointment),
		NeedQueue:   resp.NeedQueue,
		Message:     resp.Message,
	}, nil
}

// GetAppointment 获取预约详情
func (s *AppointmentService) GetAppointment(ctx context.Context, req *pb.GetAppointmentRequest) (*pb.GetAppointmentResponse, error) {
	s.log.Infof("收到获取预约请求，预约ID: %s", req.AppointmentId)

	// 这里假设传入的是预约码，实际可能需要区分ID和预约码
	resp, err := s.appointmentUC.GetAppointmentByCode(ctx, req.AppointmentId)
	if err != nil {
		s.log.Errorf("获取预约失败: %v", err)
		return nil, err
	}

	return &pb.GetAppointmentResponse{
		Appointment: s.convertAppointmentToPB(resp.Appointment),
	}, nil
}

// CancelAppointment 取消预约
func (s *AppointmentService) CancelAppointment(ctx context.Context, req *pb.CancelAppointmentRequest) (*pb.CancelAppointmentResponse, error) {
	s.log.Infof("收到取消预约请求，预约ID: %s", req.AppointmentId)

	// 解析预约ID
	appointmentID, err := strconv.ParseUint(req.AppointmentId, 10, 64)
	if err != nil {
		return &pb.CancelAppointmentResponse{
			Success: false,
			Message: "无效的预约ID",
		}, err
	}

	// 转换请求对象
	domainReq := &dto.CancelAppointmentRequest{
		AppointmentID: appointmentID,
		Reason:        req.Reason,
	}

	// 调用业务逻辑
	resp, err := s.appointmentUC.CancelAppointment(ctx, domainReq)
	if err != nil {
		s.log.Errorf("取消预约失败: %v", err)
		return &pb.CancelAppointmentResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	return &pb.CancelAppointmentResponse{
		Success: resp.Success,
		Message: resp.Message,
	}, nil
}

// GetAvailableSlots 获取可预约时段
func (s *AppointmentService) GetAvailableSlots(ctx context.Context, req *pb.GetAvailableSlotsRequest) (*pb.GetAvailableSlotsResponse, error) {
	s.log.Infof("收到获取可预约时段请求，门店ID: %s", req.StoreId)

	// 转换请求对象
	domainReq := &dto.ListAvailableSlotsRequest{
		StoreID:   req.StoreId,
		StartDate: req.StartDate,
		Days:      req.Days,
	}

	// 调用业务逻辑
	resp, err := s.appointmentUC.GetAvailableTimeSlots(ctx, domainReq)
	if err != nil {
		s.log.Errorf("获取可预约时段失败: %v", err)
		return nil, err
	}

	// 转换响应对象
	slots := make([]*pb.TimeSlot, len(resp.Slots))
	for i, slot := range resp.Slots {
		slots[i] = &pb.TimeSlot{
			Date:              slot.Date,
			StartTime:         slot.StartTime,
			EndTime:           slot.EndTime,
			Available:         slot.Available,
			AvailableRealtors: slot.AvailableRealtors,
		}
	}

	return &pb.GetAvailableSlotsResponse{
		Slots: slots,
	}, nil
}

// AcceptAppointment 经纪人接单
func (s *AppointmentService) AcceptAppointment(ctx context.Context, req *pb.AcceptAppointmentRequest) (*pb.AcceptAppointmentResponse, error) {
	s.log.Infof("收到经纪人接单请求，预约ID: %s, 经纪人ID: %s", req.AppointmentId, req.RealtorId)

	// 解析ID
	appointmentID, err := strconv.ParseUint(req.AppointmentId, 10, 64)
	if err != nil {
		return &pb.AcceptAppointmentResponse{
			Success: false,
			Message: "无效的预约ID",
		}, err
	}

	realtorID, err := strconv.ParseUint(req.RealtorId, 10, 64)
	if err != nil {
		return &pb.AcceptAppointmentResponse{
			Success: false,
			Message: "无效的经纪人ID",
		}, err
	}

	// 调用业务逻辑
	err = s.appointmentUC.AcceptAppointment(ctx, appointmentID, realtorID)
	if err != nil {
		s.log.Errorf("经纪人接单失败: %v", err)
		return &pb.AcceptAppointmentResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	return &pb.AcceptAppointmentResponse{
		Success: true,
		Message: "接单成功",
	}, nil
}

// ConfirmAppointment 经纪人确认服务
func (s *AppointmentService) ConfirmAppointment(ctx context.Context, req *pb.ConfirmAppointmentRequest) (*pb.ConfirmAppointmentResponse, error) {
	s.log.Infof("收到经纪人确认服务请求，预约ID: %s, 经纪人ID: %s", req.AppointmentId, req.RealtorId)

	// 解析ID
	appointmentID, err := strconv.ParseUint(req.AppointmentId, 10, 64)
	if err != nil {
		return &pb.ConfirmAppointmentResponse{
			Success: false,
			Message: "无效的预约ID",
		}, err
	}

	realtorID, err := strconv.ParseUint(req.RealtorId, 10, 64)
	if err != nil {
		return &pb.ConfirmAppointmentResponse{
			Success: false,
			Message: "无效的经纪人ID",
		}, err
	}

	// 调用业务逻辑
	err = s.appointmentUC.ConfirmAppointment(ctx, appointmentID, realtorID)
	if err != nil {
		s.log.Errorf("经纪人确认服务失败: %v", err)
		return &pb.ConfirmAppointmentResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	return &pb.ConfirmAppointmentResponse{
		Success: true,
		Message: "确认成功",
	}, nil
}

// CompleteAppointment 完成服务
func (s *AppointmentService) CompleteAppointment(ctx context.Context, req *pb.CompleteAppointmentRequest) (*pb.CompleteAppointmentResponse, error) {
	s.log.Infof("收到完成服务请求，预约ID: %s, 经纪人ID: %s", req.AppointmentId, req.RealtorId)

	// 解析ID
	appointmentID, err := strconv.ParseUint(req.AppointmentId, 10, 64)
	if err != nil {
		return &pb.CompleteAppointmentResponse{
			Success: false,
			Message: "无效的预约ID",
		}, err
	}

	realtorID, err := strconv.ParseUint(req.RealtorId, 10, 64)
	if err != nil {
		return &pb.CompleteAppointmentResponse{
			Success: false,
			Message: "无效的经纪人ID",
		}, err
	}

	// 调用业务逻辑
	err = s.appointmentUC.CompleteAppointment(ctx, appointmentID, realtorID, req.ServiceNotes)
	if err != nil {
		s.log.Errorf("完成服务失败: %v", err)
		return &pb.CompleteAppointmentResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	return &pb.CompleteAppointmentResponse{
		Success: true,
		Message: "服务完成",
	}, nil
}

// UpdateRealtorStatus 更新经纪人状态
func (s *AppointmentService) UpdateRealtorStatus(ctx context.Context, req *pb.UpdateRealtorStatusRequest) (*pb.UpdateRealtorStatusResponse, error) {
	s.log.Infof("收到更新经纪人状态请求，经纪人ID: %s, 状态: %s", req.RealtorId, req.Status)

	// 解析经纪人ID
	realtorID, err := strconv.ParseUint(req.RealtorId, 10, 64)
	if err != nil {
		return &pb.UpdateRealtorStatusResponse{
			Success: false,
			Message: "无效的经纪人ID",
		}, err
	}

	// 转换请求对象
	domainReq := &dto.UpdateRealtorStatusRequest{
		RealtorID: realtorID,
		Status:    domain.RealtorStatus(req.Status),
	}

	// 调用业务逻辑
	resp, err := s.appointmentUC.UpdateRealtorStatus(ctx, domainReq)
	if err != nil {
		s.log.Errorf("更新经纪人状态失败: %v", err)
		return &pb.UpdateRealtorStatusResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	return &pb.UpdateRealtorStatusResponse{
		Success: resp.Success,
		Message: resp.Message,
	}, nil
}

// 其他方法的简化实现
func (s *AppointmentService) ListAvailableStores(ctx context.Context, req *pb.ListStoresRequest) (*pb.ListStoresResponse, error) {
	// 简化实现，返回空列表
	return &pb.ListStoresResponse{
		Stores: []*pb.StoreInfo{},
	}, nil
}

func (s *AppointmentService) ListRealtorAppointments(ctx context.Context, req *pb.ListRealtorAppointmentsRequest) (*pb.ListRealtorAppointmentsResponse, error) {
	// 简化实现，返回空列表
	return &pb.ListRealtorAppointmentsResponse{
		Appointments: []*pb.AppointmentInfo{},
	}, nil
}

// 辅助方法

// convertAppointmentToPB 将领域对象转换为protobuf对象
func (s *AppointmentService) convertAppointmentToPB(appointment *domain.AppointmentInfo) *pb.AppointmentInfo {
	pbAppointment := &pb.AppointmentInfo{
		AppointmentId:        strconv.FormatUint(appointment.ID, 10),
		AppointmentCode:      appointment.AppointmentCode,
		UserId:               appointment.UserID,
		StoreId:              strconv.FormatUint(appointment.StoreID, 10),
		CustomerName:         appointment.CustomerName,
		CustomerPhone:        appointment.CustomerPhone,
		AppointmentDate:      appointment.AppointmentDate.Format("2006-01-02"),
		StartTime:            appointment.StartTime.Format("15:04"),
		EndTime:              appointment.EndTime.Format("15:04"),
		DurationMinutes:      appointment.DurationMinutes,
		Requirements:         appointment.Requirements,
		Status:               pb.AppointmentStatus(pb.AppointmentStatus_value[string(appointment.Status)]),
		QueuePosition:        appointment.QueuePosition,
		EstimatedWaitMinutes: appointment.EstimatedWaitMinutes,
		CreatedAt:            timestamppb.New(appointment.CreatedAt),
	}

	if appointment.RealtorID != nil {
		pbAppointment.RealtorId = strconv.FormatUint(*appointment.RealtorID, 10)
	}

	if appointment.RealtorInfo != nil {
		pbAppointment.RealtorName = appointment.RealtorInfo.Name
		pbAppointment.RealtorPhone = appointment.RealtorInfo.Phone
	}

	if appointment.ConfirmedAt != nil {
		pbAppointment.ConfirmedAt = timestamppb.New(*appointment.ConfirmedAt)
	}

	return pbAppointment
}
