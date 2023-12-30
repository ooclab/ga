# 部署 openai 验证网关

## 启动服务

修改 `config.yml` 。

```shell
# 启动本目录下的 etcd 服务
docker-compose up -d
# 到 ga 主目录下启动 ga 服务
GA_DEBUG=true ./ga serve -c deploy/openai/config.yml
```

## 使用

配置环境变量

```bash
# 设置自定义的 key （openssl rand -base64 32)
export OPENAI_API_KEY=cs-PR9PQZZz6Ww6P7c2hHCA0KeQmcAxJPsTsqeXZ7NXTB+TPpJLwn7vDx
# 设置 key
etcdctl put $OPENAI_API_KEY '{"token":"sk-xxx", "count":0}'
# 获取 key
etcdctl get $OPENAI_API_KEY
```

获取 key 结果示例：

```plain
cs-PR9PQZZz6Ww6P7c2hHCA0KeQmcAxJPsTsqeXZ7NXTB+TPpJLwn7vDx
{"token":"sk-xxx", "count":0}
```

![openai](./custom_keys_for_openai.png)

**注意：**

1. `OPENAI_API_KEY` 为自定义的 key，可以随意设置，但是要保证唯一性；
2. `sk-xxx` 为 openai 的 key，需要从 openai 官网获取。

## FAQ

### 如果达到调用次数限制，返回 429

```plain
*   Trying 127.0.0.1:2999...
* Connected to 127.0.0.1 (127.0.0.1) port 2999
> POST /v1/chat/completions HTTP/1.1
> Host: 127.0.0.1:2999
> User-Agent: curl/8.4.0
> Accept: */*
> Content-Type: application/json
> Authorization: Bearer cs-PR9PQZZz6Ww6P7c2hHCA0KeQmcAxJPsTsqeXZ7NXTB+TPpJLwn7vDx
> Content-Length: 220
>
< HTTP/1.1 429 Too Many Requests
< Date: Sat, 30 Dec 2023 07:08:00 GMT
< Content-Length: 0
<
* Connection #0 to host 127.0.0.1 left intact
```
