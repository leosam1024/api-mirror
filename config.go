package main

import (
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"time"
)

const (
	DefaultTimeoutDuration time.Duration = 5000
	EvnMirrorPort          string        = "MIRROR-PORT"
	PathMatchTypeExact     string        = "exact"
	PathMatchTypePrefix    string        = "prefix"
	PathMatchTypeRegexp    string        = "regexp"
)

type ServerProjectConfig struct {
	Port         int           `yaml:"port"`
	ProxyConfigs []ProxyConfig `yaml:"proxyConfig"`
}

type ProxyConfig struct {
	Desc   string            `yaml:"desc"`
	Paths  []ProxyPathConfig `yaml:"paths"`
	Hosts  []ProxyHostConfig `yaml:"hosts"`
	Filter ProxyConfigFilter `yaml:"filter"`
}

type ProxyPathConfig struct {
	Path      string `yaml:"path"`
	MatchType string `yaml:"matchType"`
}

type ProxyHostConfig struct {
	Host   string `yaml:"host"`
	Weight int    `yaml:"weight"`
}

type ProxyConfigFilter struct {
	TimeOut          int      `yaml:"timeOut"`
	LimitHosts       int      `yaml:"limitHosts"`
	LimitQps         int      `yaml:"limitQps"`
	LimitRespHeaders []string `yaml:"limitRespHeaders"`
	Limiter          *rate.Limiter
}

func (x ProxyConfig) isEmpty() bool {
	return reflect.DeepEqual(x, ProxyConfig{})
}

func (p ProxyPathConfig) isExactMatchType() bool {
	return p.MatchType == PathMatchTypeExact
}
func (p ProxyPathConfig) isPrefixMatchType() bool {
	return p.MatchType == PathMatchTypePrefix
}
func (p ProxyPathConfig) isRegexpMatchType() bool {
	return p.MatchType == PathMatchTypeRegexp
}

type ProxyHostConfigs []ProxyHostConfig

func (s ProxyHostConfigs) Len() int { return len(s) }

func (s ProxyHostConfigs) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s ProxyHostConfigs) Less(i, j int) bool { return s[i].Weight > s[j].Weight }

var ProjectConfig ServerProjectConfig

func initLog() {
	//var log = logrus.New().WithFields(logrus.Fields{})

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
		log.Error(err)
	}

	// yaml文件内容影射到结构体中
	err1 := yaml.Unmarshal(config, &ProjectConfig)
	if err1 != nil {
		log.Error("config.yaml 解析有问题", err1)
	}

	// 初始化其他信息
	for index := range ProjectConfig.ProxyConfigs {
		log.Infof("init HandleFunc desc:[%s], filter:[%+v]", ProjectConfig.ProxyConfigs[index].Desc, ProjectConfig.ProxyConfigs[index].Filter)

		// 设置默认过滤的返回响应头
		ProjectConfig.ProxyConfigs[index].Filter.LimitRespHeaders = append(ProjectConfig.ProxyConfigs[index].Filter.LimitRespHeaders, "Content-Encoding")

		// 设置限流器
		if ProjectConfig.ProxyConfigs[index].Filter.LimitQps > 0 {
			ProjectConfig.ProxyConfigs[index].Filter.Limiter = rate.NewLimiter(rate.Limit(ProjectConfig.ProxyConfigs[index].Filter.LimitQps), ProjectConfig.ProxyConfigs[index].Filter.LimitQps)
		} else {
			ProjectConfig.ProxyConfigs[index].Filter.Limiter = nil
		}
		// 设置path和匹配模式
		for i := range ProjectConfig.ProxyConfigs[index].Paths {
			// 匹配模式
			ProjectConfig.ProxyConfigs[index].Paths[i].MatchType = strings.TrimSpace(strings.ToLower(ProjectConfig.ProxyConfigs[index].Paths[i].MatchType))
			if len(ProjectConfig.ProxyConfigs[index].Paths[i].MatchType) == 0 {
				ProjectConfig.ProxyConfigs[index].Paths[i].MatchType = PathMatchTypeExact
			}
			if !strings.Contains(ProjectConfig.ProxyConfigs[index].Paths[i].MatchType, PathMatchTypeExact) &&
				!strings.Contains(ProjectConfig.ProxyConfigs[index].Paths[i].MatchType, PathMatchTypePrefix) &&
				!strings.Contains(ProjectConfig.ProxyConfigs[index].Paths[i].MatchType, PathMatchTypeRegexp) {
				log.Errorf("desc:[%s],path:[%s],匹配模式不对，matchType：[%s]", ProjectConfig.ProxyConfigs[index].Desc, ProjectConfig.ProxyConfigs[index].Paths[i].Path, ProjectConfig.ProxyConfigs[index].Paths[i].MatchType)
			}
			log.Infof("add HandleFunc success, desc:[%s], path:[%s], matchType：[%s]",
				ProjectConfig.ProxyConfigs[index].Desc,
				ProjectConfig.ProxyConfigs[index].Paths[i].Path,
				ProjectConfig.ProxyConfigs[index].Paths[i].MatchType)
		}
	}

}
