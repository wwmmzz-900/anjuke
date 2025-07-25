package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	pb "github.com/wwmmzz-900/anjuke/api/house/v3"

	"github.com/wwmmzz-900/anjuke/internal/biz"
)

// 全局WebSocketHub实例

type HouseService struct {
	pb.UnimplementedHouseServer
	uc *biz.HouseUsecase
	// 请求统计
	stats struct {
		sync.RWMutex
		totalRequests     int64
		successRequests   int64
		failedRequests    int64
		avgResponseTime   time.Duration
		lastRequestTime   time.Time
	}
}

func NewHouseService(uc *biz.HouseUsecase) *HouseService {
	service := &HouseService{
		uc: uc,
	}
	
	// 启动统计信息定期输出
	go service.logStats()
	
	return service
}

// 记录请求统计
func (s *HouseService) recordRequest(success bool, duration time.Duration) {
	s.stats.Lock()
	defer s.stats.Unlock()
	
	s.stats.totalRequests++
	s.stats.lastRequestTime = time.Now()
	
	if success {
		s.stats.successRequests++
	} else {
		s.stats.failedRequests++
	}
	
	// 计算平均响应时间（简单移动平均）
	if s.stats.totalRequests == 1 {
		s.stats.avgResponseTime = duration
	} else {
		s.stats.avgResponseTime = (s.stats.avgResponseTime + duration) / 2
	}
}

// 定期输出统计信息
func (s *HouseService) logStats() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		s.stats.RLock()
		log.Printf("房源服务统计 - 总请求: %d, 成功: %d, 失败: %d, 平均响应时间: %v",
			s.stats.totalRequests, s.stats.successRequests, s.stats.failedRequests, s.stats.avgResponseTime)
		s.stats.RUnlock()
	}
}

// 普通推荐列表（高并发优化版本）
func (s *HouseService) RecommendList(ctx context.Context, req *pb.HouseRecommendRequest) (*pb.HouseRecommendReply, error) {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		s.recordRequest(true, duration)
	}()
	
	log.Printf("接收到普通推荐请求: page=%d, pageSize=%d", req.Page, req.PageSize)

	// 参数验证和标准化
	page := int(req.Page)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// 设置请求超时
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 从业务层查询推荐房源
	houses, total, err := s.uc.RecommendList(ctx, page, pageSize)
	if err != nil {
		log.Printf("获取推荐列表失败: %v", err)
		s.recordRequest(false, time.Since(startTime))
		
		// 降级策略：返回空列表但不报错
		return &pb.HouseRecommendReply{
			Code: 0,
			Msg:  "success",
			Data: &pb.HouseRecommendData{
				Total: 0,
				List:  []*pb.HouseRecommendItem{},
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

// 个性化推荐列表（高并发优化版本）
func (s *HouseService) PersonalRecommendList(ctx context.Context, req *pb.PersonalRecommendRequest) (*pb.HouseRecommendReply, error) {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		s.recordRequest(true, duration)
	}()
	
	log.Printf("接收到个性化推荐请求: userId=%d, page=%d, pageSize=%d", req.UserId, req.Page, req.PageSize)

	// 参数验证
	if req.UserId <= 0 {
		log.Printf("无效的用户ID: %d", req.UserId)
		s.recordRequest(false, time.Since(startTime))
		return &pb.HouseRecommendReply{
			Code: 400,
			Msg:  "无效的用户ID",
			Data: &pb.HouseRecommendData{
				Total: 0,
				List:  []*pb.HouseRecommendItem{},
			},
		}, nil
	}

	// 参数标准化
	page := int(req.Page)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// 设置请求超时
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 调用业务层获取个性化推荐
	houses, total, err := s.uc.PersonalRecommendList(ctx, req.UserId, page, pageSize)
	if err != nil {
		log.Printf("获取个性化推荐失败: %v", err)
		s.recordRequest(false, time.Since(startTime))
		
		// 降级策略：返回空列表但不报错
		return &pb.HouseRecommendReply{
			Code: 0,
			Msg:  "success",
			Data: &pb.HouseRecommendData{
				Total: 0,
				List:  []*pb.HouseRecommendItem{},
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
