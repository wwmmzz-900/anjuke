# 安居客房源系统核心接口技术分析文档

## 概述

本文档深入分析安居客房源系统中四个核心接口的技术实现，包括需求分析、技术难点、解决方案、性能优化和扩展方案。

---

## 1. 房源推荐接口 (RecommendList)

### 1.1 需求分析

**业务需求：**
- 为用户提供通用的房源推荐列表
- 支持分页查询，提升用户体验
- 返回房源基本信息（标题、描述、价格、面积、户型、图片）
- 确保数据的实时性和准确性

**技术需求：**
- 高并发访问支持
- 快速响应时间（< 200ms）
- 数据一致性保证
- 容错和降级机制

### 1.2 技术难点

1. **数据量大**：房源数据可能达到百万级别
2. **查询性能**：多表关联查询的性能优化
3. **图片加载**：批量获取房源图片避免N+1问题
4. **缓存策略**：如何平衡数据实时性和性能
5. **容错处理**：数据库异常时的降级策略

### 1.3 解决方案

**架构设计：**
```go
// Service层 - 业务逻辑处理
func (s *HouseService) RecommendList(ctx context.Context, req *pb.HouseRecommendRequest) (*pb.HouseRecommendReply, error) {
    // 1. 参数验证和默认值设置
    page := int(req.Page)
    if page <= 0 { page = 1 }
    
    pageSize := int(req.PageSize)
    if pageSize <= 0 { pageSize = 10 }
    
    // 2. 调用业务层
    houses, total, err := s.uc.RecommendList(ctx, page, pageSize)
    
    // 3. 错误处理和降级
    if err != nil {
        return s.getDefaultRecommendList(), nil
    }
    
    // 4. 数据转换
    return s.convertToProtobuf(houses, total), nil
}
```

**数据层优化：**
```go
// 批量查询优化
func (r *houseRepo) GetRecommendList(ctx context.Context, page, pageSize int) ([]*biz.House, int, error) {
    // 1. 主查询 - 获取房源基本信息
    err = r.data.db.
        Table("house").
        Where("status = 'active'").
        Order("house_id DESC").
        Limit(pageSize).
        Offset((page - 1) * pageSize).
        Scan(&results).Error
    
    // 2. 批量获取图片 - 避免N+1查询
    imageMap := r.getHouseImages(houseIDs)
    
    return houses, int(total), nil
}
```

### 1.4 性能优化

1. **数据库优化**
   - 在 `status` 字段上建立索引
   - 使用 `house_id DESC` 索引优化排序
   - 分页查询使用 LIMIT/OFFSET

2. **查询优化**
   - 批量获取房源图片，减少数据库连接
   - 使用连接池管理数据库连接
   - 预编译SQL语句

3. **缓存策略**
   ```go
   // Redis缓存实现（扩展方案）
   func (r *houseRepo) GetRecommendListWithCache(ctx context.Context, page, pageSize int) {
       cacheKey := fmt.Sprintf("recommend_list:%d:%d", page, pageSize)
       
       // 1. 尝试从缓存获取
       if cached := r.redis.Get(cacheKey); cached != nil {
           return cached, nil
       }
       
       // 2. 从数据库查询
       result := r.queryFromDB(page, pageSize)
       
       // 3. 写入缓存（TTL: 5分钟）
       r.redis.Set(cacheKey, result, 5*time.Minute)
       
       return result, nil
   }
   ```

### 1.5 扩展方案

1. **多维度排序**
   - 支持按价格、面积、发布时间等多种排序方式
   - 实现排序算法的可配置化

2. **地理位置过滤**
   - 集成地理位置服务
   - 支持按距离、区域筛选

3. **实时推荐**
   - 集成推荐算法引擎
   - 基于用户行为实时调整推荐结果

---

## 2. 个性化推荐接口 (PersonalRecommendList)

### 2.1 需求分析

**业务需求：**
- 基于用户浏览历史提供个性化推荐
- 分析用户偏好（价格区间、户型、地理位置等）
- 冷启动问题处理（新用户推荐策略）
- 推荐结果的多样性和准确性平衡

**技术需求：**
- 用户行为数据收集和分析
- 实时推荐算法计算
- 推荐结果的个性化排序
- 推荐效果的评估和优化

### 2.2 技术难点

1. **冷启动问题**：新用户没有历史数据
2. **实时性要求**：用户行为变化后推荐结果的更新
3. **算法复杂度**：推荐算法的计算效率
4. **数据稀疏性**：用户行为数据不足的处理
5. **推荐多样性**：避免推荐结果过于单一

### 2.3 解决方案

**用户画像构建：**
```go
// 分析用户价格偏好
func (r *houseRepo) GetUserPricePreference(ctx context.Context, userID int64) (float64, float64, error) {
    // 分析最近20次浏览记录
    err := r.data.db.
        Table("user_behavior AS ub").
        Select("MIN(h.price) AS min_price, MAX(h.price) AS max_price").
        Joins("JOIN house h ON ub.house_id = h.house_id").
        Where("ub.user_id = ? AND ub.behavior = 'view'", userID).
        Order("ub.created_at DESC").
        Limit(20).
        Scan(&res).Error
    
    // 冷启动处理
    if res.MinPrice == 0 && res.MaxPrice == 0 {
        return 800, 5000, nil // 默认价格区间
    }
    
    return res.MinPrice, res.MaxPrice, nil
}
```

**推荐算法实现：**
```go
// 个性化推荐业务逻辑
func (uc *HouseUsecase) PersonalRecommendList(ctx context.Context, userID int64, page, pageSize int) ([]*House, int, error) {
    // 1. 获取用户偏好
    minPrice, maxPrice, err := uc.repo.GetUserPricePreference(ctx, userID)
    
    // 2. 基于偏好过滤房源
    houses, total, err := uc.repo.GetPersonalRecommendList(ctx, minPrice, maxPrice, page, pageSize)
    
    // 3. 个性化排序（可扩展）
    houses = uc.personalizedSort(houses, userID)
    
    return houses, total, nil
}
```

### 2.4 性能优化

1. **用户行为数据优化**
   ```sql
   -- 用户行为表索引优化
   CREATE INDEX idx_user_behavior_user_time ON user_behavior(user_id, created_at DESC);
   CREATE INDEX idx_user_behavior_house ON user_behavior(house_id);
   ```

2. **推荐结果缓存**
   ```go
   // 个性化推荐缓存策略
   func (uc *HouseUsecase) PersonalRecommendListWithCache(ctx context.Context, userID int64) {
       cacheKey := fmt.Sprintf("personal_recommend:%d", userID)
       
       // 缓存时间较短（1分钟），保证个性化的实时性
       if cached := uc.cache.Get(cacheKey); cached != nil {
           return cached, nil
       }
       
       result := uc.calculatePersonalRecommend(userID)
       uc.cache.Set(cacheKey, result, 1*time.Minute)
       
       return result, nil
   }
   ```

3. **异步计算**
   ```go
   // 异步更新用户画像
   func (uc *HouseUsecase) UpdateUserProfileAsync(userID int64, behavior string) {
       go func() {
           // 异步更新用户偏好模型
           uc.updateUserPreference(userID, behavior)
           
           // 预计算推荐结果
           uc.preCalculateRecommendations(userID)
       }()
   }
   ```

### 2.5 扩展方案

1. **多维度推荐**
   ```go
   // 扩展用户偏好分析
   type UserPreference struct {
       PriceRange    [2]float64  // 价格区间
       AreaRange     [2]float64  // 面积区间
       PreferredLayout []string  // 偏好户型
       LocationPrefs []int64     // 偏好地区
       Facilities    []string    // 偏好设施
   }
   ```

2. **机器学习集成**
   ```go
   // 集成推荐算法服务
   type MLRecommendService struct {
       client *grpc.ClientConn
   }
   
   func (s *MLRecommendService) GetPersonalizedRecommendations(userID int64) ([]*House, error) {
       // 调用机器学习服务获取推荐结果
       return s.client.Recommend(userID)
   }
   ```

3. **实时特征工程**
   ```go
   // 实时特征计算
   func (uc *HouseUsecase) CalculateRealTimeFeatures(userID int64) *UserFeatures {
       return &UserFeatures{
           RecentViewCount:    uc.getRecentViewCount(userID),
           AverageBrowseTime:  uc.getAverageBrowseTime(userID),
           PreferredTimeSlot:  uc.getPreferredBrowseTime(userID),
           DeviceType:         uc.getUserDeviceType(userID),
       }
   }
   ```

---

## 3. 预约看房接口 (ReserveHouse)

### 3.1 需求分析

**业务需求：**
- 用户预约房源看房功能
- 防止重复预约同一房源
- 实时通知房东和用户
- 预约状态管理和跟踪
- 预约时间冲突检测

**技术需求：**
- 事务一致性保证
- 实时消息推送
- 并发预约处理
- 数据完整性约束

### 3.2 技术难点

1. **并发控制**：多用户同时预约同一时段
2. **事务管理**：预约创建和消息推送的一致性
3. **实时通知**：WebSocket消息推送的可靠性
4. **状态管理**：预约状态的复杂流转
5. **时间冲突**：预约时间段的冲突检测

### 3.3 解决方案

**事务处理：**
```go
func (uc *HouseUsecase) ReserveHouse(ctx context.Context, req *pb.ReserveHouseRequest) error {
    // 开启数据库事务
    return uc.repo.WithTransaction(ctx, func(tx *gorm.DB) error {
        // 1. 检查重复预约
        if exists, err := uc.repo.HasReservation(ctx, req.UserId, req.HouseId); err != nil {
            return err
        } else if exists {
            return fmt.Errorf("您已预约过该房源")
        }
        
        // 2. 检查时间冲突
        if conflict, err := uc.repo.CheckTimeConflict(ctx, req.HouseId, req.ReserveTime); err != nil {
            return err
        } else if conflict {
            return fmt.Errorf("该时间段已被预约")
        }
        
        // 3. 创建预约记录
        reservation := &model.HouseReservation{
            LandlordID:  req.LandlordId,
            UserID:      req.UserId,
            UserName:    req.UserName,
            HouseID:     req.HouseId,
            HouseTitle:  req.HouseTitle,
            ReserveTime: req.ReserveTime,
            Status:      "pending",
            CreatedAt:   time.Now().Unix(),
        }
        
        return uc.repo.CreateReservation(ctx, reservation)
    })
}
```

**实时通知系统：**
```go
func (s *HouseService) ReserveHouse(ctx context.Context, req *pb.ReserveHouseRequest) (*pb.ReserveHouseReply, error) {
    // 1. 业务逻辑处理
    err := s.uc.ReserveHouse(ctx, req)
    if err != nil {
        return &pb.ReserveHouseReply{Code: 400, Msg: err.Error()}, nil
    }
    
    // 2. 异步发送通知
    go s.sendReservationNotifications(req)
    
    return &pb.ReserveHouseReply{Code: 0, Msg: "预约成功"}, nil
}

func (s *HouseService) sendReservationNotifications(req *pb.ReserveHouseRequest) {
    reservationID := time.Now().Unix()
    
    // 通知房东
    landlordMessage := map[string]interface{}{
        "type":           "reservation_created",
        "title":          "新的预约请求",
        "message":        fmt.Sprintf("用户 %s 预约了您的房源《%s》", req.UserName, req.HouseTitle),
        "reservation_id": reservationID,
        "timestamp":      time.Now().Unix(),
    }
    
    // 使用WebSocket管理器发送消息
    if err := GlobalWebSocketManager.SendMessageToUser(req.LandlordId, landlordMessage); err != nil {
        log.Printf("推送消息给房东失败: %v", err)
        // 可以考虑使用消息队列重试
    }
    
    // 通知用户
    userMessage := map[string]interface{}{
        "type":           "reservation_created",
        "title":          "预约成功",
        "message":        fmt.Sprintf("您已成功预约房源《%s》", req.HouseTitle),
        "reservation_id": reservationID,
        "timestamp":      time.Now().Unix(),
    }
    
    GlobalWebSocketManager.SendMessageToUser(req.UserId, userMessage)
}
```

### 3.4 性能优化

1. **数据库优化**
   ```sql
   -- 预约表索引优化
   CREATE INDEX idx_reservation_user_house ON house_reservations(user_id, house_id);
   CREATE INDEX idx_reservation_house_time ON house_reservations(house_id, reserve_time);
   CREATE INDEX idx_reservation_status ON house_reservations(status);
   ```

2. **并发控制**
   ```go
   // 使用分布式锁防止并发预约
   func (uc *HouseUsecase) ReserveHouseWithLock(ctx context.Context, req *pb.ReserveHouseRequest) error {
       lockKey := fmt.Sprintf("reserve_lock:%d:%s", req.HouseId, req.ReserveTime)
       
       // 获取分布式锁
       lock, err := uc.redis.Lock(lockKey, 30*time.Second)
       if err != nil {
           return fmt.Errorf("获取锁失败: %v", err)
       }
       defer lock.Unlock()
       
       // 执行预约逻辑
       return uc.ReserveHouse(ctx, req)
   }
   ```

3. **消息队列集成**
   ```go
   // 使用消息队列确保通知可靠性
   func (s *HouseService) publishReservationEvent(req *pb.ReserveHouseRequest) {
       event := &ReservationEvent{
           Type:          "reservation_created",
           ReservationID: generateReservationID(),
           LandlordID:    req.LandlordId,
           UserID:        req.UserId,
           HouseID:       req.HouseId,
           Timestamp:     time.Now().Unix(),
       }
       
       // 发布到消息队列
       s.messageQueue.Publish("reservation.events", event)
   }
   ```

### 3.5 扩展方案

1. **预约状态机**
   ```go
   type ReservationStatus string
   
   const (
       StatusPending   ReservationStatus = "pending"
       StatusConfirmed ReservationStatus = "confirmed"
       StatusCancelled ReservationStatus = "cancelled"
       StatusCompleted ReservationStatus = "completed"
   )
   
   // 状态转换规则
   func (r *Reservation) CanTransitionTo(newStatus ReservationStatus) bool {
       transitions := map[ReservationStatus][]ReservationStatus{
           StatusPending:   {StatusConfirmed, StatusCancelled},
           StatusConfirmed: {StatusCompleted, StatusCancelled},
           StatusCancelled: {},
           StatusCompleted: {},
       }
       
       allowedTransitions := transitions[r.Status]
       for _, allowed := range allowedTransitions {
           if allowed == newStatus {
               return true
           }
       }
       return false
   }
   ```

2. **智能调度系统**
   ```go
   // 预约时间智能推荐
   func (uc *HouseUsecase) SuggestAvailableTimeSlots(houseID int64, date string) ([]TimeSlot, error) {
       // 1. 获取已预约时间段
       bookedSlots := uc.repo.GetBookedTimeSlots(houseID, date)
       
       // 2. 生成可用时间段
       availableSlots := uc.generateAvailableSlots(date, bookedSlots)
       
       // 3. 根据房东偏好排序
       return uc.sortByLandlordPreference(houseID, availableSlots), nil
   }
   ```

---

## 4. 用户私聊接口 (StartChat)

### 4.1 需求分析

**业务需求：**
- 基于预约建立用户与房东的聊天会话
- 支持实时消息收发
- 消息持久化存储
- 消息状态管理（已读/未读）
- 聊天会话管理

**技术需求：**
- WebSocket实时通信
- 消息加密和安全性
- 高并发消息处理
- 消息可靠性保证
- 离线消息处理

### 4.2 技术难点

1. **实时性要求**：消息的实时传输和处理
2. **连接管理**：WebSocket连接的生命周期管理
3. **消息可靠性**：确保消息不丢失
4. **安全性**：消息内容的加密和权限控制
5. **扩展性**：支持大量并发连接

### 4.3 解决方案

**WebSocket连接管理：**
```go
// WebSocket管理器
type WebSocketManager struct {
    mutex           sync.RWMutex
    connections     map[int64]*websocket.Conn    // userID -> connection
    houseConnections map[int64]map[int64]*websocket.Conn // houseID -> userID -> connection
    sessionKeys     map[int64]string             // userID -> sessionKey
}

// 添加连接
func (wm *WebSocketManager) AddConnection(userID int64, conn *websocket.Conn) {
    wm.mutex.Lock()
    defer wm.mutex.Unlock()
    
    // 关闭旧连接
    if oldConn, exists := wm.connections[userID]; exists {
        oldConn.Close()
    }
    
    wm.connections[userID] = conn
}

// 发送消息
func (wm *WebSocketManager) SendMessageToUser(userID int64, message interface{}) error {
    wm.mutex.RLock()
    conn, exists := wm.connections[userID]
    wm.mutex.RUnlock()
    
    if !exists {
        return fmt.Errorf("用户 %d 未连接", userID)
    }
    
    jsonData, err := json.Marshal(message)
    if err != nil {
        return err
    }
    
    return conn.WriteMessage(websocket.TextMessage, jsonData)
}
```

**聊天会话管理：**
```go
func (s *ChatService) StartChat(ctx context.Context, req *pb.StartChatRequest) (*pb.StartChatReply, error) {
    // 1. 参数验证
    if req.ReservationId <= 0 || req.UserId <= 0 || req.LandlordId <= 0 {
        return &pb.StartChatReply{Code: 400, Msg: "无效的参数"}, nil
    }
    
    // 2. 检查或创建聊天会话
    var chatID string
    if exists, _ := s.chatRepo.ChatSessionExists(ctx, req.ReservationId); exists {
        session, _ := s.chatRepo.GetChatSessionByReservationID(ctx, req.ReservationId)
        chatID = session.ChatID
    } else {
        chatID, _ = s.chatRepo.CreateChatSession(ctx, req.ReservationId, req.UserId, req.LandlordId, 0)
    }
    
    // 3. 处理初始消息
    if req.InitialMessage != "" {
        s.handleInitialMessage(ctx, chatID, req)
    }
    
    return &pb.StartChatReply{
        Code: 0,
        Msg:  "发起聊天成功",
        Data: &pb.StartChatData{ChatId: chatID, Success: true},
    }, nil
}
```

**消息加密处理：**
```go
// 消息加密发送
func (s *ChatService) SendEncryptedMessage(ctx context.Context, senderID, receiverID int64, content string) error {
    // 1. 获取会话密钥
    sessionKey, exists := GlobalWebSocketManager.GetSessionKey(senderID)
    if !exists {
        return fmt.Errorf("会话密钥不存在")
    }
    
    // 2. 加密消息内容
    key := GenerateKey(sessionKey)
    encryptedContent, err := EncryptMessage([]byte(content), key)
    if err != nil {
        return fmt.Errorf("消息加密失败: %v", err)
    }
    
    // 3. 发送加密消息
    message := map[string]interface{}{
        "type":      "chat",
        "from":      senderID,
        "to":        receiverID,
        "content":   encryptedContent,
        "encrypted": true,
        "timestamp": time.Now().Unix(),
        "sequence":  GlobalSequenceManager.GetNextSequence(senderID),
    }
    
    return GlobalWebSocketManager.SendMessageToUser(receiverID, message)
}
```

### 4.4 性能优化

1. **连接池管理**
   ```go
   // WebSocket连接池优化
   type ConnectionPool struct {
       maxConnections int
       activeConns    map[int64]*websocket.Conn
       connQueue      chan *websocket.Conn
       mutex          sync.RWMutex
   }
   
   func (cp *ConnectionPool) AcquireConnection(userID int64) (*websocket.Conn, error) {
       cp.mutex.RLock()
       if conn, exists := cp.activeConns[userID]; exists {
           cp.mutex.RUnlock()
           return conn, nil
       }
       cp.mutex.RUnlock()
       
       // 从连接池获取连接
       select {
       case conn := <-cp.connQueue:
           cp.mutex.Lock()
           cp.activeConns[userID] = conn
           cp.mutex.Unlock()
           return conn, nil
       default:
           return nil, fmt.Errorf("连接池已满")
       }
   }
   ```

2. **消息批处理**
   ```go
   // 批量处理消息
   type MessageBatcher struct {
       messages []ChatMessage
       batchSize int
       ticker   *time.Ticker
       mutex    sync.Mutex
   }
   
   func (mb *MessageBatcher) AddMessage(msg ChatMessage) {
       mb.mutex.Lock()
       defer mb.mutex.Unlock()
       
       mb.messages = append(mb.messages, msg)
       
       if len(mb.messages) >= mb.batchSize {
           go mb.flushMessages()
       }
   }
   
   func (mb *MessageBatcher) flushMessages() {
       mb.mutex.Lock()
       messages := make([]ChatMessage, len(mb.messages))
       copy(messages, mb.messages)
       mb.messages = mb.messages[:0]
       mb.mutex.Unlock()
       
       // 批量写入数据库
       mb.batchInsertMessages(messages)
   }
   ```

3. **消息队列集成**
   ```go
   // 使用消息队列处理离线消息
   func (s *ChatService) HandleOfflineMessage(userID int64, message ChatMessage) {
       // 检查用户是否在线
       if !GlobalWebSocketManager.IsUserOnline(userID) {
           // 用户离线，将消息放入队列
           s.messageQueue.Push(fmt.Sprintf("offline_messages:%d", userID), message)
           
           // 发送推送通知
           s.pushNotificationService.SendNotification(userID, "您有新消息")
       } else {
           // 用户在线，直接发送
           GlobalWebSocketManager.SendMessageToUser(userID, message)
       }
   }
   ```

### 4.5 扩展方案

1. **多媒体消息支持**
   ```go
   // 扩展消息类型
   type MessageType int
   
   const (
       TextMessage     MessageType = 0
       ImageMessage    MessageType = 1
       VoiceMessage    MessageType = 2
       VideoMessage    MessageType = 3
       LocationMessage MessageType = 4
       FileMessage     MessageType = 5
   )
   
   // 多媒体消息处理
   func (s *ChatService) HandleMultimediaMessage(ctx context.Context, msgType MessageType, content []byte) error {
       switch msgType {
       case ImageMessage:
           return s.handleImageMessage(ctx, content)
       case VoiceMessage:
           return s.handleVoiceMessage(ctx, content)
       case VideoMessage:
           return s.handleVideoMessage(ctx, content)
       default:
           return s.handleTextMessage(ctx, string(content))
       }
   }
   ```

2. **聊天机器人集成**
   ```go
   // AI聊天助手
   type ChatBot struct {
       nlpService *NLPService
       knowledge  *KnowledgeBase
   }
   
   func (cb *ChatBot) ProcessMessage(message string) (string, error) {
       // 1. 意图识别
       intent, err := cb.nlpService.RecognizeIntent(message)
       if err != nil {
           return "", err
       }
       
       // 2. 知识库查询
       response, err := cb.knowledge.Query(intent)
       if err != nil {
           return "抱歉，我没有理解您的问题", nil
       }
       
       return response, nil
   }
   ```

3. **消息审核系统**
   ```go
   // 消息内容审核
   type MessageModerator struct {
       sensitiveWords []string
       aiModerator    *AIModerationService
   }
   
   func (mm *MessageModerator) ModerateMessage(content string) (bool, string, error) {
       // 1. 敏感词过滤
       if mm.containsSensitiveWords(content) {
           return false, "消息包含敏感词", nil
       }
       
       // 2. AI内容审核
       result, err := mm.aiModerator.Moderate(content)
       if err != nil {
           return true, "", err // 审核失败时允许通过
       }
       
       if !result.IsAppropriate {
           return false, result.Reason, nil
       }
       
       return true, "", nil
   }
   ```

---

## 总结

### 技术架构优势

1. **分层架构**：清晰的职责分离，便于维护和扩展
2. **微服务设计**：支持独立部署和扩展
3. **事件驱动**：异步处理提升系统性能
4. **容错机制**：完善的错误处理和降级策略

### 性能特点

1. **高并发支持**：通过连接池、缓存、异步处理等技术
2. **低延迟**：WebSocket实时通信，消息队列异步处理
3. **高可用**：数据库事务、分布式锁、消息重试机制
4. **可扩展**：水平扩展支持，负载均衡

### 未来优化方向

1. **智能化**：集成AI推荐算法、智能客服
2. **实时性**：流式处理、事件溯源
3. **安全性**：端到端加密、身份认证
4. **监控**：全链路监控、性能分析

这套系统架构为房源推荐和用户交互提供了完整的解决方案，具备良好的扩展性和维护性。