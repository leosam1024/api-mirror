<h1 align="center">
  <br>api-mirror<br>
</h1>

<h4 align="center">一个十分简单的转发工具</h4>

指定一个URL后，可以同时对多个API地址发起请求，返回最快的哪一个。

## Features

- 支持自定义多组转发配置
- 支持指定多个并发域名，并为每个域名配置权重
- 支持为每组转发配置指定超时时间、限流器、并发域名数量和过滤header头

## CONFIG

**配置示例**

~~~
# http服务启动端口
port: 8899

proxyConfig:
  - desc: "订阅转换服务"
    paths: # URL中PATH 不包含参数 。 可以配置多个
      - path: "/sub"   # URL中PATH 不包含参数 。 可以配置多个
        # matchType: "exact"  # paths匹配模式, exact：精确匹配、prefix：前缀匹配、regexp：正则匹配
      - path: "/version"
        matchType: "exact"
      - path: "/readconf"
        matchType: "exact"
      - path: "/updateconf"
        matchType: "exact"
      - path: "/getruleset"
        matchType: "exact"
      - path: "/getprofile"
        matchType: "exact"
      - path: "/qx-script"
        matchType: "exact"
      - path: "/qx-rewrite"
        matchType: "exact"
      - path: "/render"
        matchType: "exact"
      - path: "/convert"
        matchType: "exact"
      - path: "/getlocal"
        matchType: "exact"
    hosts: # 要并发访问的host
      - host: "https://id9.cc"
        weight: 1
      - host: "https://sub.xeton.dev"
        weight: 1
      - host: "https://api.dler.io"
        weight: 1
      - host: "https://sub.maoxiongnet.com"
        weight: 1
    filter:
      timeOut: 5000     # 超时时间 单位毫秒
      limitQps: 20      # 限流器，每秒只允许limiterQps访问,超过就
      limitHosts: 3     # 如果hosts数量过多，则随机选limit个进行请求
      limitRespHeaders: # 要过滤掉的响应头header，大小写敏感
        - "Set-Cookie"
~~~

**使用**

1. 启动服务后。
2. 访问xxx:port/sub?xxxx网址
3. 后端就会把网址替换为多个域名地址，并发访问，并返回最快响应的

## License

This software is released under the GPL-3.0 license.

