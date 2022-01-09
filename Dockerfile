FROM golang:alpine as builder

RUN apk add --no-cache make
WORKDIR /api-mirror
COPY . /api-mirror
ENV GOPROXY=https://goproxy.io,direct
RUN go mod download && go build

FROM alpine:latest
LABEL anme="api-mirror"
ENV MIRRPR-PROT=8899
EXPOSE 8899
COPY --from=builder /api-mirror /
ENTRYPOINT ["/api-mirror"]
