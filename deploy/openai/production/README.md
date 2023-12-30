# 部署生产环境

进入 `deploy/openai/production` 目录

1. 查看 docker-compose.yml 文件，修改配置
2. 查看 ga.yml 文件，修改配置
3. 启动服务

```bash
docker compose up -d
```

## FAQ

### 配置 key

```bash
export OPENAI_API_KEY=cs-yyy
# 设置 key
docker compose exec etcd etcdctl put $OPENAI_API_KEY '{"token":"sk-xxx", "count":0}'
# 获取 key 配置
docker compose exec etcd etcdctl get $OPENAI_API_KEY
```
