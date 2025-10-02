# Introduction
A high-performance WebSocket gateway service.
# Architecture

# Components

## 基础组件
+ fiber 框架
+ samber/do 依赖注入库
+ viper 解析 config
+ slog 日志库

## 认证与会话组件
+ jwt 认证
+ session 会话管理

## 定义 api 协议
+ buf 管理 protobuf
+ 定义 message 协议和 rpc 服务
+ 定义 node 协议，用于网关节点的服务注册、负载均衡、路由（连接重定向）等

## 用户连接的抽象和封装
Link: 用户连接的抽象，它封装了底层的网络连接（如 WebSocket、TCP），并与一个用户会话 (Session) 绑定。它提供了面向业务的、统一的连接操作接口

### 升级 HTTP 到 WebSocket
因为我们需要在每个连接绑定一个用户会话，所以要自己处理 HTTP 到 WebSocket 的升级过程。

### WebSocket消息读取器和消息写入器
读取器和写入器负责从 net.Conn 读取和写入消息。