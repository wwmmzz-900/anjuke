package biz

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
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

// UserBase 用户基础信息表
type UserBase struct {
	UserId     int64  `json:"user_id"`     // 用户唯一ID
	Name       string `json:"name"`        // 用户昵称/姓名
	RealName   string `json:"real_name"`   // 真实姓名
	Phone      string `json:"phone"`       // 手机号
	Password   string `json:"password"`    // 密码（加密存储）
	Avatar     string `json:"avatar"`      // 头像URL
	RoleId     int64  `json:"role_id"`     // 角色id
	Sex        string `json:"sex"`         // 用户性别
	RealStatus int8   `json:"real_status"` // 用户实名状态(1: 已实名2:未实名 )
	Status     int8   `json:"status"`      // 状态（0禁用1正常）
}

func (UserBase) TableName() string {
	return "user_base"
}

// RealName 实名认证表
type RealName struct {
	Id        uint64    `json:"id"`
	UserId    uint32    `json:"user_id"`    // 用户id
	Name      string    `json:"name"`       // 姓名
	IdCard    string    `json:"id_card"`    // 身份证号码
	CreatedAt time.Time `json:"created_at"` // 创建时间
}

func (RealName) TableName() string {
	return "real_name"
}

// QueryCertifyResult 查询认证结果
type QueryCertifyResult struct {
	Passed       bool   `json:"passed"`         // 认证是否通过
	Status       string `json:"status"`         // 认证状态
	RealName     string `json:"real_name"`      // 真实姓名
	IdCardNumber string `json:"id_card_number"` // 身份证号
	FailReason   string `json:"fail_reason"`    // 失败原因
	Message      string `json:"message"`        // 响应消息
}

// LoginLog 登录日志表
type LoginLog struct {
	Id          int64     `json:"id"`           // 日志ID
	UserId      int64     `json:"user_id"`      // 用户ID
	IpAddress   string    `json:"ip_address"`   // IP地址
	UserAgent   string    `json:"user_agent"`   // 用户代理
	DeviceInfo  string    `json:"device_info"`  // 设备信息
	Location    string    `json:"location"`     // 登录地点
	LoginStatus int8      `json:"login_status"` // 登录状态 (1:成功 0:失败)
	FailReason  string    `json:"fail_reason"`  // 失败原因
	LoginTime   time.Time `json:"login_time"`   // 登录时间
	CreatedAt   time.Time `json:"created_at"`   // 创建时间
}

func (LoginLog) TableName() string {
	return "login_log"
}

// AccountOperation 账号操作记录表
type AccountOperation struct {
	Id        int64     `json:"id"`         // 记录ID
	UserId    int64     `json:"user_id"`    // 用户ID
	AdminId   int64     `json:"admin_id"`   // 操作管理员ID
	Operation string    `json:"operation"`  // 操作类型 (freeze/unfreeze/reset_password)
	Reason    string    `json:"reason"`     // 操作原因
	OldStatus int8      `json:"old_status"` // 操作前状态
	NewStatus int8      `json:"new_status"` // 操作后状态
	IpAddress string    `json:"ip_address"` // 操作IP
	CreatedAt time.Time `json:"created_at"` // 创建时间
}

func (AccountOperation) TableName() string {
	return "account_operation"
}

// UserRepo  is a user repo.
type UserRepo interface {
	//CreateUser(context.Context, *User) (*User, error)
	//GetUser(ctx context.Context, phone string) (*User, error)
	SendSms(ctx context.Context, phone, source string) error                                                         // 验证码
	CreateUser(ctx context.Context, user *UserBase) (*UserBase, error)                                               // 用户添加
	GetUser(ctx context.Context, phone string) (*UserBase, error)                                                    // 根据手机号查询用户
	GetUserID(ctx context.Context, id int64) (*UserBase, error)                                                      // 根据用户id查询用户
	Login(ctx context.Context, phone, password, name string) (*UserBase, error)                                      // 登录
	VerifySmsCode(ctx context.Context, phone, code string) error                                                     // 登录验证码校验
	UpdateSmsCode(ctx context.Context, phone, code string) error                                                     // 密码修改验证码校验
	DeleteUpdateSmsCode(ctx context.Context, phone string) error                                                     // 删除密码修改验证码
	UpdateUserInfo(ctx context.Context, base *UserBase) error                                                        // 更新用户信息
	GetUserByName(ctx context.Context, name string) (*UserBase, error)                                               // 通过用户名查询用户
	FaceCertify(ctx context.Context, userID int64, realName, idCardNumber, returnURL string) (string, string, error) // 人脸识别实名认证
	SaveRealName(ctx context.Context, rn *RealName) error                                                            // 保存实名认证信息
	CertifyNotify(ctx context.Context, notifyData string) error                                                      // 实名认证回调处理
	QueryCertify(ctx context.Context, certifyID string, userID int64) (*QueryCertifyResult, error)                   // 查询认证结果
	ResetPassword(ctx context.Context, phone, smsCode, newPassword string) error                                     // 密码重置
	FreezeAccount(ctx context.Context, userID, adminID int64, reason, ipAddress string) error                        // 账号冻结
	UnfreezeAccount(ctx context.Context, userID, adminID int64, reason, ipAddress string) error                      // 账号解冻
	SaveLoginLog(ctx context.Context, log *LoginLog) error                                                           // 保存登录日志
	GetLoginLogs(ctx context.Context, userID int64, page, pageSize int32) ([]*LoginLog, int32, error)                // 获取登录日志
	SaveAccountOperation(ctx context.Context, op *AccountOperation) error                                            // 保存账号操作记录
	VerifyResetPasswordSmsCode(ctx context.Context, phone, code string) error                                        // 密码重置验证码校验
}

// UserUsecase is a user usecase.
type UserUsecase struct {
	repo          UserRepo
	log           *log.Helper
	jwtSecret     string
	tokenExpire   time.Duration
	refreshExpire time.Duration
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

// todo:用户添加
func (uc *UserUsecase) CreateUser(ctx context.Context, user *UserBase) (*UserBase, error) {
	return uc.repo.CreateUser(ctx, user)
}

// todo:根据手机号查询用户
func (uc *UserUsecase) GetUser(ctx context.Context, phone string) (*UserBase, error) {
	uc.log.WithContext(ctx).Infof("GetUser: %v", phone)
	return uc.repo.GetUser(ctx, phone)
}

// todo:根据用户信息查询用户
func (uc *UserUsecase) GetUserID(ctx context.Context, id int64) (*UserBase, error) {
	uc.log.WithContext(ctx).Infof("GetUserID: %v", id)
	return uc.repo.GetUserID(ctx, id)
}

// todo:通过用户名查询用户
func (uc *UserUsecase) GetUserByName(ctx context.Context, name string) (*UserBase, error) {
	return uc.repo.GetUserByName(ctx, name)
}

// todo:登录
func (uc *UserUsecase) Login(ctx context.Context, phone, password, name string) (*UserBase, error) {
	return uc.repo.Login(ctx, phone, password, name)
}

// todo:验证码
func (uc *UserUsecase) SendSms(ctx context.Context, phone, source string) error {
	err := uc.repo.SendSms(ctx, phone, source)
	return err
}

// todo:登录验证码校验
func (uc *UserUsecase) VerifySmsCode(ctx context.Context, phone, code string) error {
	return uc.repo.VerifySmsCode(ctx, phone, code)
}

// todo:密码修改验证码校验
func (uc *UserUsecase) UpdateSmsCode(ctx context.Context, phone, code string) error {
	return uc.repo.UpdateSmsCode(ctx, phone, code)
}

// todo:删除密码修改验证码
func (uc *UserUsecase) DeleteUpdateSmsCode(ctx context.Context, phone string) error {
	return uc.repo.DeleteUpdateSmsCode(ctx, phone)
}

// todo:更新用户信息
func (uc *UserUsecase) UpdateUserInfo(ctx context.Context, base *UserBase) error {

	_, err := uc.repo.GetUserID(ctx, base.UserId)
	if err != nil {
		return err
	}
	return uc.repo.UpdateUserInfo(ctx, base)
}

// todo:人脸识别实名认证
func (uc *UserUsecase) FaceCertify(ctx context.Context, userID int64, realName, idCardNumber, returnURL string) (string, string, error) {
	uc.log.Infof("开始人脸识别实名认证，用户ID: %d, 真实姓名: %s, 身份证号: %s", userID, realName, idCardNumber)
	return uc.repo.FaceCertify(ctx, userID, realName, idCardNumber, returnURL)
}

// todo:支付宝实名认证回调
func (uc *UserUsecase) CertifyNotify(ctx context.Context, notifyData string) error {
	uc.log.Info("收到支付宝实名认证回调")
	return uc.repo.CertifyNotify(ctx, notifyData)
}

// todo:查询实名认证结果
func (uc *UserUsecase) QueryCertify(ctx context.Context, certifyID string, userID int64) (*QueryCertifyResult, error) {
	uc.log.Infof("查询实名认证结果，认证ID: %s, 用户ID: %d", certifyID, userID)
	return uc.repo.QueryCertify(ctx, certifyID, userID)
}

// todo:保存实名认证信息
func (uc *UserUsecase) SaveRealName(ctx context.Context, rn *RealName) error {
	return uc.repo.SaveRealName(ctx, rn)
}

// todo:密码重置
func (uc *UserUsecase) ResetPassword(ctx context.Context, phone, smsCode, newPassword string) error {
	uc.log.Infof("用户密码重置，手机号: %s", phone)
	return uc.repo.ResetPassword(ctx, phone, smsCode, newPassword)
}

// todo:账号冻结
func (uc *UserUsecase) FreezeAccount(ctx context.Context, userID, adminID int64, reason, ipAddress string) error {
	uc.log.Infof("冻结账号，用户ID: %d, 管理员ID: %d, 原因: %s", userID, adminID, reason)
	return uc.repo.FreezeAccount(ctx, userID, adminID, reason, ipAddress)
}

// todo:账号解冻
func (uc *UserUsecase) UnfreezeAccount(ctx context.Context, userID, adminID int64, reason, ipAddress string) error {
	uc.log.Infof("解冻账号，用户ID: %d, 管理员ID: %d, 原因: %s", userID, adminID, reason)
	return uc.repo.UnfreezeAccount(ctx, userID, adminID, reason, ipAddress)
}

// todo:保存登录日志
func (uc *UserUsecase) SaveLoginLog(ctx context.Context, log *LoginLog) error {
	return uc.repo.SaveLoginLog(ctx, log)
}

// todo:获取登录日志
func (uc *UserUsecase) GetLoginLogs(ctx context.Context, userID int64, page, pageSize int32) ([]*LoginLog, int32, error) {
	uc.log.Infof("获取登录日志，用户ID: %d, 页码: %d, 每页数量: %d", userID, page, pageSize)
	return uc.repo.GetLoginLogs(ctx, userID, page, pageSize)
}

// todo:保存账号操作记录
func (uc *UserUsecase) SaveAccountOperation(ctx context.Context, op *AccountOperation) error {
	return uc.repo.SaveAccountOperation(ctx, op)
}

// todo:密码重置验证码校验
func (uc *UserUsecase) VerifyResetPasswordSmsCode(ctx context.Context, phone, code string) error {
	return uc.repo.VerifyResetPasswordSmsCode(ctx, phone, code)
}

// todo:生成JWT令牌和刷新令牌
func (uc *UserUsecase) GenerateTokens(userID int64) (string, string, error) {
	// 创建JWT令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(uc.tokenExpire).Unix(),
	})

	// 生成签名字符串
	tokenString, err := token.SignedString([]byte(uc.jwtSecret))
	if err != nil {
		return "", "", err
	}

	// 创建刷新令牌
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(uc.refreshExpire).Unix(),
	})

	// 生成刷新令牌字符串
	refreshTokenString, err := refreshToken.SignedString([]byte(uc.jwtSecret))
	if err != nil {
		return "", "", err
	}

	return tokenString, refreshTokenString, nil
}

func Md5(pwd string) string {
	h := md5.New()
	h.Write([]byte(pwd))
	return hex.EncodeToString(h.Sum(nil))
}
