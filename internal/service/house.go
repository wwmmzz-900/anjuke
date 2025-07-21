package service

import (
	"context"
	"fmt"
	v3 "github.com/wwmmzz-900/anjuke/api/house/v3"

	"github.com/wwmmzz-900/anjuke/internal/biz"
)

// 全局WebSocketHub实例
var WsHub = NewWebSocketHub()

type HouseService struct {
	v3.UnimplementedHouseServer
	uc *biz.HouseUsecase
}

func NewHouseService(uc *biz.HouseUsecase) *HouseService {
	return &HouseService{uc: uc}
}

// 预约看房接口实现
func (s *HouseService) ReserveHouse(ctx context.Context, req *v3.ReserveHouseRequest) (*v3.ReserveHouseReply, error) {
	// 1. 业务逻辑：保存预约信息
	err := s.uc.ReserveHouse(ctx, req)
	if err != nil {
		return &v3.ReserveHouseReply{Success: false, Message: "预约失败"}, err
	}

	// 2. 推送消息给房东
	landlordID := req.GetLandlordId()
	userName := req.GetUserName()
	houseTitle := req.GetHouseTitle()
	message := fmt.Sprintf("用户 %s 预约了您的房源《%s》", userName, houseTitle)
	go WsHub.SendToUser(landlordID, message)

	return &v3.ReserveHouseReply{Success: true, Message: "预约成功"}, nil
}
