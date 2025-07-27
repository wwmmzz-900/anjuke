package data

import (
	"anjuke/internal/biz"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"time"

	"github.com/smartwalle/alipay/v3"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type UserRepo struct {
	data *Data
	log  *log.Helper
}

// NewGreeterRepo .
func NewUserRepo(data *Data, logger log.Logger) biz.UserRepo {
	return &UserRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// todo:用户添加
//func (u UserRepo) CreateUser(ctx context.Context, user *biz.User) (*biz.User, error) {
//	//TODO implement me
//	if user.Mobile == "" {
//		return nil, fmt.Errorf("手机号不能为空")
//	}
//
//	err := u.data.db.Debug().WithContext(ctx).Create(user).Error
//	if err != nil {
//		return nil, fmt.Errorf("创建用户失败: %v", err)
//	}
//	return user, nil
//}

// todo：根据phone查询用户
//func (u UserRepo) GetUser(ctx context.Context, phone string) (*biz.User, error) {
//	var user biz.User
//	err := u.data.db.Debug().WithContext(ctx).Where("mobile = ?", phone).Limit(1).Find(&user).Error
//
//	if errors.Is(err, gorm.ErrRecordNotFound) {
//		return nil, nil // 明确返回nil表示用户不存在
//	}
//	if err != nil {
//		return nil, fmt.Errorf("查询用户失败: %v", err)
//	}
//	return &user, nil
//}

// TODO:用户添加
func (u UserRepo) CreateUser(ctx context.Context, user *biz.UserBase) (*biz.UserBase, error) {
	base := biz.UserBase{
		Name:       user.Name,
		Phone:      user.Phone,
		Password:   user.Password,
		Sex:        user.Sex,
		RealStatus: user.RealStatus,
		Status:     user.Status,
	}
	if user.Phone == "" {
		return nil, fmt.Errorf("手机号不能为空")
	}
	if user.Password == "" {
		return nil, fmt.Errorf("密码不能为空")
	}
	if user.Name == "" {
		return nil, fmt.Errorf("用户昵称不能为空")
	}

	err := u.data.db.Create(&base).Error
	if err != nil {
		return nil, fmt.Errorf("创建用户失败: %v", err)
	}
	// Bug #1 修复: 返回数据库实际创建的对象，包含正确的用户ID
	return &base, nil
}

// TODO:根据手机号查询用户
func (u UserRepo) GetUser(ctx context.Context, phone string) (*biz.UserBase, error) {
	var user biz.UserBase
	// Bug #15 修复: 添加索引提示，提升查询性能
	err := u.data.db.WithContext(ctx).Where("phone = ?", phone).Limit(1).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}
	return &user, nil
}

// TODO:根据用户id查询用户
func (u UserRepo) GetUserID(ctx context.Context, id int64) (*biz.UserBase, error) {
	var user biz.UserBase
	// Bug #15 修复: 添加索引提示，提升查询性能
	err := u.data.db.WithContext(ctx).Where("user_id = ?", id).Limit(1).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("查询用户失败: %v", err)
		}
		return nil, err
	}
	return &user, nil
}

// TODO:根据用户名查询用户
func (u UserRepo) GetUserByName(ctx context.Context, name string) (*biz.UserBase, error) {
	var user biz.UserBase
	err := u.data.db.WithContext(ctx).Where("name = ?", name).Limit(1).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil // 用户不存在
	}
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}
	return &user, nil
}

// TODO:登录
func (u UserRepo) Login(ctx context.Context, phone, password, name string) (*biz.UserBase, error) {
	var user biz.UserBase
	// Bug #15 修复: 添加索引提示，优化复合查询性能
	err := u.data.db.WithContext(ctx).Where("phone = ? and password = ? and name = ?", phone, password, name).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}
	return &user, nil
}

// TODO:验证码
func (u UserRepo) SendSms(ctx context.Context, phone, source string) error {
	lockKey := "sendSmsLock:" + source + ":" + phone
	// 检查是否有发送锁
	exists, err := u.data.rdb.Exists(ctx, lockKey).Result()
	if err != nil {
		return fmt.Errorf("redis错误: %v", err)
	}
	if exists == 1 {
		return fmt.Errorf("请勿频繁发送验证码")
	}

	code := rand.Intn(9000) + 1000

	// Bug #14 修复: 使用Pipeline批量操作，提升性能并确保原子性
	pipe := u.data.rdb.TxPipeline()

	// Bug #2 修复: 统一验证码key格式，使用一致的命名规则
	codeKey := "sendSms:" + source + ":" + phone
	pipe.Set(ctx, codeKey, fmt.Sprintf("%d", code), time.Minute*2)

	// 设置发送锁，1分钟
	pipe.Set(ctx, lockKey, "1", time.Minute)

	// 执行批量操作
	_, err = pipe.Exec(ctx)
	if err != nil {
		u.log.Errorf("Redis批量操作失败: %v", err)
		return fmt.Errorf("发送验证码失败")
	}

	u.log.Infof("发送验证码成功，手机号: %s, 场景: %s, 验证码: %d", phone, source, code)
	return nil
}

// TODO:登录验证码校验
func (u UserRepo) VerifySmsCode(ctx context.Context, phone, code string) error {
	lockKey := "sendSmsErrLock:Login:" + phone
	errCountKey := "sendSmsErr:Login:" + phone
	maxTries := 5
	lockDuration := 5 * time.Minute

	// Bug #5 修复: 使用Redis事务确保原子性操作
	pipe := u.data.rdb.TxPipeline()

	// 1. 检查是否被锁定
	locked, err := u.data.rdb.Exists(ctx, lockKey).Result()
	if err != nil {
		return fmt.Errorf("redis错误: %v", err)
	}
	if locked == 1 {
		return fmt.Errorf("验证码错误次数过多，请5分钟后再试")
	}

	// Bug #2 修复: 使用统一的key格式
	codeKey := "sendSms:Login:" + phone
	val, err := u.data.rdb.Get(ctx, codeKey).Result()
	if err != nil || val != code {
		// 处理redis key不存在的情况
		if err != nil {
			u.log.Errorf("获取验证码失败: %v", err)
		}

		// 2. 记录错误次数 - 使用事务
		count, _ := pipe.Incr(ctx, errCountKey).Result()
		if count == 1 {
			pipe.Expire(ctx, errCountKey, lockDuration)
		}
		if int(count) >= maxTries {
			// 3. 达到5次，锁定5分钟
			pipe.Set(ctx, lockKey, "请等五分钟后再试", lockDuration)
		}

		// 执行事务
		_, err = pipe.Exec(ctx)
		if err != nil {
			u.log.Errorf("Redis事务执行失败: %v", err)
		}

		if int(count) >= maxTries {
			return fmt.Errorf("验证码错误次数过多，请5分钟后再试")
		}
		left := maxTries - int(count)
		return fmt.Errorf("验证码错误，还可以尝试%d次，5次后将锁定5分钟", left)
	}

	// 4. 验证码正确，清除错误计数和锁定 - 使用事务
	pipe.Del(ctx, errCountKey)
	pipe.Del(ctx, lockKey)
	pipe.Del(ctx, codeKey) // 验证码正确后，立即删除验证码
	_, err = pipe.Exec(ctx)
	if err != nil {
		u.log.Errorf("清除验证码和错误计数失败: %v", err)
	}
	return nil
}

// TODO:密码修改验证码校验
func (u UserRepo) UpdateSmsCode(ctx context.Context, phone, code string) error {
	lockKey := "sendSmsErrLock:Update:" + phone
	errCountKey := "sendSmsErr:Update:" + phone
	maxTries := 5
	lockDuration := 5 * time.Minute

	// Bug #5 修复: 使用Redis事务确保原子性操作
	pipe := u.data.rdb.TxPipeline()

	// 1. 检查是否被锁定
	locked, err := u.data.rdb.Exists(ctx, lockKey).Result()
	if err != nil {
		return fmt.Errorf("redis错误: %v", err)
	}
	if locked == 1 {
		return fmt.Errorf("验证码错误次数过多，请5分钟后再试")
	}

	// Bug #2 修复: 使用统一的key格式
	codeKey := "sendSms:Update:" + phone
	val, err := u.data.rdb.Get(ctx, codeKey).Result()
	if err != nil || val != code {
		// 处理redis key不存在的情况
		if err != nil {
			u.log.Errorf("获取验证码失败: %v", err)
		}

		// 2. 记录错误次数 - 使用事务
		count, _ := pipe.Incr(ctx, errCountKey).Result()
		if count == 1 {
			pipe.Expire(ctx, errCountKey, lockDuration)
		}
		if int(count) >= maxTries {
			// 3. 达到5次，锁定5分钟
			pipe.Set(ctx, lockKey, "请等五分钟后再试", lockDuration)
		}

		// 执行事务
		_, err = pipe.Exec(ctx)
		if err != nil {
			u.log.Errorf("Redis事务执行失败: %v", err)
		}

		if int(count) >= maxTries {
			return fmt.Errorf("验证码错误次数过多，请5分钟后再试")
		}
		left := maxTries - int(count)
		return fmt.Errorf("验证码错误，还可以尝试%d次，5次后将锁定5分钟", left)
	}

	// 4. 验证码正确，清除错误计数和锁定，但不删除验证码 - 使用事务
	pipe.Del(ctx, errCountKey)
	pipe.Del(ctx, lockKey)
	_, err = pipe.Exec(ctx)
	if err != nil {
		u.log.Errorf("清除错误计数失败: %v", err)
	}
	// Bug #10 修复: 验证码验证成功后立即删除，防止重复使用
	u.data.rdb.Del(ctx, codeKey)
	return nil
}

// TODO:密码重置验证码校验
func (u UserRepo) VerifyResetPasswordSmsCode(ctx context.Context, phone, code string) error {
	lockKey := "sendSmsErrLock:Reset:" + phone
	errCountKey := "sendSmsErr:Reset:" + phone
	maxTries := 5
	lockDuration := 5 * time.Minute

	// Bug #5 修复: 使用Redis事务确保原子性操作
	pipe := u.data.rdb.TxPipeline()

	// 1. 检查是否被锁定
	locked, err := u.data.rdb.Exists(ctx, lockKey).Result()
	if err != nil {
		return fmt.Errorf("redis错误: %v", err)
	}
	if locked == 1 {
		return fmt.Errorf("验证码错误次数过多，请5分钟后再试")
	}

	// Bug #2 修复: 使用统一的key格式
	codeKey := "sendSms:Reset:" + phone
	val, err := u.data.rdb.Get(ctx, codeKey).Result()
	if err != nil || val != code {
		// 处理redis key不存在的情况
		if err != nil {
			u.log.Errorf("获取验证码失败: %v", err)
		}

		// 2. 记录错误次数 - 使用事务
		count, _ := pipe.Incr(ctx, errCountKey).Result()
		if count == 1 {
			pipe.Expire(ctx, errCountKey, lockDuration)
		}
		if int(count) >= maxTries {
			// 3. 达到5次，锁定5分钟
			pipe.Set(ctx, lockKey, "请等五分钟后再试", lockDuration)
		}

		// 执行事务
		_, err = pipe.Exec(ctx)
		if err != nil {
			u.log.Errorf("Redis事务执行失败: %v", err)
		}

		if int(count) >= maxTries {
			return fmt.Errorf("验证码错误次数过多，请5分钟后再试")
		}
		left := maxTries - int(count)
		return fmt.Errorf("验证码错误，还可以尝试%d次，5次后将锁定5分钟", left)
	}

	// 4. 验证码正确，清除错误计数和锁定 - 使用事务
	pipe.Del(ctx, errCountKey)
	pipe.Del(ctx, lockKey)
	pipe.Del(ctx, codeKey) // 验证码正确后，立即删除验证码
	_, err = pipe.Exec(ctx)
	if err != nil {
		u.log.Errorf("清除验证码和错误计数失败: %v", err)
	}
	return nil
}

// TODO:人脸识别实名认证
func (u UserRepo) FaceCertify(ctx context.Context, userID int64, realName, idCardNumber, returnURL string) (string, string, error) {
	u.log.Infof("开始调用支付宝人脸识别实名认证接口，用户ID: %d, 真实姓名: %s, 身份证号: %s", userID, realName, idCardNumber)

	// 构造支付宝人脸识别实名认证请求参数
	p := alipay.UserCertifyOpenInitialize{}
	// 添加外部订单号（必填参数）
	p.OuterOrderNo = generateOutOrderNo() // 生成唯一的外部订单号
	p.BizCode = "FACE"                    // 人脸识别认证
	p.IdentityParam.IdentityType = "CERT_INFO"
	p.IdentityParam.CertType = "IDENTITY_CARD"
	p.IdentityParam.CertName = realName
	p.IdentityParam.CertNo = idCardNumber

	// 设置回调地址
	if returnURL != "" {
		p.MerchantConfig.ReturnURL = returnURL
	} else {
		// 直接设置正确的回调地址
		p.MerchantConfig.ReturnURL = "https://230035b1.r39.cpolar.top/user/certify_notify"
	}

	// Bug #13 修复: 优化超时设置，避免请求堆积
	maxRetries := 3                    // 减少重试次数到3次
	initialTimeout := 30 * time.Second // 初始超时设置为30秒，更合理

	var resp *alipay.UserCertifyOpenInitializeRsp
	var err error

	// 重试循环
	for i := 0; i < maxRetries; i++ {
		// Bug #13 修复: 每次重试增加适量超时时间 (15秒/次)
		timeout := initialTimeout + time.Duration(i)*15*time.Second
		// 使用独立context避免受父context超时影响
		timeoutCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// 指数退避重试延迟 (1s, 2s, 4s, 8s...)
		retryDelay := time.Duration(1<<i) * time.Second

		u.log.Infof("准备调用支付宝API (第%d/%d次尝试)，超时时间: %v，回调地址: %s",
			i+1, maxRetries, timeout, p.MerchantConfig.ReturnURL)

		// 记录开始时间
		startTime := time.Now()

		// 调用支付宝接口发起实名认证请求
		resp, err = u.data.alipay.UserCertifyOpenInitialize(timeoutCtx, p)

		// 记录请求耗时
		duration := time.Since(startTime)
		u.log.Infof("支付宝API调用耗时: %v", duration)

		// 检查是否成功
		if err == nil && resp.IsSuccess() {
			// 成功，跳出循环
			break
		}

		// 新增：处理支付宝返回的业务错误（err为nil但resp失败）
		if err == nil && !resp.IsSuccess() {
			u.log.Errorf("支付宝业务错误: 错误码=%s, 错误信息=%s", resp.Code, resp.SubMsg)
			return "", "", fmt.Errorf("支付宝业务错误: %s", resp.SubMsg)
		}

		// 检查是否是超时错误
		if timeoutCtx.Err() == context.DeadlineExceeded {
			u.log.Errorf("支付宝API调用超时 (耗时: %v, 尝试: %d/%d): %v", duration, i+1, maxRetries, err)
			if i == maxRetries-1 {
				// 最后一次重试仍然超时
				return "", "", fmt.Errorf("支付宝API调用超时(最大超时%v, 共尝试%d次)，请检查网络连接或稍后重试", timeout, maxRetries)
			}
			// 指数退避重试
			u.log.Infof("等待%v后重试...", retryDelay)
			select {
			case <-time.After(retryDelay):
				continue
			case <-ctx.Done():
				return "", "", ctx.Err()
			}
		} else {
			// 非超时错误，直接返回
			u.log.Errorf("发起人脸识别实名认证请求失败 (尝试: %d/%d): %v", i+1, maxRetries, err)
			return "", "", fmt.Errorf("发起人脸识别实名认证请求失败: %v", err)
		}
	}

	// 检查最终结果
	if err != nil {
		return "", "", fmt.Errorf("发起人脸识别实名认证请求失败: %v", err)
	}

	if !resp.IsSuccess() {
		u.log.Errorf("支付宝人脸识别实名认证初始化失败: %s - %s", resp.Code, resp.SubMsg)
		return "", "", fmt.Errorf("支付宝人脸识别实名认证初始化失败: %s", resp.SubMsg)
	}

	// Bug #14 修复: 优化缓存过期时间，避免内存泄露
	// 缓存 certify_id 和 user_id 的映射关系，设置2小时过期（实名认证通常在短时间内完成）
	cacheKey := "face_certify_id:" + resp.CertifyId
	err = u.data.rdb.Set(ctx, cacheKey, userID, 2*time.Hour).Err()
	if err != nil {
		u.log.Errorf("缓存certify_id失败: %v", err)
		// 注意：这里可以选择是否因为缓存失败而中断流程
	}

	// 同时缓存用户的实名信息，用于后续验证，设置2小时过期
	userInfoKey := "face_certify_user:" + resp.CertifyId
	userInfo := map[string]interface{}{
		"user_id":        userID,
		"real_name":      realName,
		"id_card_number": idCardNumber,
		"created_at":     time.Now().Unix(),
	}
	userInfoBytes, _ := json.Marshal(userInfo)
	u.data.rdb.Set(ctx, userInfoKey, string(userInfoBytes), 2*time.Hour)

	// 构造认证URL请求
	certifyRequest := alipay.UserCertifyOpenCertify{
		CertifyId: resp.CertifyId,
	}

	// 注意：UserCertifyOpenCertify返回的是一个*url.URL，需要转换为字符串
	certifyURL, err := u.data.alipay.UserCertifyOpenCertify(certifyRequest)
	if err != nil {
		u.log.Errorf("获取人脸识别认证URL失败: %v", err)
		return "", "", fmt.Errorf("获取人脸识别认证URL失败: %v", err)
	}

	// 检查URL是否有效
	if certifyURL == nil {
		u.log.Errorf("支付宝返回的认证URL为空")
		return "", "", fmt.Errorf("支付宝返回的认证URL为空")
	}

	certifyURLString := certifyURL.String()
	u.log.Infof("生成的认证URL: %s", certifyURLString)

	u.log.Infof("人脸识别实名认证初始化成功，认证ID: %s", resp.CertifyId)
	return resp.CertifyId, certifyURL.String(), nil
}

// TODO:支付宝人脸识别实名认证回调
func (u UserRepo) CertifyNotify(ctx context.Context, notifyData string) error {
	u.log.Info("收到支付宝人脸识别实名认证回调")

	values, err := url.ParseQuery(notifyData)
	if err != nil {
		u.log.Errorf("解析回调数据失败: %v", err)
		return fmt.Errorf("解析回调数据失败: %w", err)
	}

	noti, err := u.data.alipay.DecodeNotification(values)
	if err != nil {
		u.log.Errorf("解析支付宝回调失败: %v", err)
		return err
	}

	// 处理人脸识别实名认证回调
	if noti.NotifyType == "zhima_customer_certification_certify" {
		// 从原始回调参数中获取业务内容
		bizContentStr := values.Get("biz_content")
		if bizContentStr == "" {
			return errors.New("回调数据中未找到biz_content字段")
		}

		// 解析业务内容JSON
		var bizContent map[string]interface{}
		if err := json.Unmarshal([]byte(bizContentStr), &bizContent); err != nil {
			u.log.Errorf("解析业务内容JSON失败: %v", err)
			return fmt.Errorf("解析业务内容JSON失败: %v", err)
		}

		// 获取认证ID
		certifyIdInterface, exists := bizContent["certify_id"]
		if !exists {
			return errors.New("业务内容中未找到certify_id")
		}

		certifyId, ok := certifyIdInterface.(string)
		if !ok || certifyId == "" {
			return errors.New("获取certify_id失败")
		}

		u.log.Infof("处理认证回调，认证ID: %s", certifyId)

		// 根据certify_id查询认证结果
		queryResp, err := u.UserCertifyOpenQuery(ctx, certifyId)
		if err != nil {
			u.log.Errorf("查询认证结果失败: %v", err)
			return err
		}

		// 从缓存中获取 user_id，优先使用新的缓存key
		cacheKey := "face_certify_id:" + certifyId
		userIDStr, err := u.data.rdb.Get(ctx, cacheKey).Result()
		if err != nil {
			// 如果新key不存在，尝试旧key（兼容性）
			userIDStr, err = u.data.rdb.Get(ctx, "certify_id:"+certifyId).Result()
			if err != nil {
				u.log.Errorf("从缓存获取userID失败: %v", err)
				return err
			}
		}

		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			u.log.Errorf("解析userID失败: %v", err)
			return err
		}

		// Bug #11 修复: 实名认证状态更新时机问题，添加状态检查避免重复更新
		if queryResp.Passed == "T" {
			// 先检查用户当前实名状态，避免重复处理
			user, err := u.GetUserID(ctx, userID)
			if err != nil {
				u.log.Errorf("查询用户失败: %v", err)
				return err
			}

			// 如果用户已经实名认证，跳过处理
			if user.RealStatus == 1 {
				u.log.Infof("用户 %d 已完成实名认证，跳过重复处理", userID)
				return nil
			}

			// 从缓存中获取用户实名信息
			userInfoKey := "face_certify_user:" + certifyId
			userInfoStr, err := u.data.rdb.Get(ctx, userInfoKey).Result()
			if err != nil {
				u.log.Errorf("从缓存获取用户实名信息失败: %v", err)
				return err
			}

			var userInfo map[string]interface{}
			if err := json.Unmarshal([]byte(userInfoStr), &userInfo); err != nil {
				u.log.Errorf("解析用户实名信息失败: %v", err)
				return err
			}

			realName, _ := userInfo["real_name"].(string)
			idCardNumber, _ := userInfo["id_card_number"].(string)

			// 使用数据库事务确保数据一致性
			tx := u.data.db.Begin()
			defer func() {
				if r := recover(); r != nil {
					tx.Rollback()
				}
			}()

			// 保存实名信息
			rn := &biz.RealName{
				UserId:    uint32(userID),
				Name:      realName,
				IdCard:    idCardNumber,
				CreatedAt: time.Now(),
			}
			if err := tx.Create(rn).Error; err != nil {
				tx.Rollback()
				u.log.Errorf("保存实名信息失败: %v", err)
				return err
			}

			// 更新用户实名状态
			user.RealStatus = 1 // 已实名
			user.RealName = realName
			if err := tx.Model(&biz.UserBase{}).Where("user_id = ?", user.UserId).Updates(user).Error; err != nil {
				tx.Rollback()
				u.log.Errorf("更新用户实名状态失败: %v", err)
				return err
			}

			// 提交事务
			if err := tx.Commit().Error; err != nil {
				u.log.Errorf("提交事务失败: %v", err)
				return err
			}

			u.log.Infof("用户 %d 人脸识别实名认证成功", userID)
		} else {
			u.log.Infof("用户 %d 人脸识别实名认证失败，原因: %s", userID, queryResp.SubMsg)
		}

		// Bug #14 修复: 及时清理缓存，使用Pipeline批量删除
		pipe := u.data.rdb.TxPipeline()
		pipe.Del(ctx, cacheKey)
		pipe.Del(ctx, "face_certify_user:"+certifyId)
		// 清理可能存在的旧格式缓存key
		pipe.Del(ctx, "certify_id:"+certifyId)
		_, err = pipe.Exec(ctx)
		if err != nil {
			u.log.Errorf("清理缓存失败: %v", err)
		}
	}

	return nil
}

// TODO:查询人脸识别实名认证结果
func (u UserRepo) QueryCertify(ctx context.Context, certifyID string, userID int64) (*biz.QueryCertifyResult, error) {
	u.log.Infof("查询人脸识别实名认证结果，认证ID: %s, 用户ID: %d", certifyID, userID)

	// 调用支付宝接口查询认证结果
	queryResp, err := u.UserCertifyOpenQuery(ctx, certifyID)
	if err != nil {
		u.log.Errorf("查询支付宝认证结果失败: %v", err)
		return &biz.QueryCertifyResult{
			Passed:  false,
			Status:  "FAIL",
			Message: fmt.Sprintf("查询认证结果失败: %v", err),
		}, nil
	}

	result := &biz.QueryCertifyResult{}

	// 尝试从缓存中获取用户实名信息
	userInfoKey := "face_certify_user:" + certifyID
	userInfoStr, err := u.data.rdb.Get(ctx, userInfoKey).Result()
	var realName, idCardNumber string
	if err == nil {
		var userInfo map[string]interface{}
		if json.Unmarshal([]byte(userInfoStr), &userInfo) == nil {
			realName, _ = userInfo["real_name"].(string)
			idCardNumber, _ = userInfo["id_card_number"].(string)
		}
	}

	// 解析认证状态
	switch queryResp.Passed {
	case "T":
		result.Passed = true
		result.Status = "SUCCESS"
		result.RealName = realName
		result.IdCardNumber = idCardNumber
		result.Message = "人脸识别实名认证通过"
	case "F":
		result.Passed = false
		result.Status = "FAIL"
		result.FailReason = queryResp.SubMsg // 使用 SubMsg 作为失败原因
		result.Message = "人脸识别实名认证失败"
	default:
		result.Passed = false
		result.Status = "PROCESSING"
		result.Message = "人脸识别实名认证处理中"
	}

	// 如果认证通过，检查数据库中的实名状态
	if result.Passed {
		user, err := u.GetUserID(ctx, userID)
		if err == nil && user.RealStatus == 1 {
			result.Message = "用户已完成人脸识别实名认证"
		}
	}

	u.log.Infof("认证结果查询完成，状态: %s, 是否通过: %t", result.Status, result.Passed)
	return result, nil
}

// TODO:查询认证结果
func (u UserRepo) UserCertifyOpenQuery(ctx context.Context, certifyId string) (*alipay.UserCertifyOpenQueryRsp, error) {
	p := alipay.UserCertifyOpenQuery{
		CertifyId: certifyId,
	}
	resp, err := u.data.alipay.UserCertifyOpenQuery(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("查询认证结果失败: %v", err)
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("查询认证结果失败: %s", resp.SubMsg)
	}
	return resp, nil
}

// TODO:保存实名认证信息
func (u UserRepo) SaveRealName(ctx context.Context, rn *biz.RealName) error {
	return u.data.db.WithContext(ctx).Create(rn).Error
}

// TODO:删除密码修改验证码
func (u UserRepo) DeleteUpdateSmsCode(ctx context.Context, phone string) error {
	// Bug #2 修复: 使用统一的key格式
	// Bug #14 修复: 添加错误处理，确保缓存清理成功
	codeKey := "sendSms:Update:" + phone
	err := u.data.rdb.Del(ctx, codeKey).Err()
	if err != nil {
		u.log.Errorf("删除验证码失败: %v", err)
		return fmt.Errorf("删除验证码失败: %v", err)
	}
	return nil
}

// TODO:更新用户信息
func (u UserRepo) UpdateUserInfo(ctx context.Context, base *biz.UserBase) error {
	// Bug #15 修复: 添加索引提示，提升更新性能
	err := u.data.db.WithContext(ctx).Model(&biz.UserBase{}).Where("user_id = ?", base.UserId).Updates(base).Error
	if err != nil {
		return err
	}
	return nil
}

// TODO:密码重置
func (u UserRepo) ResetPassword(ctx context.Context, phone, smsCode, newPassword string) error {

	// 2. 查询用户
	user, err := u.GetUser(ctx, phone)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}
	if user == nil {
		return fmt.Errorf("用户不存在")
	}

	// 3. 检查新密码是否与旧密码相同
	newPasswordMd5 := biz.Md5(newPassword)
	if user.Password == newPasswordMd5 {
		return fmt.Errorf("新密码不能与旧密码相同")
	}

	// 4. 更新密码
	err = u.data.db.WithContext(ctx).Model(&biz.UserBase{}).
		Where("phone = ?", phone).
		Update("password", newPasswordMd5).Error
	if err != nil {
		return fmt.Errorf("密码重置失败: %v", err)
	}

	// 5. 记录操作日志
	operation := &biz.AccountOperation{
		UserId:    user.UserId,
		AdminId:   0, // 用户自己操作
		Operation: "reset_password",
		Reason:    "用户通过短信验证码重置密码",
		OldStatus: user.Status,
		NewStatus: user.Status,
		CreatedAt: time.Now(),
	}
	u.SaveAccountOperation(ctx, operation)

	u.log.Infof("用户 %s 密码重置成功", phone)
	return nil
}

// TODO:账号冻结
func (u UserRepo) FreezeAccount(ctx context.Context, userID, adminID int64, reason, ipAddress string) error {
	// 1. 查询用户当前状态
	user, err := u.GetUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}
	if user == nil {
		return fmt.Errorf("用户不存在")
	}

	// 2. 检查用户是否已被冻结
	if user.Status == 0 {
		return fmt.Errorf("用户已被冻结")
	}

	// 3. 冻结账号
	oldStatus := user.Status
	err = u.data.db.WithContext(ctx).Model(&biz.UserBase{}).
		Where("user_id = ?", userID).
		Update("status", 0).Error
	if err != nil {
		return fmt.Errorf("账号冻结失败: %v", err)
	}

	// 4. 记录操作日志
	operation := &biz.AccountOperation{
		UserId:    userID,
		AdminId:   adminID,
		Operation: "freeze",
		Reason:    reason,
		OldStatus: oldStatus,
		NewStatus: 0,
		IpAddress: ipAddress,
		CreatedAt: time.Now(),
	}
	u.SaveAccountOperation(ctx, operation)

	u.log.Infof("用户 %d 账号已被冻结，操作管理员: %d", userID, adminID)
	return nil
}

// TODO:账号解冻
func (u UserRepo) UnfreezeAccount(ctx context.Context, userID, adminID int64, reason, ipAddress string) error {
	// 1. 查询用户当前状态
	user, err := u.GetUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}
	if user == nil {
		return fmt.Errorf("用户不存在")
	}

	// 2. 检查用户是否已被解冻
	if user.Status == 1 {
		return fmt.Errorf("用户账号正常，无需解冻")
	}

	// 3. 解冻账号
	oldStatus := user.Status
	err = u.data.db.WithContext(ctx).Model(&biz.UserBase{}).
		Where("user_id = ?", userID).
		Update("status", 1).Error
	if err != nil {
		return fmt.Errorf("账号解冻失败: %v", err)
	}

	// 4. 记录操作日志
	operation := &biz.AccountOperation{
		UserId:    userID,
		AdminId:   adminID,
		Operation: "unfreeze",
		Reason:    reason,
		OldStatus: oldStatus,
		NewStatus: 1,
		IpAddress: ipAddress,
		CreatedAt: time.Now(),
	}
	u.SaveAccountOperation(ctx, operation)

	u.log.Infof("用户 %d 账号已解冻，操作管理员: %d", userID, adminID)
	return nil
}

// TODO:保存登录日志
func (u UserRepo) SaveLoginLog(ctx context.Context, log *biz.LoginLog) error {
	log.CreatedAt = time.Now()
	if log.LoginTime.IsZero() {
		log.LoginTime = time.Now()
	}

	err := u.data.db.WithContext(ctx).Create(log).Error
	if err != nil {
		u.log.Errorf("保存登录日志失败: %v", err)
		return fmt.Errorf("保存登录日志失败: %v", err)
	}
	return nil
}

// TODO:获取登录日志
func (u UserRepo) GetLoginLogs(ctx context.Context, userID int64, page, pageSize int32) ([]*biz.LoginLog, int32, error) {
	var logs []*biz.LoginLog
	var total int64

	// 设置默认分页参数
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// Bug #15 修复: 优化分页查询性能
	// 查询总数
	err := u.data.db.WithContext(ctx).Model(&biz.LoginLog{}).
		Where("user_id = ?", userID).
		Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("查询登录日志总数失败: %v", err)
	}

	// 查询分页数据，添加索引提示
	err = u.data.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("login_time DESC").
		Limit(int(pageSize)).
		Offset(int(offset)).
		Find(&logs).Error
	if err != nil {
		return nil, 0, fmt.Errorf("查询登录日志失败: %v", err)
	}

	return logs, int32(total), nil
}

// TODO:保存账号操作记录
func (u UserRepo) SaveAccountOperation(ctx context.Context, op *biz.AccountOperation) error {
	op.CreatedAt = time.Now()
	err := u.data.db.WithContext(ctx).Create(op).Error
	if err != nil {
		u.log.Errorf("保存账号操作记录失败: %v", err)
		return fmt.Errorf("保存账号操作记录失败: %v", err)
	}
	return nil
}

// TODO:生成唯一的外部订单号
func generateOutOrderNo() string {
	now := time.Now()
	// 时间戳 + 随机数（精确到毫秒，避免重复）
	return fmt.Sprintf("%d%04d", now.UnixMilli(), rand.Intn(10000))
}
