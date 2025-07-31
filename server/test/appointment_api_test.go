package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// 预约API测试
// 测试MySQL版本的预约系统API接口

const (
	baseURL = "http://localhost:8000" // 测试服务器地址
)

// TestCreateAppointment 测试创建预约
func TestCreateAppointment(t *testing.T) {
	// 构建预约请求
	appointmentReq := map[string]interface{}{
		"store_id":         "1",
		"customer_name":    "张三",
		"customer_phone":   "13800138000",
		"appointment_date": time.Now().AddDate(0, 0, 1).Format("2006-01-02"), // 明天
		"start_time":       "14:00",
		"duration_minutes": 60,
		"requirements":     "需要了解二手房购买流程",
	}

	// 发送请求
	resp, err := sendPostRequest("/api/v1/appointment/appointments", appointmentReq)
	if err != nil {
		t.Fatalf("发送创建预约请求失败: %v", err)
	}

	// 检查响应
	if resp.StatusCode != http.StatusOK {
		t.Errorf("期望状态码 200，实际得到 %d", resp.StatusCode)
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	// 验证响应内容
	if appointment, ok := result["appointment"].(map[string]interface{}); ok {
		if code, exists := appointment["appointment_code"].(string); exists {
			t.Logf("预约创建成功，预约码: %s", code)
		} else {
			t.Error("响应中缺少预约码")
		}
	} else {
		t.Error("响应中缺少预约信息")
	}
}

// TestGetAppointmentByCode 测试根据预约码查询预约
func TestGetAppointmentByCode(t *testing.T) {
	// 首先创建一个预约
	appointmentCode := createTestAppointment(t)
	if appointmentCode == "" {
		t.Fatal("创建测试预约失败")
	}

	// 根据预约码查询
	url := fmt.Sprintf("/api/v1/appointment/appointments/%s", appointmentCode)
	resp, err := sendGetRequest(url)
	if err != nil {
		t.Fatalf("发送查询请求失败: %v", err)
	}

	// 检查响应
	if resp.StatusCode != http.StatusOK {
		t.Errorf("期望状态码 200，实际得到 %d", resp.StatusCode)
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	// 验证响应内容
	if appointment, ok := result["appointment"].(map[string]interface{}); ok {
		if code, exists := appointment["appointment_code"].(string); exists && code == appointmentCode {
			t.Logf("预约查询成功，预约码: %s", code)
		} else {
			t.Error("预约码不匹配")
		}
	} else {
		t.Error("响应中缺少预约信息")
	}
}

// TestGetAvailableSlots 测试获取可预约时段
func TestGetAvailableSlots(t *testing.T) {
	// 构建查询参数
	storeID := "1"
	startDate := time.Now().Format("2006-01-02")
	days := 7

	url := fmt.Sprintf("/api/v1/appointment/stores/%s/slots?start_date=%s&days=%d",
		storeID, startDate, days)

	// 发送请求
	resp, err := sendGetRequest(url)
	if err != nil {
		t.Fatalf("发送查询请求失败: %v", err)
	}

	// 检查响应
	if resp.StatusCode != http.StatusOK {
		t.Errorf("期望状态码 200，实际得到 %d", resp.StatusCode)
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	// 验证响应内容
	if slots, ok := result["slots"].([]interface{}); ok {
		t.Logf("获取到 %d 个可预约时段", len(slots))

		// 检查第一个时段的结构
		if len(slots) > 0 {
			if slot, ok := slots[0].(map[string]interface{}); ok {
				requiredFields := []string{"date", "start_time", "end_time", "available"}
				for _, field := range requiredFields {
					if _, exists := slot[field]; !exists {
						t.Errorf("时段信息缺少字段: %s", field)
					}
				}
			}
		}
	} else {
		t.Error("响应中缺少时段信息")
	}
}

// TestCancelAppointment 测试取消预约
func TestCancelAppointment(t *testing.T) {
	// 首先创建一个预约
	appointmentCode := createTestAppointment(t)
	if appointmentCode == "" {
		t.Fatal("创建测试预约失败")
	}

	// 取消预约
	cancelReq := map[string]interface{}{
		"reason": "临时有事，需要取消",
	}

	url := fmt.Sprintf("/api/v1/appointment/appointments/%s/cancel", appointmentCode)
	resp, err := sendPostRequest(url, cancelReq)
	if err != nil {
		t.Fatalf("发送取消请求失败: %v", err)
	}

	// 检查响应
	if resp.StatusCode != http.StatusOK {
		t.Errorf("期望状态码 200，实际得到 %d", resp.StatusCode)
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	// 验证响应内容
	if success, ok := result["success"].(bool); ok && success {
		t.Log("预约取消成功")
	} else {
		t.Error("预约取消失败")
	}
}

// TestRealtorStatusUpdate 测试经纪人状态更新
func TestRealtorStatusUpdate(t *testing.T) {
	// 更新经纪人状态为在线
	statusReq := map[string]interface{}{
		"realtor_id": "1",
		"status":     "online",
	}

	resp, err := sendPostRequest("/api/v1/appointment/realtor/status", statusReq)
	if err != nil {
		t.Fatalf("发送状态更新请求失败: %v", err)
	}

	// 检查响应
	if resp.StatusCode != http.StatusOK {
		t.Errorf("期望状态码 200，实际得到 %d", resp.StatusCode)
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	// 验证响应内容
	if success, ok := result["success"].(bool); ok && success {
		t.Log("经纪人状态更新成功")
	} else {
		t.Error("经纪人状态更新失败")
	}
}

// 辅助函数

// createTestAppointment 创建测试预约并返回预约码
func createTestAppointment(t *testing.T) string {
	appointmentReq := map[string]interface{}{
		"store_id":         "1",
		"customer_name":    "测试用户",
		"customer_phone":   "13900139000",
		"appointment_date": time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
		"start_time":       "15:00",
		"duration_minutes": 60,
		"requirements":     "测试预约",
	}

	resp, err := sendPostRequest("/api/v1/appointment/appointments", appointmentReq)
	if err != nil {
		t.Logf("创建测试预约失败: %v", err)
		return ""
	}

	if resp.StatusCode != http.StatusOK {
		t.Logf("创建测试预约失败，状态码: %d", resp.StatusCode)
		return ""
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Logf("解析创建预约响应失败: %v", err)
		return ""
	}

	if appointment, ok := result["appointment"].(map[string]interface{}); ok {
		if code, exists := appointment["appointment_code"].(string); exists {
			return code
		}
	}

	return ""
}

// sendGetRequest 发送GET请求
func sendGetRequest(path string) (*http.Response, error) {
	url := baseURL + path
	return http.Get(url)
}

// sendPostRequest 发送POST请求
func sendPostRequest(path string, data interface{}) (*http.Response, error) {
	url := baseURL + path

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return http.Post(url, "application/json", bytes.NewBuffer(jsonData))
}

// BenchmarkCreateAppointment 预约创建性能测试
func BenchmarkCreateAppointment(b *testing.B) {
	appointmentReq := map[string]interface{}{
		"store_id":         "1",
		"customer_name":    "性能测试用户",
		"customer_phone":   "13800138888",
		"appointment_date": time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
		"start_time":       "16:00",
		"duration_minutes": 60,
		"requirements":     "性能测试",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 每次测试使用不同的手机号避免重复
		appointmentReq["customer_phone"] = fmt.Sprintf("138%08d", i)

		resp, err := sendPostRequest("/api/v1/appointment/appointments", appointmentReq)
		if err != nil {
			b.Fatalf("发送请求失败: %v", err)
		}
		resp.Body.Close()
	}
}

// TestConcurrentAppointments 并发预约测试
func TestConcurrentAppointments(t *testing.T) {
	concurrency := 10
	results := make(chan error, concurrency)

	// 启动多个并发预约
	for i := 0; i < concurrency; i++ {
		go func(index int) {
			appointmentReq := map[string]interface{}{
				"store_id":         "1",
				"customer_name":    fmt.Sprintf("并发用户%d", index),
				"customer_phone":   fmt.Sprintf("139%08d", index),
				"appointment_date": time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
				"start_time":       "17:00",
				"duration_minutes": 60,
				"requirements":     "并发测试",
			}

			resp, err := sendPostRequest("/api/v1/appointment/appointments", appointmentReq)
			if err != nil {
				results <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				results <- fmt.Errorf("状态码错误: %d", resp.StatusCode)
				return
			}

			results <- nil
		}(i)
	}

	// 收集结果
	successCount := 0
	for i := 0; i < concurrency; i++ {
		if err := <-results; err != nil {
			t.Logf("并发预约失败: %v", err)
		} else {
			successCount++
		}
	}

	t.Logf("并发预约测试完成，成功: %d/%d", successCount, concurrency)

	if successCount == 0 {
		t.Error("所有并发预约都失败了")
	}
}
