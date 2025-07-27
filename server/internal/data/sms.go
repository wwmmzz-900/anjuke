package data

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"time"

	"anjuke/server/internal/domain"

	"github.com/go-redis/redis/v8"
)

var (
	ErrSmsIntervalLimit = errors.New("请勿频繁操作，请稍后再试")
	ErrSmsMobileLimit   = errors.New("该手机号今日短信发送次数已达上限")
	ErrSmsDeviceLimit   = errors.New("该设备今日短信发送次数已达上限")
	ErrSmsIPLimit       = errors.New("该IP今日短信发送次数已达上限")
	ErrSmsCodeExpired   = errors.New("验证码已过期")
	ErrSmsCodeInvalid   = errors.New("验证码错误")
	ErrSmsCodeNotFound  = errors.New("验证码不存在")
)

// SmsSender 是一个通用的短信发送接口，可以用于适配不同的短信服务商。
type SmsSender interface {
	SendSms(ctx context.Context, phone, scene string) (string, error)
	VerifySms(ctx context.Context, phone, code, scene string) (bool, error)
}

// SmsTemplate 短信模板配置
type SmsTemplate struct {
	TemplateId string
	Content    string
}

// SmsCodeInfo 验证码信息
// 由于key中已经包含了phone和scene信息，这里可以简化存储内容
type SmsCodeInfo struct {
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ShumaiSmsSender 是数脉短信服务的具体实现，它实现了 SmsSender 接口。
type ShumaiSmsSender struct {
	ApiUrl    string
	AppCode   string
	rdb       redis.Cmdable
	templates map[string]SmsTemplate
	MockMode  bool // 模拟模式，用于开发测试
}

// NewShumaiSmsSender 是 ShumaiSmsSender 的构造函数。
func NewShumaiSmsSender(appCode string, rdb redis.Cmdable) *ShumaiSmsSender {
	// 初始化不同场景的短信模板
	templates := map[string]SmsTemplate{
		"register":       {TemplateId: "eca797c0c9ac334318e8d3900ef73ac5", Content: "注册验证码"},
		"login":          {TemplateId: "eca797c0c9ac334318e8d3900ef73ac6", Content: "登录验证码"},
		"reset_password": {TemplateId: "eca797c0c9ac334318e8d3900ef73ac7", Content: "重置密码验证码"},
		"bind_phone":     {TemplateId: "eca797c0c9ac334318e8d3900ef73ac8", Content: "绑定手机验证码"},
		"change_phone":   {TemplateId: "eca797c0c9ac334318e8d3900ef73ac9", Content: "更换手机验证码"},
		"real_name":      {TemplateId: "eca797c0c9ac334318e8d3900ef73aca", Content: "实名认证验证码"},
	}

	// 检查是否启用模拟模式（可以通过环境变量或配置文件控制）
	mockMode := appCode == "mock" || appCode == ""

	return &ShumaiSmsSender{
		ApiUrl:    "https://smssend.shumaidata.com/sms/send",
		AppCode:   appCode,
		rdb:       rdb,
		templates: templates,
		MockMode:  mockMode, // 如果AppCode为空或为"mock"，则启用模拟模式
	}
}

// SendSms 实现了具体的短信发送逻辑，通过 HTTP 调用数脉短信 API。
func (s *ShumaiSmsSender) SendSms(ctx context.Context, phone, scene string) (string, error) {
	// 1. 获取对应场景的模板
	template, exists := s.templates[scene]
	if !exists {
		return "", fmt.Errorf("不支持的短信场景: %s", scene)
	}

	// 2. 生成6位数字验证码
	code, err := s.generateCode()
	if err != nil {
		return "", fmt.Errorf("生成验证码失败: %v", err)
	}

	// 3. 将验证码信息存储到Redis，key格式简化为：sms:code:{scene}:{phone}
	// 这样每个场景+手机号组合只能有一个有效验证码，新的会覆盖旧的
	codeInfo := SmsCodeInfo{
		Code:      code,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	codeData, err := json.Marshal(codeInfo)
	if err != nil {
		return "", fmt.Errorf("序列化验证码信息失败: %v", err)
	}

	// 存储验证码，key格式：sms:code:{scene}:{phone}
	codeKey := fmt.Sprintf("sms:code:%s:%s", scene, phone)
	err = s.rdb.Set(ctx, codeKey, codeData, 5*time.Minute).Err()
	if err != nil {
		return "", fmt.Errorf("存储验证码失败: %v", err)
	}

	// 4. 模拟模式：跳过实际的短信发送
	if s.MockMode {
		log.Printf("🔧 模拟模式 - 短信发送成功，场景: %s, 手机号: %s, 验证码: %s", scene, phone, code)
		return fmt.Sprintf("%s发送成功（模拟模式）", template.Content), nil
	}

	// 5. 调用短信API发送验证码
	params := url.Values{}
	params.Set("templateId", template.TemplateId)
	params.Set("receive", phone)
	params.Set("param", code) // 将验证码作为模板参数

	urlStr := s.ApiUrl + "?" + params.Encode()
	log.Printf("发送短信请求 - URL: %s, 场景: %s, 手机号: %s, 验证码: %s", urlStr, scene, phone, code)

	// 构造 HTTP POST 请求
	req, err := http.NewRequestWithContext(ctx, "POST", urlStr, nil)
	if err != nil {
		return "", fmt.Errorf("创建HTTP请求失败: %v", err)
	}
	req.Header.Set("Authorization", "APPCODE "+s.AppCode)

	// 创建HTTP客户端，增加超时时间和连接配置
	client := &http.Client{
		Timeout: 30 * time.Second, // 增加到30秒
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			DisableKeepAlives:   false,
		},
	}

	// 重试机制：最多重试3次
	var resp *http.Response
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		log.Printf("发送短信请求，第%d次尝试", attempt)
		resp, err = client.Do(req)
		if err == nil {
			break // 成功，跳出重试循环
		}
		lastErr = err
		if attempt < 3 {
			// 等待一段时间后重试，递增延迟
			waitTime := time.Duration(attempt) * 2 * time.Second
			log.Printf("请求失败，%v后重试: %v", waitTime, err)
			time.Sleep(waitTime)
		}
	}

	if resp == nil {
		return "", fmt.Errorf("发送HTTP请求失败，已重试3次: %v", lastErr)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != 200 {
		// 发送失败时删除已存储的验证码
		s.rdb.Del(ctx, codeKey)
		return "", fmt.Errorf("短信发送失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	log.Printf("短信发送成功，场景: %s, 手机号: %s", scene, phone)
	return fmt.Sprintf("%s发送成功", template.Content), nil
}

// VerifySms 验证短信验证码
func (s *ShumaiSmsSender) VerifySms(ctx context.Context, phone, code, scene string) (bool, error) {
	// 1. 直接构造key进行查询：sms:code:{scene}:{phone}
	// 这样设计更简单，每个场景+手机号组合只有一个验证码
	codeKey := fmt.Sprintf("sms:code:%s:%s", scene, phone)
	codeData, err := s.rdb.Get(ctx, codeKey).Result()
	if err != nil {
		if err == redis.Nil {
			return false, ErrSmsCodeNotFound
		}
		return false, fmt.Errorf("获取验证码失败: %v", err)
	}

	// 2. 反序列化验证码信息
	var codeInfo SmsCodeInfo
	err = json.Unmarshal([]byte(codeData), &codeInfo)
	if err != nil {
		return false, fmt.Errorf("解析验证码信息失败: %v", err)
	}

	// 3. 检查验证码是否过期（虽然Redis会自动过期，但这里做双重检查）
	if time.Now().After(codeInfo.ExpiresAt) {
		// 删除过期的验证码
		s.rdb.Del(ctx, codeKey)
		return false, ErrSmsCodeExpired
	}

	// 4. 验证验证码是否正确
	if codeInfo.Code != code {
		return false, ErrSmsCodeInvalid
	}

	// 5. 验证成功，删除验证码（一次性使用）
	s.rdb.Del(ctx, codeKey)

	log.Printf("短信验证码验证成功，场景: %s, 手机号: %s", scene, phone)
	return true, nil
}

// generateCode 生成6位数字验证码
func (s *ShumaiSmsSender) generateCode() (string, error) {
	code := ""
	for i := 0; i < 6; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code += num.String()
	}
	return code, nil
}

// NewSmsRepo 是一个 wire provider，它将具体的短信发送者（*ShumaiSmsSender）
// 绑定到 domain.SmsRepo 接口。这样 biz 层就可以无感知地使用具体的实现。
func NewSmsRepo(sender *ShumaiSmsSender) domain.SmsRepo {
	return sender
}

// SmsRiskControl 封装了所有基于 Redis 的短信发送风控校验逻辑。
// 这个方法被 UserRepo 调用，是数据层内部的一个可复用能力。
func (d *Data) SmsRiskControl(ctx context.Context, mobile, deviceID, ip string) error {
	dateStr := time.Now().Format("20060102")

	// 1. 60秒内只能发一次
	intervalKey := fmt.Sprintf("sms:interval:%s", mobile)
	ok, err := d.rdb.SetNX(ctx, intervalKey, 1, 60*time.Second).Result()
	if err != nil {
		return err
	}
	if !ok {
		return ErrSmsIntervalLimit
	}

	// 2. 单日最多5次
	countKey := fmt.Sprintf("sms:count:%s:%s", mobile, dateStr)
	count, err := d.rdb.Incr(ctx, countKey).Result()
	if err != nil {
		return err
	}
	if count == 1 {
		d.rdb.Expire(ctx, countKey, 24*time.Hour)
	}
	if count > 5 {
		return ErrSmsMobileLimit
	}

	// 3. 设备单日最多10次
	if deviceID != "" {
		deviceKey := fmt.Sprintf("sms:device:%s:%s", deviceID, dateStr)
		deviceCount, _ := d.rdb.Incr(ctx, deviceKey).Result()
		if deviceCount == 1 {
			d.rdb.Expire(ctx, deviceKey, 24*time.Hour)
		}
		if deviceCount > 10 {
			return ErrSmsDeviceLimit
		}
	}

	// 4. IP单日最多10次
	if ip != "" {
		ipKey := fmt.Sprintf("sms:ip:%s:%s", ip, dateStr)
		ipCount, _ := d.rdb.Incr(ctx, ipKey).Result()
		if ipCount == 1 {
			d.rdb.Expire(ctx, ipKey, 24*time.Hour)
		}
		if ipCount > 10 {
			return ErrSmsIPLimit
		}
	}

	return nil
}
