# debug middleware

方便开发过程中，输出完整的 HTTP Request 数据。

## build `.so`

```
go build -buildmode=plugin
```

默认情况下 debug 不会主动输出请求数据。触发 debug 输出 HTTP 请求的完整数据条件：
1. 加载 debug middleware
2. 下面任意一条件：
   - ga 命令行使用 `-v` 或 `--verbose` 开启 debug 模式
   - export GA_DEBUG=true
   - URL Query String 增加 `debug=true`
