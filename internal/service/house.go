package service

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/wwmmzz-900/anjuke/api/house/v3"

	"github.com/wwmmzz-900/anjuke/internal/biz"
)

// 全局WebSocketHub实例

type HouseService struct {
	pb.UnimplementedHouseServer
	uc *biz.HouseUsecase
}

func NewHouseService(uc *biz.HouseUsecase) *HouseService {
	return &HouseService{
		uc: uc,
	}
}

// 普通推荐列表
func (s *HouseService) RecommendList(ctx context.Context, req *pb.HouseRecommendRequest) (*pb.HouseRecommendReply, error) {
	// 模拟推荐数据
	items := []*pb.HouseRecommendItem{
		{
			HouseId:     1,
			Title:       "精装修两室一厅",
			Description: "地铁口附近，交通便利，精装修",
			Price:       3500.0,
			Area:        85.5,
			Layout:      "2室1厅1卫",
			ImageUrl:    "https://example.com/house1.jpg",
		},
		{
			HouseId:     2,
			Title:       "温馨三室两厅",
			Description: "小区环境优美，配套设施完善",
			Price:       4200.0,
			Area:        120.0,
			Layout:      "3室2厅2卫",
			ImageUrl:    "https://example.com/house2.jpg",
		},
	}

	// 返回符合三范式的响应
	return &pb.HouseRecommendReply{
		Code: 0,
		Msg:  "success",
		Data: &pb.HouseRecommendData{
			Total: int64(len(items)),
			List:  items,
		},
	}, nil
}

// 个性化推荐列表
func (s *HouseService) PersonalRecommendList(ctx context.Context, req *pb.PersonalRecommendRequest) (*pb.HouseRecommendReply, error) {
	log.Printf("为用户 %d 生成个性化推荐", req.UserId)

	// 调用业务层获取个性化推荐
	houses, total, err := s.uc.PersonalRecommendList(ctx, req.UserId, int(req.Page), int(req.PageSize))
	if err != nil {
		log.Printf("获取个性化推荐失败: %v", err)
		// 如果业务层失败，返回默认推荐
		items := []*pb.HouseRecommendItem{
			{
				HouseId:     3,
				Title:       "个性化推荐-豪华公寓",
				Description: "根据您的浏览记录推荐",
				Price:       5800.0,
				Area:        150.0,
				Layout:      "3室2厅2卫",
				ImageUrl:    "https://example.com/house3.jpg",
			},
		}
		return &pb.HouseRecommendReply{
			Code: 0,
			Msg:  "success",
			Data: &pb.HouseRecommendData{
				Total: int64(len(items)),
				List:  items,
			},
		}, nil
	}

	// 转换为protobuf格式
	items := make([]*pb.HouseRecommendItem, 0, len(houses))
	for _, house := range houses {
		items = append(items, &pb.HouseRecommendItem{
			HouseId:     house.HouseID,
			Title:       house.Title,
			Description: house.Description,
			Price:       house.Price,
			Area:        house.Area,
			Layout:      house.Layout,
			ImageUrl:    house.ImageURL,
		})
	}

	log.Printf("成功获取到 %d 条个性化推荐", len(items))
	return &pb.HouseRecommendReply{
		Code: 0,
		Msg:  "success",
		Data: &pb.HouseRecommendData{
			Total: int64(total),
			List:  items,
		},
	}, nil
}

// 预约看房接口实现
func (s *HouseService) ReserveHouse(ctx context.Context, req *pb.ReserveHouseRequest) (*pb.ReserveHouseReply, error) {
	// 1. 业务逻辑：保存预约信息
	err := s.uc.ReserveHouse(ctx, req)
	if err != nil {
		return &pb.ReserveHouseReply{
			Code: 400,
			Msg:  fmt.Sprintf("预约失败: %v", err),
			Data: &pb.ReserveHouseData{
				Success: false,
			},
		}, nil
	}

	// 2. 推送消息给房东和预约用户
	landlordID := req.GetLandlordId()
	userName := req.GetUserName()
	houseTitle := req.GetHouseTitle()
	houseID := req.GetHouseId()
	
	// 通知房东
	landlordMessage := map[string]interface{}{
		"type":          "reservation_created",
		"title":         "新的预约请求",
		"message":       fmt.Sprintf("用户 %s 预约了您的房源《%s》", userName, houseTitle),
		"house_id":      houseID,
		"user_id":       req.GetUserId(),
		"user_name":     userName,
		"reserve_time":  req.GetReserveTime(),
		"timestamp":     time.Now().Unix(),
	}
	pushToUser(houseID, landlordID, landlordMessage)
	
	// 通知预约用户
	userMessage := map[string]interface{}{
		"type":          "reservation_created",
		"title":         "预约成功",
		"message":       fmt.Sprintf("您已成功预约房源《%s》，请等待房东确认", houseTitle),
		"house_id":      houseID,
		"landlord_id":   landlordID,
		"reserve_time":  req.GetReserveTime(),
		"timestamp":     time.Now().Unix(),
	}
	pushToUser(houseID, req.GetUserId(), userMessage)
	
	log.Printf("WebSocket消息已发送 - 房东 %d 和用户 %d", landlordID, req.GetUserId())

	// 模拟生成预约ID
	reservationID := time.Now().Unix()

	return &pb.ReserveHouseReply{
		Code: 0,
		Msg:  "预约成功",
		Data: &pb.ReserveHouseData{
			Success:       true,
			ReservationId: reservationID,
		},
	}, nil
}

// 发起在线聊天
func (s *HouseService) StartChat(ctx context.Context, req *pb.StartChatRequest) (*pb.StartChatReply, error) {
	// 简化实现：生成聊天ID并返回成功
	chatID := fmt.Sprintf("chat_%d_%d_%d", req.ReservationId, req.UserId, time.Now().Unix())
	
	log.Printf("发起聊天: 预约ID=%d, 用户ID=%d, 房东ID=%d, 聊天ID=%s", 
		req.ReservationId, req.UserId, req.LandlordId, chatID)
	
	// 如果有初始消息，通过WebSocket发送
	if req.InitialMessage != "" {
		message := map[string]interface{}{
			"type":           "chat_started",
			"title":          "聊天已开始",
			"content":        req.InitialMessage,
			"chat_id":        chatID,
			"reservation_id": req.ReservationId,
			"timestamp":      time.Now().Unix(),
		}
		
		// 通知房东和用户
		// 这里需要知道房源ID才能推送消息，暂时记录日志
		log.Printf("聊天消息: %+v", message)
	}
	
	return &pb.StartChatReply{
		Code: 0,
		Msg:  "聊天发起成功",
		Data: &pb.StartChatData{
			ChatId:  chatID,
			Success: true,
		},
	}, nil
}
