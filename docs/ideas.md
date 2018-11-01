# Idea


## 支持可插拔的中间件（已完成）

ga 的功能可以精简如下：
1. 运行多个 http forwarder (以后可以是 grpc forwarder, ...)
2. 每个不同的 http forwarder 配置加载相应的中间件

中间件 (middleware) 也有可能是其他功能：
1. 将一种协议转换为另外一种协议（如 grpc <-> http）（不仅仅本协议内操作）

和其他程序配合：
1. 不使用 service name 概念，仅仅使用 `key` prefix，与 etcd 集成（方便加载/更新配置）
2. 和 traefix 等 gateway 配合（当然可能都是使用 etcd），利用其已经成熟的功能实现灵活架构扩展


## 和 otunnel 配合使用

可以让开发者使用任何私人环境部署后端服务，通过 otunnel 链接到线上测试环境。
真正实现“分布式开发/协作”。
