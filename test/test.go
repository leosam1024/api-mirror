package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	defaultTimeoutDuration time.Duration = 5000
)

func main() {
	go spinner(100 * time.Millisecond)
	const n = 45
	fibN := fib(n) // slow
	fmt.Printf("\rFibonacci(%d) = %d\n", n, fibN)
}

func spinner(delay time.Duration) {
	for {
		for _, r := range `-\|/` {
			fmt.Printf("\r%c", r)
			time.Sleep(delay)
		}
	}
}

func fib(x int) int {
	if x < 2 {
		return x
	}
	return fib(x-1) + fib(x-2)
}

func getUrl(host string) {
	timestamp := time.Now().UnixMilli()
	url := host + "/sub?target=clash&new_name=true&url=https%3A%2F%2Fmiaona.xyz%2Fapi%2Fv1%2Fclient%2Fsubscribe%3Ftoken%3Da1cbe1c8b23631265eb919151911f247&configs=https%3A%2F%2Fraw.githubusercontent.com%2FACL4SSR%2FACL4SSR%2Fmaster%2FClash%2Fconfig%2FACL4SSR_Online_Mini_MultiCountry.ini&include=(%e6%b8%af%7c%e6%97%a5%7c%e7%be%8e)"
	content := getRequest(url, 5000)
	timestamp2 := time.Now().UnixMilli()
	fmt.Println(timestamp2 - timestamp)
	fmt.Println(len(content))
}

func getRequest(url string, timeOut time.Duration) string {

	client := &http.Client{Timeout: timeOut * time.Millisecond}

	resp, err := client.Get(url)

	if err != nil {
		log.Fatalln("ERROR", err)
		return ""
	}
	defer resp.Body.Close()

	result, _ := ioutil.ReadAll(resp.Body)
	return string(result)
}

func mirroredQuery() (string, int) {
	hosts := [...]string{"https://www.baidu.com", "https://www.google.com"}

	responses := make(chan string, len(hosts))
	fastUrls := make(chan string, len(hosts))
	for i := 0; i < len(hosts); i++ {
		url := hosts[i]
		go func() {
			responses <- getRequest(url, defaultTimeoutDuration)
			fastUrls <- url
		}()
	}

	fastUrl := <-fastUrls
	response := <-responses
	return fastUrl, len(response)
}

func getRequest2(url string, timeOut time.Duration) (string, http.Header) {

	client := &http.Client{Timeout: timeOut * time.Millisecond}

	req, err := http.NewRequest("GET", "http://httpbin.org/user-agent", nil)
	if err != nil {
		//panic(err)
		println(err.Error())
		return "", nil
	}
	req.Header.Set("User-Agent", "Golang_Spider_Bot/3.0")

	resp, err := client.Do(req)
	if err != nil {
		//panic(err)
		println(err.Error())
		return "", nil
	}
	defer resp.Body.Close()

	result, _ := ioutil.ReadAll(resp.Body)

	return string(result), resp.Header
}
