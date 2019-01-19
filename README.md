# Guardian Angel

**G**uardian **A**ngel is a lightweight middleware for service-oriented architecture.

What is the `service-oriented` here ?
1. `big services` : export a http(s) api
2. `small services`
   - microservices / nanoservices
   - serverless / FaaS

We don't want to argument the term, we just want to do right job.
:-)

![](./docs/attachments/ga-current-arch.png)

Current Support:

- \[x] custom your middleware with golang plugin ( [third-party-middleware](https://github.com/urfave/negroni#third-party-middleware) )

Current Middlewares:

- \[x] [hello](./middlewares/hello) : for example
- \[x] [jwt](./middlewares/jwt) : decode JWT (from `Authorization` in HTTP Request Header), and set `X-User-Id` header by user id
- \[x] [addauth](./middlewares/addauth) : manage `access_token` and `refresh_token` , add JWT with `Authorization` in HTTP Request Header auto
- \[x] [openapi](./middlewares/openapi) : authorize permissions and validate request args with OpenAPI 2.0 Spec Document of the backend service
- \[x] [debug](./middlewares/debug) : print http request data when debug is enable
- \[ ] openapi-response : check the response for testing

------

## 目录

- [为什么开发 ga ？](./docs/reason.md)
- [ideas](./docs/ideas.md)
- [istio](./docs/istio.md)
- [与 es 配合架构思考](./docs/add-es.md)
- [authz](./docs/authz.md)
- [docker build image](./docs/docker.md)

ga 架构示例：

- [ga 架构一](./docs/arch-design/arch1.md)


### 进展

#### 已完成

- \[x] 支持启动任意数量的 forwarder ，每个 forwarder 可以自由搭配任意数量的 middlewares
- \[x] 支持 golang plugin 方式写自定义的 middleware 。 可以查看示例： [hello](https://github.com/ooclab/ga/tree/master/middlewares/hello)

#### TODO


## Contact

- info@ooclab.com
