package main

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"math/rand"
	http "net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
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

	http.HandleFunc("/", indexHandler)

	for i := 0; i < len(configs); i++ {
		config := configs[i]
		if len(config.Paths) == 0 {
			continue
		}
		log.Infof("start add http HandleFunc success, desc:[%s], filter:[%+v]", config.Desc, config.Filter)
		for _, pathConfig := range config.Paths {
			if len(pathConfig.Path) == 0 {
				continue
			}
			http.HandleFunc(pathConfig.Path, proxyHandler)
			log.Infof("add http HandleFunc success, desc:[%s], path:[%s]", config.Desc, pathConfig.Path)
		}
	}
	log.Infof("Starting http listen on port:[%d], cost:[%d ms]", port, time.Now().UnixMilli()-ProjectStartTime)

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
	requestURI := request.URL.RequestURI()
	// favicon
	if requestURI == "/favicon.ico" {
		return
	}

	// 首页
	matchIndex, _ := regexp.MatchString(`^(/|/index.html|index)$`, requestURI)
	if matchIndex {
		fmt.Fprintf(writer, "api-mirror running...")
		return
	}

	// 其他未找到代理器的报错
	log.Warnf("not find HandleFunc, path:[%s]", requestURI)
	fmt.Fprintf(writer, "not find HandleFunc, path:[%s]", requestURI)
}

// proxyHandler 转发Handler  并发请求多个网址，返回最快的
func proxyHandler(writer http.ResponseWriter, request *http.Request) {
	// 1. 准备数据
	t := time.Now().UnixMilli()
	configs := ProjectConfig.ProxyConfigs
	path := request.URL.Path
	requestURI := request.RequestURI

	// 2. 根据path查找出符合的配置项来
	var config = findProxyConfig(configs, path)
	if config.Paths == nil {
		writer.WriteHeader(404)
		fmt.Fprintf(writer, "未匹配合适Handler：Path：[%s]", requestURI)
		log.Warnf("未匹配合适Handler：Path：[%s]", requestURI)
		return
	}
	// 3. 开启限流器，现在高并发请求
	if config.Filter.Limiter != nil && !config.Filter.Limiter.Allow() {
		fmt.Fprintf(writer, "QPS超出系统限制，请稍后再试")
		log.Warnf("限流成功，每秒限流QPS：[%.1f],Path：[%s]", config.Filter.Limiter.Limit(), requestURI)
		return
	}

	// 4. 请求代理，活动最快的结果
	host, responseBodyByte, statusCode, header := mirroredQuery(request, config)

	// 5. 处理响应 顺序不能乱
	// 5.1 处理响应 => 先处理响应头
	copyHeader(writer.Header(), header, config.Filter.LimitRespHeaders)
	// 5.2 处理响应 => 在处理返回code
	writer.WriteHeader(statusCode)
	// 5.3 处理响应 => 处理响应体
	writer.Write(responseBodyByte)

	// 6. 打印日志
	if len(responseBodyByte) < 10 && string(responseBodyByte) == "httpError" {
		log.Warnf("请求失败，耗时%d毫秒，Limit：[%d]，使用HOST：[%+v]，Path：[%s]", time.Now().UnixMilli()-t, config.Filter.LimitHosts, config.Hosts, requestURI)
	} else {
		log.Infof("请求成功，耗时%d毫秒，Limit：[%d]，使用HOST：[%s]，Path：[%s]", time.Now().UnixMilli()-t, config.Filter.LimitHosts, host, requestURI)
	}
}

func findProxyConfig(configs []ProxyConfig, path string) ProxyConfig {
	var proxyConfig ProxyConfig

	for i := 0; i < len(configs); i++ {
		config := configs[i]
		if len(config.Paths) == 0 {
			continue
		}
		for j := 0; j < len(config.Paths); j++ {
			pathConfig := config.Paths[j]
			if len(pathConfig.Path) == 0 {
				continue
			}
			find := false
			if pathConfig.isExactMatchType() && pathConfig.Path == path {
				find = true
			}
			if pathConfig.isPrefixMatchType() && strings.HasPrefix(pathConfig.Path, path) {
				find = true
			}
			if pathConfig.isPrefixMatchType() {
				matched, _ := regexp.MatchString(pathConfig.Path, path)
				find = matched
			}
			if find {
				// 深复制一份
				copyHosts := make([]ProxyHostConfig, len(config.Hosts))
				copy(copyHosts, config.Hosts)
				proxyConfig = ProxyConfig{
					Desc:   config.Desc,
					Paths:  config.Paths,
					Hosts:  copyHosts,
					Filter: config.Filter,
				}
				break
			}
		}
	}

	if proxyConfig.isEmpty() {
		return proxyConfig
	}

	// 如果hosts的数量 超出Limit 。 则从 Hosts 随机取出Limit个
	if proxyConfig.Filter.LimitHosts < len(proxyConfig.Hosts) {
		// 随机打乱一下
		rand.Seed(time.Now().Unix())
		rand.Shuffle(
			len(proxyConfig.Hosts),
			func(i, j int) {
				proxyConfig.Hosts[i], proxyConfig.Hosts[j] = proxyConfig.Hosts[j], proxyConfig.Hosts[i]
			},
		)

		// 按权重排序
		for i := range proxyConfig.Hosts {
			proxyConfig.Hosts[i].Weight = proxyConfig.Hosts[i].Weight * rand.Intn(100)
		}
		sort.Sort(ProxyHostConfigs(proxyConfig.Hosts))

		// 选取前LimitHosts个
		proxyConfig.Hosts = proxyConfig.Hosts[0:proxyConfig.Filter.LimitHosts]
	}

	return proxyConfig
}

func copyHeader(wHeader http.Header, respHeader http.Header, limitRespHeaders []string) {
	if len(respHeader) == 0 {
		return
	}

	for header, value := range respHeader {
		// 删除要过滤的
		if len(limitRespHeaders) > 0 && contains(header, limitRespHeaders) {
			wHeader.Del(header)
			continue
		}

		// 要转发的
		//if len(wHeader.Get(header)) > 0 {
		//// 	wHeader.Del(header)
		//}
		for _, v := range value {
			wHeader.Add(header, v)
		}
	}
}

func mirroredQuery(request *http.Request, config ProxyConfig) (string, []byte, int, http.Header) {
	// 准备数据
	method := request.Method
	requestURI := request.RequestURI
	hosts := config.Hosts
	timeOut := time.Duration(config.Filter.TimeOut)
	if timeOut <= 0 {
		timeOut = DefaultTimeoutDuration
	}
	requestBodyBytes := getRequestBody(request)
	// 并发执行请求
	hostChan := make(chan string, len(hosts))
	responseChan := make(chan *http.Response, len(hosts))
	responseBodyChan := make(chan []byte, len(hosts))

	for i := 0; i < len(hosts); i++ {
		num := i
		go func() {
			url := hosts[num].Host + requestURI
			body, response := getRequestByAll(url, method, request.Header, requestBodyBytes, timeOut)
			hostChan <- hosts[num].Host
			responseChan <- response
			responseBodyChan <- body
		}()
	}

	// 判断结果
	var responseBodyByte []byte
	var host string
	var response *http.Response
	for i := 0; i < len(hosts); i++ {
		// 接收最先返回的
		host = <-hostChan
		responseBodyByte = <-responseBodyChan
		response = <-responseChan
		// 判断返回结果
		ok := true
		ok = ok && response != nil && response.StatusCode < 500 && response.StatusCode > 99
		ok = ok && string(responseBodyByte) != "httpError"

		// 如果符合条件，直接结束
		if ok {
			break
		}
	}
	if response == nil {
		response = &http.Response{
			StatusCode: 404,
		}
	}
	return host, responseBodyByte, response.StatusCode, response.Header
}

func getRequestBody(request *http.Request) []byte {
	// 把request的内容读取出来
	var bodyBytes []byte
	if request.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(request.Body)
	}
	// 把刚刚读出来的再写进去
	request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	return bodyBytes
}
func getRequestByAll(url string, method string, requestHeader http.Header, requestBodyBytes []byte, timeOut time.Duration) ([]byte, *http.Response) {
	if len(method) == 0 {
		method = "GET"
	}
	if timeOut <= 0 {
		timeOut = DefaultTimeoutDuration
	}
	requestBody := ioutil.NopCloser(bytes.NewBuffer(requestBodyBytes))
	req, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		//panic(err)
		log.Error(err.Error())
		return []byte("httpError"), nil
	}

	if requestHeader != nil && len(requestHeader) > 0 {
		for key, headers := range requestHeader {
			// go不支持br编码，所以不要透传accept-encoding
			if strings.ToLower(key) == "accept-encoding" || len(headers) == 0 {
				continue
			}
			req.Header.Set(key, headers[0])
		}
	}

	client := &http.Client{Timeout: timeOut * time.Millisecond}
	resp, err := client.Do(req)
	if err != nil {
		//panic(err)
		log.Error(err.Error())
		return []byte("httpError"), nil
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	return bodyBytes, resp
}

// IsNum 判断是否是数字
func IsNum(s string) bool {
	match, _ := regexp.MatchString(`^[\+-]?\d+$`, s)
	return match
}

func contains(target string, strArray []string) bool {
	if len(strArray) == 0 {
		return false
	}
	for _, element := range strArray {
		if target == element {
			return true
		}
	}
	return false
}
