package service

import (
	commonv1 "anjuke/server/api/common/v1"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// BuildSuccessResponse 构建成功响应
func BuildSuccessResponse(msg string, data proto.Message) (*commonv1.BaseResponse, error) {
	response := &commonv1.BaseResponse{
		Code: 0,
		Msg:  msg,
	}

	if data != nil {
		anyData, err := anypb.New(data)
		if err != nil {
			return nil, err
		}
		response.Data = anyData
	}

	return response, nil
}

// BuildErrorResponse 构建错误响应
func BuildErrorResponse(code int32, msg string) *commonv1.BaseResponse {
	return &commonv1.BaseResponse{
		Code: code,
		Msg:  msg,
		Data: nil,
	}
}
