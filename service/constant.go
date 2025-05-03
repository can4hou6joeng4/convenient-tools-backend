package service

import "github.com/can4hou6joeng4/convenient-tools-project-v1-backend/models"

// 常量定义，避免循环导入
const (
	HttpHeaderUserAgent   = "User-Agent"
	HttpHeaderReferer     = "Referer"
	HttpHeaderContentType = "Content-Type"
	HttpHeaderCookie      = "Cookie"
	DefaultUserAgent      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	MobileUserAgent       = "Mozilla/5.0 (iPhone; CPU iPhone OS 16_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.6 Mobile/15E148 Safari/604.1 Edg/122.0.0.0"
)

// 文件类型常量
const (
	FileTypeImage = "image" // 图片文件
	FileTypeVideo = "video" // 视频文件
	FileTypeAudio = "audio" // 音频文件
	FileTypeOther = "other" // 其他文件
)

// 视频渠道来源
const (
	SourceDouYin   = "douyin"   // 抖音
	SourceKuaiShou = "kuaishou" // 快手
	SourceWeiBo    = "weibo"    // 微博
	SourceXiGua    = "xigua"    // 西瓜
	SourceRedBook  = "redbook"  // 小红书
)

// 视频渠道映射信息
var VideoSourceInfoMapping = map[string]models.VideoSourceInfo{
	SourceDouYin: {
		VideoShareUrlDomain: []string{"v.douyin.com", "www.iesdouyin.com", "www.douyin.com"},
		VideoShareUrlParser: Douyin{},
		VideoIdParser:       Douyin{},
	},
	SourceXiGua: {
		VideoShareUrlDomain: []string{"v.ixigua.com"},
		VideoShareUrlParser: XiGua{},
		VideoIdParser:       XiGua{},
	},
	SourceRedBook: {
		VideoShareUrlDomain: []string{
			"www.xiaohongshu.com",
			"xhslink.com",
		},
		VideoShareUrlParser: RedBook{},
	},
	/* 暂时注释掉未实现的解析器，避免编译错误
	SourceKuaiShou: {
		VideoShareUrlDomain: []string{"v.kuaishou.com"},
		VideoShareUrlParser: kuaiShou{},
	},

	SourceWeiBo: {
		VideoShareUrlDomain: []string{"weibo.com"},
		VideoShareUrlParser: weiBo{},
		VideoIdParser:       weiBo{},
	},

	*/
}
