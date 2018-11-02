# Docker

## docker image build

约束：

1. golang 使用 `CGO_ENABLED=0` 编译，不支持 golang plugin
2. alpine busybox 环境不支持动态编译的 golang 程序
