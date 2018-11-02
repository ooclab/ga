# addauth middleware

用途：

1. 内部服务访问其他服务，且需要统一的验证时，此 middleware 可以自动完成验证

流程：

1. [authn](https://github.com/ooclab/ga.authn) 完成验证（ 管理 app_id / app_secret ；保存 access_token, refresh_token ; 自动刷新 access_token ）
2. 启动 forwarder 监听端口（如 `2998`），接受内部请求
3. 内部请求需要访问内部其他服务时，为其自动添加名为 `Authorization: Bearer {access_token}` 的 HTTP Request Header
4. 转发 HTTP Request 到真正的服务

说明：

1. [authn](https://github.com/ooclab/ga.authn) 是我们设计的一种用户验证服务，实际环境中，可以修改代码，实现各种用户认证服务。
