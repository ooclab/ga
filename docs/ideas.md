# Idea


## 为旧的无验证服务提供自由的验证支持

比如：
1. docker registry , 加一个 auth 模块就可以支持账号系统
2. 各种静态 Web 页面（如文档），可以加上用户验证（参考 kong + oidc / openid）
3. 上面的静态资源还有下载站等

用户及权限模块可以直接对接 [dex](https://github.com/dexidp/dex) / [keycloak](https://www.keycloak.org/)


## 解绑 ga 和后端服务的关系

ga 已经是一个 forwarder 抽象的管理者，但是由于之前的设计，
ga 的 openapi middleware 目前和后端服务是一一绑定关系：

> 意味着使用 openapi middleware 的 ga 还只是和一个后端服务绑定，

这样的话，ga 就不能是自由扩展。

下一步计划：

> 将 ga 的所有 middleware 和后端服务解绑（目前只有 openapi 依赖后端的
OpenAPI 2.0 Spec 文档）

这样 ga 可以直接将 traefix / kong 当作后端, 支持水平扩展。比如 openapi middleware
可以根据 URL PathPrefix (服务名参数也不需要，自行判断) 来加载不同的 Spec 对象，完成其流程。

对于小型的应用，ga 可以替换 traefix / kong ，管理后端服务，只要加上 PathPrefix 和负载均衡即可。

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
