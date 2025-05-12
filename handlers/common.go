package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
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
// @Description 从远程服务器获取媒体文件并转发给客户端，解决小程序域名限制问题，支持视频格式转换
// @Tags tools
// @Accept json
// @Produce octet-stream
// @Param url query string true "媒体文件URL"
// @Param type query string false "媒体类型(video/image)"
// @Param format query string false "输出格式(mp4/webm/mov)"
// @Success 200 {file} binary "媒体文件"
// @Failure 400 {object} map[string]interface{}
// @Failure 405 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /tools/media-proxy [get]
func (h *CommonHandler) ProxyMedia(ctx *fiber.Ctx) error {
	// 获取请求信息
	mediaURL := ctx.Query("url")
	mediaType := ctx.Query("type", "video")
	format := ctx.Query("format", "")
	userAgent := ctx.Get("User-Agent")

	// 记录请求信息
	log.Infof("媒体代理请求: URL=%s, 类型=%s, 格式=%s, UA=%s", mediaURL, mediaType, format, userAgent)

	// 只允许GET请求
	if ctx.Method() != "GET" {
		return ctx.Status(fiber.StatusMethodNotAllowed).JSON(fiber.Map{
			"error": "Method Not Allowed",
		})
	}

	// 验证URL参数
	if mediaURL == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "缺少url参数",
		})
	}

	// 验证URL格式和协议
	if !strings.HasPrefix(mediaURL, "http://") && !strings.HasPrefix(mediaURL, "https://") {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "无效的URL协议",
		})
	}

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("重定向次数过多")
			}
			return nil
		},
	}

	// 根据媒体类型处理请求
	switch mediaType {
	case "video":
		return h.handleVideoProxy(ctx, client, mediaURL, format)
	case "image":
		return h.handleImageProxy(ctx, client, mediaURL)
	default:
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "不支持的媒体类型",
		})
	}
}

// handleVideoProxy 处理视频代理请求
func (h *CommonHandler) handleVideoProxy(ctx *fiber.Ctx, client *http.Client, url string, format string) error {
	// 创建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("创建请求失败: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("创建请求失败: %v", err),
		})
	}

	// 添加用户代理头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("获取视频失败: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("获取视频失败: %v", err),
		})
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		log.Errorf("源服务器响应错误: %s", resp.Status)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("源服务器响应错误: %s", resp.Status),
		})
	}

	// 读取响应内容
	videoData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("读取视频数据失败: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("读取视频数据失败: %v", err),
		})
	}

	// 检查视频数据的有效性
	if len(videoData) < 1024 {
		log.Errorf("视频数据无效或太小: %d bytes", len(videoData))
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "视频数据无效或太小",
		})
	}

	// 检测视频格式，如果需要且不是MP4，则转换为MP4
	needConversion := format == "mp4" || (format == "" && !isMP4(videoData, resp.Header.Get("Content-Type")))

	if needConversion {
		// 使用FFmpeg进行格式转换
		convertedData, err := convertToMP4(videoData)
		if err != nil {
			log.Errorf("视频格式转换失败: %v", err)
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("视频格式转换失败: %v", err),
			})
		}
		videoData = convertedData

		// 设置正确的Content-Type和文件扩展名
		ctx.Set("Content-Type", "video/mp4")
		ctx.Set("Content-Disposition", `attachment; filename="video.mp4"`)
	} else {
		// 保持原始格式
		ctx.Set("Content-Type", resp.Header.Get("Content-Type"))
	}

	// 设置其他响应头
	ctx.Set("X-Content-Type-Options", "nosniff")
	ctx.Set("Accept-Ranges", "bytes")
	ctx.Set("Content-Length", fmt.Sprintf("%d", len(videoData)))
	ctx.Set("Access-Control-Allow-Origin", "*")
	ctx.Set("Cache-Control", "public, max-age=3600") // 缓存1小时

	// 返回视频数据
	return ctx.Send(videoData)
}

// handleImageProxy 处理图片代理请求
func (h *CommonHandler) handleImageProxy(ctx *fiber.Ctx, client *http.Client, url string) error {
	// 创建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("创建请求失败: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("创建请求失败: %v", err),
		})
	}

	// 添加用户代理头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("获取图片失败: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("获取图片失败: %v", err),
		})
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		log.Errorf("源服务器响应错误: %s", resp.Status)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("源服务器响应错误: %s", resp.Status),
		})
	}

	// 设置响应头
	ctx.Set("Content-Type", resp.Header.Get("Content-Type"))
	ctx.Set("X-Content-Type-Options", "nosniff")
	ctx.Set("Access-Control-Allow-Origin", "*")
	ctx.Set("Cache-Control", "public, max-age=3600") // 缓存1小时

	// 将图片流式传输给客户端
	_, err = io.Copy(ctx.Response().BodyWriter(), resp.Body)
	if err != nil {
		log.Errorf("传输图片数据失败: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("传输图片数据失败: %v", err),
		})
	}

	return nil
}

// isMP4 检查视频是否为MP4格式
func isMP4(data []byte, contentType string) bool {
	// 检查内容类型
	if strings.Contains(strings.ToLower(contentType), "mp4") {
		return true
	}

	// 检查文件头部标记
	if len(data) > 4 {
		// MP4文件的标记通常是 ftyp
		return bytes.Contains(data[:50], []byte("ftyp"))
	}

	return false
}

// convertToMP4 将视频转换为MP4格式
func convertToMP4(videoData []byte) ([]byte, error) {
	// 创建临时输入文件
	tempInFile, err := os.CreateTemp("", "video-in-*")
	if err != nil {
		return nil, fmt.Errorf("创建临时输入文件失败: %v", err)
	}
	defer os.Remove(tempInFile.Name())

	// 写入原始视频数据
	if _, err = tempInFile.Write(videoData); err != nil {
		return nil, fmt.Errorf("写入临时文件失败: %v", err)
	}
	tempInFile.Close()

	// 创建临时输出文件
	tempOutFile, err := os.CreateTemp("", "video-out-*.mp4")
	if err != nil {
		return nil, fmt.Errorf("创建临时输出文件失败: %v", err)
	}
	defer os.Remove(tempOutFile.Name())
	tempOutFile.Close()

	// 使用FFmpeg进行转换
	cmd := exec.Command("ffmpeg",
		"-i", tempInFile.Name(),
		"-c:v", "libx264", // 使用H.264编码
		"-preset", "fast", // 快速编码
		"-c:a", "aac", // 音频使用AAC编码
		"-y", // 覆盖输出文件
		tempOutFile.Name())

	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("FFmpeg转换失败: %v, 输出: %s", err, string(output))
	}

	// 读取转换后的文件
	return os.ReadFile(tempOutFile.Name())
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
