package biz

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	"time"
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

// todo:用户基础表
type UserBase struct {
	gorm.Model
	Name       string `json:"name"`        // 用户昵称/姓名
	RealName   string `json:"real_name"`   // 真实姓名
	Phone      string `json:"phone"`       // 手机号
	Password   string `json:"password"`    // 密码（加密存储）
	Avatar     string `json:"avatar"`      // 头像URL
	RoleId     int64  `json:"role_id"`     // 角色id
	Sex        string `json:"sex"`         // 用户性别
	RealAtatus int8   `json:"real_atatus"` // 用户实名状态(1: 已实名2:未实名 )
	Status     int8   `json:"status"`      // 状态（0禁用1正常）
}

func (*UserBase) TableName() string {
	return "user_base"
}

// UserRepo todo: 用户绑定表
type UserBinding struct {
	Id        int64     `json:"id"`         // 绑定ID
	UserId    int64     `json:"user_id"`    // 用户ID
	Type      string    `json:"type"`       // 绑定类型(phone/email/wechat/alipay/github等)
	Value     string    `json:"value"`      // 绑定值(手机号/邮箱/第三方openid等)
	Extra     string    `json:"extra"`      // 额外信息(如第三方昵称、头像等)
	CreatedAt time.Time `json:"created_at"` // 创建时间
}

func (*UserBinding) TableName() string {
	return "user_binding"
}

// UserRepo  is a user repo.
type UserRepo interface {
	//CreateUser(context.Context, *User) (*User, error)
	//GetUser(ctx context.Context, phone string) (*User, error)
	BindPhone(ctx context.Context, uid int64, phone string, binding *UserBinding) error
	CheckPhoneExists(ctx context.Context, phone string) (bool, error)
	Store(ctx context.Context, Soures, Phone, code string, expire time.Duration) error
	Get(ctx context.Context, Soures, Phone string) (string, error)
	Delete(ctx context.Context, Soures, Phone string) error
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

// todo:用户添加
//func (uc *UserUsecase) CreateUser(ctx context.Context, g *User) (*User, error) {
//	uc.log.WithContext(ctx).Infof("CreateUser: %v", g.NickName)
//	return uc.repo.CreateUser(ctx, g)
//}

// todo:根据手机号查询用户
//	func (uc *UserUsecase) GetUser(ctx context.Context, phone string) (*User, error) {
//		uc.log.WithContext(ctx).Infof("GetUser: %v", phone)
//		return uc.repo.GetUser(ctx, phone)
//	}

func (uc *UserUsecase) BindPhone(ctx context.Context, uid int64, phone string, bing *UserBinding) error {
	// 检查是否已被其他账号绑定
	if exists, err := uc.repo.CheckPhoneExists(ctx, phone); err != nil {
		return err
	} else if exists {
		return errors.New("该手机号已被绑定")
	}
	return uc.repo.BindPhone(ctx, uid, phone, bing)
}

func (uc *UserUsecase) Store(ctx context.Context, Soures, Phone, code string, expire time.Duration) error {
	return uc.repo.Store(ctx, Soures, Phone, code, expire)
}

func (uc *UserUsecase) Get(ctx context.Context, Soures, Phone string) (string, error) {
	get, err := uc.repo.Get(ctx, Soures, Phone)
	if err != nil {
		return "", err
	}
	fmt.Println(get)
	return get, nil
}

func (uc *UserUsecase) Delete(ctx context.Context, Soures, Phone string) error {
	return uc.repo.Delete(ctx, Soures, Phone)
}
