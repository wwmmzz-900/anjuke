package model

import (
	"time"

	"gorm.io/gorm"
)

// OrderStatus 订单状态枚举
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"   // 待处理
	OrderStatusActive    OrderStatus = "active"    // 活动中
	OrderStatusCompleted OrderStatus = "completed" // 已完成
	OrderStatusCancelled OrderStatus = "cancelled" // 已取消
)

// RentalOrder 租房订单表结构体
type RentalOrder struct {
	gorm.Model
	OrderNo      string      `json:"orderNo" form:"orderNo" gorm:"comment:订单号;column:order_no;size:64;uniqueIndex;"`         // 订单号
	HouseId      *uint       `json:"houseId" form:"houseId" gorm:"comment:房源ID;column:house_id;size:19;index;"`              // 房源ID
	TenantId     *uint       `json:"tenantId" form:"tenantId" gorm:"comment:租客ID;column:tenant_id;size:19;index;"`           // 租客ID
	LandlordId   *uint       `json:"landlordId" form:"landlordId" gorm:"comment:房东ID;column:landlord_id;size:19;index;"`     // 房东ID
	TenantPhone  string      `json:"tenantPhone" form:"tenantPhone" gorm:"comment:租客手机号;column:tenant_phone;size:20;index;"` // 租客手机号
	RentStart    *time.Time  `json:"rentStart" form:"rentStart" gorm:"comment:租期开始;column:rent_start;"`                      // 租期开始
	RentEnd      *time.Time  `json:"rentEnd" form:"rentEnd" gorm:"comment:租期结束;column:rent_end;"`                            // 租期结束
	RentAmount   *float64    `json:"rentAmount" form:"rentAmount" gorm:"comment:租金;column:rent_amount;size:10;"`             // 租金
	Deposit      *float64    `json:"deposit" form:"deposit" gorm:"comment:押金;column:deposit;size:10;"`                       // 押金
	Status       OrderStatus `json:"status" form:"status" gorm:"comment:订单状态;column:status;size:20;index;"`                  // 订单状态
	SignedAt     *time.Time  `json:"signedAt" form:"signedAt" gorm:"comment:签约时间;column:signed_at;"`                         // 签约时间
	CancelReason string      `json:"cancelReason" form:"cancelReason" gorm:"comment:取消原因;column:cancel_reason;size:255;"`    // 取消原因
	CancelledAt  *time.Time  `json:"cancelledAt" form:"cancelledAt" gorm:"comment:取消时间;column:cancelled_at;"`                // 取消时间
}

// TableName RentalOrder自定义表名
func (RentalOrder) TableName() string {
	return "rental_order"
}
