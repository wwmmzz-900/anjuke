// Package service 提供了对外暴露的 gRPC 和 HTTP 服务。
// 这一层扮演了适配器（Adapter）的角色，负责将外部请求（gRPC/HTTP）
// 转换为内部的业务调用（biz.Usecase），并将业务逻辑的执行结果封装成外部响应。
// 它不包含任何业务逻辑，只做参数校验、格式转换和调用委托。
package service

import (
	commonv1 "anjuke/server/api/common/v1"
	v2 "anjuke/server/api/user/v2"
	"anjuke/server/internal/biz"
	"anjuke/server/internal/domain"
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/go-kratos/kratos/v2/transport"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/protobuf/types/known/anypb"
)

// UserService 实现了 API 定义的用户服务。
// 它依赖 biz.UserUsecase 来完成实际的业务处理。
type UserService struct {
	v2.UnimplementedUserServer
	uc *biz.UserUsecase
}

// NewUserService 是 UserService 的构造函数。
func NewUserService(uc *biz.UserUsecase) *UserService {
	return &UserService{
		uc: uc,
	}
}

// RealName 是处理实名认证请求的 gRPC/HTTP 入口。
func (s *UserService) RealName(ctx context.Context, req *v2.RealNameRequest) (*commonv1.BaseResponse, error) {

	// 将请求参数转换为 biz 层的领域模型
	_, err := s.uc.RealName(ctx, &domain.RealName{
		UserId: uint32(req.UserId),
		Name:   req.Name,
		IdCard: req.IdCard,
	})
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 1,
			Msg:  err.Error(),
			Data: nil,
		}, nil
	}

	// 构建响应数据
	data := &v2.RealNameData{
		UserId: req.UserId,
		Name:   req.Name,
		Status: "verified",
	}

	anyData, err := anypb.New(data)
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 1,
			Msg:  "数据序列化失败",
			Data: nil,
		}, nil
	}

	return &commonv1.BaseResponse{
		Code: 0,
		Msg:  "实名认证成功",
		Data: anyData,
	}, nil
}

// UpdateUserStatus 是处理更新用户状态请求的 gRPC/HTTP 入口。
func (s *UserService) UpdateUserStatus(ctx context.Context, req *v2.UpdateUserStatusRequest) (*commonv1.BaseResponse, error) {

	// 将请求参数转换为 biz 层的领域模型
	_, err := s.uc.UpdateUserStatus(ctx, &domain.UserBase{
		UserId:     req.UserId,
		RealStatus: domain.RealNameUnverified,
	})
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 1,
			Msg:  err.Error(),
			Data: nil,
		}, nil
	}

	// 构建响应数据
	data := &v2.UpdateUserStatusData{
		UserId: req.UserId,
		Status: "updated",
	}

	anyData, err := anypb.New(data)
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 1,
			Msg:  "数据序列化失败",
			Data: nil,
		}, nil
	}

	return &commonv1.BaseResponse{
		Code: 0,
		Msg:  "用户状态更新成功",
		Data: anyData,
	}, nil
}

// SendSms 是处理发送短信请求的 gRPC/HTTP 入口。
func (s *UserService) SendSms(ctx context.Context, req *v2.SendSmsRequest) (*commonv1.BaseResponse, error) {
	log.Printf("SendSms参数: Phone=%s, DeviceID=%s, Scene=%s", req.Phone, req.DeviceId, req.Scene)

	// 参数验证
	if req.Phone == "" {
		return &commonv1.BaseResponse{
			Code: 1,
			Msg:  "缺少参数: phone",
			Data: nil,
		}, nil
	}
	if req.Scene == "" {
		return &commonv1.BaseResponse{
			Code: 1,
			Msg:  "缺少参数: scene",
			Data: nil,
		}, nil
	}

	// 从 context 中获取 ip 等元数据
	var ip string
	if tr, ok := transport.FromServerContext(ctx); ok {
		ip = tr.RequestHeader().Get("X-Real-IP")
	}

	// 调用业务层方法，并传递所有必要参数
	result, err := s.uc.SendSms(ctx, req.Phone, req.DeviceId, ip, req.Scene)
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 1,
			Msg:  err.Error(),
			Data: nil,
		}, nil
	}

	// 构建响应数据
	data := &v2.SendSmsData{
		Phone:      req.Phone,
		Scene:      req.Scene,
		ExpireTime: time.Now().Add(5 * time.Minute).Unix(), // 5分钟后过期
	}

	anyData, err := anypb.New(data)
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 1,
			Msg:  "数据序列化失败",
			Data: nil,
		}, nil
	}

	return &commonv1.BaseResponse{
		Code: 0,
		Msg:  result,
		Data: anyData,
	}, nil
}

// VerifySms 是处理验证短信验证码请求的 gRPC/HTTP 入口。
func (s *UserService) VerifySms(ctx context.Context, req *v2.VerifySmsRequest) (*commonv1.BaseResponse, error) {
	log.Printf("VerifySms参数: Phone=%s, Code=%s, Scene=%s", req.Phone, req.Code, req.Scene)

	// 参数验证
	if req.Phone == "" {
		return &commonv1.BaseResponse{
			Code: 1,
			Msg:  "缺少参数: phone",
			Data: nil,
		}, nil
	}
	if req.Code == "" {
		return &commonv1.BaseResponse{
			Code: 1,
			Msg:  "缺少参数: code",
			Data: nil,
		}, nil
	}
	if req.Scene == "" {
		return &commonv1.BaseResponse{
			Code: 1,
			Msg:  "缺少参数: scene",
			Data: nil,
		}, nil
	}

	// 调用业务层方法进行验证
	success, err := s.uc.VerifySms(ctx, req.Phone, req.Code, req.Scene)
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 1,
			Msg:  err.Error(),
			Data: nil,
		}, nil
	}

	// 构建响应数据
	data := &v2.VerifySmsData{
		Phone:   req.Phone,
		Success: success,
		Scene:   req.Scene,
	}

	anyData, err := anypb.New(data)
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 1,
			Msg:  "数据序列化失败",
			Data: nil,
		}, nil
	}

	msg := "验证失败"
	code := int32(1)
	if success {
		msg = "验证成功"
		code = 0
	}

	return &commonv1.BaseResponse{
		Code: code,
		Msg:  msg,
		Data: anyData,
	}, nil
}

// CreateUser 是处理用户注册请求的 gRPC/HTTP 入口。
func (s *UserService) CreateUser(ctx context.Context, req *v2.CreateUserRequest) (*commonv1.BaseResponse, error) {
	// 参数验证
	if req.Mobile == "" {
		return BuildErrorResponse(1, "手机号不能为空"), nil
	}
	if req.NickName == "" {
		return BuildErrorResponse(1, "昵称不能为空"), nil
	}
	if req.Password == "" {
		return BuildErrorResponse(1, "密码不能为空"), nil
	}
	if req.Code == "" {
		return BuildErrorResponse(1, "验证码不能为空"), nil
	}

	// 验证手机号格式
	if len(req.Mobile) != 11 {
		return BuildErrorResponse(1, "手机号格式不正确"), nil
	}

	// 验证密码长度
	if len(req.Password) < 6 || len(req.Password) > 20 {
		return BuildErrorResponse(1, "密码长度应为6-20个字符"), nil
	}

	// 验证短信验证码
	verified, err := s.uc.VerifySms(ctx, req.Mobile, req.Code, "register")
	if err != nil {
		return BuildErrorResponse(1, fmt.Sprintf("验证码验证失败: %v", err)), nil
	}
	if !verified {
		return BuildErrorResponse(1, "验证码错误或已过期"), nil
	}

	// 调用业务层创建用户
	user, err := s.uc.CreateUser(ctx, req.Mobile, req.NickName, req.Password)
	if err != nil {
		return BuildErrorResponse(1, fmt.Sprintf("创建用户失败: %v", err)), nil
	}

	// 构建响应数据
	data := &v2.CreateUserData{
		UserId:   fmt.Sprintf("%d", user.UserId),
		Mobile:   user.Phone,
		NickName: user.Name,
	}

	// 构建成功响应
	resp, err := BuildSuccessResponse("用户创建成功", data)
	if err != nil {
		return BuildErrorResponse(1, "响应构建失败"), nil
	}
	return resp, nil
}

// UploadFile 是为了满足 gRPC 接口生成的桩函数，我们不直接使用它。
// 实际的文件上传逻辑由我们手动注册的 HTTP Handler 处理。
func (s *UserService) UploadFile(ctx context.Context, body *httpbody.HttpBody) (*httpbody.HttpBody, error) {
	return nil, fmt.Errorf("not implemented")
}

// UploadToMinioWithProgress 支持进度回调的文件上传
func (s *UserService) UploadToMinioWithProgress(ctx context.Context, fileName string, reader io.Reader, size int64, contentType string, progressCallback func(uploaded, total int64)) (string, error) {
	url, err := s.uc.UploadToMinioWithProgress(ctx, fileName, reader, size, contentType, progressCallback)
	if err != nil {
		return "", fmt.Errorf("上传失败: %v", err)
	}
	return url, nil
}

// DeleteFromMinio 删除 MinIO 文件
func (s *UserService) DeleteFromMinio(ctx context.Context, objectName string) error {
	return s.uc.DeleteFromMinio(ctx, objectName)
}

// Login 是处理用户登录请求的 gRPC/HTTP 入口。
func (s *UserService) Login(ctx context.Context, req *v2.LoginRequest) (*commonv1.BaseResponse, error) {
	log.Printf("Login参数: LoginType=%s, Mobile=%s", req.LoginType, req.Mobile)

	// 参数验证
	if req.Mobile == "" {
		return BuildErrorResponse(1, "手机号不能为空"), nil
	}
	if req.LoginType == "" {
		return BuildErrorResponse(1, "登录类型不能为空"), nil
	}

	// 根据登录类型进行不同的参数验证
	switch req.LoginType {
	case "password":
		if req.Password == "" {
			return BuildErrorResponse(1, "密码不能为空"), nil
		}
	case "sms":
		if req.Code == "" {
			return BuildErrorResponse(1, "验证码不能为空"), nil
		}
	default:
		return BuildErrorResponse(1, "不支持的登录类型"), nil
	}

	// 调用业务层进行登录
	user, token, err := s.uc.Login(ctx, req.LoginType, req.Mobile, req.Password, req.Code)
	if err != nil {
		return BuildErrorResponse(1, err.Error()), nil
	}

	// 构建响应数据
	data := &v2.LoginData{
		UserId:     fmt.Sprintf("%d", user.UserId),
		Mobile:     user.Phone,
		NickName:   user.Name,
		Token:      token,
		ExpireTime: time.Now().Add(24 * time.Hour).Unix(), // 24小时后过期
	}

	// 构建成功响应
	resp, err := BuildSuccessResponse("登录成功", data)
	if err != nil {
		return BuildErrorResponse(1, "响应构建失败"), nil
	}
	return resp, nil
}

// GetFileList 是处理获取文件列表请求的 gRPC/HTTP 入口。
func (s *UserService) GetFileList(ctx context.Context, req *v2.GetFileListRequest) (*commonv1.BaseResponse, error) {
	log.Printf("GetFileList参数: Page=%d, PageSize=%d, Keyword=%s", req.Page, req.PageSize, req.Keyword)

	// 调用业务层获取文件列表
	fileInfos, total, err := s.uc.GetFileList(ctx, req.Page, req.PageSize, req.Keyword)
	if err != nil {
		log.Printf("获取文件列表失败: %v", err)
		return BuildErrorResponse(1, "获取文件列表失败"), nil
	}

	// 转换为proto格式
	files := make([]*v2.FileInfo, 0, len(fileInfos))
	for _, fileInfo := range fileInfos {
		files = append(files, &v2.FileInfo{
			Id:         fileInfo.ETag, // 使用ETag作为ID
			Name:       fileInfo.Name,
			Size:       fileInfo.Size,
			Type:       fileInfo.ContentType,
			Url:        fileInfo.URL,
			UploadTime: fileInfo.LastModified.Format("2006-01-02 15:04:05"),
			Status:     "success",
			ObjectName: fileInfo.Name,
		})
	}

	// 构建响应数据
	data := &v2.GetFileListData{
		List:  files,
		Total: total,
	}

	// 构建成功响应
	resp, err := BuildSuccessResponse("获���文件列表成功", data)
	if err != nil {
		return BuildErrorResponse(1, "响应构建失败"), nil
	}
	return resp, nil
}

// GetUploadStats 是处理获取上传统计请求的 gRPC/HTTP 入口。
func (s *UserService) GetUploadStats(ctx context.Context, req *v2.GetUploadStatsRequest) (*commonv1.BaseResponse, error) {
	log.Printf("GetUploadStats请求")

	// 调用业务层获取上传统计
	stats, err := s.uc.GetUploadStats(ctx)
	if err != nil {
		log.Printf("获取上传统计失败: %v", err)
		return BuildErrorResponse(1, "获取上传统计失败"), nil
	}

	// 构建响应数据
	data := &v2.GetUploadStatsData{
		TotalUploads:   int32(stats["totalUploads"].(int32)),
		SuccessUploads: int32(stats["successUploads"].(int32)),
		TotalSize:      stats["totalSize"].(int64),
		TodayUploads:   int32(stats["todayUploads"].(int32)),
	}

	// 构建成功响应
	resp, err := BuildSuccessResponse("获取上传统计成功", data)
	if err != nil {
		return BuildErrorResponse(1, "响应构建失败"), nil
	}
	return resp, nil
}

// DeleteFile 是处理删除文件请求的 gRPC/HTTP 入口。
func (s *UserService) DeleteFile(ctx context.Context, req *v2.DeleteFileRequest) (*commonv1.BaseResponse, error) {
	log.Printf("DeleteFile参数: ObjectName=%s", req.ObjectName)

	// 调用业务层删除文件
	err := s.uc.DeleteFile(ctx, req.ObjectName)
	if err != nil {
		log.Printf("删除文件失败: %v", err)
		return BuildErrorResponse(1, "删除文件失败: "+err.Error()), nil
	}

	// 构建响应数据
	data := &v2.DeleteFileData{
		ObjectName: req.ObjectName,
		Success:    true,
	}

	// 构建成功响应
	resp, err := BuildSuccessResponse("删除文件成功", data)
	if err != nil {
		return BuildErrorResponse(1, "响应构建失败"), nil
	}
	return resp, nil
}

// TestMinioConnection 测试MinIO连接
func (s *UserService) TestMinioConnection(ctx context.Context, req *v2.GetUploadStatsRequest) (*commonv1.BaseResponse, error) {
	log.Printf("TestMinioConnection请求")

	// 尝试列出文件来测试连接
	files, total, err := s.uc.GetFileList(ctx, 1, 1, "")
	if err != nil {
		log.Printf("MinIO连接测试失败: %v", err)
		return BuildErrorResponse(1, "MinIO连接失败: "+err.Error()), nil
	}

	// 构建响应数据
	data := &v2.GetUploadStatsData{
		TotalUploads:   total,
		SuccessUploads: int32(len(files)),
		TotalSize:      0,
		TodayUploads:   0,
	}

	// 构建成功响应
	resp, err := BuildSuccessResponse("MinIO连接测试成功", data)
	if err != nil {
		return BuildErrorResponse(1, "响应构建失败"), nil
	}
	return resp, nil
}
