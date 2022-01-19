package main

import (
	"flag"
	"os"
	"strconv"
	"time"
)

var ProjectStartTime = time.Now().UnixMilli()

func main() {
	// 获取运行参数
	var configFilePath = ""
	var httpServerPort = 0
	flag.StringVar(&configFilePath, "c", "", "path to config file: config.yaml")
	flag.IntVar(&httpServerPort, "p", 0, "path to config port: http.port")
	flag.Parse()

	// 初始化日志  -> config.go
	initLog()

	// 初始化配置 -> config.go
	configFilePath = getConfigFilePath(configFilePath)
	initConfig(configFilePath)

	// 初始化web服务 -> handler.go
	httpServerPort = getHttpServerPort(httpServerPort)
	startWeb(httpServerPort)
}

// 获取配置文件路径
func getConfigFilePath(configFilePath string) string {
	// 解析配置文件路径 运行参数 =》 环境变量 =》 兜底
	if len(configFilePath) == 0 {
		// 从环境变量里取
		configFilePath = os.Getenv(EvnMirrorConfigFile)
	}
	if len(configFilePath) == 0 {
		configFilePath = "config.yaml"
	}
	return configFilePath
}

// 获取http服务启动端口号
func getHttpServerPort(httpPort int) int {
	if httpPort <= 0 {
		// 如果在环境变量里定义了端口号，则用环境变量中的
		httpPort, _ = strconv.Atoi(os.Getenv(EvnMirrorPort))
	}
	if httpPort <= 0 {
		httpPort = ProjectConfig.Port
	}
	return httpPort
}
