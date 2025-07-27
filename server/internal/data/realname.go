package data

import (
	"anjuke/server/internal/conf"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	gourl "net/url"
	"strings"
	"time"

	"github.com/hashicorp/go-uuid"
)

type RealNameSDK struct {
	conf *conf.Data
}

func NewRealNameSDK(conf *conf.Data) *RealNameSDK {
	return &RealNameSDK{conf: conf}
}

func calcAuthorization(secretId string, secretKey string) (auth string, datetime string, err error) {
	timeLocation, _ := time.LoadLocation("Etc/GMT")
	datetime = time.Now().In(timeLocation).Format("Mon, 02 Jan 2006 15:04:05 GMT")
	signStr := fmt.Sprintf("x-date: %s", datetime)

	// hmac-sha1
	mac := hmac.New(sha1.New, []byte(secretKey))
	mac.Write([]byte(signStr))
	sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	auth = fmt.Sprintf("{\"id\":\"%s\", \"x-date\":\"%s\", \"signature\":\"%s\"}",
		secretId, datetime, sign)

	return auth, datetime, nil
}

func urlencode(params map[string]string) string {
	var p = gourl.Values{}
	for k, v := range params {
		p.Add(k, v)
	}
	return p.Encode()
}

type RealNameResp struct {
	ErrorCode int    `json:"error_code"`
	Reason    string `json:"reason"`
	Result    struct {
		IsOk bool `json:"isok"`
		// 其他字段可按需补充
	} `json:"result"`
}

// RealName 实名认证
func (sdk *RealNameSDK) RealName(name string, idCard string) (bool, error) {
	if sdk.conf == nil {
		return false, fmt.Errorf("RealNameSDK.conf is nil (配置未注入)")
	}
	if sdk.conf.TencentYunRealName == nil {
		return false, fmt.Errorf("RealNameSDK.conf.TencentYunRealName is nil (配置项缺失)")
	}
	secretId := sdk.conf.TencentYunRealName.SecretId
	secretKey := sdk.conf.TencentYunRealName.SecretKey
	auth, _, _ := calcAuthorization(secretId, secretKey)

	method := "POST"
	reqID, err := uuid.GenerateUUID()
	if err != nil {
		return false, err
	}
	headers := map[string]string{"Authorization": auth, "request-id": reqID}

	queryParams := make(map[string]string)
	bodyParams := make(map[string]string)
	bodyParams["cardNo"] = idCard
	bodyParams["realName"] = name
	bodyParamStr := urlencode(bodyParams)
	url := "https://ap-beijing.cloudmarket-apigw.com/service-18c38npd/idcard/VerifyIdcardv2"

	if len(queryParams) > 0 {
		url = fmt.Sprintf("%s?%s", url, urlencode(queryParams))
	}

	bodyMethods := map[string]bool{"POST": true, "PUT": true, "PATCH": true}
	var body io.Reader = nil
	if bodyMethods[method] {
		body = strings.NewReader(bodyParamStr)
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return false, err
	}

	// 明确设置 Content-Type，避免被自动修改
	if bodyMethods[method] {
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	for k, v := range headers {
		request.Header.Set(k, v)
	}

	// 调试：打印请求头信息
	// fmt.Printf("Content-Type: %s\n", request.Header.Get("Content-Type"))
	// fmt.Printf("Authorization: %s\n", request.Header.Get("Authorization"))

	response, err := client.Do(request)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return false, err
	}
	fmt.Printf("实名认证原始响应: %s\n", string(bodyBytes))

	var resp RealNameResp
	if err := json.Unmarshal(bodyBytes, &resp); err != nil {
		return false, fmt.Errorf("解析实名认证响应失败: %v", err)
	}

	if resp.ErrorCode != 0 {
		return false, fmt.Errorf("接口调用失败: %s", resp.Reason)
	}
	if !resp.Result.IsOk {
		return false, fmt.Errorf("实名认证未通过: 身份证号码与姓名不匹配")
	}
	return true, nil
}
