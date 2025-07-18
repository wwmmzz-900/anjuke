package biz

import (
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

// UserRepo  is a user repo.
type UserRepo interface {
	//CreateUser(context.Context, *User) (*User, error)
	//GetUser(ctx context.Context, phone string) (*User, error)
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
//func (uc *UserUsecase) GetUser(ctx context.Context, phone string) (*User, error) {
//	uc.log.WithContext(ctx).Infof("GetUser: %v", phone)
//	return uc.repo.GetUser(ctx, phone)
//}
