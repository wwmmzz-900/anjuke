package service

import (
	"context"
	"errors"
	"testing"
	"time"

	pb "anjuke/server/api/points/v5"
	"anjuke/server/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPointsUsecase 积分用例的模拟实现
type MockPointsUsecase struct {
	mock.Mock
}

func (m *MockPointsUsecase) GetUserPoints(ctx context.Context, userID uint64) (*domain.UserPoints, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserPoints), args.Error(1)
}

func (m *MockPointsUsecase) GetPointsHistory(ctx context.Context, userID uint64, page, pageSize int32, pointsType string) ([]*domain.PointsRecord, int32, error) {
	args := m.Called(ctx, userID, page, pageSize, pointsType)
	return args.Get(0).([]*domain.PointsRecord), args.Get(1).(int32), args.Error(2)
}

func (m *MockPointsUsecase) CheckIn(ctx context.Context, userID uint64) (*domain.CheckInResult, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CheckInResult), args.Error(1)
}

func (m *MockPointsUsecase) EarnPointsByConsume(ctx context.Context, userID uint64, orderID string, amount int64) (*domain.EarnResult, error) {
	args := m.Called(ctx, userID, orderID, amount)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.EarnResult), args.Error(1)
}

func (m *MockPointsUsecase) UsePoints(ctx context.Context, userID uint64, points int64, orderID, description string) (*domain.UseResult, error) {
	args := m.Called(ctx, userID, points, orderID, description)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UseResult), args.Error(1)
}

func TestPointsService_GetUserPoints(t *testing.T) {
	tests := []struct {
		name       string
		req        *pb.GetUserPointsRequest
		setupMocks func(*MockPointsUsecase)
		wantCode   int32
		wantMsg    string
	}{
		{
			name: "查询积分成功",
			req: &pb.GetUserPointsRequest{
				UserId: 123,
			},
			setupMocks: func(uc *MockPointsUsecase) {
				uc.On("GetUserPoints", mock.Anything, uint64(123)).Return(&domain.UserPoints{
					UserID:      123,
					TotalPoints: 500,
				}, nil)
			},
			wantCode: 0,
			wantMsg:  "查询成功",
		},
		{
			name: "用户ID为空",
			req: &pb.GetUserPointsRequest{
				UserId: 0,
			},
			setupMocks: func(uc *MockPointsUsecase) {},
			wantCode:   1,
			wantMsg:    "缺少参数: user_id",
		},
		{
			name: "查询失败",
			req: &pb.GetUserPointsRequest{
				UserId: 123,
			},
			setupMocks: func(uc *MockPointsUsecase) {
				uc.On("GetUserPoints", mock.Anything, uint64(123)).Return(nil, errors.New("数据库错误"))
			},
			wantCode: 1,
			wantMsg:  "数据库错误",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟对象
			uc := new(MockPointsUsecase)

			// 设置模拟行为
			tt.setupMocks(uc)

			// 创建服务
			service := NewPointsServiceWithInterface(uc)

			// 执行测试
			resp, err := service.GetUserPoints(context.Background(), tt.req)

			// 验证结果
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.wantCode, resp.Code)
			assert.Contains(t, resp.Msg, tt.wantMsg)

			// 验证模拟对象的调用
			uc.AssertExpectations(t)
		})
	}
}

func TestPointsService_GetPointsHistory(t *testing.T) {
	tests := []struct {
		name       string
		req        *pb.GetPointsHistoryRequest
		setupMocks func(*MockPointsUsecase)
		wantCode   int32
		wantMsg    string
	}{
		{
			name: "查询积分明细成功",
			req: &pb.GetPointsHistoryRequest{
				UserId:   123,
				Page:     1,
				PageSize: 20,
				Type:     "earn",
			},
			setupMocks: func(uc *MockPointsUsecase) {
				records := []*domain.PointsRecord{
					{
						ID:          1,
						UserID:      123,
						Type:        "checkin",
						Points:      10,
						Description: "签到获得积分",
						CreatedAt:   time.Now(),
					},
				}
				uc.On("GetPointsHistory", mock.Anything, uint64(123), int32(1), int32(20), "earn").Return(records, int32(1), nil)
			},
			wantCode: 0,
			wantMsg:  "查询成功",
		},
		{
			name: "用户ID为空",
			req: &pb.GetPointsHistoryRequest{
				UserId: 0,
			},
			setupMocks: func(uc *MockPointsUsecase) {},
			wantCode:   1,
			wantMsg:    "缺少参数: user_id",
		},
		{
			name: "查询失败",
			req: &pb.GetPointsHistoryRequest{
				UserId: 123,
			},
			setupMocks: func(uc *MockPointsUsecase) {
				uc.On("GetPointsHistory", mock.Anything, uint64(123), int32(0), int32(0), "").Return([]*domain.PointsRecord{}, int32(0), errors.New("数据库错误"))
			},
			wantCode: 1,
			wantMsg:  "数据库错误",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟对象
			uc := new(MockPointsUsecase)

			// 设置模拟行为
			tt.setupMocks(uc)

			// 创建服务
			service := NewPointsServiceWithInterface(uc)

			// 执行测试
			resp, err := service.GetPointsHistory(context.Background(), tt.req)

			// 验证结果
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.wantCode, resp.Code)
			assert.Contains(t, resp.Msg, tt.wantMsg)

			// 验证模拟对象的调用
			uc.AssertExpectations(t)
		})
	}
}

func TestPointsService_CheckIn(t *testing.T) {
	tests := []struct {
		name       string
		req        *pb.CheckInRequest
		setupMocks func(*MockPointsUsecase)
		wantCode   int32
		wantMsg    string
	}{
		{
			name: "签到成功",
			req: &pb.CheckInRequest{
				UserId: 123,
			},
			setupMocks: func(uc *MockPointsUsecase) {
				uc.On("CheckIn", mock.Anything, uint64(123)).Return(&domain.CheckInResult{
					PointsEarned:    10,
					TotalPoints:     510,
					ConsecutiveDays: 1,
				}, nil)
			},
			wantCode: 0,
			wantMsg:  "签到成功",
		},
		{
			name: "用户ID为空",
			req: &pb.CheckInRequest{
				UserId: 0,
			},
			setupMocks: func(uc *MockPointsUsecase) {},
			wantCode:   1,
			wantMsg:    "缺少参数: user_id",
		},
		{
			name: "签到失败",
			req: &pb.CheckInRequest{
				UserId: 123,
			},
			setupMocks: func(uc *MockPointsUsecase) {
				uc.On("CheckIn", mock.Anything, uint64(123)).Return(nil, errors.New("今日已签到"))
			},
			wantCode: 1,
			wantMsg:  "今日已签到",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟对象
			uc := new(MockPointsUsecase)

			// 设置模拟行为
			tt.setupMocks(uc)

			// 创建服务
			service := NewPointsServiceWithInterface(uc)

			// 执行测试
			resp, err := service.CheckIn(context.Background(), tt.req)

			// 验证结果
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.wantCode, resp.Code)
			assert.Contains(t, resp.Msg, tt.wantMsg)

			// 验证模拟对象的调用
			uc.AssertExpectations(t)
		})
	}
}

func TestPointsService_EarnPointsByConsume(t *testing.T) {
	tests := []struct {
		name       string
		req        *pb.EarnPointsByConsumeRequest
		setupMocks func(*MockPointsUsecase)
		wantCode   int32
		wantMsg    string
	}{
		{
			name: "消费获取积分成功",
			req: &pb.EarnPointsByConsumeRequest{
				UserId:  123,
				OrderId: "ORDER123",
				Amount:  10000,
			},
			setupMocks: func(uc *MockPointsUsecase) {
				uc.On("EarnPointsByConsume", mock.Anything, uint64(123), "ORDER123", int64(10000)).Return(&domain.EarnResult{
					PointsEarned: 100,
					TotalPoints:  600,
				}, nil)
			},
			wantCode: 0,
			wantMsg:  "消费获得积分成功",
		},
		{
			name: "用户ID为空",
			req: &pb.EarnPointsByConsumeRequest{
				UserId:  0,
				OrderId: "ORDER123",
				Amount:  10000,
			},
			setupMocks: func(uc *MockPointsUsecase) {},
			wantCode:   1,
			wantMsg:    "缺少参数: user_id",
		},
		{
			name: "订单ID为空",
			req: &pb.EarnPointsByConsumeRequest{
				UserId:  123,
				OrderId: "",
				Amount:  10000,
			},
			setupMocks: func(uc *MockPointsUsecase) {},
			wantCode:   1,
			wantMsg:    "缺少参数: order_id",
		},
		{
			name: "消费金额为0",
			req: &pb.EarnPointsByConsumeRequest{
				UserId:  123,
				OrderId: "ORDER123",
				Amount:  0,
			},
			setupMocks: func(uc *MockPointsUsecase) {},
			wantCode:   1,
			wantMsg:    "消费金额必须大于0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟对象
			uc := new(MockPointsUsecase)

			// 设置模拟行为
			tt.setupMocks(uc)

			// 创建服务
			service := NewPointsServiceWithInterface(uc)

			// 执行测试
			resp, err := service.EarnPointsByConsume(context.Background(), tt.req)

			// 验证结果
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.wantCode, resp.Code)
			assert.Contains(t, resp.Msg, tt.wantMsg)

			// 验证模拟对象的调用
			uc.AssertExpectations(t)
		})
	}
}

func TestPointsService_UsePoints(t *testing.T) {
	tests := []struct {
		name       string
		req        *pb.UsePointsRequest
		setupMocks func(*MockPointsUsecase)
		wantCode   int32
		wantMsg    string
	}{
		{
			name: "使用积分成功",
			req: &pb.UsePointsRequest{
				UserId:      123,
				Points:      100,
				OrderId:     "ORDER123",
				Description: "积分抵扣",
			},
			setupMocks: func(uc *MockPointsUsecase) {
				uc.On("UsePoints", mock.Anything, uint64(123), int64(100), "ORDER123", "积分抵扣").Return(&domain.UseResult{
					PointsUsed:     100,
					AmountDeducted: 1000,
					TotalPoints:    400,
				}, nil)
			},
			wantCode: 0,
			wantMsg:  "积分使用成功",
		},
		{
			name: "用户ID为空",
			req: &pb.UsePointsRequest{
				UserId: 0,
				Points: 100,
			},
			setupMocks: func(uc *MockPointsUsecase) {},
			wantCode:   1,
			wantMsg:    "缺少参数: user_id",
		},
		{
			name: "积分数量为0",
			req: &pb.UsePointsRequest{
				UserId: 123,
				Points: 0,
			},
			setupMocks: func(uc *MockPointsUsecase) {},
			wantCode:   1,
			wantMsg:    "使用积分数量必须大于0",
		},
		{
			name: "积分数量不是10的倍数",
			req: &pb.UsePointsRequest{
				UserId: 123,
				Points: 15,
			},
			setupMocks: func(uc *MockPointsUsecase) {},
			wantCode:   1,
			wantMsg:    "积分数量必须是10的倍数",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟对象
			uc := new(MockPointsUsecase)

			// 设置模拟行为
			tt.setupMocks(uc)

			// 创建服务
			service := NewPointsServiceWithInterface(uc)

			// 执行测试
			resp, err := service.UsePoints(context.Background(), tt.req)

			// 验证结果
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.wantCode, resp.Code)
			assert.Contains(t, resp.Msg, tt.wantMsg)

			// 验证模拟对象的调用
			uc.AssertExpectations(t)
		})
	}
}
