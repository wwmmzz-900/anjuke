package biz

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
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
	UserId     int64  `json:"user_id"`     // 用户唯一ID
	Name       string `json:"name"`        // 用户昵称/姓名
	RealName   string `json:"real_name"`   // 真实姓名
	Phone      string `json:"phone"`       // 手机号
	Email      string `json:"email"`       // 邮箱
	Password   string `json:"password"`    // 密码（加密存储）
	Avatar     string `json:"avatar"`      // 头像URL
	RoleId     int64  `json:"role_id"`     // 角色id
	Sex        string `json:"sex"`         // 用户性别
	RealStatus int8   `json:"real_status"` // 用户实名状态(1: 已实名2:未实名 )
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
	CreateUser(ctx context.Context, user *UserBase) (*UserBase, error)
	GetUserByAccount(ctx context.Context, account string) (*UserBase, error)
	BindPhone(ctx context.Context, uid int64, phone string, binding *UserBinding) error
	CheckPhoneExists(ctx context.Context, phone string) (bool, error)
	Store(ctx context.Context, Soures, Phone, code string, expire time.Duration) error
	Get(ctx context.Context, Soures, Phone string) (string, error)
	Delete(ctx context.Context, Soures, Phone string) error
	CheckEmailExists(ctx context.Context, email string) (bool, error)
	BindEmail(ctx context.Context, uid int64, email string, binding *UserBinding) error
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

func (uc *UserUsecase) BindEmail(ctx context.Context, uid int64, email string, bing *UserBinding) error {
	// 检查是否已被其他账号绑定
	if exists, err := uc.repo.CheckEmailExists(ctx, email); err != nil {
		return err
	} else if exists {
		return errors.New("该邮箱已被绑定")
	}
	return uc.repo.BindEmail(ctx, uid, email, bing)
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

// LoginOrRegister 用户登录注册一体化
func (uc *UserUsecase) LoginOrRegister(ctx context.Context, account, password, name, sex string) (*UserBase, bool, error) {
	// 先尝试查找用户
	user, err := uc.repo.GetUserByAccount(ctx, account)
	if err != nil {
		return nil, false, fmt.Errorf("查询用户失败: %v", err)
	}

	// 用户存在，验证密码
	if user != nil {
		if !uc.verifyPassword(user.Password, password) {
			return nil, false, errors.New("密码错误")
		}
		return user, false, nil // 登录成功，不是新用户
	}

	// 用户不存在，创建新用户
	hashedPassword, err := uc.hashPassword(password)
	if err != nil {
		return nil, false, fmt.Errorf("密码加密失败: %v", err)
	}

	newUser := &UserBase{
		Name:       name,
		Password:   hashedPassword,
		Sex:        sex,
		RealStatus: 2, // 未实名
		Status:     1, // 正常状态
	}

	// 判断账号类型并设置相应字段
	if uc.isEmail(account) {
		newUser.Email = account
	} else {
		newUser.Phone = account
	}

	createdUser, err := uc.repo.CreateUser(ctx, newUser)
	if err != nil {
		return nil, false, fmt.Errorf("创建用户失败: %v", err)
	}

	return createdUser, true, nil // 注册成功，是新用户
}

// hashPassword 密码加密（简化版，实际应使用bcrypt）
func (uc *UserUsecase) hashPassword(password string) (string, error) {
	// 这里使用简单的加密方式，实际项目中应该使用bcrypt
	return fmt.Sprintf("hashed_%s", password), nil
}

// verifyPassword 验证密码
func (uc *UserUsecase) verifyPassword(hashedPassword, password string) bool {
	// 简化版密码验证
	expectedHash := fmt.Sprintf("hashed_%s", password)
	return hashedPassword == expectedHash
}

// isEmail 判断是否为邮箱格式
func (uc *UserUsecase) isEmail(account string) bool {
	// 简单的邮箱格式判断，检查是否包含@符号
	for i := 0; i < len(account); i++ {
		if account[i] == '@' {
			return true
		}
	}
	return false
}
