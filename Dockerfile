FROM golang:alpine3.21

WORKDIR /src/app

# 设置中国镜像源
ENV GOPROXY=https://goproxy.cn,direct

RUN apk add --no-cache git

RUN go install github.com/cosmtrek/air@v1.52.0

COPY . .

RUN go mod tidy