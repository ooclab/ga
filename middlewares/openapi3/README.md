# openapi middleware

`openapi` middleware 处理所有请求的权限和参数。

用途：

1. 外部服务访问内部服务，需要校验其是否有相应的权限时，此 middleware 可以根据 OpenAPI Spec 文档（目前支持 OpenAPI 3）校验其权限和参数。

流程：

![](./openapi-middleware-design.png)

1. 启动 forwarder 监听端口（如 `2999`），接受外部请求
1. 加载 openapi3 middleware （读取后端服务的 OpenAPI 3 Spec 文档）
2. 根据当前请求 Method, URL 匹配权限名称，查询权限：1. spec 文件自身是否需要验证信息；2. TODO: 查询外部 api auth 服务（外部服务扩展，可以跳过该步骤），决定当前用户是否有权限访问该接口
3. 如果通过权限校验，则继续校验请求参数是否正确（TODO: 支持 spec 文件中添加信息判断是否需要跳过参数校验）
4. 如果以上都通过，转发 HTTP Request 到后端真正的服务
