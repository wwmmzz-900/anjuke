package server

import (
	v6 "anjuke/server/api/customer/v6"
	v1 "anjuke/server/api/helloworld/v1"
	v3 "anjuke/server/api/house/v3"
	v5 "anjuke/server/api/points/v5"
	v4 "anjuke/server/api/transaction/v4"
	userv2 "anjuke/server/api/user/v2"
	"anjuke/server/internal/conf"
	"anjuke/server/internal/service"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"anjuke/server/internal/data"

	uploadv1 "anjuke/server/api/upload/v1"

	"anjuke/server/internal/domain"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
)

// Server 是HTTP服务器的包装
type Server struct {
	log         *log.Helper
	minioClient *data.MinioClient
}

// 生成随机字符串，用于生成上传ID
func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, greeter *service.GreeterService, user *service.UserService, house *service.HouseService, transaction *service.TransactionService, points *service.PointsService, customer *service.CustomerService, minioClient *data.MinioClient, logger log.Logger) *kratoshttp.Server {
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
	v3.RegisterHouseHTTPServer(srv, house)
	v4.RegisterTransactionHTTPServer(srv, transaction)
	v5.RegisterPointsHTTPServer(srv, points)
	v6.RegisterCustomerHTTPServer(srv, customer)

	// 添加一个支持multipart/form-data的上传接口
	// 添加WebSocket进度通知端点
	srv.HandleFunc("/api/upload/progress", s.WebSocketHandler)

	srv.HandleFunc("/api/upload/smart", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// 生成唯一的上传ID，用于跟踪上传进度
		uploadID := GenerateUploadID()

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
		logger.Log(log.LevelInfo, "msg", fmt.Sprintf("开始处理上传: 文件=%s, 大小=%d, 类型=%s",
			filename, header.Size, contentType))

		// 使用uploadService处理上传
		uploadService := service.NewUploadService(data.NewMinioRepo(minioClient), logger)

		// 大文件处理策略
		if header.Size > 5*1024*1024 { // 5MB以上的文件
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

			// 直接将文件流传递给MinIO进行处理，避免将整个文件加载到内存
			url, err := uploadService.MinioRepo().SmartUpload(
				uploadCtx, filename, file, header.Size, contentType, progressTracker)

			if err != nil {
				logger.Log(log.LevelError, "msg", "大文件上传失败", "error", err)
				// 更新进度状态为失败
				if GlobalProgressHub != nil {
					GlobalProgressHub.UpdateProgress(uploadID, 0, "上传失败")
				}
				http.Error(w, "上传失败: "+err.Error(), http.StatusInternalServerError)
				return
			}

			// 更新进度为完成
			if GlobalProgressHub != nil {
				GlobalProgressHub.UpdateProgress(uploadID, 100, "上传完成")
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
			return
		}

		// 小文件处理 - 读取到内存中
		if GlobalProgressHub != nil {
			GlobalProgressHub.UpdateProgress(uploadID, 10, "读取文件")
		}

		fileData, err := io.ReadAll(file)
		if err != nil {
			if GlobalProgressHub != nil {
				GlobalProgressHub.UpdateProgress(uploadID, 0, "上传失败")
			}
			http.Error(w, "读取文件失败: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if GlobalProgressHub != nil {
			GlobalProgressHub.UpdateProgress(uploadID, 50, "处理中")
		}

		// 创建SimpleUploadRequest对象
		req := &uploadv1.SimpleUploadRequest{
			Filename:    filename,
			ContentType: contentType,
			FileData:    fileData,
		}

		// 创建带有10分钟超时的上下文
		uploadCtx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
		defer cancel()

		// 调用上传服务
		resp, err := uploadService.SimpleUpload(uploadCtx, req)
		if err != nil {
			if GlobalProgressHub != nil {
				GlobalProgressHub.UpdateProgress(uploadID, 0, "上传失败")
			}
			http.Error(w, "上传失败: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 更新进度为完成
		if GlobalProgressHub != nil {
			GlobalProgressHub.UpdateProgress(uploadID, 100, "上传完成")
		}

		// 返回成功响应
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"msg":  "上传成功",
			"data": map[string]interface{}{
				"url":      resp.Url,
				"uploadID": uploadID, // 返回上传ID，前端可用于跟踪进度
			},
		})
	})

	// 手动注册文件删除路由
	srv.HandleFunc("/user/deleteFile", func(w http.ResponseWriter, r *http.Request) {
		objectName := r.URL.Query().Get("filename")
		if objectName == "" && r.Method == "POST" {
			var req struct {
				Filename string `json:"filename"`
			}
			err := json.NewDecoder(r.Body).Decode(&req)
			if err == nil {
				objectName = req.Filename
			}
		}
		if objectName == "" {
			http.Error(w, "缺少文件名", http.StatusBadRequest)
			return
		}
		// 文件名格式校验（可选，简单示例）
		if len(objectName) < 5 || len(objectName) > 128 {
			http.Error(w, "文件名长度非法", http.StatusBadRequest)
			return
		}
		// 日志打印，便于排查
		fmt.Printf("[删除文件] objectName: %s\n", objectName)
		err := user.DeleteFromMinio(r.Context(), objectName)
		if err != nil {
			http.Error(w, "删除失败: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"msg":  "删除成功",
			"data": map[string]string{"objectName": objectName},
		})
	})

	// 多文件上传接口
	srv.HandleFunc("/user/uploadFiles", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(32 << 20) // 32MB
		if err != nil {
			http.Error(w, "表单解析失败", http.StatusBadRequest)
			return
		}
		files := r.MultipartForm.File["files"]
		var results []map[string]interface{}
		for _, header := range files {
			file, err := header.Open()
			if err != nil {
				continue
			}
			defer file.Close()
			contentType := header.Header.Get("Content-Type")
			url, err := user.UploadToMinioWithProgress(r.Context(), header.Filename, file, header.Size, contentType, nil)
			if err != nil {
				continue
			}
			results = append(results, map[string]interface{}{
				"url":          url,
				"filename":     header.Filename,
				"size":         header.Size,
				"content_type": contentType,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "msg": "上传成功", "data": results})
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
			files, err = minioClient.SearchFiles(r.Context(), keyword, maxKeys)
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

		stats, err := minioClient.GetFileStats(r.Context(), "")
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

	uploadService := service.NewUploadService(data.NewMinioRepo(minioClient), logger)
	uploadv1.RegisterUploadServiceHTTPServer(srv, uploadService)

	return srv
}
