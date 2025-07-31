package service

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	v2 "anjuke/server/api/user/v2"
	"anjuke/server/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserUsecase 用户用例的模拟实现
type MockUserUsecase struct {
	mock.Mock
}

func (m *MockUserUsecase) SendSms(ctx context.Context, phone, deviceID, ip, scene string) (string, error) {
	args := m.Called(ctx, phone, deviceID, ip, scene)
	return args.String(0), args.Error(1)
}

func (m *MockUserUsecase) VerifySms(ctx context.Context, phone, code, scene string) (bool, error) {
	args := m.Called(ctx, phone, code, scene)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserUsecase) RealName(ctx context.Context, user *domain.RealName) (*domain.RealName, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RealName), args.Error(1)
}

func (m *MockUserUsecase) UpdateUserStatus(ctx context.Context, user *domain.UserBase) (*domain.UserBase, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserBase), args.Error(1)
}

func (m *MockUserUsecase) CreateUser(ctx context.Context, phone, name, password string) (*domain.UserBase, error) {
	args := m.Called(ctx, phone, name, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserBase), args.Error(1)
}

func (m *MockUserUsecase) Login(ctx context.Context, loginType, mobile, password, code string) (*domain.UserBase, string, error) {
	args := m.Called(ctx, loginType, mobile, password, code)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*domain.UserBase), args.String(1), args.Error(2)
}

func (m *MockUserUsecase) GetFileList(ctx context.Context, page, pageSize int32, keyword string) ([]domain.FileInfo, int32, error) {
	args := m.Called(ctx, page, pageSize, keyword)
	return args.Get(0).([]domain.FileInfo), args.Get(1).(int32), args.Error(2)
}

func (m *MockUserUsecase) GetUploadStats(ctx context.Context) (map[string]interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockUserUsecase) DeleteFile(ctx context.Context, objectName string) error {
	args := m.Called(ctx, objectName)
	return args.Error(0)
}

func (m *MockUserUsecase) UploadToMinioWithProgress(ctx context.Context, fileName string, reader io.Reader, size int64, contentType string, progressCallback func(uploaded, total int64)) (string, error) {
	args := m.Called(ctx, fileName, reader, size, contentType, progressCallback)
	return args.String(0), args.Error(1)
}

func (m *MockUserUsecase) DeleteFromMinio(ctx context.Context, objectName string) error {
	args := m.Called(ctx, objectName)
	return args.Error(0)
}

func TestUserService_SendSms(t *testing.T) {
	tests := []struct {
		name       string
		req        *v2.SendSmsRequest
		setupMocks func(*MockUserUsecase)
		wantCode   int32
		wantMsg    string
	}{
		{
			name: "发送短信成功",
			req: &v2.SendSmsRequest{
				Phone:    "13812345678",
				DeviceId: "device123",
				Scene:    "register",
			},
			setupMocks: func(uc *MockUserUsecase) {
				uc.On("SendSms", mock.Anything, "13812345678", "device123", "", "register").Return("短信发送成功", nil)
			},
			wantCode: 0,
			wantMsg:  "短信发送成功",
		},
		{
			name: "手机号为空",
			req: &v2.SendSmsRequest{
				Phone:    "",
				DeviceId: "device123",
				Scene:    "register",
			},
			setupMocks: func(uc *MockUserUsecase) {},
			wantCode:   1,
			wantMsg:    "缺少参数: phone",
		},
		{
			name: "场景为空",
			req: &v2.SendSmsRequest{
				Phone:    "13812345678",
				DeviceId: "device123",
				Scene:    "",
			},
			setupMocks: func(uc *MockUserUsecase) {},
			wantCode:   1,
			wantMsg:    "缺少参数: scene",
		},
		{
			name: "发送失败",
			req: &v2.SendSmsRequest{
				Phone:    "13812345678",
				DeviceId: "device123",
				Scene:    "register",
			},
			setupMocks: func(uc *MockUserUsecase) {
				uc.On("SendSms", mock.Anything, "13812345678", "device123", "", "register").Return("", errors.New("发送频率过高"))
			},
			wantCode: 1,
			wantMsg:  "发送频率过高",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟对象
			uc := new(MockUserUsecase)

			// 设置模拟行为
			tt.setupMocks(uc)

			// 创建服务
			service := NewUserServiceWithInterface(uc)

			// 执行测试
			resp, err := service.SendSms(context.Background(), tt.req)

			// 验证结果
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.wantCode, resp.Code)
			assert.Contains(t, resp.Msg, tt.wantMsg)

			// 验证模拟对象的调用
			uc.AssertExpectations(t)
		})
	}
}

func TestUserService_VerifySms(t *testing.T) {
	tests := []struct {
		name       string
		req        *v2.VerifySmsRequest
		setupMocks func(*MockUserUsecase)
		wantCode   int32
		wantMsg    string
	}{
		{
			name: "验证成功",
			req: &v2.VerifySmsRequest{
				Phone: "13812345678",
				Code:  "123456",
				Scene: "register",
			},
			setupMocks: func(uc *MockUserUsecase) {
				uc.On("VerifySms", mock.Anything, "13812345678", "123456", "register").Return(true, nil)
			},
			wantCode: 0,
			wantMsg:  "验证成功",
		},
		{
			name: "验证失败",
			req: &v2.VerifySmsRequest{
				Phone: "13812345678",
				Code:  "123456",
				Scene: "register",
			},
			setupMocks: func(uc *MockUserUsecase) {
				uc.On("VerifySms", mock.Anything, "13812345678", "123456", "register").Return(false, nil)
			},
			wantCode: 1,
			wantMsg:  "验证失败",
		},
		{
			name: "手机号为空",
			req: &v2.VerifySmsRequest{
				Phone: "",
				Code:  "123456",
				Scene: "register",
			},
			setupMocks: func(uc *MockUserUsecase) {},
			wantCode:   1,
			wantMsg:    "缺少参数: phone",
		},
		{
			name: "验证码为空",
			req: &v2.VerifySmsRequest{
				Phone: "13812345678",
				Code:  "",
				Scene: "register",
			},
			setupMocks: func(uc *MockUserUsecase) {},
			wantCode:   1,
			wantMsg:    "缺少参数: code",
		},
		{
			name: "场景为空",
			req: &v2.VerifySmsRequest{
				Phone: "13812345678",
				Code:  "123456",
				Scene: "",
			},
			setupMocks: func(uc *MockUserUsecase) {},
			wantCode:   1,
			wantMsg:    "缺少参数: scene",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟对象
			uc := new(MockUserUsecase)

			// 设置模拟行为
			tt.setupMocks(uc)

			// 创建服务
			service := NewUserServiceWithInterface(uc)

			// 执行测试
			resp, err := service.VerifySms(context.Background(), tt.req)

			// 验证结果
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.wantCode, resp.Code)
			assert.Contains(t, resp.Msg, tt.wantMsg)

			// 验证模拟对象的调用
			uc.AssertExpectations(t)
		})
	}
}

func TestUserService_CreateUser(t *testing.T) {
	tests := []struct {
		name       string
		req        *v2.CreateUserRequest
		setupMocks func(*MockUserUsecase)
		wantCode   int32
		wantMsg    string
	}{
		{
			name: "创建用户成功",
			req: &v2.CreateUserRequest{
				Mobile:   "13812345678",
				NickName: "张三",
				Password: "123456",
				Code:     "123456",
			},
			setupMocks: func(uc *MockUserUsecase) {
				uc.On("VerifySms", mock.Anything, "13812345678", "123456", "register").Return(true, nil)
				uc.On("CreateUser", mock.Anything, "13812345678", "张三", "123456").Return(&domain.UserBase{
					UserId: 123,
					Phone:  "13812345678",
					Name:   "张三",
				}, nil)
			},
			wantCode: 0,
			wantMsg:  "用户创建成功",
		},
		{
			name: "手机号为空",
			req: &v2.CreateUserRequest{
				Mobile:   "",
				NickName: "张三",
				Password: "123456",
				Code:     "123456",
			},
			setupMocks: func(uc *MockUserUsecase) {},
			wantCode:   1,
			wantMsg:    "手机号不能为空",
		},
		{
			name: "昵称为空",
			req: &v2.CreateUserRequest{
				Mobile:   "13812345678",
				NickName: "",
				Password: "123456",
				Code:     "123456",
			},
			setupMocks: func(uc *MockUserUsecase) {},
			wantCode:   1,
			wantMsg:    "昵称不能为空",
		},
		{
			name: "密码为空",
			req: &v2.CreateUserRequest{
				Mobile:   "13812345678",
				NickName: "张三",
				Password: "",
				Code:     "123456",
			},
			setupMocks: func(uc *MockUserUsecase) {},
			wantCode:   1,
			wantMsg:    "密码不能为空",
		},
		{
			name: "验证码为空",
			req: &v2.CreateUserRequest{
				Mobile:   "13812345678",
				NickName: "张三",
				Password: "123456",
				Code:     "",
			},
			setupMocks: func(uc *MockUserUsecase) {},
			wantCode:   1,
			wantMsg:    "验证码不能为空",
		},
		{
			name: "手机号格式不正确",
			req: &v2.CreateUserRequest{
				Mobile:   "138123456",
				NickName: "张三",
				Password: "123456",
				Code:     "123456",
			},
			setupMocks: func(uc *MockUserUsecase) {},
			wantCode:   1,
			wantMsg:    "手机号格式不正确",
		},
		{
			name: "密码长度不符合要求",
			req: &v2.CreateUserRequest{
				Mobile:   "13812345678",
				NickName: "张三",
				Password: "123",
				Code:     "123456",
			},
			setupMocks: func(uc *MockUserUsecase) {},
			wantCode:   1,
			wantMsg:    "密码长度应为6-20个字符",
		},
		{
			name: "验证码错误",
			req: &v2.CreateUserRequest{
				Mobile:   "13812345678",
				NickName: "张三",
				Password: "123456",
				Code:     "123456",
			},
			setupMocks: func(uc *MockUserUsecase) {
				uc.On("VerifySms", mock.Anything, "13812345678", "123456", "register").Return(false, nil)
			},
			wantCode: 1,
			wantMsg:  "验证码错误或已过期",
		},
		{
			name: "创建用户失败",
			req: &v2.CreateUserRequest{
				Mobile:   "13812345678",
				NickName: "张三",
				Password: "123456",
				Code:     "123456",
			},
			setupMocks: func(uc *MockUserUsecase) {
				uc.On("VerifySms", mock.Anything, "13812345678", "123456", "register").Return(true, nil)
				uc.On("CreateUser", mock.Anything, "13812345678", "张三", "123456").Return(nil, errors.New("该手机号已注册"))
			},
			wantCode: 1,
			wantMsg:  "创建用户失败: 该手机号已注册",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟对象
			uc := new(MockUserUsecase)

			// 设置模拟行为
			tt.setupMocks(uc)

			// 创建服务
			service := NewUserServiceWithInterface(uc)

			// 执行测试
			resp, err := service.CreateUser(context.Background(), tt.req)

			// 验证结果
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.wantCode, resp.Code)
			assert.Contains(t, resp.Msg, tt.wantMsg)

			// 验证模拟对象的调用
			uc.AssertExpectations(t)
		})
	}
}

func TestUserService_Login(t *testing.T) {
	tests := []struct {
		name       string
		req        *v2.LoginRequest
		setupMocks func(*MockUserUsecase)
		wantCode   int32
		wantMsg    string
	}{
		{
			name: "密码登录成功",
			req: &v2.LoginRequest{
				LoginType: "password",
				Mobile:    "13812345678",
				Password:  "123456",
			},
			setupMocks: func(uc *MockUserUsecase) {
				uc.On("Login", mock.Anything, "password", "13812345678", "123456", "").Return(&domain.UserBase{
					UserId: 123,
					Phone:  "13812345678",
					Name:   "张三",
				}, "token_123", nil)
			},
			wantCode: 0,
			wantMsg:  "登录成功",
		},
		{
			name: "短信登录成功",
			req: &v2.LoginRequest{
				LoginType: "sms",
				Mobile:    "13812345678",
				Code:      "123456",
			},
			setupMocks: func(uc *MockUserUsecase) {
				uc.On("Login", mock.Anything, "sms", "13812345678", "", "123456").Return(&domain.UserBase{
					UserId: 123,
					Phone:  "13812345678",
					Name:   "张三",
				}, "token_123", nil)
			},
			wantCode: 0,
			wantMsg:  "登录成功",
		},
		{
			name: "手机号为空",
			req: &v2.LoginRequest{
				LoginType: "password",
				Mobile:    "",
				Password:  "123456",
			},
			setupMocks: func(uc *MockUserUsecase) {},
			wantCode:   1,
			wantMsg:    "手机号不能为空",
		},
		{
			name: "登录类型为空",
			req: &v2.LoginRequest{
				LoginType: "",
				Mobile:    "13812345678",
				Password:  "123456",
			},
			setupMocks: func(uc *MockUserUsecase) {},
			wantCode:   1,
			wantMsg:    "登录类型不能为空",
		},
		{
			name: "密码登录时密码为空",
			req: &v2.LoginRequest{
				LoginType: "password",
				Mobile:    "13812345678",
				Password:  "",
			},
			setupMocks: func(uc *MockUserUsecase) {},
			wantCode:   1,
			wantMsg:    "密码不能为空",
		},
		{
			name: "短信登录时验证码为空",
			req: &v2.LoginRequest{
				LoginType: "sms",
				Mobile:    "13812345678",
				Code:      "",
			},
			setupMocks: func(uc *MockUserUsecase) {},
			wantCode:   1,
			wantMsg:    "验证码不能为空",
		},
		{
			name: "不支持的登录类型",
			req: &v2.LoginRequest{
				LoginType: "invalid",
				Mobile:    "13812345678",
			},
			setupMocks: func(uc *MockUserUsecase) {},
			wantCode:   1,
			wantMsg:    "不支持的登录类型",
		},
		{
			name: "登录失败",
			req: &v2.LoginRequest{
				LoginType: "password",
				Mobile:    "13812345678",
				Password:  "123456",
			},
			setupMocks: func(uc *MockUserUsecase) {
				uc.On("Login", mock.Anything, "password", "13812345678", "123456", "").Return(nil, "", errors.New("手机号或密码错误"))
			},
			wantCode: 1,
			wantMsg:  "手机号或密码错误",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟对象
			uc := new(MockUserUsecase)

			// 设置模拟行为
			tt.setupMocks(uc)

			// 创建服务
			service := NewUserServiceWithInterface(uc)

			// 执行测试
			resp, err := service.Login(context.Background(), tt.req)

			// 验证结果
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.wantCode, resp.Code)
			assert.Contains(t, resp.Msg, tt.wantMsg)

			// 验证模拟对象的调用
			uc.AssertExpectations(t)
		})
	}
}

func TestUserService_DeleteFile(t *testing.T) {
	tests := []struct {
		name       string
		req        *v2.DeleteFileRequest
		setupMocks func(*MockUserUsecase)
		wantCode   int32
		wantMsg    string
	}{
		{
			name: "删除文件成功",
			req: &v2.DeleteFileRequest{
				ObjectName: "test.jpg",
			},
			setupMocks: func(uc *MockUserUsecase) {
				uc.On("DeleteFile", mock.Anything, "test.jpg").Return(nil)
			},
			wantCode: 0,
			wantMsg:  "删除文件成功",
		},
		{
			name: "删除文件失败",
			req: &v2.DeleteFileRequest{
				ObjectName: "test.jpg",
			},
			setupMocks: func(uc *MockUserUsecase) {
				uc.On("DeleteFile", mock.Anything, "test.jpg").Return(errors.New("文件不存在"))
			},
			wantCode: 1,
			wantMsg:  "删除文件失败: 文件不存在",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟对象
			uc := new(MockUserUsecase)

			// 设置模拟行为
			tt.setupMocks(uc)

			// 创建服务
			service := NewUserServiceWithInterface(uc)

			// 执行测试
			resp, err := service.DeleteFile(context.Background(), tt.req)

			// 验证结果
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.wantCode, resp.Code)
			assert.Contains(t, resp.Msg, tt.wantMsg)

			// 验证模拟对象的调用
			uc.AssertExpectations(t)
		})
	}
}

func TestUserService_UploadToMinioWithProgress(t *testing.T) {
	tests := []struct {
		name        string
		fileName    string
		size        int64
		contentType string
		setupMocks  func(*MockUserUsecase)
		wantErr     bool
		expectedErr string
	}{
		{
			name:        "上传文件成功",
			fileName:    "test.jpg",
			size:        1024,
			contentType: "image/jpeg",
			setupMocks: func(uc *MockUserUsecase) {
				uc.On("UploadToMinioWithProgress", mock.Anything, "test.jpg", mock.Anything, int64(1024), "image/jpeg", mock.Anything).Return("http://example.com/test.jpg", nil)
			},
			wantErr: false,
		},
		{
			name:        "上传文件失败",
			fileName:    "test.jpg",
			size:        1024,
			contentType: "image/jpeg",
			setupMocks: func(uc *MockUserUsecase) {
				uc.On("UploadToMinioWithProgress", mock.Anything, "test.jpg", mock.Anything, int64(1024), "image/jpeg", mock.Anything).Return("", errors.New("上传失败"))
			},
			wantErr:     true,
			expectedErr: "上传失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟对象
			uc := new(MockUserUsecase)

			// 设置模拟行为
			tt.setupMocks(uc)

			// 创建服务
			service := NewUserServiceWithInterface(uc)

			// 创建测试数据
			reader := strings.NewReader("test file content")

			// 执行测试
			url, err := service.UploadToMinioWithProgress(context.Background(), tt.fileName, reader, tt.size, tt.contentType, nil)

			// 验证结果
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, url)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, url)
			}

			// 验证模拟对象的调用
			uc.AssertExpectations(t)
		})
	}
}

// 响应构建函数的测试已在response_test.go中实现
