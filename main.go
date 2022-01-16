package main

import "time"

var ProjectStartTime = time.Now().UnixMilli()

func main() {

	// 初始化日志  -> config.go
	initLog()

	// 初始化配置 -> config.go
	initConfig("config.yaml")

	// 初始化web服务 -> handler.go
	startWeb()
}
