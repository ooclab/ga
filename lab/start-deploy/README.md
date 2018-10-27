# 部署 ga 架构

如何部署面向服务的软件环境


## 启动

第一次启动服务，需要创建 traefik 容器网络：

```
docker network create ga-deploy-traefik
```
