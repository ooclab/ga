# 使用 ga 作为 openai 的网关

## 为什么要使用 ga 作为 openai 的网关

1. ga 可以作为 openai 的网关，可以将 openai 的服务聚合成一个虚拟服务池，方便客户端调用；
2. 只要网关所在服务器能够访问 openai 服务器，而我们的服务可以访问网关，就可以实现访问 openai 服务。

## ga 作为 openai 的网关的配置

创建目录，并创建 `docker-compose.yml` 文件：

```yaml
# docker-compose.yml
services:
  openai:
    image: ooclab/ga:v0.9.13.3
    volumes:
      - "$PWD/ga.yml:/etc/ga/config.yml"
    ports:
      - "1.2.3.4:2999:2999"
    restart: unless-stopped
```

说明：`1.2.3.4` 为网关所在服务器的 IP 地址（指定地址绑定更安全）。

创建 `ga.yml` 文件：

```yaml
# ga.yml
version: "2"

# start the forward server
listen: ":2999"

services:

  httpbin:
    path_prefix: /httpbin
    backend: https://httpbin.org
    middlewares:
    - name: logger
    - name: debug

  openai:
    path_prefix: /v1
    backend: https://api.openai.com/v1
    middlewares:
    - name: logger
    - name: debug
```

启动 ga 服务：

```bash
docker-compose up -d
```

查看日志：

```bash
docker-compose logs -f
```

### 测试

配置环境变量：

```bash
export OPENAI_BASE_URL=http://1.2.3.4:2999/v1
export OPENAI_API_KEY=sk-xxx
```

说明：环境变量 `OPENAI_BASE_URL` 和 `OPENAI_API_KEY` 配置可以让我们在测试 python openai sdk 时，不需要修改代码。

测试 stream 返回：

```bash
curl $OPENAI_BASE_URL/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [
      {
        "role": "system",
        "content": "You are a helpful assistant."
      },
      {
        "role": "user",
        "content": "Hello!"
      }
    ],
    "stream": true
  }'
```

测试 TTS 返回：

```bash
curl $OPENAI_BASE_URL/audio/speech \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "tts-1-hd",
    "input": "Courage is what it takes to stand up and speak; courage is also what it takes to sit down and listen.",
    "voice": "alloy"
  }' \
  --output speech-hd.mp3
```

## TODO

- [ ] 支持 tokens 统计、限流
- [ ] 支持 api key 认证（自定义）
- [ ] 支持其他 LLMs （私有化部署）
