# **G** uardian **A** ngel

A lightweight middleware for service-oriented architecture

![](./docs/attachments/ga-current-arch.png)

## 简介

ga 是一个缩写，可以有多种含义：
1. **G**uardian **A**ngel (守护神)
2. **G**eneral **A**gent (总代)

## 目录

- [为什么开发 ga ？](./docs/reason.md)
- [ideas](./docs/ideas.md)
- [istio](./docs/istio.md)
- [与 es 配合架构思考](./docs/add-es.md)
- [authz](./docs/authz.md)

ga 架构示例：

- [ga 架构一](./docs/arch-design/arch1.md)


### 进展

#### 已完成

- \[x] 支持启动任意数量的 forwarder ，每个 forwarder 可以自由搭配任意数量的 middlewares
- \[x] 支持 golang plugin 方式写自定义的 middleware 。 可以查看示例： [hello](https://github.com/ooclab/ga/tree/master/middlewares/hello)

#### TODO


## Contact

- info@ooclab.com
