FROM golang:alpine AS base 

ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64
#环境变量设置
WORKDIR /app
COPY . .
#设置工作目录和复制文件

RUN go mod download
#下载依赖
RUN go build -o ./cmd/user/main ./app/user/cmd/main.go
#构建GO程序，将生成的二进制文件输出到目录

FROM alpine AS final
RUN set -eux && sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk update && apk add tzdata
COPY --from=base /app/config/ /config/
COPY --from=base /app/cmd/user/main .
EXPOSE 10002
CMD ["./main"]
#设置时区，复制配置文件，复制二进制文件，暴露端口，启动程序
