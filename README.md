<h1 align="center">
  <br>api-mirror<br>
</h1>

<h4 align="center">一个十分简单的转发工具</h4>


指定一个URL后，可以同时对多个API地址发起请求，返回最快的哪一个。 

## Features

- 自定义转发配置
- 可以指定超时见见
- 可以指定多个域名


## CONFIG
**配置示例**
~~~
# http服务启动端口
port: 8899

proxyConfig:
  - desc: "订阅转换服务"
    path: "/sub"                    # URL中PATH 不包含参数
    timeOut: 5000                   # 超时时间 单位毫秒
    hosts: # 要并发访问的host
      - "https://sub.id9.cc"
      - "https://sub.xeton.dev"
      - "https://api.dler.io"
      - "https://sub.maoxiongnet.com"

~~~
**使用**
1. 启动服务后。
2. 访问xxx:port/sub?xxxx网址
3. 后端就会把网址替换为多个域名地址，并发访问，并返回最快响应的

## License

This software is released under the GPL-3.0 license.

