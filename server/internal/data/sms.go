package data

import (
	"context"

	"anjuke/server/internal/domain"

	"github.com/go-kratos/kratos/v2/log"
)

// ShumaiSmsSender 数脉短信发送器
type ShumaiSmsSender struct {
	log *log.Helper
}

// NewShumaiSmsSender 创建数脉短信发送器
func NewShumaiSmsSender(logger log.Logger) *ShumaiSmsSender {
	return &ShumaiSmsSender{
		log: log.NewHelper(logger),
	}
}

// SendSms 发送短信验证码
func (s *ShumaiSmsSender) SendSms(ctx context.Context, phone, code string) error {
	// 简化实现：模拟短信发送
	s.log.Infof("发送短信: phone=%s, code=%s", phone, code)

	// 模拟发送成功
	return nil
}

// VerifySms 验证短信验证码
func (s *ShumaiSmsSender) VerifySms(ctx context.Context, phone, code string) (bool, error) {
	// 简化实现：模拟验证成功
	s.log.Infof("验证短信: phone=%s, code=%s", phone, code)

	// 模拟验证成功（在实际项目中应该从Redis或数据库验证）
	return true, nil
}

// NewSmsRepo 创建短信仓储
func NewSmsRepo(sender *ShumaiSmsSender) domain.SmsRepo {
	return sender
}

// SmsRiskControl 简化的短信发送风控校验逻辑
func (d *Data) SmsRiskControl(ctx context.Context, mobile, deviceID, ip string) error {
	// 简化实现：暂时不做风控限制
	// 在生产环境中，这里应该实现基于Redis的风控逻辑
	d.log.Infof("短信风控检查: mobile=%s, deviceID=%s, ip=%s", mobile, deviceID, ip)
	return nil
}
