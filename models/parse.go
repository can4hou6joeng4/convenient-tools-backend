package models

// videoShareUrlParser 根据视频分享地址解析
type VideoShareUrlParser interface {
	ParseShareUrl(shareUrl string) (*VideoParseInfo, error)
}

// videoIdParser 根据视频ID解析
type VideoIdParser interface {
	ParseVideoID(videoId string) (*VideoParseInfo, error)
}

// VideoParseInfo 视频解析信息
type VideoParseInfo struct {
	Author struct {
		Uid    string `json:"uid"`    // 作者id
		Name   string `json:"name"`   // 作者名称
		Avatar string `json:"avatar"` // 作者头像
	} `json:"author"`
	Title    string   `json:"title"`     // 描述
	VideoUrl string   `json:"video_url"` // 视频播放地址
	MusicUrl string   `json:"music_url"` // 音乐播放地址
	CoverUrl string   `json:"cover_url"` // 视频封面地址
	Images   []string `json:"images"`    // 图集图片地址列表
}

// BatchParseItem 批量解析时, 单条解析格式
type BatchParseItem struct {
	ParseInfo *VideoParseInfo // 视频解析信息
	Error     error           // 错误, 如果单条解析失败时, 记录error信息
}

// 视频渠道信息
type VideoSourceInfo struct {
	VideoShareUrlDomain []string            // 视频分享地址域名
	VideoShareUrlParser VideoShareUrlParser // 视频分享地址解析方法
	VideoIdParser       VideoIdParser       // 视频id解析方法, 有些渠道可能没有id解析方法
}
