package main

import (
	"flag"
	"os"
	"time"
)

var ProjectStartTime = time.Now().UnixMilli()

func main() {
	// 解析配置文件路径
	var configFilePath = ""
	flag.StringVar(&configFilePath, "c", "", "path to config file: config.yaml")
	flag.Parse()
	if len(configFilePath) == 0 {
		// 从环境变量里取
		configFilePath = os.Getenv("CONFIG_FILE")
	}
	if len(configFilePath) == 0 {
		configFilePath = "config.yaml"
	}

	// 初始化日志  -> config.go
	initLog()

	// 初始化配置 -> config.go
	initConfig(configFilePath)

	// 初始化web服务 -> handler.go
	startWeb()
}
