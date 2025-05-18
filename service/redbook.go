package service

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/can4hou6joeng4/convenient-tools-project-v1-backend/models"
	"github.com/tidwall/gjson"

	"github.com/go-resty/resty/v2"
)

type RedBook struct{}

func (r RedBook) ParseShareUrl(shareUrl string) (*models.VideoParseInfo, error) {
	client := resty.New()
	videoRes, err := client.R().
		SetHeader(HttpHeaderUserAgent, "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 Edg/129.0.0.0").
		Get(shareUrl)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`window.__INITIAL_STATE__\s*=\s*(.*?)</script>`)
	findRes := re.FindSubmatch(videoRes.Body())
	if len(findRes) < 2 {
		return nil, errors.New("parse video json info from html fail")
	}

	jsonBytes := bytes.TrimSpace(findRes[1])

	nodeId := gjson.GetBytes(jsonBytes, "note.currentNoteId").String()
	data := gjson.GetBytes(jsonBytes, fmt.Sprintf("note.noteDetailMap.%s.note", nodeId))

	videoUrl := data.Get("video.media.stream.h264.0.masterUrl").String()

	// 获取图集图片地址
	imagesObjArr := data.Get("imageList").Array()
	images := make([]string, 0, len(imagesObjArr))
	if len(videoUrl) <= 0 {
		for _, imageItem := range imagesObjArr {
			// 检查是否为livePhoto
			isLivePhoto := imageItem.Get("livePhoto").Bool()

			if isLivePhoto {
				// 如果是livePhoto，获取视频URL
				livePhotoUrl := imageItem.Get("stream.h264.0.masterUrl").String()
				if len(livePhotoUrl) > 0 {
					images = append(images, livePhotoUrl)
				} else {
					// 尝试从全局视频数据获取
					livePhotoUrl = data.Get("video.media.stream.h264.0.masterUrl").String()
					if len(livePhotoUrl) > 0 {
						images = append(images, livePhotoUrl)
					}
				}
			} else {
				// 处理普通图片
				// 首先尝试使用原始URL
				imageUrl := imageItem.Get("urlDefault").String()
				if len(imageUrl) > 0 {
					// 保留原始URL作为主URL
					images = append(images, imageUrl)

					// 同时提取图片ID并创建备用URL
					var imgId string
					if strings.Contains(imageUrl, "!") {
						// 处理形如 xxx/yyy!format 的URL
						imgId = strings.Split(imageUrl[strings.LastIndex(imageUrl, "/")+1:], "!")[0]
					} else {
						// 处理其他格式URL
						imgId = imageUrl[strings.LastIndex(imageUrl, "/")+1:]
					}

					// 如果链接中带有 spectrum/ , 替换域名时需要带上
					spectrumStr := ""
					if strings.Contains(imageUrl, "spectrum") {
						spectrumStr = "spectrum/"
					}

					// 创建备用URL
					backupUrl := fmt.Sprintf("https://ci.xiaohongshu.com/%s%s?imageView2/2/w/0/format/jpg", spectrumStr, imgId)
					fmt.Println("Original URL:", imageUrl)
					fmt.Println("Backup URL:", backupUrl)

					// 添加备用URL到images数组
					// 注意：这里我们不再将备用URL加入images数组，因为前端可能不支持多URL
					// 但我们保留它作为日志输出以便调试
				}
			}
		}
	}

	parseInfo := &models.VideoParseInfo{
		Title:    data.Get("title").String(),
		VideoUrl: data.Get("video.media.stream.h264.0.masterUrl").String(),
		CoverUrl: data.Get("imageList.0.urlDefault").String(),
		Images:   images,
	}
	parseInfo.Author.Uid = data.Get("user.userId").String()
	parseInfo.Author.Name = data.Get("user.nickname").String()
	parseInfo.Author.Avatar = data.Get("user.avatar").String()

	return parseInfo, nil
}
