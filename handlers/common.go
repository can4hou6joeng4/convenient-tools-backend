package handlers

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/can4hou6joeng4/convenient-tools-project-v1-backend/config"
	"github.com/can4hou6joeng4/convenient-tools-project-v1-backend/models"
	"github.com/can4hou6joeng4/convenient-tools-project-v1-backend/repositories"
	"github.com/can4hou6joeng4/convenient-tools-project-v1-backend/service"
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

// ProxyMedia 代理媒体文件
// @Summary 代理媒体文件
// @Description 从远程服务器获取媒体文件并转发给客户端，解决小程序域名限制问题
// @Tags tools
// @Accept json
// @Produce octet-stream
// @Param url query string true "媒体文件URL"
// @Param type query string false "媒体类型(video/image)"
// @Success 200 {file} binary "媒体文件"
// @Failure 400 {object} map[string]interface{}
// @Failure 405 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /tools/media-proxy [get]
func (h *CommonHandler) ProxyMedia(ctx *fiber.Ctx) error {
	// 只允许GET请求
	if ctx.Method() != "GET" {
		return ctx.Status(fiber.StatusMethodNotAllowed).JSON(fiber.Map{
			"error": "Method Not Allowed",
		})
	}

	// 获取请求参数
	mediaURL := ctx.Query("url")
	mediaType := ctx.Query("type", "video") // 默认为视频

	// 验证URL参数
	if mediaURL == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing url parameter",
		})
	}

	// 验证URL格式和协议
	if !strings.HasPrefix(mediaURL, "http://") && !strings.HasPrefix(mediaURL, "https://") {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid URL protocol",
		})
	}

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// 允许最多10次重定向
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	// 创建请求
	req, err := http.NewRequest("GET", mediaURL, nil)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create request: %v", err),
		})
	}

	// 添加用户代理头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to fetch media: %v", err),
		})
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Remote server returned status code %d", resp.StatusCode),
		})
	}

	// 设置Content-Type
	contentType := getMediaContentType(mediaType, mediaURL, resp.Header.Get("Content-Type"))
	ctx.Set("Content-Type", contentType)

	// 设置其他响应头
	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		ctx.Set("Content-Length", contentLength)
	}
	ctx.Set("Access-Control-Allow-Origin", "*")
	ctx.Set("Cache-Control", "public, max-age=86400") // 缓存24小时

	// 将媒体文件流式传输给客户端
	_, err = io.Copy(ctx.Response().BodyWriter(), resp.Body)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to stream media: %v", err),
		})
	}

	return nil
}

// getMediaContentType 获取媒体文件的Content-Type
func getMediaContentType(mediaType, url, originalContentType string) string {
	// 如果原始服务器提供了Content-Type，优先使用
	if originalContentType != "" && originalContentType != "application/octet-stream" {
		return originalContentType
	}

	// 将URL转为小写以便于扩展名匹配
	urlLower := strings.ToLower(url)
	extension := strings.ToLower(filepath.Ext(urlLower))

	// 根据媒体类型和文件扩展名确定Content-Type
	if mediaType == "image" {
		switch extension {
		case ".jpg", ".jpeg":
			return "image/jpeg"
		case ".png":
			return "image/png"
		case ".gif":
			return "image/gif"
		case ".webp":
			return "image/webp"
		case ".svg":
			return "image/svg+xml"
		default:
			return "image/jpeg" // 默认图片类型
		}
	} else if mediaType == "video" {
		switch extension {
		case ".mp4":
			return "video/mp4"
		case ".webm":
			return "video/webm"
		case ".ogg", ".ogv":
			return "video/ogg"
		case ".mov":
			return "video/quicktime"
		default:
			return "video/mp4" // 默认视频类型
		}
	}

	// 默认二进制流
	return "application/octet-stream"
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
	commonRouter.Get("/media-proxy", handler.ProxyMedia)
}
