package service

import (
	"mime"
	"net/url"
	"path/filepath"
	"strings"

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
				"sec-ch-ua":                 `"Not/A)Brand";v="99", "Google Chrome";v="115", "Chromium";v="115"`,
				"sec-ch-ua-mobile":          "?0",
				"sec-ch-ua-platform":        `"iOS"`,
				"sec-fetch-dest":            "document",
				"sec-fetch-mode":            "navigate",
				"sec-fetch-site":            "none",
				"sec-fetch-user":            "?1",
				"upgrade-insecure-requests": "1",
				"X-Forwarded-For":           "223.104.41.25", // 中国IP，减少地域限制
				"cookie":                    "xsecappid=xhs-pc-web;a1=187016c31dflbva806wl9ygzq62zk16q44z8yc73v50000327781;webId=7de93376af2a82d7f57ca338cef4bb73;gid=yYq288KSdqyYyYq288KS0ydWKhfiy4DhVFd2Y9qh2k2xFj98F6h299844qYKYiq8dy8jvKSS;web_session=030037a2b4dfc8fc8bbcaf9f1d9879e6c8faa3",
			},
		},
		"xiaohongshu.com": {
			RefererURL:    "https://www.xiaohongshu.com/",
			OriginURL:     "https://www.xiaohongshu.com",
			UserAgentType: "mobile",
			ExtraHeaders: map[string]string{
				"sec-ch-ua":                 `"Not/A)Brand";v="99", "Google Chrome";v="115", "Chromium";v="115"`,
				"sec-ch-ua-mobile":          "?0",
				"sec-ch-ua-platform":        `"iOS"`,
				"sec-fetch-dest":            "document",
				"sec-fetch-mode":            "navigate",
				"sec-fetch-site":            "none",
				"sec-fetch-user":            "?1",
				"upgrade-insecure-requests": "1",
				"X-Forwarded-For":           "223.104.41.25", // 中国IP，减少地域限制
				"cookie":                    "xsecappid=xhs-pc-web;a1=187016c31dflbva806wl9ygzq62zk16q44z8yc73v50000327781;webId=7de93376af2a82d7f57ca338cef4bb73;gid=yYq288KSdqyYyYq288KS0ydWKhfiy4DhVFd2Y9qh2k2xFj98F6h299844qYKYiq8dy8jvKSS;web_session=030037a2b4dfc8fc8bbcaf9f1d9879e6c8faa3",
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
		return MobileUserAgent
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
