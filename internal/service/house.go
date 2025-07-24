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
	log.Printf("接收到普通推荐请求: page=%d, pageSize=%d", req.Page, req.PageSize)
	
	// 确保页码和每页数量有效
	page := int(req.Page)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 10
	}
	
	// 从数据库查询推荐房源
	houses, total, err := s.uc.RecommendList(ctx, page, pageSize)
	if err != nil {
		log.Printf("获取推荐列表失败: %v", err)
		// 如果查询失败，返回默认数据
		items := []*pb.HouseRecommendItem{
			{
				// 不再设置HouseId字段
				Title:       "精装修两室一厅",
				Description: "地铁口附近，交通便利，精装修",
				Price:       3500.0,
				Area:        85.5,
				Layout:      "2室1厅1卫",
				ImageUrl:    "https://example.com/house1.jpg",
			},
			{
				// 不再设置HouseId字段
				Title:       "温馨三室两厅",
				Description: "小区环境优美，配套设施完善",
				Price:       4200.0,
				Area:        120.0,
				Layout:      "3室2厅2卫",
				ImageUrl:    "https://example.com/house2.jpg",
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
			// 不再设置HouseId字段
			Title:       house.Title,
			Description: house.Description,
			Price:       house.Price,
			Area:        house.Area,
			Layout:      house.Layout,
			ImageUrl:    house.ImageURL,
		})
	}

	log.Printf("成功获取到 %d 条推荐", len(items))
	return &pb.HouseRecommendReply{
		Code: 0,
		Msg:  "success",
		Data: &pb.HouseRecommendData{
			Total: int64(total),
			List:  items,
		},
	}, nil
}

// 个性化推荐列表
func (s *HouseService) PersonalRecommendList(ctx context.Context, req *pb.PersonalRecommendRequest) (*pb.HouseRecommendReply, error) {
	log.Printf("接收到个性化推荐请求: userId=%d, page=%d, pageSize=%d", req.UserId, req.Page, req.PageSize)

	// 确保页码和每页数量有效
	page := int(req.Page)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 10
	}

	// 注意：这里不需要验证码验证，直接处理请求
	// 如果前端API需要验证码，应该在API网关层处理，而不是在这个服务中

	// 调用业务层获取个性化推荐
	houses, total, err := s.uc.PersonalRecommendList(ctx, req.UserId, page, pageSize)
	if err != nil {
		log.Printf("获取个性化推荐失败: %v", err)
		// 如果业务层失败，返回默认推荐
		items := []*pb.HouseRecommendItem{
			{
				// 不再设置HouseId字段
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
			// 不再设置HouseId字段
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
	userID := req.GetUserId()
	userName := req.GetUserName()
	houseTitle := req.GetHouseTitle()
	houseID := req.GetHouseId()
	
	// 生成预约ID（实际项目中应该从数据库获取）
	reservationID := time.Now().Unix()

	// 通知房东
	landlordMessage := map[string]interface{}{
		"type":           "reservation_created",
		"title":          "新的预约请求",
		"message":        fmt.Sprintf("用户 %s 预约了您的房源《%s》", userName, houseTitle),
		"house_id":       houseID,
		"user_id":        userID,
		"user_name":      userName,
		"reserve_time":   req.GetReserveTime(),
		"reservation_id": reservationID,
		"timestamp":      time.Now().Unix(),
		"sequence":       GlobalSequenceManager.GetNextSequence(0), // 系统消息使用0作为发送者ID
	}
	
	// 使用全局WebSocket管理器推送消息给房东
	if err := GlobalWebSocketManager.SendMessageToUser(landlordID, landlordMessage); err != nil {
		log.Printf("推送消息给房东 %d 失败: %v", landlordID, err)
	} else {
		log.Printf("成功推送消息给房东 %d", landlordID)
	}

	// 通知预约用户
	userMessage := map[string]interface{}{
		"type":           "reservation_created",
		"title":          "预约成功",
		"message":        fmt.Sprintf("您已成功预约房源《%s》，请等待房东确认", houseTitle),
		"house_id":       houseID,
		"landlord_id":    landlordID,
		"reserve_time":   req.GetReserveTime(),
		"reservation_id": reservationID,
		"timestamp":      time.Now().Unix(),
		"sequence":       GlobalSequenceManager.GetNextSequence(0), // 系统消息使用0作为发送者ID
	}
	
	// 使用全局WebSocket管理器推送消息给用户
	if err := GlobalWebSocketManager.SendMessageToUser(userID, userMessage); err != nil {
		log.Printf("推送消息给用户 %d 失败: %v", userID, err)
	} else {
		log.Printf("成功推送消息给用户 %d", userID)
	}

	log.Printf("WebSocket消息已发送 - 房东 %d 和用户 %d", landlordID, userID)

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
	log.Printf("接收到发起聊天请求: %+v", req)

	// 参数验证
	if req.ReservationId <= 0 {
		log.Printf("无效的预约ID: %d", req.ReservationId)
		return &pb.StartChatReply{
			Code: 400,
			Msg:  "无效的预约ID",
			Data: &pb.StartChatData{
				ChatId:  "",
				Success: false,
			},
		}, nil
	}

	if req.UserId <= 0 {
		log.Printf("无效的用户ID: %d", req.UserId)
		return &pb.StartChatReply{
			Code: 400,
			Msg:  "无效的用户ID",
			Data: &pb.StartChatData{
				ChatId:  "",
				Success: false,
			},
		}, nil
	}

	if req.LandlordId <= 0 {
		log.Printf("无效的房东ID: %d", req.LandlordId)
		return &pb.StartChatReply{
			Code: 400,
			Msg:  "无效的房东ID",
			Data: &pb.StartChatData{
				ChatId:  "",
				Success: false,
			},
		}, nil
	}

	// 生成聊天ID
	chatID := fmt.Sprintf("chat_%d_%d_%d", req.ReservationId, req.UserId, time.Now().Unix())

	log.Printf("发起聊天成功: 预约ID=%d, 用户ID=%d, 房东ID=%d, 聊天ID=%s",
		req.ReservationId, req.UserId, req.LandlordId, chatID)

	// 如果有初始消息，通过WebSocket发送
	if req.InitialMessage != "" {
		// 发送给房东的消息
		landlordMessage := map[string]interface{}{
			"type":           "chat_started",
			"title":          "新的聊天请求",
			"content":        req.InitialMessage,
			"message":        req.InitialMessage, // 兼容旧版本
			"chat_id":        chatID,
			"reservation_id": req.ReservationId,
			"from":           req.UserId,
			"to":             req.LandlordId,
			"timestamp":      time.Now().Unix(),
			"sequence":       GlobalSequenceManager.GetNextSequence(req.UserId),
		}

		// 使用全局WebSocket管理器发送消息给房东
		if err := GlobalWebSocketManager.SendMessageToUser(req.LandlordId, landlordMessage); err != nil {
			log.Printf("发送聊天消息给房东 %d 失败: %v", req.LandlordId, err)
		} else {
			log.Printf("成功发送聊天消息给房东 %d", req.LandlordId)
		}
		
		// 发送确认消息给用户
		userMessage := map[string]interface{}{
			"type":           "chat_started",
			"title":          "聊天已发起",
			"message":        "您的消息已发送给房东",
			"chat_id":        chatID,
			"reservation_id": req.ReservationId,
			"timestamp":      time.Now().Unix(),
		}
		
		// 使用全局WebSocket管理器发送确认消息给用户
		if err := GlobalWebSocketManager.SendMessageToUser(req.UserId, userMessage); err != nil {
			log.Printf("发送确认消息给用户 %d 失败: %v", req.UserId, err)
		}
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
