# openai 中间件

以下场景使用：

1. 如果你想使用自定义 key 来验证 openai 请求，并分发任意的 key 给不同的 app ；
2. 如果你想统计不同 app 的 tokens 或其他调用信息；
3. 如果你想限制不同 app 的使用频率（比如每天 <100 请求）；
4. 对，你可以把自己的 openai 网关部署到公网上（比如我们之前是不能地🤣）；
5. 当然，对于某些环境，用 ga 创建一个 openai 的网关是必要的。

btw, ga 是透明网关，因此 openai api 自身的限制配额和用户账户有关。

## 使用

- [配置 openai middle 示例](../../deploy/openai/README.md)
- [生厂环境部署示例](../../deploy/openai/production/README.md)

## TODO

- [ ] 每天对每个 token 的调用次数清零
- [ ] 结合 prometheus 统计调用次数
- [ ] 支持多个后端 openai key 的负载均衡(failover, round-robin, leastconn, ip_hash)
- [ ] 前端 key 的动态管理（暴力尝试，限制次数，限制时间，限制IP等）
