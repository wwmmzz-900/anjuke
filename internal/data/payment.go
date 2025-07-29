package data

import (
	"anjuke/internal/biz"
	"context"
	"fmt"
	"github.com/smartwalle/alipay/v3"
	"time"
)

type alipayRepo struct{}

func NewAlipayRepo() biz.AlipayBizRepo {
	return &alipayRepo{}
}

const AppId = "9021000142690952"
const PrivateKey = "MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQCHOdxmStOy2g30p4AvhSZ+Ox1vI3ESDpLPtAcp+xDWdOGfa3yh1gFukIr3wrVzxacrTUR5HYyr0g3GiVUlff9dCnh24JpfqMQ8GEz8niu72rIIIBmFN8m4Ejd2qJeHeWou5Uf3TVcH6ygtzVfeWPybssa0lJfw4TPWb+d2rPP5xtt2rFktLXdsE7vI4OFOEM3tRSfuLfG242OXrAK+FxCEPQ9dT9FJklYvCLJ8RKKqLstMuBWOUzzc7otplmVWFokElH71LVEcz26mXZX1ExEvPMw47BFYZ6Y38VLJeYMslS1bqaTAXgUBD8PoKsDCxycYM11eMHidaqUxA48suYNvAgMBAAECggEALeMNjkyv/84M1EuOrRFy2Xz35QHS2bzGRuMhVzaSJSPueCmCVmyHedxku+R/rHSS4JfMt4i2dovGDuwFT76szAbEkBpxaCqdxIK+hS6rSojQxv8VieY/dk4AMizNlrQ1uwok3J+K++3paXl36sSpm7ATy61szdmtvIOmuNfBxq0c2yu+E+2nzJPsIPtX18v2nG/X4Kp+4ZCBorcPssm54RicfKmfmGA4zGxmvhUczuhHiEGIvJFbwMjP6rOwjGArfWYrlOZ1mSTXi9/W7Ji1LDXWrXEDgwmQ/RUBvBgZP1XUr2R6/PJaiFWGwCYcYkip+WJ1n4fu06+0M9EOgJ88EQKBgQDDHPbNQcJhgzI2Dw7taZJs1/K5+NrE6CvFgYzkuHp/FKY3dW/zm5rlJSUzKYCwHC/meoDaZGJ1aXN+9o73VTHsXlW+SsgWR0mZ9GlnIB7QHSuL17lJbHLnQ1tU4+PqNAIstT5yocToy22p4rpmnyosF9St0Ow+c8+z0z+oTp3B6QKBgQCxbK82ZPUwnP12tGeFarqxnJ7lQXHkdU3w//RKWox0w2W9rgpW9nSUOKUyPO4dWn46FEo2nZ3ZEsGSAlZ8iqEKxYOFeUDQVgS5O39+v/YzxqUpXYehn3xKkZDQVXnorotKK5Rj74Ba+QaGfAqe22np0IMqPU10hInZjg1dAworlwKBgAhbWjrKYT/59ZGZLYN/rRTaXvwWK5CZfR51gQpe2GhPAxuG/SeK96Ru5dv+IBPq8SZHAvPXrtvmi1rZxp/TV1MPa06+Nzm1DfL5I/aVypwRU8cmkzoQ2g8LtIK7TAzA84LktGsGgL+TzvuiyWcR1CWVU7eqJiQ6o5/JIYXc8CbZAoGAETv7cQ8xef1l6YfwnlcVt3b9QEuxIn36ijRyqF5PUnBAi8JCItxhypwN/+lHP/awWDfsVY3N7W4S+3naqNJWflNdSTPUBei1IMEUy10eLz1WgcQiDqMNUbj+Fh6XbvC1ewjsqyBymWOjLKET7wZlLV8hvpKh2XWeZlGUHrrS3BUCgYAOBx/sUXCCT0FAPy/bU//eEgnuZ8Q1ABTX77fafokqHfCBEn6tNHtGFRPowp8B0hEw7rq1dTmLkmwML2TrIzAcit0qqzGZg5/DVW83faL1G69Swxl0Xc8Vc8328+61W4mKWuiEOC9EBr3sHr0A2VEpGmdyQAQCHNqbQB5o+a39eQ=="
const NotifyURL = "http://3ee55b16.r9.cpolar.top"
const ReturnURL = "https://www.baidu.com"

// 定义 alipayRepo 结构体的 AlipayPay 方法，接收上下文、订单号和金额作为参数，返回支付跳转链接和错误信息
func (r *alipayRepo) AlipayPay(ctx context.Context, orderId string, amount string) (string, error) {
	// 从常量 PrivateKey 复制私钥到局部变量 privateKey
	var privateKey = PrivateKey // 必须，上一步中使用 RSA签名验签工具 生成的私钥
	// 使用 AppId、私钥和是否为沙箱环境（false 表示正式环境）创建支付宝客户端实例
	var client, err = alipay.New(AppId, privateKey, false)
	// 检查创建客户端实例是否出错
	if err != nil {
		// 若出错，打印错误信息
		fmt.Println(err)
		// 返回空字符串和错误信息
		return "", err
	}
	// 创建一个 TradeWapPay 类型的对象 p，用于封装手机网页支付请求参数
	var p = alipay.TradeWapPay{}
	// 设置支付结果异步通知地址
	p.NotifyURL = NotifyURL
	// 设置支付结果同步跳转地址
	p.ReturnURL = ReturnURL
	// 设置商品描述，这里为空字符串
	p.Subject = ""
	// 设置商户订单号
	p.OutTradeNo = orderId
	// 设置订单总金额
	p.TotalAmount = amount
	// 设置产品码，固定为 QUICK_WAP_WAY 表示手机网页支付
	p.ProductCode = "QUICK_WAP_WAY"

	// 调用客户端的 TradeWapPay 方法发起支付请求，获取支付跳转链接
	url, err := client.TradeWapPay(p)
	// 检查发起支付请求是否出错
	if err != nil {
		// 若出错，打印错误信息
		fmt.Println(err)
	}

	// 注释说明 payURL 是用于打开支付宝支付页面的 URL
	// 将 url 转换为字符串类型赋值给 payURL
	var payURL = url.String()
	// 返回支付跳转链接和 nil 错误信息
	return payURL, nil
}

func (r *alipayRepo) AlipayQuery(ctx context.Context, orderId string) (string, string, string, error) {
	// 模拟查询支付结果
	status := "SUCCESS"
	tradeNo := "2024060100000000"
	payTime := time.Now().Format("2006-01-02 15:04:05")
	return status, tradeNo, payTime, nil
}
