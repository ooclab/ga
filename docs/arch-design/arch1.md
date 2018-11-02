# ga 架构一

每个 ga 和一个后端服务对应。

优势：
- 后端服务网络可以完全隔离，ga 和 traefik 在一个网络

网络架构

![](./attachments/ga-design-arch1.png)

`ga serve` 服务可以有：
- `2999`: 转发所有外部请求到内部服务（中间件保证 token 校验、权限校验等）
- `2998`: 转发内部服务请求其他服务的请求（中间件可以为请求自动添加 token ，以满足校验）

![](./../attachments/ga-serve-design.png)
