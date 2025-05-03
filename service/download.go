package service

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

// GetFileType 根据Content-Type和文件名判断文件类型
func GetFileType(contentType string, filename string) string {
	// 优先使用Content-Type判断
	if contentType != "" {
		contentType = strings.ToLower(contentType)
		if strings.HasPrefix(contentType, "image/") {
			return FileTypeImage
		}
		if strings.HasPrefix(contentType, "video/") {
			return FileTypeVideo
		}
		if strings.HasPrefix(contentType, "audio/") {
			return FileTypeAudio
		}
	}

	// 使用文件扩展名判断
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg":
		return FileTypeImage
	case ".mp4", ".webm", ".ogg", ".mov", ".avi", ".flv", ".mkv":
		return FileTypeVideo
	case ".mp3", ".wav", ".flac", ".aac", ".m4a":
		return FileTypeAudio
	default:
		// 尝试使用MIME类型匹配
		if ext != "" {
			mimeType := mime.TypeByExtension(ext)
			if strings.HasPrefix(mimeType, "image/") {
				return FileTypeImage
			}
			if strings.HasPrefix(mimeType, "video/") {
				return FileTypeVideo
			}
			if strings.HasPrefix(mimeType, "audio/") {
				return FileTypeAudio
			}
		}
		return FileTypeOther
	}
}

// FormatFilename 格式化下载文件名，确保文件名合法
func FormatFilename(filename string) string {
	// 移除不安全的文件名字符
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	filename = strings.ReplaceAll(filename, ":", "_")
	filename = strings.ReplaceAll(filename, "*", "_")
	filename = strings.ReplaceAll(filename, "?", "_")
	filename = strings.ReplaceAll(filename, "\"", "_")
	filename = strings.ReplaceAll(filename, "<", "_")
	filename = strings.ReplaceAll(filename, ">", "_")
	filename = strings.ReplaceAll(filename, "|", "_")

	// 如果文件名为空，使用默认名称
	if filename == "" {
		return "download"
	}

	return filename
}

// WeChatMiniProgramDownloadHelper 微信小程序下载辅助函数
type WeChatMiniProgramDownloadHelper struct{}

// GetDownloadPermissionInfo 获取微信小程序下载所需权限信息
func (w *WeChatMiniProgramDownloadHelper) GetDownloadPermissionInfo(fileType string) map[string]interface{} {
	result := map[string]interface{}{
		"needPermission": false,
		"permissionType": "",
		"apiName":        "",
	}

	switch fileType {
	case FileTypeImage:
		result["needPermission"] = true
		result["permissionType"] = "scope.writePhotosAlbum"
		result["apiName"] = "wx.saveImageToPhotosAlbum"
	case FileTypeVideo:
		result["needPermission"] = true
		result["permissionType"] = "scope.writePhotosAlbum"
		result["apiName"] = "wx.saveVideoToPhotosAlbum"
	}

	return result
}

// GenerateWeChatDownloadConfig 生成微信小程序下载配置
func GenerateWeChatDownloadConfig(url string, filename string, contentType string) map[string]interface{} {
	// 格式化文件名
	safeFilename := FormatFilename(filename)

	// 确定文件类型
	fileType := GetFileType(contentType, safeFilename)

	// 创建辅助器
	helper := WeChatMiniProgramDownloadHelper{}

	// 获取权限信息
	permissionInfo := helper.GetDownloadPermissionInfo(fileType)

	// 返回配置信息
	return map[string]interface{}{
		"url":            url,
		"filename":       safeFilename,
		"fileType":       fileType,
		"contentType":    contentType,
		"permissionInfo": permissionInfo,
		"needPreview":    fileType == FileTypeImage && !strings.Contains(contentType, "svg"),
		"canDirectSave":  fileType == FileTypeImage || fileType == FileTypeVideo,
		"needCustomSave": fileType != FileTypeImage && fileType != FileTypeVideo,
	}
}

// VideoHostConfig 视频平台配置信息
type VideoHostConfig struct {
	RefererURL    string            // 引用地址
	OriginURL     string            // 源地址
	ExtraHeaders  map[string]string // 额外的请求头
	UserAgentType string            // 用户代理类型：mobile或desktop
}

// GetVideoHostConfigs 获取所有支持的视频平台配置
func GetVideoHostConfigs() map[string]VideoHostConfig {
	return map[string]VideoHostConfig{
		// 小红书
		"xhscdn.com": {
			RefererURL:    "https://www.xiaohongshu.com/",
			OriginURL:     "https://www.xiaohongshu.com",
			UserAgentType: "mobile",
			ExtraHeaders: map[string]string{
				"sec-ch-ua":          `"Not A(Brand";v="99", "Microsoft Edge";v="121", "Chromium";v="121"`,
				"sec-ch-ua-mobile":   "?1",
				"sec-ch-ua-platform": `"Android"`,
				"sec-fetch-dest":     "video",
				"sec-fetch-mode":     "no-cors",
				"sec-fetch-site":     "cross-site",
				"sec-fetch-user":     "?1",
				"pragma":             "no-cache",
				"cache-control":      "no-cache",
				"X-Forwarded-For":    "223.104.41.25", // 中国IP，减少地域限制
				"cookie":             "xsecappid=xhs-pc-web; webId=681bec21e8314215e1583c6dd3071b10; gid=yYqxYyK8yDyjyYqxYyK8Sx8xJKSFdf48qCfAkF87q82y28082WxK8K8488q8d2qyYKJi8S; timestamp2=1680c7d0b3f36e96c2e3f7a2d; timestamp2.sig=nElj5VX3cHwUi3DlsR-eNEtdcZWx-FTLOts6u2zXkTg",
			},
		},
		"xiaohongshu.com": {
			RefererURL:    "https://www.xiaohongshu.com/",
			OriginURL:     "https://www.xiaohongshu.com",
			UserAgentType: "mobile",
			ExtraHeaders: map[string]string{
				"sec-ch-ua":          `"Not A(Brand";v="99", "Microsoft Edge";v="121", "Chromium";v="121"`,
				"sec-ch-ua-mobile":   "?1",
				"sec-ch-ua-platform": `"Android"`,
				"sec-fetch-dest":     "document",
				"sec-fetch-mode":     "navigate",
				"sec-fetch-site":     "none",
				"sec-fetch-user":     "?1",
				"pragma":             "no-cache",
				"cache-control":      "no-cache",
				"X-Forwarded-For":    "223.104.41.25", // 中国IP，减少地域限制
				"cookie":             "xsecappid=xhs-pc-web; webId=681bec21e8314215e1583c6dd3071b10; gid=yYqxYyK8yDyjyYqxYyK8Sx8xJKSFdf48qCfAkF87q82y28082WxK8K8488q8d2qyYKJi8S; timestamp2=1680c7d0b3f36e96c2e3f7a2d; timestamp2.sig=nElj5VX3cHwUi3DlsR-eNEtdcZWx-FTLOts6u2zXkTg",
			},
		},
		// 抖音
		"douyin.com": {
			RefererURL:    "https://www.douyin.com/",
			OriginURL:     "https://www.douyin.com",
			UserAgentType: "mobile",
			ExtraHeaders: map[string]string{
				"sec-fetch-dest":  "empty",
				"sec-fetch-mode":  "cors",
				"sec-fetch-site":  "same-site",
				"X-Forwarded-For": "14.30.49.112", // 中国IP，减少地域限制
			},
		},
		"douyinvod.com": {
			RefererURL:    "https://www.douyin.com/",
			OriginURL:     "https://www.douyin.com",
			UserAgentType: "mobile",
			ExtraHeaders: map[string]string{
				"X-Forwarded-For": "14.30.49.112", // 中国IP，减少地域限制
			},
		},
		// 针对其他平台可以继续添加...

		// 默认配置，用于未特别处理的平台
		"default": {
			UserAgentType: "mobile",
			ExtraHeaders: map[string]string{
				"Accept":          "*/*",
				"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
				"Connection":      "keep-alive",
			},
		},
	}
}

// GetVideoHostConfig 根据主机名获取视频平台配置
func GetVideoHostConfig(host string) VideoHostConfig {
	configs := GetVideoHostConfigs()

	// 检查是否有完全匹配的配置
	if config, exists := configs[host]; exists {
		return config
	}

	// 检查是否有部分匹配的配置
	for key, config := range configs {
		if key != "default" && strings.Contains(host, key) {
			return config
		}
	}

	// 返回默认配置
	return configs["default"]
}

// GetUserAgent 根据配置类型返回对应的User-Agent
func GetUserAgent(userAgentType string) string {
	if userAgentType == "mobile" {
		// 使用微信小程序User-Agent，更容易被接受
		return "Mozilla/5.0 (iPhone; CPU iPhone OS 16_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.6 Mobile/15E148 Safari/604.1 Edg/122.0.0.0"
	}
	return DefaultUserAgent
}

// HandleVideoRedirects 处理视频URL的重定向
// 返回: 最终URL, 是否需要继续处理重定向, 错误
func HandleVideoRedirects(client *resty.Client, url string) (string, bool, error) {
	// 不跟随重定向，以便我们可以从302响应中获取真实URL
	client.SetRedirectPolicy(resty.NoRedirectPolicy())

	// 设置用户代理
	host := ExtractHost(url)
	config := GetVideoHostConfig(host)

	// 发送HEAD请求获取重定向信息
	resp, err := client.R().
		SetHeader(HttpHeaderUserAgent, GetUserAgent(config.UserAgentType)).
		Head(url)

	// 重置客户端以便后续请求使用
	client.SetRedirectPolicy(resty.FlexibleRedirectPolicy(10))

	// 检查响应
	if err != nil {
		// 如果是重定向错误，我们可以从中获取目标URL
		if resp != nil && (resp.StatusCode() == 301 || resp.StatusCode() == 302 || resp.StatusCode() == 307 || resp.StatusCode() == 308) {
			location := resp.Header().Get("Location")
			if location != "" {
				return location, true, nil
			}
		}
		return url, false, err
	}

	// 没有重定向
	return url, false, nil
}

// ExtractHost 从URL中提取主机名
func ExtractHost(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	return parsedURL.Host
}

// 新增函数: 对小红书特殊处理
// IsXiaohongshuHost 检查是否是小红书相关域名
func IsXiaohongshuHost(host string) bool {
	return strings.Contains(host, "xhscdn.com") || strings.Contains(host, "xiaohongshu.com")
}

// HandleXiaohongshuVideo 对小红书视频进行特殊处理
func HandleXiaohongshuVideo(client *resty.Client, url string) (string, error) {
	// 检查URL格式
	if !strings.Contains(url, "xhscdn.com") && !strings.Contains(url, "xiaohongshu.com") {
		return url, nil
	}

	// 如果已经是替代域名，不需要再处理
	if strings.Contains(url, "sns-video-hw.xhscdn.com") ||
		strings.Contains(url, "sns-video-qc.xhscdn.com") ||
		strings.Contains(url, "sns-video.xhscdn.com") {
		return url, nil
	}

	// 尝试直接替换域名访问原始视频
	if strings.Contains(url, "sns-video-bd.xhscdn.com") {
		// 替换为备用域名尝试
		alternateUrls := []string{
			strings.Replace(url, "sns-video-bd.xhscdn.com", "sns-video-hw.xhscdn.com", 1),
			strings.Replace(url, "sns-video-bd.xhscdn.com", "sns-video-qc.xhscdn.com", 1),
			strings.Replace(url, "sns-video-bd.xhscdn.com", "sns-video.xhscdn.com", 1),
		}

		// 测试备用URL是否有效
		for _, altUrl := range alternateUrls {
			// 使用一个带有超时的上下文
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// 准备请求
			req, err := http.NewRequestWithContext(ctx, "HEAD", altUrl, nil)
			if err != nil {
				continue
			}

			// 添加关键请求头
			req.Header.Set(HttpHeaderUserAgent, MobileUserAgent)
			req.Header.Set(HttpHeaderReferer, "https://www.xiaohongshu.com/")
			req.Header.Set("Accept", "*/*")
			req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
			req.Header.Set("X-Forwarded-For", "223.104.41.25")

			// 执行请求
			httpClient := &http.Client{
				Timeout: 5 * time.Second,
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					// 不自动跟随重定向
					return http.ErrUseLastResponse
				},
			}

			resp, err := httpClient.Do(req)
			if err == nil {
				defer resp.Body.Close()

				// 检查状态码
				if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusPartialContent {
					return altUrl, nil
				} else if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
					// 处理重定向
					location := resp.Header.Get("Location")
					if location != "" {
						return location, nil
					}
				}
			}
		}

		// 如果所有尝试都失败，返回原始URL但使用hw域名
		return strings.Replace(url, "sns-video-bd.xhscdn.com", "sns-video-hw.xhscdn.com", 1), nil
	}

	return url, nil
}

// 新增：流式代理下载小红书视频
func StreamXiaohongshuVideo(url string, w http.ResponseWriter, r *http.Request) error {
	// 替换为可能有效的域名
	if strings.Contains(url, "sns-video-bd.xhscdn.com") {
		url = strings.Replace(url, "sns-video-bd.xhscdn.com", "sns-video-hw.xhscdn.com", 1)
	}

	// 创建客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 准备请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// 设置请求头
	req.Header.Set(HttpHeaderUserAgent, MobileUserAgent)
	req.Header.Set(HttpHeaderReferer, "https://www.xiaohongshu.com/")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("X-Forwarded-For", "223.104.41.25")
	req.Header.Set("Origin", "https://www.xiaohongshu.com")

	// 添加微信环境的Range头(如果存在)
	if r.Header.Get("Range") != "" {
		req.Header.Set("Range", r.Header.Get("Range"))
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("视频源返回状态码: %d", resp.StatusCode)
	}

	// 设置响应头
	for k, v := range resp.Header {
		if k != "Server" && k != "Date" && k != "Connection" {
			w.Header().Set(k, v[0])
		}
	}

	// 必要的跨域和缓存控制头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Range, Origin, Content-Type, Accept")
	w.Header().Set("Cache-Control", "public, max-age=86400")

	// 设置状态码
	w.WriteHeader(resp.StatusCode)

	// 拷贝内容
	_, err = io.Copy(w, resp.Body)
	return err
}
