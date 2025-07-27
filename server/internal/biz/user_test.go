package biz

import (
	"context"
	"errors"
	"strings"
	"testing"

	"anjuke/server/internal/domain"
	"anjuke/server/internal/mocks"
	"anjuke/server/internal/testutil"

	"github.com/go-kratos/kratos/v2/log"
)

func TestUserUsecase_SendSms(t *testing.T) {
	tests := []struct {
		name           string
		phone          string
		deviceID       string
		ip             string
		scene          string
		mockUserRepo   func() *mocks.MockUserRepo
		mockSmsRepo    func() *mocks.MockSmsRepo
		expectedResult string
		expectError    bool
	}{
		{
			name:     "成功发送短信",
			phone:    "13800138000",
			deviceID: "device123",
			ip:       "192.168.1.1",
			scene:    "login",
			mockUserRepo: func() *mocks.MockUserRepo {
				return &mocks.MockUserRepo{
					SmsRiskControlFunc: func(ctx context.Context, phone, deviceID, ip string) error {
						return nil
					},
				}
			},
			mockSmsRepo: func() *mocks.MockSmsRepo {
				return &mocks.MockSmsRepo{
					SendSmsFunc: func(ctx context.Context, phone, scene string) (string, error) {
						return "短信发送成功", nil
					},
				}
			},
			expectedResult: "短信发送成功",
			expectError:    false,
		},
		{
			name:     "无效场景",
			phone:    "13800138000",
			deviceID: "device123",
			ip:       "192.168.1.1",
			scene:    "invalid_scene",
			mockUserRepo: func() *mocks.MockUserRepo {
				return &mocks.MockUserRepo{}
			},
			mockSmsRepo: func() *mocks.MockSmsRepo {
				return &mocks.MockSmsRepo{}
			},
			expectedResult: "",
			expectError:    true,
		},
		{
			name:     "风控失败",
			phone:    "13800138000",
			deviceID: "device123",
			ip:       "192.168.1.1",
			scene:    "login",
			mockUserRepo: func() *mocks.MockUserRepo {
				return &mocks.MockUserRepo{
					SmsRiskControlFunc: func(ctx context.Context, phone, deviceID, ip string) error {
						return errors.New("风控检查失败")
					},
				}
			},
			mockSmsRepo: func() *mocks.MockSmsRepo {
				return &mocks.MockSmsRepo{}
			},
			expectedResult: "",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &UserUsecase{
				repo:      tt.mockUserRepo(),
				smsRepo:   tt.mockSmsRepo(),
				minioRepo: &mocks.MockMinioRepo{},
				log:       log.NewHelper(testutil.MockLogger()),
			}

			result, err := uc.SendSms(context.Background(), tt.phone, tt.deviceID, tt.ip, tt.scene)

			testutil.AssertError(t, err, tt.expectError, "SendSms error check")
			if !tt.expectError {
				testutil.AssertEqual(t, tt.expectedResult, result, "SendSms result")
			}
		})
	}
}

func TestUserUsecase_VerifySms(t *testing.T) {
	tests := []struct {
		name           string
		phone          string
		code           string
		scene          string
		mockSmsRepo    func() *mocks.MockSmsRepo
		expectedResult bool
		expectError    bool
	}{
		{
			name:  "验证成功",
			phone: "13800138000",
			code:  "123456",
			scene: "login",
			mockSmsRepo: func() *mocks.MockSmsRepo {
				return &mocks.MockSmsRepo{
					VerifySmsFunc: func(ctx context.Context, phone, code, scene string) (bool, error) {
						return true, nil
					},
				}
			},
			expectedResult: true,
			expectError:    false,
		},
		{
			name:  "验证码错误",
			phone: "13800138000",
			code:  "654321",
			scene: "login",
			mockSmsRepo: func() *mocks.MockSmsRepo {
				return &mocks.MockSmsRepo{
					VerifySmsFunc: func(ctx context.Context, phone, code, scene string) (bool, error) {
						return false, nil
					},
				}
			},
			expectedResult: false,
			expectError:    false,
		},
		{
			name:  "参数为空",
			phone: "",
			code:  "123456",
			scene: "login",
			mockSmsRepo: func() *mocks.MockSmsRepo {
				return &mocks.MockSmsRepo{}
			},
			expectedResult: false,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &UserUsecase{
				repo:      &mocks.MockUserRepo{},
				smsRepo:   tt.mockSmsRepo(),
				minioRepo: &mocks.MockMinioRepo{},
				log:       log.NewHelper(testutil.MockLogger()),
			}

			result, err := uc.VerifySms(context.Background(), tt.phone, tt.code, tt.scene)

			testutil.AssertError(t, err, tt.expectError, "VerifySms error check")
			if !tt.expectError {
				testutil.AssertEqual(t, tt.expectedResult, result, "VerifySms result")
			}
		})
	}
}

func TestUserUsecase_RealName(t *testing.T) {
	tests := []struct {
		name         string
		user         *domain.RealName
		mockUserRepo func() *mocks.MockUserRepo
		expectError  bool
	}{
		{
			name: "实名认证成功",
			user: &domain.RealName{
				UserId: 1,
				Name:   "张三",
				IdCard: "110101199001011234",
			},
			mockUserRepo: func() *mocks.MockUserRepo {
				return &mocks.MockUserRepo{
					RealNameFunc: func(ctx context.Context, user *domain.RealName) (*domain.RealName, error) {
						return user, nil
					},
				}
			},
			expectError: false,
		},
		{
			name: "用户ID为空",
			user: &domain.RealName{
				UserId: 0,
				Name:   "张三",
				IdCard: "110101199001011234",
			},
			mockUserRepo: func() *mocks.MockUserRepo {
				return &mocks.MockUserRepo{}
			},
			expectError: true,
		},
		{
			name: "姓名为空",
			user: &domain.RealName{
				UserId: 1,
				Name:   "",
				IdCard: "110101199001011234",
			},
			mockUserRepo: func() *mocks.MockUserRepo {
				return &mocks.MockUserRepo{}
			},
			expectError: true,
		},
		{
			name: "身份证号长度错误",
			user: &domain.RealName{
				UserId: 1,
				Name:   "张三",
				IdCard: "1101011990",
			},
			mockUserRepo: func() *mocks.MockUserRepo {
				return &mocks.MockUserRepo{}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &UserUsecase{
				repo:      tt.mockUserRepo(),
				smsRepo:   &mocks.MockSmsRepo{},
				minioRepo: &mocks.MockMinioRepo{},
				log:       log.NewHelper(testutil.MockLogger()),
			}

			result, err := uc.RealName(context.Background(), tt.user)

			testutil.AssertError(t, err, tt.expectError, "RealName error check")
			if !tt.expectError {
				testutil.AssertNotNil(t, result, "RealName result")
			}
		})
	}
}

func TestUserUsecase_CreateUser(t *testing.T) {
	tests := []struct {
		name         string
		phone        string
		userName     string
		password     string
		mockUserRepo func() *mocks.MockUserRepo
		expectError  bool
	}{
		{
			name:     "创建用户成功",
			phone:    "13800138000",
			userName: "测试用户",
			password: "123456",
			mockUserRepo: func() *mocks.MockUserRepo {
				return &mocks.MockUserRepo{
					CheckPhoneExistsFunc: func(ctx context.Context, phone string) (bool, error) {
						return false, nil
					},
					CreateUserFunc: func(ctx context.Context, user *domain.UserBase) (*domain.UserBase, error) {
						user.UserId = 1
						return user, nil
					},
				}
			},
			expectError: false,
		},
		{
			name:     "手机号已存在",
			phone:    "13800138000",
			userName: "测试用户",
			password: "123456",
			mockUserRepo: func() *mocks.MockUserRepo {
				return &mocks.MockUserRepo{
					CheckPhoneExistsFunc: func(ctx context.Context, phone string) (bool, error) {
						return true, nil
					},
				}
			},
			expectError: true,
		},
		{
			name:     "参数为空",
			phone:    "",
			userName: "测试用户",
			password: "123456",
			mockUserRepo: func() *mocks.MockUserRepo {
				return &mocks.MockUserRepo{}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &UserUsecase{
				repo:      tt.mockUserRepo(),
				smsRepo:   &mocks.MockSmsRepo{},
				minioRepo: &mocks.MockMinioRepo{},
				log:       log.NewHelper(testutil.MockLogger()),
			}

			result, err := uc.CreateUser(context.Background(), tt.phone, tt.userName, tt.password)

			testutil.AssertError(t, err, tt.expectError, "CreateUser error check")
			if !tt.expectError {
				testutil.AssertNotNil(t, result, "CreateUser result")
				testutil.AssertEqual(t, tt.phone, result.Phone, "CreateUser phone")
			}
		})
	}
}

func TestUserUsecase_Login(t *testing.T) {
	tests := []struct {
		name         string
		loginType    string
		mobile       string
		password     string
		code         string
		mockUserRepo func() *mocks.MockUserRepo
		mockSmsRepo  func() *mocks.MockSmsRepo
		expectError  bool
	}{
		{
			name:      "密码登录成功",
			loginType: "password",
			mobile:    "13800138000",
			password:  "123456",
			code:      "",
			mockUserRepo: func() *mocks.MockUserRepo {
				return &mocks.MockUserRepo{
					GetUserByPhoneAndPasswordFunc: func(ctx context.Context, phone, password string) (*domain.UserBase, error) {
						return &domain.UserBase{
							UserId: 1,
							Phone:  phone,
							Name:   "测试用户",
							Status: domain.UserStatusNormal,
						}, nil
					},
				}
			},
			mockSmsRepo: func() *mocks.MockSmsRepo {
				return &mocks.MockSmsRepo{}
			},
			expectError: false,
		},
		{
			name:      "短信登录成功",
			loginType: "sms",
			mobile:    "13800138000",
			password:  "",
			code:      "123456",
			mockUserRepo: func() *mocks.MockUserRepo {
				return &mocks.MockUserRepo{
					GetUserByPhoneFunc: func(ctx context.Context, phone string) (*domain.UserBase, error) {
						return &domain.UserBase{
							UserId: 1,
							Phone:  phone,
							Name:   "测试用户",
							Status: domain.UserStatusNormal,
						}, nil
					},
				}
			},
			mockSmsRepo: func() *mocks.MockSmsRepo {
				return &mocks.MockSmsRepo{
					VerifySmsFunc: func(ctx context.Context, phone, code, scene string) (bool, error) {
						return true, nil
					},
				}
			},
			expectError: false,
		},
		{
			name:      "不支持的登录类型",
			loginType: "invalid",
			mobile:    "13800138000",
			password:  "123456",
			code:      "",
			mockUserRepo: func() *mocks.MockUserRepo {
				return &mocks.MockUserRepo{}
			},
			mockSmsRepo: func() *mocks.MockSmsRepo {
				return &mocks.MockSmsRepo{}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &UserUsecase{
				repo:      tt.mockUserRepo(),
				smsRepo:   tt.mockSmsRepo(),
				minioRepo: &mocks.MockMinioRepo{},
				log:       log.NewHelper(testutil.MockLogger()),
			}

			user, token, err := uc.Login(context.Background(), tt.loginType, tt.mobile, tt.password, tt.code)

			testutil.AssertError(t, err, tt.expectError, "Login error check")
			if !tt.expectError {
				testutil.AssertNotNil(t, user, "Login user result")
				testutil.AssertEqual(t, tt.mobile, user.Phone, "Login user phone")
				if !strings.Contains(token, "token_") {
					t.Errorf("Login token format error: %s", token)
				}
			}
		})
	}
}

func TestUserUsecase_GetFileList(t *testing.T) {
	tests := []struct {
		name          string
		page          int32
		pageSize      int32
		keyword       string
		mockMinioRepo func() *mocks.MockMinioRepo
		expectError   bool
	}{
		{
			name:     "获取文件列表成功",
			page:     1,
			pageSize: 10,
			keyword:  "",
			mockMinioRepo: func() *mocks.MockMinioRepo {
				return &mocks.MockMinioRepo{
					ListFilesFunc: func(ctx context.Context, prefix string, maxKeys int) ([]domain.FileInfo, error) {
						return []domain.FileInfo{
							{Name: "test1.txt", Size: 1024},
							{Name: "test2.txt", Size: 2048},
						}, nil
					},
				}
			},
			expectError: false,
		},
		{
			name:     "搜索文件成功",
			page:     1,
			pageSize: 10,
			keyword:  "test",
			mockMinioRepo: func() *mocks.MockMinioRepo {
				return &mocks.MockMinioRepo{
					SearchFilesFunc: func(ctx context.Context, keyword string, maxKeys int) ([]domain.FileInfo, error) {
						return []domain.FileInfo{
							{Name: "test1.txt", Size: 1024},
						}, nil
					},
				}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &UserUsecase{
				repo:      &mocks.MockUserRepo{},
				smsRepo:   &mocks.MockSmsRepo{},
				minioRepo: tt.mockMinioRepo(),
				log:       log.NewHelper(testutil.MockLogger()),
			}

			files, total, err := uc.GetFileList(context.Background(), tt.page, tt.pageSize, tt.keyword)

			testutil.AssertError(t, err, tt.expectError, "GetFileList error check")
			if !tt.expectError {
				testutil.AssertNotNil(t, files, "GetFileList files result")
				if total <= 0 {
					t.Errorf("GetFileList total should be > 0, got %d", total)
				}
			}
		})
	}
}
