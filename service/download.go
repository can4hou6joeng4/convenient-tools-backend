package service

import (
	"mime"
	"path/filepath"
	"strings"
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
