FROM golang:alpine3.21

WORKDIR /src/app

# 设置中国镜像源
ENV GOPROXY=https://goproxy.cn,direct

# 安装 FFmpeg 及必要依赖
RUN apk add --no-cache \
    ffmpeg \
    libavcodec \
    libavformat \
    libavutil \
    libswscale \
    libavfilter \
    libswresample

RUN go install github.com/cosmtrek/air@v1.52.0

# 复制项目文件
COPY . .

# 下载依赖
RUN go mod tidy
