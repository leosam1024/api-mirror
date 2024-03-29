FROM golang:alpine as builder

RUN apk add --no-cache make
WORKDIR /api-mirror
COPY . /api-mirror
ENV GOPROXY=https://goproxy.io,direct
RUN go mod download && go build

FROM alpine:latest
LABEL anme="api-mirror"
# 如果在环境变量里定义了端口号，则用环境变量中的
# ENV MIRROR_PORT=8899
# 如果在环境变量里定义了配置文件路径且未运行参数指定，则用环境变量中的
# ENV MIRROR_CONFIG_FILE=8899
EXPOSE 8899
COPY --from=builder /api-mirror /
ENTRYPOINT ["/api-mirror"]
