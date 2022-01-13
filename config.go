package main

import (
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"reflect"
	"time"
)

const (
	DefaultTimeoutDuration time.Duration = 5000
	EvnMirrorPort          string        = "MIRROR-PORT"
)

type ServerProjectConfig struct {
	Port         int           `yaml:"port"`
	ProxyConfigs []ProxyConfig `yaml:"proxyConfig"`
}

type ProxyConfig struct {
	Desc   string            `yaml:"desc"`
	Paths  []string          `yaml:"paths"`
	Hosts  []string          `yaml:"hosts"`
	Filter ProxyConfigFilter `yaml:"filter"`
}

type ProxyConfigFilter struct {
	TimeOut    int      `yaml:"timeOut"`
	LimitHosts int      `yaml:"limitHosts"`
	LimitQps   int      `yaml:"limitQps"`
	RespHeader []string `yaml:"respHeader"`
	Limiter    *rate.Limiter
}

func (x ProxyConfig) isEmpty() bool {
	return reflect.DeepEqual(x, ProxyConfig{})
}

var ProjectConfig ServerProjectConfig

//var log = logrus.New().WithFields(logrus.Fields{})

func initLog() {
	// 设置日志格式为json格式
	// log.SetFormatter(&log.JSONFormatter{})

	// force colors on for TextFormatter
	log.SetFormatter(&log.TextFormatter{
		EnvironmentOverrideColors: true,
		ForceColors:               true,
		FullTimestamp:             true,
		TimestampFormat:           "2006-01-02 15:04:05",
		// DisableSorting:true,
	})

	// 设置将日志输出到标准输出（默认的输出为stderr，标准错误）
	// 日志消息输出可以是任意的io.writer类型
	log.SetOutput(os.Stdout)
	// logfile, _ := os.OpenFile("./app.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	// logrus.SetOutput(logfile)

	// 设置日志级别为warn以上
	// logrus.SetLevel(logrus.InfoLevel)
	log.SetLevel(log.DebugLevel)

	// 设置在输出日志中添加文件名和方法信息：
	//log.SetReportCaller(true)
}

func initConfig(configFile string) {

	config, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Info(err)
	}

	//yaml文件内容影射到结构体中
	err1 := yaml.Unmarshal(config, &ProjectConfig)
	if err1 != nil {
		log.Error("config.yaml 解析有问题", err1)
	}
	for index, configs := range ProjectConfig.ProxyConfigs {
		// 设置默认返回响应头
		ProjectConfig.ProxyConfigs[index].Filter.RespHeader = append(configs.Filter.RespHeader, "Content-Type")
		// 设置限流器
		if ProjectConfig.ProxyConfigs[index].Filter.LimitQps > 0 {
			ProjectConfig.ProxyConfigs[index].Filter.Limiter = rate.NewLimiter(rate.Limit(ProjectConfig.ProxyConfigs[index].Filter.LimitQps), ProjectConfig.ProxyConfigs[index].Filter.LimitQps)
		} else {
			ProjectConfig.ProxyConfigs[index].Filter.Limiter = nil
		}
	}

	//通过访问结构体成员获取yaml文件中对应的key-value
	// fmt.Println(serverConfig.Port)
	//fmt.Println(ProjectConfig.ProxyConfigs)
}
