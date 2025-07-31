package dto

import "anjuke/server/internal/domain"

// 预约相关的请求和响应对象

// CreateAppointmentRequest 创建预约请求
type CreateAppointmentRequest struct {
	UserID          int64  `json:"user_id" validate:"required"`
	StoreID         string `json:"store_id" validate:"required"`
	CustomerName    string `json:"customer_name" validate:"required,min=2,max=50"`
	CustomerPhone   string `json:"customer_phone" validate:"required,phone"`
	AppointmentDate string `json:"appointment_date" validate:"required,date"`
	StartTime       string `json:"start_time" validate:"required,time"`
	DurationMinutes int32  `json:"duration_minutes" validate:"required,oneof=30 60 90"`
	Requirements    string `json:"requirements" validate:"max=500"`
}

// CreateAppointmentResponse 创建预约响应
type CreateAppointmentResponse struct {
	Appointment *domain.AppointmentInfo `json:"appointment"`
	NeedQueue   bool                    `json:"need_queue"`
	Message     string                  `json:"message"`
}

// GetAppointmentResponse 获取预约响应
type GetAppointmentResponse struct {
	Appointment *domain.AppointmentInfo `json:"appointment"`
}

// ListAvailableSlotsRequest 获取可预约时段请求
type ListAvailableSlotsRequest struct {
	StoreID   string `json:"store_id" validate:"required"`
	StartDate string `json:"start_date" validate:"date"`
	Days      int32  `json:"days" validate:"min=1,max=7"`
}

// ListAvailableSlotsResponse 获取可预约时段响应
type ListAvailableSlotsResponse struct {
	Slots []*domain.TimeSlot `json:"slots"`
}

// CancelAppointmentRequest 取消预约请求
type CancelAppointmentRequest struct {
	AppointmentID uint64 `json:"appointment_id" validate:"required"`
	Reason        string `json:"reason" validate:"max=200"`
}

// CancelAppointmentResponse 取消预约响应
type CancelAppointmentResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// UpdateRealtorStatusRequest 更新经纪人状态请求
type UpdateRealtorStatusRequest struct {
	RealtorID uint64               `json:"realtor_id" validate:"required"`
	Status    domain.RealtorStatus `json:"status" validate:"required,oneof=online offline busy"`
}

// UpdateRealtorStatusResponse 更新经纪人状态响应
type UpdateRealtorStatusResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
