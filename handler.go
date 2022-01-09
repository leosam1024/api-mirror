package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

// 初始化web服务
func startWeb() {
	// 设置路由器
	port := ProjectConfig.Port
	configs := ProjectConfig.ProxyConfigs
	// 如果在环境变量里定义了端口号，则用环境变量中的
	if len(os.Getenv(EvnMirrorPort)) > 0 && IsNum(os.Getenv(EvnMirrorPort)) {
		port, _ = strconv.Atoi(os.Getenv(EvnMirrorPort))
	}

	for i := 0; i < len(configs); i++ {
		http.HandleFunc(configs[0].Path, proxyHandler)
	}

	http.HandleFunc("/", indexHandler)

	log.Info("Starting listen on ", port)

	// 启动web服务
	err := http.ListenAndServe(
		":"+strconv.Itoa(port),
		nil,
	)
	if err != nil {
		log.Error("ERROR", err)
	}

}

// indexHandler 首页Handler
func indexHandler(writer http.ResponseWriter, request *http.Request) {
	if request.URL.RequestURI() == "/favicon.ico" {
		return
	}
	fmt.Fprintf(writer, "api-mirror running...")
}

// proxyHandler 转发Handler  并发请求多个网址，返回最快的
func proxyHandler(writer http.ResponseWriter, request *http.Request) {
	t := time.Now().UnixMilli()

	configs := ProjectConfig.ProxyConfigs
	method := request.Method
	userAgent := request.UserAgent()
	path := request.URL.Path
	requestURI := request.RequestURI
	var config = findProxyConfig(configs, path)
	//urls := []string{"http://m2.auto.itc.cn/car/theme/newdb/images/favicon.ico", "https://www.google.com"}

	timeout := time.Duration(config.TimeOut)

	content, host, head := mirroredQuery(config.Hosts, requestURI, method, userAgent, timeout)

	if len(head["Content-Type"]) > 0 {
		writer.Header().Set("Content-Type", head["Content-Type"][0])
	}
	fmt.Fprintf(writer, content)

	log.Infof("请求成功，耗时%d毫秒，Limit：[%d]，使用HOST：[%s]，Path：[%s]",
		time.Now().UnixMilli()-t, config.Limit, host, requestURI)
}

func findProxyConfig(configs []ProxyConfig, path string) ProxyConfig {
	var proxyConfig ProxyConfig
	for i := 0; i < len(configs); i++ {
		config := configs[i]
		if path == config.Path {
			// 深复制一份
			copyHosts := make([]string, len(config.Hosts))
			copy(copyHosts, config.Hosts)
			proxyConfig = ProxyConfig{
				Desc:    config.Desc,
				Path:    config.Path,
				TimeOut: config.TimeOut,
				Limit:   config.Limit,
				Hosts:   copyHosts,
			}
			break
		}
	}
	if proxyConfig.isEmpty() {
		return proxyConfig
	}

	// 如果hosts的数量 超出Limit 。 则从 Hosts 随机取出Limit个
	if proxyConfig.Limit < len(proxyConfig.Hosts) {
		rand.Seed(time.Now().Unix())
		rand.Shuffle(
			len(proxyConfig.Hosts),
			func(i, j int) { proxyConfig.Hosts[i], proxyConfig.Hosts[j] = proxyConfig.Hosts[j], proxyConfig.Hosts[i] },
		)
		proxyConfig.Hosts = proxyConfig.Hosts[0:proxyConfig.Limit]
	}

	return proxyConfig
}

func mirroredQuery(hosts []string, requestURI string, method string, userAgent string, timeOut time.Duration) (string, string, http.Header) {
	if timeOut <= 0 {
		timeOut = DefaultTimeoutDuration
	}

	responses := make(chan string, len(hosts))
	hostChans := make(chan string, len(hosts))
	headers := make(chan http.Header, len(hosts))

	for i := 0; i < len(hosts); i++ {
		i := i
		go func() {
			url := hosts[i] + requestURI
			content, header := getRequestByAll(url, method, userAgent, timeOut)
			responses <- content
			headers <- header
			hostChans <- hosts[i]
		}()
	}

	response := <-responses
	host := <-hostChans
	header := <-headers

	return response, host, header
}

func getRequestByAll(url string, method string, userAgent string, timeOut time.Duration) (string, http.Header) {
	if len(method) == 0 {
		method = "GET"
	}
	if len(userAgent) == 0 {
		userAgent = "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36"
	}
	if timeOut <= 0 {
		timeOut = DefaultTimeoutDuration
	}

	client := &http.Client{Timeout: timeOut * time.Millisecond}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		//panic(err)
		log.Error(err.Error())
		return "", nil
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		//panic(err)
		log.Error(err.Error())
		return "", nil
	}
	defer resp.Body.Close()

	result, _ := ioutil.ReadAll(resp.Body)

	return string(result), resp.Header
}

func IsNum(s string) bool {
	match, _ := regexp.MatchString(`^[\+-]?\d+$`, s)
	return match
}
