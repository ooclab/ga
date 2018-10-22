# Auth

ga 环境中 authz 架构图如下：

![](./attachments/ga-authz-arch.png)

1. 管理员（admin）运行初始化程序（`ga permission`）调用 `auth` 服务提供的接口更新权限。考虑 golang openapi 包处理 SwaggerUI Spec 的 Path 一致性，这里会使用 ga 子命令处理 Spec 文件的权限处理，并调用 auth 接口上传配置。
2. authz 和 authn 分开，这样有利于 authz 功能更内聚。
