package server

import (
	appointmentv1 "anjuke/server/api/appointment/v1"
	companyv1 "anjuke/server/api/company/v1"
	v1 "anjuke/server/api/helloworld/v1"
	v5 "anjuke/server/api/points/v5"
	userv2 "anjuke/server/api/user/v2"
	"anjuke/server/internal/common"
	"anjuke/server/internal/conf"
	"anjuke/server/internal/domain"
	"anjuke/server/internal/service"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"anjuke/server/internal/data"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
)

// 等待WebSocket连接建立的函数
func waitForWebSocketConnection(uploadID string, timeout time.Duration) {
	if GlobalProgressHub == nil {
		return
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		GlobalProgressHub.mu.Lock()
		if conns, exists := GlobalProgressHub.connections[uploadID]; exists && len(conns) > 0 {
			GlobalProgressHub.mu.Unlock()
			return
		}
		GlobalProgressHub.mu.Unlock()
		time.Sleep(100 * time.Millisecond)
	}
}

// Server 是HTTP服务器的包装
type Server struct {
	log         *log.Helper
	minioClient *data.MinioClient
}

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, greeter *service.GreeterService, user *service.UserService, points *service.PointsService, company *service.CompanyService, store *service.StoreService, realtor *service.RealtorService, appointment *service.AppointmentService, minioClient *data.MinioClient, logger log.Logger) *kratoshttp.Server {
	// 初始化随机数种子
	rand.Seed(time.Now().UnixNano())

	// 初始化服务器实例
	s := &Server{
		log:         log.NewHelper(logger),
		minioClient: minioClient,
	}

	// 初始化全局进度Hub
	InitProgressHub(logger)
	var opts = []kratoshttp.ServerOption{
		kratoshttp.Middleware(
			recovery.Recovery(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, kratoshttp.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, kratoshttp.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, kratoshttp.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := kratoshttp.NewServer(opts...)
	v1.RegisterGreeterHTTPServer(srv, greeter)
	userv2.RegisterUserHTTPServer(srv, user)
	v5.RegisterPointsHTTPServer(srv, points)
	companyv1.RegisterCompanyHTTPServer(srv, company)
	companyv1.RegisterStoreHTTPServer(srv, store)
	companyv1.RegisterRealtorHTTPServer(srv, realtor)
	appointmentv1.RegisterAppointmentServiceHTTPServer(srv, appointment)

	// 添加一个支持multipart/form-data的上传接口
	// 添加WebSocket进度通知端点
	srv.HandleFunc("/api/upload/progress", s.WebSocketHandler)

	// 统一文件上传接口
	srv.HandleFunc("/api/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// 从请求中获取uploadID，如果没有提供则自动生成
		uploadID := r.FormValue("uploadID")
		if uploadID == "" {
			uploadID = common.GenerateUploadID()
		}

		// 解析multipart表单
		err := r.ParseMultipartForm(32 << 20) // 32MB限制
		if err != nil {
			http.Error(w, "无法解析multipart表单: "+err.Error(), http.StatusBadRequest)
			return
		}

		// 获取文件
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "无法获取上传文件: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		// 等待WebSocket连接建立（最多等待3秒）
		waitForWebSocketConnection(uploadID, 3*time.Second)

		// 初始化进度为0
		if GlobalProgressHub != nil {
			GlobalProgressHub.UpdateProgress(uploadID, 0, "准备上传")
		}

		// 获取文件名和content type
		filename := header.Filename
		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		// 日志输出
		logger.Log(log.LevelInfo, "msg", fmt.Sprintf("开始上传: %s (%.2fMB)",
			filename, float64(header.Size)/(1024*1024)))

		// 创建带有30分钟超时的上下文
		uploadCtx, cancel := context.WithTimeout(r.Context(), 30*time.Minute)
		defer cancel()

		// 创建一个进度跟踪函数
		progressTracker := func(uploaded, total int64) {
			if GlobalProgressHub != nil {
				// 计算上传百分比进度
				progress := float64(uploaded) / float64(total) * 100
				status := "上传中"

				if uploaded >= total {
					status = "处理中" // 文件已上传完，但服务器仍在处理
				}

				GlobalProgressHub.UpdateProgress(uploadID, progress, status)
			}
		}

		// 直接调用MinIO进行上传，统一处理所有文件大小
		url, err := minioClient.SmartUpload(
			uploadCtx, filename, file, header.Size, contentType, progressTracker)

		if err != nil {
			logger.Log(log.LevelError, "msg", "上传失败", "error", err, "filename", filename)
			// 更新进度状态为失败
			if GlobalProgressHub != nil {
				GlobalProgressHub.UpdateProgress(uploadID, 0, "上传失败")
			}

			// 返回JSON格式的错误响应
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 500,
				"msg":  "上传失败: " + err.Error(),
				"data": nil,
			})
			return
		}

		// 更新进度为完成
		if GlobalProgressHub != nil {
			GlobalProgressHub.UpdateProgress(uploadID, 100, "上传完成")
			// 延迟发送最终状态，确保前端能收到，然后关闭连接
			go func() {
				time.Sleep(2 * time.Second)
				GlobalProgressHub.UpdateProgress(uploadID, 100, "处理完成")
				// 再等待1秒后清理连接
				time.Sleep(1 * time.Second)
				GlobalProgressHub.CleanupUpload(uploadID)
			}()
		}

		// 返回成功响应
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"msg":  "上传成功",
			"data": map[string]interface{}{
				"url":      url,
				"uploadID": uploadID, // 返回上传ID，前端可用于跟踪进度
			},
		})
	})

	// 手动注册文件删除路由
	srv.HandleFunc("/user/deleteFile", func(w http.ResponseWriter, r *http.Request) {
		objectName := r.URL.Query().Get("filename")
		if objectName == "" && r.Method == "POST" {
			var req struct {
				Filename   string `json:"filename"`
				ObjectName string `json:"objectName"`
			}
			err := json.NewDecoder(r.Body).Decode(&req)
			if err == nil {
				if req.ObjectName != "" {
					objectName = req.ObjectName
				} else {
					objectName = req.Filename
				}
			}
		}
		if objectName == "" {
			http.Error(w, "缺少文件名", http.StatusBadRequest)
			return
		}
		// 文件名格式校验
		if len(objectName) < 1 || len(objectName) > 255 {
			http.Error(w, "文件名长度非法", http.StatusBadRequest)
			return
		}

		// 通过正确的调用链路：HTTP -> Service -> Usecase -> Repo
		err := user.DeleteFromMinio(r.Context(), objectName)
		if err != nil {
			logger.Log(log.LevelError, "msg", "删除文件失败", "error", err, "objectName", objectName)
			http.Error(w, "删除失败: "+err.Error(), http.StatusInternalServerError)
			return
		}

		logger.Log(log.LevelInfo, "msg", "文件删除成功", "objectName", objectName)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"msg":  "删除成功",
			"data": map[string]string{"objectName": objectName},
		})
	})

	srv.HandleFunc("/admin/cleanupIncompleteUploads", func(w http.ResponseWriter, r *http.Request) {
		prefix := r.URL.Query().Get("prefix")
		olderThanStr := r.URL.Query().Get("older_than")
		olderThan := 24 * time.Hour
		if olderThanStr != "" {
			if d, err := time.ParseDuration(olderThanStr); err == nil {
				olderThan = d
			}
		}
		if minioClient == nil {
			http.Error(w, "minioClient 未初始化", http.StatusInternalServerError)
			return
		}
		err := minioClient.CleanupIncompleteUploads(r.Context(), prefix, olderThan)
		if err != nil {
			http.Error(w, "清理失败: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "msg": "清理成功"})
	})

	// 文件列表接口
	srv.HandleFunc("/user/fileList", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// 获取查询参数
		keyword := r.URL.Query().Get("keyword")
		pageStr := r.URL.Query().Get("page")
		pageSizeStr := r.URL.Query().Get("pageSize")

		// 设置默认值
		page := 1
		pageSize := 10
		if pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}
		if pageSizeStr != "" {
			if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
				pageSize = ps
			}
		}

		// 计算最大返回数量（简单分页）
		maxKeys := page * pageSize

		var files []domain.FileInfo
		var err error

		if keyword != "" {
			// 搜索文件
			files, _, err = minioClient.SearchFiles(r.Context(), keyword, int32(page), int32(pageSize))
		} else {
			// 列出所有文件
			files, err = minioClient.ListFiles(r.Context(), "", maxKeys)
		}

		if err != nil {
			http.Error(w, "获取文件列表失败: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 简单分页处理
		total := len(files)
		start := (page - 1) * pageSize
		end := start + pageSize

		if start >= total {
			files = []domain.FileInfo{}
		} else {
			if end > total {
				end = total
			}
			files = files[start:end]
		}

		// 转换为前端需要的格式
		var fileList []map[string]interface{}
		for i, file := range files {
			fileList = append(fileList, map[string]interface{}{
				"id":          start + i + 1, // 简单的ID生成
				"name":        file.Name,
				"fileName":    file.Name,
				"size":        file.Size,
				"fileSize":    file.Size,
				"type":        file.ContentType,
				"mimeType":    file.ContentType,
				"url":         file.URL,
				"downloadUrl": file.URL,
				"uploadTime":  file.LastModified.Format("2006-01-02 15:04:05"),
				"createTime":  file.LastModified.Format("2006-01-02 15:04:05"),
				"status":      "success",
				"objectName":  file.Name,
			})
		}

		response := map[string]interface{}{
			"code": 0,
			"msg":  "获取成功",
			"data": map[string]interface{}{
				"list":     fileList,
				"total":    total,
				"page":     page,
				"pageSize": pageSize,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// 文件统计接口
	srv.HandleFunc("/user/uploadStats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		stats, err := minioClient.GetFileStats(r.Context())
		if err != nil {
			http.Error(w, "获取统计信息失败: "+err.Error(), http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"code": 0,
			"msg":  "获取成功",
			"data": stats,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// 文件详情接口
	srv.HandleFunc("/user/fileInfo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		objectName := r.URL.Query().Get("objectName")
		if objectName == "" {
			http.Error(w, "objectName 参数不能为空", http.StatusBadRequest)
			return
		}

		fileInfo, err := minioClient.GetFileInfo(r.Context(), objectName)
		if err != nil {
			http.Error(w, "获取文件信息失败: "+err.Error(), http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"code": 0,
			"msg":  "获取成功",
			"data": map[string]interface{}{
				"name":       fileInfo.Name,
				"size":       fileInfo.Size,
				"type":       fileInfo.ContentType,
				"url":        fileInfo.URL,
				"uploadTime": fileInfo.LastModified.Format("2006-01-02 15:04:05"),
				"etag":       fileInfo.ETag,
				"objectName": fileInfo.Name,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// 健康检查接口
	srv.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"version":   "1.0.0",
		})
	})

	return srv
}
