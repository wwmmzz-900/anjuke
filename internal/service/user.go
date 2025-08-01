package service

import (
	"context"
	v2 "github.com/wwmmzz-900/anjuke/api/user/v2"
	"github.com/wwmmzz-900/anjuke/internal/biz"
	"github.com/wwmmzz-900/anjuke/internal/model"
)

type UserService struct {
	v2.UnimplementedUserServer
	v2uc *biz.UserUsecase
}

func NewUserService(v2uc *biz.UserUsecase) *UserService {
	return &UserService{
		v2uc: v2uc,
	}
}

// BloggerProfileService 博主主页服务层
type BloggerProfileService struct {
	v2.UnimplementedBloggerProfileServer
	uc biz.BloggerProfileUsecaseInterface
}

// NewBloggerProfileService 创建博主主页服务层实例
func NewBloggerProfileService(uc biz.BloggerProfileUsecaseInterface) *BloggerProfileService {
	return &BloggerProfileService{
		uc: uc,
	}
}

// GetBloggerProfile 获取博主主页信息
func (s *BloggerProfileService) GetBloggerProfile(ctx context.Context, req *v2.GetBloggerProfileRequest) (*v2.GetBloggerProfileResponse, error) {
	// 参数验证
	if err := s.uc.ValidateUserId(req.UserId); err != nil {
		return nil, err
	}
	
	// 记录访问日志
	if err := s.recordAccessLog(ctx, req); err != nil {
		// 访问日志记录失败不影响主流程，只记录错误日志
		// 这里可以考虑使用日志记录，但不返回错误
	}
	
	// 调用业务逻辑层
	profile, err := s.uc.GetBloggerProfile(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	
	// 转换为响应格式
	resp := s.convertToProfileResponse(profile)
	return resp, nil
}

// GetBloggerHouses 获取博主房源列表
func (s *BloggerProfileService) GetBloggerHouses(ctx context.Context, req *v2.GetBloggerHousesRequest) (*v2.GetBloggerHousesResponse, error) {
	// 参数验证
	if err := s.uc.ValidateUserId(req.UserId); err != nil {
		return nil, err
	}
	
	// 构建分页参数
	pagination := &model.PaginationParams{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
		SortBy:   req.SortBy,
		Order:    req.Order,
	}
	pagination.ValidatePaginationParams()
	
	// 构建过滤参数
	filter := &model.HouseFilterParams{
		Status:    req.Status,
		MinPrice:  req.MinPrice,
		MaxPrice:  req.MaxPrice,
		MinArea:   req.MinArea,
		MaxArea:   req.MaxArea,
		Layout:    req.Layout,
		Keyword:   req.Keyword,
		RegionId:  req.RegionId,
	}
	filter.ValidateHouseFilterParams()
	
	// 调用业务逻辑层
	houses, paginationResult, err := s.uc.GetBloggerHouses(ctx, req.UserId, pagination, filter)
	if err != nil {
		return nil, err
	}
	
	// 转换为响应格式
	resp := &v2.GetBloggerHousesResponse{
		Houses:   s.convertToHouseInfoList(houses),
		Total:    paginationResult.Total,
		Page:     int32(paginationResult.Page),
		PageSize: int32(paginationResult.PageSize),
		Pages:    int32(paginationResult.Pages),
		HasNext:  paginationResult.HasNextPage(),
		HasPrev:  paginationResult.HasPrevPage(),
	}
	
	return resp, nil
}

// convertToProfileResponse 转换博主主页信息为响应格式
func (s *BloggerProfileService) convertToProfileResponse(profile *model.BloggerProfile) *v2.GetBloggerProfileResponse {
	resp := &v2.GetBloggerProfileResponse{
		// 用户基础信息
		UserId:      profile.UserInfo.UserId,
		Name:        profile.UserInfo.Name,
		RealName:    profile.UserInfo.RealName,
		PhoneMasked: profile.UserInfo.Phone, // 已经在usecase中脱敏
		Email:       profile.UserInfo.Email,
		Avatar:      profile.UserInfo.GetDefaultAvatar(),
		Sex:         string(profile.UserInfo.Sex),
		RealStatus:  int32(profile.UserInfo.RealStatus),
		Status:      int32(profile.UserInfo.Status),
		CreatedAt:   profile.UserInfo.CreatedAt.Unix(),
		UpdatedAt:   profile.UserInfo.UpdatedAt.Unix(),
		
		// 最近房源列表
		RecentHouses: s.convertToHouseInfoList(profile.RecentHouses),
	}
	
	return resp
}

// convertToHouseInfoList 转换房源列表为响应格式
func (s *BloggerProfileService) convertToHouseInfoList(houses []*model.House) []*v2.HouseInfo {
	var result []*v2.HouseInfo
	for _, house := range houses {
		houseInfo := &v2.HouseInfo{
			HouseId:   house.HouseId,
			Title:     house.Title,
			Address:   house.Address,
			Price:     house.Price,
			Area:      house.Area,
			Layout:    house.Layout,
			Status:    string(house.Status),
			CreatedAt: house.CreatedAt.Unix(),
		}
		result = append(result, houseInfo)
	}
	return result
}

// recordAccessLog 记录访问日志
func (s *BloggerProfileService) recordAccessLog(ctx context.Context, req *v2.GetBloggerProfileRequest) error {
	// 从context中提取请求信息（这些信息通常由HTTP中间件设置）
	accessLogRequest := &model.AccessLogRequest{
		BloggerID:     req.UserId,
		VisitorID:     s.extractVisitorID(ctx),
		VisitorIP:     s.extractVisitorIP(ctx),
		UserAgent:     s.extractUserAgent(ctx),
		Referer:       s.extractReferer(ctx),
		RequestPath:   s.extractRequestPath(ctx),
		RequestMethod: "GET", // gRPC调用默认为GET
		SessionID:     s.extractSessionID(ctx),
	}
	
	// 调用业务逻辑层记录访问日志
	return s.uc.RecordAccessLog(ctx, accessLogRequest)
}

// extractVisitorID 从context中提取访问者ID
func (s *BloggerProfileService) extractVisitorID(ctx context.Context) int64 {
	// 从context中提取用户ID，如果未登录则返回0
	if userID, ok := ctx.Value("user_id").(int64); ok {
		return userID
	}
	return 0 // 未登录用户
}

// extractVisitorIP 从context中提取访问者IP
func (s *BloggerProfileService) extractVisitorIP(ctx context.Context) string {
	if ip, ok := ctx.Value("client_ip").(string); ok {
		return ip
	}
	return "unknown"
}

// extractUserAgent 从context中提取User-Agent
func (s *BloggerProfileService) extractUserAgent(ctx context.Context) string {
	if ua, ok := ctx.Value("user_agent").(string); ok {
		return ua
	}
	return ""
}

// extractReferer 从context中提取Referer
func (s *BloggerProfileService) extractReferer(ctx context.Context) string {
	if referer, ok := ctx.Value("referer").(string); ok {
		return referer
	}
	return ""
}

// extractRequestPath 从context中提取请求路径
func (s *BloggerProfileService) extractRequestPath(ctx context.Context) string {
	if path, ok := ctx.Value("request_path").(string); ok {
		return path
	}
	return "/api/v2/blogger/profile" // 默认路径
}

// extractSessionID 从context中提取会话ID
func (s *BloggerProfileService) extractSessionID(ctx context.Context) string {
	if sessionID, ok := ctx.Value("session_id").(string); ok {
		return sessionID
	}
	return ""
}

// todo:用户登录注册一体化
//func (s *UserService) CreateUser(ctx context.Context, req *v2.CreateUserRequest) (*v2.CreateUserReply, error) {
//	user, err := s.v2uc.GetUser(ctx, req.Mobile)
//	if err != nil {
//		return nil, fmt.Errorf("查询失败: %v", err)
//	}
//
//	// 用户不存在时才创建
//	if user == nil || user.Mobile == "" {
//		_, err := s.v2uc.CreateUser(ctx, &biz.User{
//			Mobile:   req.Mobile,
//			NickName: req.NickName,
//			Password: req.Password, // 注意：密码应该加密
//			Birthday: 0,            // 设置默认值
//			Gender:   0,            // 设置默认值
//			Grade:    0,            // 设置默认值
//		})
//		if err != nil {
//			return nil, fmt.Errorf("创建用户失败: %v", err)
//		}
//		return &v2.CreateUserReply{
//			Success: "注册成功",
//		}, nil
//	}
//
//	// 用户已存在，检查密码
//	if user.Password != req.Password { // 注意：实际应该对比加密后的密码
//		return nil, fmt.Errorf("密码错误")
//	}
//	return &v2.CreateUserReply{
//		Success: "登录成功",
//	}, nil
//}
