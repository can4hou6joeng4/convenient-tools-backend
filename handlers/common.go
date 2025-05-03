package handlers

import (
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/can4hou6joeng4/convenient-tools-project-v1-backend/config"
	"github.com/can4hou6joeng4/convenient-tools-project-v1-backend/models"
	"github.com/can4hou6joeng4/convenient-tools-project-v1-backend/repositories"
	"github.com/can4hou6joeng4/convenient-tools-project-v1-backend/service"
	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/redis/go-redis/v9"
	"github.com/tencentyun/cos-go-sdk-v5"
)

type CommonHandler struct {
	redis      *redis.Client
	cos        *cos.Client
	repository *repositories.ToolRepository
	config     *config.EnvConfig
}

// GetTools godoc
// @Summary 获取所有工具列表
// @Description 获取系统中所有已注册的工具
// @Tags tools
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /tools/list [get]
func (h *CommonHandler) GetTools(ctx *fiber.Ctx) error {
	tools, err := h.repository.GetAllTools()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Get tools list failed",
		})
	}
	if len(tools) == 0 {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "No tools found",
		})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Get tools list success",
		"data":    tools,
	})
}

// CreateTool godoc
// @Summary 创建新工具
// @Description 创建一个新的工具
// @Tags tools
// @Accept json
// @Produce json
// @Param tool body models.Tool true "工具信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /tools [post]
func (h *CommonHandler) CreateTool(ctx *fiber.Ctx) error {
	tool := &models.Tool{}
	if err := ctx.BodyParser(tool); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "fail",
			"message": "Invalid request body",
		})
	}
	if err := h.repository.CreateTool(tool); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "fail",
			"message": "Create tool failed",
		})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Create tool success",
	})
}

// Upload godoc
// @Summary 上传文件
// @Description 上传文件到服务器
// @Tags file
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "要上传的文件"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /file/upload [post]
func (h *CommonHandler) Upload(ctx *fiber.Ctx) error {
	file, err := ctx.FormFile("file")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "fail",
			"message": "Invalid file",
		})
	}
	fileType := strings.Split(file.Filename, ".")[1]
	if fileType != "pdf" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "fail",
			"message": "Invalid file type Please upload a pdf file",
		})
	}
	open, err := file.Open()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "fail",
			"message": "Open file failed",
		})
	}
	fileUrl := time.Now().Format("20060102") + "/" + time.Now().Format("150405-") + file.Filename
	_, err = h.cos.Object.Put(ctx.Context(), fileUrl, open, nil)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "fail",
			"message": "Upload file failed",
		})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Upload file success",
		"data": map[string]string{
			"url":  fileUrl,
			"name": file.Filename,
		},
	})
}

func (h *CommonHandler) ParseShareUrl(ctx *fiber.Ctx) error {
	// 从JSON请求体中获取URL
	var req struct {
		URL string `json:"url"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "fail",
			"message": "Invalid request format",
		})
	}

	if req.URL == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "fail",
			"message": "URL is required",
		})
	}

	parseInfo, err := service.ParseVideoShareUrlByRegexp(req.URL)
	if err != nil {
		log.Errorf("fail parse %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "fail",
			"message": "Parse URL fail",
		})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Parse URL success",
		"data":    parseInfo,
	})
}

// DownloadFile 处理文件下载请求
// @Summary 媒体文件下载代理
// @Description 代理下载视频、图片等媒体资源，解决小程序环境中的合法域名限制问题
// @Tags tools
// @Accept json
// @Produce octet-stream
// @Param url query string true "需要代理下载的媒体URL"
// @Param filename query string false "下载文件的文件名"
// @Param forceDownload query boolean false "对图片类型是否强制下载而非预览，默认false"
// @Success 200 {file} binary "文件内容"
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /tools/download [get]
func (h *CommonHandler) DownloadFile(ctx *fiber.Ctx) error {
	// 获取请求参数
	url := ctx.Query("url")
	filename := ctx.Query("filename")
	forceDownload := ctx.QueryBool("forceDownload", false)

	// 验证参数
	if url == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "fail",
			"message": "URL is required",
		})
	}

	// 如果没有提供文件名，尝试从URL中提取
	if filename == "" {
		// 从URL中获取文件名
		filename = filepath.Base(url)
		// 如果URL不包含有效的文件名，使用默认名称
		if filename == "." || filename == "/" || filename == "" {
			filename = "download"
		}
	}

	// 使用resty客户端
	client := resty.New()

	// 获取主机名进行特殊处理
	host := service.ExtractHost(url)

	// 针对小红书视频的特殊处理
	if service.IsXiaohongshuHost(host) {
		// 尝试获取备用URL
		alternateURL, err := service.HandleXiaohongshuVideo(client, url)
		if err == nil && alternateURL != url {
			log.Infof("Using alternate URL for xiaohongshu video: %s instead of %s", alternateURL, url)
			url = alternateURL
			// 更新主机名
			host = service.ExtractHost(url)
		}
	}

	// 检查是否有重定向，获取最终URL
	finalURL, redirected, err := service.HandleVideoRedirects(client, url)
	if err != nil && !redirected {
		log.Errorf("Error handling redirects: %v", err)
	} else if redirected {
		log.Infof("URL redirected from %s to %s", url, finalURL)
		url = finalURL
		// 更新主机名
		host = service.ExtractHost(url)
	}

	// 获取针对该平台的特定配置
	hostConfig := service.GetVideoHostConfig(host)

	// 准备请求
	req := client.R().
		SetHeader(service.HttpHeaderUserAgent, service.GetUserAgent(hostConfig.UserAgentType)).
		SetDoNotParseResponse(true) // 不解析响应体，以便流式传输

	// 添加通用请求头
	req.SetHeader("Accept", "*/*")
	req.SetHeader("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.SetHeader("Connection", "keep-alive")

	// 如果有特定的Referer和Origin，添加它们
	if hostConfig.RefererURL != "" {
		req.SetHeader(service.HttpHeaderReferer, hostConfig.RefererURL)
	}

	if hostConfig.OriginURL != "" {
		req.SetHeader("Origin", hostConfig.OriginURL)
	}

	// 添加特定平台的额外请求头
	for key, value := range hostConfig.ExtraHeaders {
		req.SetHeader(key, value)
	}

	// 添加特殊处理，直接复用微信小程序的User-Agent
	if service.IsXiaohongshuHost(host) {
		// 尝试完全匹配微信请求的User-Agent
		req.SetHeader(service.HttpHeaderUserAgent, ctx.Get(service.HttpHeaderUserAgent))
		// 添加微信请求头
		req.SetHeader("Referer", ctx.Get("Referer"))
	}

	// 添加超时设置
	client.SetTimeout(30 * time.Second)
	client.SetRetryCount(2) // 失败时重试

	// 记录请求信息，便于调试
	log.Infof("Downloading file from %s with host %s, using agent: %s", url, host, req.Header.Get(service.HttpHeaderUserAgent))

	// 发送请求获取文件内容
	resp, err := req.Get(url)

	if err != nil {
		log.Errorf("download file error: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "fail",
			"message": "Failed to download file: " + err.Error(),
		})
	}

	// 获取原始响应并确保关闭
	rawResponse := resp.RawResponse
	if rawResponse == nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "fail",
			"message": "Invalid response from source",
		})
	}
	defer resp.RawBody().Close()

	// 检查响应状态
	if rawResponse.StatusCode != http.StatusOK {
		log.Errorf("file source responded with status: %d, headers: %v", rawResponse.StatusCode, rawResponse.Header)

		// 针对小红书视频，尝试替换URL再试一次
		if service.IsXiaohongshuHost(host) && rawResponse.StatusCode == http.StatusBadGateway {
			// 返回自定义错误提示
			return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"status":  "fail",
				"message": "小红书视频服务器拒绝访问，请复制视频链接在浏览器中尝试",
				"details": fmt.Sprintf("服务器返回状态码: %d", rawResponse.StatusCode),
			})
		}

		return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"status":  "fail",
			"message": fmt.Sprintf("File source responded with status: %d", rawResponse.StatusCode),
		})
	}

	// 确定文件类型（Content-Type）
	contentType := rawResponse.Header.Get("Content-Type")
	// 如果源服务器没有提供Content-Type，尝试从文件扩展名猜测
	if contentType == "" || contentType == "application/octet-stream" {
		ext := filepath.Ext(filename)
		if ext != "" {
			// 使用http包提供的MIME类型查找
			contentType = mime.TypeByExtension(ext)
		}
		// 如果仍然无法确定，使用通用二进制类型
		if contentType == "" {
			contentType = "application/octet-stream"
		}
	}

	// 设置响应头
	ctx.Set("Content-Type", contentType)

	// 确定是下载还是预览
	isImage := strings.HasPrefix(contentType, "image/")
	if forceDownload || !isImage {
		// 强制下载或非图片内容
		ctx.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	} else {
		// 图片内容默认预览
		ctx.Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filename))
	}

	// 如果源响应包含Content-Length，设置它
	if cl := rawResponse.Header.Get("Content-Length"); cl != "" {
		ctx.Set("Content-Length", cl)
	}

	// 将文件内容流式传输到客户端
	return ctx.SendStream(resp.RawBody())
}

// ProxyVideo 实现视频代理功能，解决跨域问题
func (h *CommonHandler) ProxyVideo(ctx *fiber.Ctx) error {
	// 获取要代理的视频URL
	videoURL := ctx.Query("url")
	if videoURL == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "fail",
			"message": "Video URL is required",
		})
	}

	// 使用resty客户端
	client := resty.New()

	// 获取主机名进行特殊处理
	host := service.ExtractHost(videoURL)

	// 针对小红书视频的特殊处理
	if service.IsXiaohongshuHost(host) {
		// 尝试获取备用URL
		alternateURL, err := service.HandleXiaohongshuVideo(client, videoURL)
		if err == nil && alternateURL != videoURL {
			log.Infof("Using alternate URL for xiaohongshu video: %s instead of %s", alternateURL, videoURL)
			videoURL = alternateURL
			// 更新主机名
			host = service.ExtractHost(videoURL)
		}
	}

	// 检查是否有重定向，获取最终URL
	finalURL, redirected, err := service.HandleVideoRedirects(client, videoURL)
	if err != nil && !redirected {
		log.Errorf("Error handling redirects: %v", err)
	} else if redirected {
		log.Infof("URL redirected from %s to %s", videoURL, finalURL)
		videoURL = finalURL
		// 更新主机名
		host = service.ExtractHost(videoURL)
	}

	// 获取针对该平台的特定配置
	hostConfig := service.GetVideoHostConfig(host)

	// 准备请求
	req := client.R().
		SetHeader(service.HttpHeaderUserAgent, service.GetUserAgent(hostConfig.UserAgentType)).
		SetDoNotParseResponse(true) // 不解析响应体，便于流式传输

	// 添加通用请求头
	req.SetHeader("Accept", "*/*")
	req.SetHeader("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.SetHeader("Connection", "keep-alive")

	// 如果有特定的Referer和Origin，添加它们
	if hostConfig.RefererURL != "" {
		req.SetHeader(service.HttpHeaderReferer, hostConfig.RefererURL)
	}

	if hostConfig.OriginURL != "" {
		req.SetHeader("Origin", hostConfig.OriginURL)
	}

	// 添加特定平台的额外请求头
	for key, value := range hostConfig.ExtraHeaders {
		req.SetHeader(key, value)
	}

	// 添加特殊处理，直接复用微信小程序的User-Agent
	if service.IsXiaohongshuHost(host) {
		// 尝试完全匹配微信请求的User-Agent
		req.SetHeader(service.HttpHeaderUserAgent, ctx.Get(service.HttpHeaderUserAgent))
		// 添加微信请求头
		req.SetHeader("Referer", ctx.Get("Referer"))
	}

	// 添加超时设置
	client.SetTimeout(30 * time.Second)
	client.SetRetryCount(2) // 失败时重试

	// 记录请求信息，便于调试
	log.Infof("Proxying video from %s with host %s, using agent: %s", videoURL, host, req.Header.Get(service.HttpHeaderUserAgent))

	// 发送请求
	resp, err := req.Get(videoURL)

	if err != nil {
		log.Errorf("fetch video error: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "fail",
			"message": "Failed to fetch video: " + err.Error(),
		})
	}

	// 获取原始响应
	rawResponse := resp.RawResponse
	if rawResponse == nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "fail",
			"message": "Invalid response from video source",
		})
	}
	defer resp.RawBody().Close()

	// 检查响应状态
	if rawResponse.StatusCode != http.StatusOK {
		log.Errorf("video source responded with status: %d, headers: %v", rawResponse.StatusCode, rawResponse.Header)

		// 针对小红书视频，返回更友好的错误提示
		if service.IsXiaohongshuHost(host) && rawResponse.StatusCode == http.StatusBadGateway {
			return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"status":  "fail",
				"message": "小红书视频服务器拒绝访问，请使用其他方式下载或播放",
				"details": fmt.Sprintf("服务器返回状态码: %d", rawResponse.StatusCode),
			})
		}

		return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"status":  "fail",
			"message": fmt.Sprintf("Video source responded with status: %d", rawResponse.StatusCode),
		})
	}

	// 设置响应头，保留源服务器的内容类型和长度
	contentType := rawResponse.Header.Get("Content-Type")
	if contentType == "" {
		// 如果原始服务器没有提供Content-Type，基于URL猜测
		ext := filepath.Ext(videoURL)
		switch strings.ToLower(ext) {
		case ".mp4":
			contentType = "video/mp4"
		case ".webm":
			contentType = "video/webm"
		case ".ogg":
			contentType = "video/ogg"
		case ".mov":
			contentType = "video/quicktime"
		default:
			contentType = "application/octet-stream"
		}
	}

	// 设置响应头
	ctx.Set("Content-Type", contentType)
	if cl := rawResponse.Header.Get("Content-Length"); cl != "" {
		ctx.Set("Content-Length", cl)
	}

	// 支持范围请求，对于视频很重要
	ctx.Set("Accept-Ranges", "bytes")

	// 设置缓存控制
	ctx.Set("Cache-Control", "public, max-age=86400") // 缓存一天

	// 支持跨域请求
	ctx.Set("Access-Control-Allow-Origin", "*")
	ctx.Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	ctx.Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

	// 将视频流直接传输给客户端
	return ctx.SendStream(resp.RawBody())
}

// GetWeChatDownloadConfig 获取微信小程序下载配置信息
// @Summary 获取微信小程序下载配置信息
// @Description 根据文件URL生成微信小程序下载所需的配置信息，包括权限要求和API调用建议
// @Tags tools
// @Accept json
// @Produce json
// @Param url query string true "需要下载的媒体URL"
// @Param filename query string false "下载文件的文件名"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /tools/wechat-download-config [get]
func (h *CommonHandler) GetWeChatDownloadConfig(ctx *fiber.Ctx) error {
	// 获取请求参数
	url := ctx.Query("url")
	filename := ctx.Query("filename")

	// 验证参数
	if url == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "fail",
			"message": "URL is required",
		})
	}

	// 如果没有提供文件名，尝试从URL中提取
	if filename == "" {
		// 从URL中获取文件名
		filename = filepath.Base(url)
		// 如果URL不包含有效的文件名，使用默认名称
		if filename == "." || filename == "/" || filename == "" {
			filename = "download"
		}
	}

	// 使用resty客户端发送HEAD请求获取Content-Type
	client := resty.New()
	resp, err := client.R().
		SetHeader(service.HttpHeaderUserAgent, service.DefaultUserAgent).
		Head(url)

	if err != nil {
		log.Errorf("head request error: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "fail",
			"message": "Failed to get file information",
		})
	}

	// 获取Content-Type
	contentType := resp.Header().Get("Content-Type")

	// 如果无法获取Content-Type，尝试从文件扩展名猜测
	if contentType == "" {
		ext := filepath.Ext(filename)
		if ext != "" {
			// 使用http包提供的MIME类型查找
			contentType = mime.TypeByExtension(ext)
		}
		// 如果仍然无法确定，使用通用二进制类型
		if contentType == "" {
			contentType = "application/octet-stream"
		}
	}

	// 生成微信小程序下载配置
	config := service.GenerateWeChatDownloadConfig(url, filename, contentType)

	// 返回配置信息
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Get WeChat download config success",
		"data":    config,
	})
}

// GetXiaohongshuVideoInfo 获取小红书视频信息
// @Summary 获取小红书视频直接链接
// @Description 尝试解析小红书视频地址，获取可以直接播放的视频链接
// @Tags tools
// @Accept json
// @Produce json
// @Param url query string true "小红书视频URL"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /tools/xiaohongshu/video-info [get]
func (h *CommonHandler) GetXiaohongshuVideoInfo(ctx *fiber.Ctx) error {
	// 获取视频URL
	videoURL := ctx.Query("url")
	if videoURL == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "fail",
			"message": "URL is required",
		})
	}

	// 验证是否是小红书链接
	host := service.ExtractHost(videoURL)
	if !service.IsXiaohongshuHost(host) && !strings.Contains(videoURL, "xiaohongshu") && !strings.Contains(videoURL, "xhslink") {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "fail",
			"message": "Not a Xiaohongshu URL",
		})
	}

	// 解析小红书链接
	client := resty.New()

	// 1. 尝试处理重定向
	finalURL, redirected, err := service.HandleVideoRedirects(client, videoURL)
	if err != nil && !redirected {
		log.Errorf("Error handling redirects: %v", err)
	} else if redirected {
		log.Infof("URL redirected from %s to %s", videoURL, finalURL)
		videoURL = finalURL
	}

	// 2. 尝试小红书特殊处理
	alternateURL, err := service.HandleXiaohongshuVideo(client, videoURL)
	if err == nil && alternateURL != videoURL {
		videoURL = alternateURL
	}

	// 获取针对该平台的特定配置
	hostConfig := service.GetVideoHostConfig(service.ExtractHost(videoURL))

	// 准备请求
	req := client.R().
		SetHeader(service.HttpHeaderUserAgent, service.GetUserAgent(hostConfig.UserAgentType))

	// 添加通用请求头
	for key, value := range hostConfig.ExtraHeaders {
		req.SetHeader(key, value)
	}

	// 添加Referer
	if hostConfig.RefererURL != "" {
		req.SetHeader(service.HttpHeaderReferer, hostConfig.RefererURL)
	}

	// 直接复用微信小程序的User-Agent
	req.SetHeader(service.HttpHeaderUserAgent, ctx.Get(service.HttpHeaderUserAgent))
	req.SetHeader("Referer", ctx.Get("Referer"))

	// 检查URL是否可访问
	resp, err := req.Head(videoURL)

	var videoStatus string
	if err != nil || (resp != nil && resp.StatusCode() != http.StatusOK) {
		videoStatus = "不可直接访问"
	} else {
		videoStatus = "可直接访问"
	}

	// 返回信息
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "获取小红书视频信息成功",
		"data": map[string]interface{}{
			"original_url":   ctx.Query("url"),
			"final_url":      videoURL,
			"video_status":   videoStatus,
			"download_url":   fmt.Sprintf("/api/tools/download?url=%s", url.QueryEscape(videoURL)),
			"proxy_url":      fmt.Sprintf("/api/tools/proxy?url=%s", url.QueryEscape(videoURL)),
			"is_xiaohongshu": true,
			"alternate_domains": []string{
				strings.Replace(videoURL, "sns-video-bd.xhscdn.com", "sns-video.xhscdn.com", 1),
				strings.Replace(videoURL, "sns-video-bd.xhscdn.com", "sns-video-qc.xhscdn.com", 1),
				strings.Replace(videoURL, "sns-video-bd.xhscdn.com", "sns-video-hw.xhscdn.com", 1),
			},
		},
	})
}

func NewCommonHandler(router fiber.Router, repository *repositories.ToolRepository, redis *redis.Client, cos *cos.Client, config *config.EnvConfig) {
	handler := &CommonHandler{
		redis:      redis,
		cos:        cos,
		repository: repository,
		config:     config,
	}
	commonRouter := router.Group("/tools")
	commonRouter.Post("/parse", handler.ParseShareUrl)
	commonRouter.Get("/list", handler.GetTools)
	commonRouter.Post("/", handler.CreateTool)
	commonRouter.Post("/file/upload", handler.Upload)
	commonRouter.Get("/download", handler.DownloadFile)
	// 添加视频代理路由
	commonRouter.Get("/proxy", handler.ProxyVideo)
	// 添加微信小程序下载配置路由
	commonRouter.Get("/wechat-download-config", handler.GetWeChatDownloadConfig)
	// 添加小红书视频信息路由
	commonRouter.Get("/xiaohongshu/video-info", handler.GetXiaohongshuVideoInfo)
}
