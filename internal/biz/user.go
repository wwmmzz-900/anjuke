package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	bloggerErrors "github.com/wwmmzz-900/anjuke/internal/errors"
	"github.com/wwmmzz-900/anjuke/internal/model"
)

// todo:这个结构体就是数据库的结构体
//type User struct {
//	gorm.Model
//	Mobile   string // 账号或手机号
//	NickName string // 昵称
//	Password string // 密码
//	Birthday int32  // 生日
//	Gender   int32  // 性别（0男 1女）
//	Grade    int32  // 等级（0普通游客 1会员 2商家 3管理）
//}

// UserRepo  is a user repo.
type UserRepo interface {
	//CreateUser(context.Context, *User) (*User, error)
	//GetUser(ctx context.Context, phone string) (*User, error)
}

// BloggerProfileRepo 博主主页数据访问接口
type BloggerProfileRepo interface {
	// GetUserById 根据用户ID获取用户基础信息
	GetUserById(ctx context.Context, userId int64) (*model.UserBase, error)

	// GetUserHouseStats 获取用户房源统计信息
	GetUserHouseStats(ctx context.Context, userId int64) (*model.HouseStatistics, error)

	// GetUserInteractStats 获取用户互动统计信息
	GetUserInteractStats(ctx context.Context, userId int64) (*model.InteractStatistics, error)

	// GetUserRecentHouses 获取用户最近发布的房源
	GetUserRecentHouses(ctx context.Context, userId int64, limit int) ([]*model.House, error)

	// GetUserHouses 分页获取用户房源列表
	GetUserHouses(ctx context.Context, userId int64, pagination *model.PaginationParams, filter *model.HouseFilterParams) ([]*model.House, *model.PaginationResult, error)

	// CreateAccessLog 创建访问日志记录
	CreateAccessLog(ctx context.Context, accessLog *model.BloggerProfileAccessLog) error

	// UpdateAccessStats 更新访问统计数据
	UpdateAccessStats(ctx context.Context, bloggerID int64) error

	// GetAccessStats 获取访问统计数据
	GetAccessStats(ctx context.Context, bloggerID int64) (*model.BloggerProfileAccessStats, error)
}

// UserUsecase is a user usecase.
type UserUsecase struct {
	repo UserRepo
	log  *log.Helper
}

// NewUserUsecase new a User usecase.
func NewUserUsecase(repo UserRepo, logger log.Logger) *UserUsecase {
	return &UserUsecase{repo: repo, log: log.NewHelper(logger)}
}

// BloggerProfileUsecaseInterface 博主主页业务逻辑层接口
type BloggerProfileUsecaseInterface interface {
	GetBloggerProfile(ctx context.Context, userId int64) (*model.BloggerProfile, error)
	GetBloggerHouses(ctx context.Context, userId int64, pagination *model.PaginationParams, filter *model.HouseFilterParams) ([]*model.House, *model.PaginationResult, error)
	ValidateUserId(userId int64) error
	RecordAccessLog(ctx context.Context, request *model.AccessLogRequest) error
	GetAccessStats(ctx context.Context, bloggerID int64) (*model.BloggerProfileAccessStats, error)
}

// BloggerProfileUsecase 博主主页业务逻辑层
type BloggerProfileUsecase struct {
	repo BloggerProfileRepo
	log  *log.Helper
}

// NewBloggerProfileUsecase 创建博主主页业务逻辑层实例
func NewBloggerProfileUsecase(repo BloggerProfileRepo, logger log.Logger) *BloggerProfileUsecase {
	return &BloggerProfileUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// GetBloggerProfile 获取博主主页信息
func (uc *BloggerProfileUsecase) GetBloggerProfile(ctx context.Context, userId int64) (*model.BloggerProfile, error) {
	uc.log.WithContext(ctx).Infof("GetBloggerProfile: userId=%d", userId)

	// 1. 获取用户基础信息
	userInfo, err := uc.repo.GetUserById(ctx, userId)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("GetUserById failed: %v", err)
		return nil, bloggerErrors.WrapUserNotFoundError(userId)
	}

	// 2. 检查用户状态
	if userInfo.DeletedAt != nil {
		return nil, bloggerErrors.WrapUserNotFoundError(userId)
	}

	if userInfo.Status != model.UserStatusActive {
		return nil, bloggerErrors.WrapUserDisabledError(userId)
	}

	// 3. 获取最近房源信息
	recentHouses, err := uc.repo.GetUserRecentHouses(ctx, userId, 6)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("GetUserRecentHouses failed: %v", err)
		return nil, bloggerErrors.WrapHouseQueryError(err)
	}

	// 4. 脱敏处理
	userInfo.Phone = uc.maskPhone(userInfo.Phone)

	// 5. 构建返回结果
	profile := &model.BloggerProfile{
		UserInfo:      userInfo,
		HouseStats:    nil, // 不再获取房源统计
		InteractStats: nil, // 不再获取互动统计
		RecentHouses:  recentHouses,
	}

	return profile, nil
}

// GetBloggerHouses 获取博主房源列表
func (uc *BloggerProfileUsecase) GetBloggerHouses(ctx context.Context, userId int64, pagination *model.PaginationParams, filter *model.HouseFilterParams) ([]*model.House, *model.PaginationResult, error) {
	uc.log.WithContext(ctx).Infof("GetBloggerHouses: userId=%d, pagination=%+v, filter=%+v", userId, pagination, filter)

	// 参数验证和默认值设置
	if pagination == nil {
		pagination = &model.PaginationParams{
			Page:     model.DefaultPage,
			PageSize: model.DefaultPageSize,
		}
	}
	pagination.ValidatePaginationParams()

	if filter == nil {
		filter = &model.HouseFilterParams{}
	}
	filter.ValidateHouseFilterParams()

	// 检查用户是否存在
	userInfo, err := uc.repo.GetUserById(ctx, userId)
	if err != nil {
		return nil, nil, bloggerErrors.WrapUserNotFoundError(userId)
	}

	if userInfo.DeletedAt != nil {
		return nil, nil, bloggerErrors.WrapUserNotFoundError(userId)
	}

	if userInfo.Status != model.UserStatusActive {
		return nil, nil, bloggerErrors.WrapUserDisabledError(userId)
	}

	// 获取房源列表
	houses, paginationResult, err := uc.repo.GetUserHouses(ctx, userId, pagination, filter)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("GetUserHouses failed: %v", err)
		return nil, nil, bloggerErrors.WrapHouseQueryError(err)
	}

	return houses, paginationResult, nil
}

// maskPhone 手机号脱敏处理
func (uc *BloggerProfileUsecase) maskPhone(phone string) string {
	if len(phone) != 11 {
		return phone
	}
	return phone[:3] + "****" + phone[7:]
}

// ValidateUserId 验证用户ID
func (uc *BloggerProfileUsecase) ValidateUserId(userId int64) error {
	if userId <= 0 {
		return bloggerErrors.WrapInvalidUserIdError(userId)
	}
	return nil
}

// RecordAccessLog 记录访问日志
func (uc *BloggerProfileUsecase) RecordAccessLog(ctx context.Context, request *model.AccessLogRequest) error {
	uc.log.WithContext(ctx).Infof("RecordAccessLog: bloggerID=%d, visitorID=%d, ip=%s",
		request.BloggerID, request.VisitorID, request.VisitorIP)

	// 验证请求参数
	if err := request.Validate(); err != nil {
		uc.log.WithContext(ctx).Errorf("AccessLogRequest validation failed: %v", err)
		return bloggerErrors.WrapInvalidUserIdError(request.BloggerID)
	}

	// 转换为访问日志模型
	accessLog := request.ToAccessLog()

	// 记录开始时间用于计算响应时间
	startTime := time.Now()

	// 创建访问日志记录
	err := uc.repo.CreateAccessLog(ctx, accessLog)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("CreateAccessLog failed: %v", err)
		return err
	}

	// 计算响应时间
	responseTime := time.Since(startTime).Milliseconds()
	accessLog.ResponseTime = responseTime

	// 异步更新访问统计数据（避免影响主流程性能）
	go uc.updateAccessStatsAsync(request.BloggerID)

	uc.log.WithContext(ctx).Infof("AccessLog recorded successfully: id=%d, responseTime=%dms",
		accessLog.ID, responseTime)

	return nil
}

// updateAccessStatsAsync 异步更新访问统计数据
func (uc *BloggerProfileUsecase) updateAccessStatsAsync(bloggerID int64) {
	// 创建新的context避免超时问题
	ctx := context.Background()

	// 添加延迟避免频繁更新
	time.Sleep(100 * time.Millisecond)

	if err := uc.repo.UpdateAccessStats(ctx, bloggerID); err != nil {
		uc.log.WithContext(ctx).Errorf("UpdateAccessStats failed: %v", err)
	} else {
		uc.log.WithContext(ctx).Infof("AccessStats updated successfully for bloggerID=%d", bloggerID)
	}
}

// GetAccessStats 获取访问统计数据
func (uc *BloggerProfileUsecase) GetAccessStats(ctx context.Context, bloggerID int64) (*model.BloggerProfileAccessStats, error) {
	uc.log.WithContext(ctx).Infof("GetAccessStats: bloggerID=%d", bloggerID)

	// 验证博主ID
	if err := uc.ValidateUserId(bloggerID); err != nil {
		return nil, err
	}

	// 获取访问统计数据
	stats, err := uc.repo.GetAccessStats(ctx, bloggerID)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("GetAccessStats failed: %v", err)
		return nil, bloggerErrors.WrapStatsQueryError(err)
	}

	return stats, nil
}

// todo:用户添加
//func (uc *UserUsecase) CreateUser(ctx context.Context, g *User) (*User, error) {
//	uc.log.WithContext(ctx).Infof("CreateUser: %v", g.NickName)
//	return uc.repo.CreateUser(ctx, g)
//}

// todo:根据手机号查询用户
//func (uc *UserUsecase) GetUser(ctx context.Context, phone string) (*User, error) {
//	uc.log.WithContext(ctx).Infof("GetUser: %v", phone)
//	return uc.repo.GetUser(ctx, phone)
//}
