# bot_msg

`bot_msg` 既可以作为 OneBot v11 SDK 使用，也可以作为一个基于 Gin 的 HTTP 服务运行。

服务模式下，调用方请求 `bot_msg` 暴露的 HTTP API，`bot_msg` 再通过 OneBot v11 HTTP/HTTPS 或 WS/WSS API 转发到实际机器人实现端。

## 架构

```text
cmd/botmsg                 程序入口、启动与优雅关闭
internal/config            服务环境变量配置
internal/event             收消息后的事件分发 pipeline
internal/service           服务业务编排与返回模型
internal/server            Gin 路由、中间件、HTTP action handler
pkg/onebot                 对外 SDK：配置、消息模型、HTTP/WS transport、事件订阅
docs                       Swagger 生成产物
```

核心关系：

```text
外部调用方 -> bot_msg HTTP 服务 -> pkg/onebot SDK -> OneBot v11 实现端 -> QQ
```

## 快速启动

WS/WSS 模式：

```bash
BOTMSG_ONEBOT_WS_URL=ws://dev-server:3001 \
BOTMSG_ONEBOT_ACCESS_TOKEN=your_token \
BOTMSG_ONEBOT_HEARTBEAT_SECONDS=30 \
BOTMSG_ONEBOT_RECONNECT_SECONDS=5 \
go run ./cmd/botmsg
```

HTTP/HTTPS 模式：

```bash
BOTMSG_ONEBOT_HTTP_URL=http://dev-server:3000 \
BOTMSG_ONEBOT_ACCESS_TOKEN=your_token \
go run ./cmd/botmsg
```

默认服务监听：

```text
:8080
```

## 服务配置

### bot_msg 服务自身配置

| 环境变量 | 默认值 | 说明 |
| --- | --- | --- |
| `BOTMSG_HTTP_ADDR` | `:8080` | `bot_msg` 自身 HTTP 服务监听地址 |
| `BOTMSG_ACCESS_TOKEN` | 空 | 调用 `bot_msg` HTTP API 时使用的 token，空表示不鉴权 |
| `BOTMSG_SELF_ID` | `10000` | `get_login_info` 返回的本地展示账号 |
| `BOTMSG_NICKNAME` | `bot_msg` | `get_login_info` 返回的本地展示昵称 |
| `BOTMSG_LOG_LEVEL` | `info` | 日志级别，支持 `debug`、`info`、`warn`、`error` |

### OneBot 连接配置

| 环境变量 | 默认值 | 说明 |
| --- | --- | --- |
| `BOTMSG_ONEBOT_HTTP_URL` | 空 | OneBot HTTP/HTTPS API 地址，例如 `http://dev-server:3000` |
| `BOTMSG_ONEBOT_WS_URL` | `ws://127.0.0.1:3001/api` | OneBot WS/WSS 地址，例如 `ws://dev-server:3001` |
| `BOTMSG_ONEBOT_WS_ENDPOINT` | 空 | 便捷写法，会把 `http://` 转成 `ws://`，把 `https://` 转成 `wss://` |
| `BOTMSG_ONEBOT_ACCESS_TOKEN` | 空 | 调用 OneBot 实现端时使用的 token |
| `BOTMSG_ONEBOT_TIMEOUT_SECONDS` | `10` | 单次 OneBot API 调用超时时间 |
| `BOTMSG_ONEBOT_HEARTBEAT_SECONDS` | `30` | WS 模式心跳间隔，会定期调用 `get_status` |
| `BOTMSG_ONEBOT_RECONNECT_SECONDS` | `5` | WS 断线后的重连间隔 |

### HTTP 和 WS 同时配置时的行为

如果同时配置：

```bash
BOTMSG_ONEBOT_HTTP_URL=http://dev-server:3000
BOTMSG_ONEBOT_WS_URL=ws://dev-server:3001
```

行为是：

- 主动 API 调用优先走 HTTP/HTTPS。
- WS/WSS 仍会启动，用于接收事件和消息。
- `get_status` 在服务层会认为 HTTP transport 可用；WS 连接状态仍会通过 SDK 日志体现。

如果只配置 `BOTMSG_ONEBOT_HTTP_URL`，且不配置 `BOTMSG_ONEBOT_WS_URL`，服务不会启动 WS 后台连接，也不会接收消息事件。

## 鉴权说明

有两个 token，含义不同：

| 配置 | 用途 |
| --- | --- |
| `BOTMSG_ACCESS_TOKEN` | 保护 `bot_msg` 自己的 HTTP API |
| `BOTMSG_ONEBOT_ACCESS_TOKEN` | 调用 OneBot 实现端时使用 |

调用 `bot_msg` 时可以这样传 `BOTMSG_ACCESS_TOKEN`：

```bash
curl http://127.0.0.1:8080/get_status \
  -H 'Authorization: Bearer service_token'
```

也支持 query 参数：

```bash
curl 'http://127.0.0.1:8080/get_status?access_token=service_token'
```

## 已支持 API

- `send_private_msg`
- `send_group_msg`
- `send_msg`
- `delete_msg`
- `get_msg`
- `get_login_info`
- `get_status`
- `get_version_info`
- `get_friend_list`
- `get_group_list`
- `get_group_member_list`

接口同时支持根路径和 `/api` 前缀：

```text
POST /send_private_msg
POST /api/send_private_msg
GET  /get_status
GET  /api/get_status
```

## 请求示例

发送私聊：

```bash
curl -X POST http://127.0.0.1:8080/send_private_msg \
  -H 'Content-Type: application/json' \
  -d '{"user_id":123456,"message":"你好 OneBot"}'
```

发送群聊：

```bash
curl -X POST http://127.0.0.1:8080/send_group_msg \
  -H 'Content-Type: application/json' \
  -d '{"group_id":123456,"message":"你好 OneBot"}'
```

消息段格式：

```bash
curl -X POST http://127.0.0.1:8080/send_group_msg \
  -H 'Content-Type: application/json' \
  -d '{
    "group_id": 123456,
    "message": [
      {
        "type": "text",
        "data": {
          "text": "hello"
        }
      }
    ]
  }'
```

获取状态：

```bash
curl http://127.0.0.1:8080/get_status
```

获取群列表：

```bash
curl http://127.0.0.1:8080/get_group_list
```

获取好友列表：

```bash
curl http://127.0.0.1:8080/get_friend_list
```

获取群成员列表：

```bash
curl -X POST http://127.0.0.1:8080/get_group_member_list \
  -H 'Content-Type: application/json' \
  -d '{"group_id":123456}'
```

## SDK 用法

```go
package main

import (
    "context"
    "log/slog"
    "os"
    "time"

    "bot_msg/pkg/onebot"
)

func main() {
    ctx := context.Background()

    client := onebot.NewClient(
        onebot.WithHTTP("http://dev-server:3000"),
        onebot.WithWebSocket("ws://dev-server:3001"),
        onebot.WithAccessToken("your_token"),
        onebot.WithTimeout(10*time.Second),
        onebot.WithHeartbeatInterval(30*time.Second),
        onebot.WithReconnectInterval(5*time.Second),
        onebot.WithLogger(slog.New(slog.NewJSONHandler(os.Stdout, nil))),
    )

    client.OnMessage(func(ctx context.Context, event onebot.MessageEvent) {
        slog.Info("收到消息",
            "message_type", event.MessageType,
            "user_id", event.UserID,
            "group_id", event.GroupID,
            "text", event.Message.PlainText(),
        )
    })

    client.Start(ctx)

    _, _ = client.SendGroupMessage(ctx, 123456, onebot.TextMessage("hello"))
}
```

SDK 配置项：

| 配置 | 说明 |
| --- | --- |
| `WithHTTP` | 配置 OneBot HTTP/HTTPS API 地址 |
| `WithWebSocket` | 配置 OneBot WS/WSS 地址 |
| `WithAccessToken` | 配置 OneBot 访问 token |
| `WithTimeout` | 配置 API 调用超时 |
| `WithHeartbeatInterval` | 配置 WS 心跳间隔 |
| `WithReconnectInterval` | 配置 WS 重连间隔 |
| `WithEventBuffer` | 配置事件通道缓冲大小 |
| `WithLogger` | 配置 SDK 内部日志 |

## 收消息行为

服务模式下，如果启用了 WS/WSS，SDK 会接收 OneBot 事件。

当前默认行为：

- 收到消息后进入 `internal/event` 分发 pipeline。
- 默认只打印结构化日志。
- 不自动回复。
- 不落库。
- 不转发 webhook。

后续扩展入口：

```go
eventDispatcher.UseMessageHandler(...)
```

相关文件：

```text
internal/event/dispatcher.go
internal/event/log_handler.go
```

## Swagger 文档

生成文档：

```bash
make swagger
```

清理文档：

```bash
make clean-swagger
```

生成产物：

```text
docs/docs.go
docs/swagger.json
docs/swagger.yaml
```

## Makefile

```bash
make swagger        # 生成 Swagger 文档
make clean-swagger  # 删除 docs 目录
make test           # 运行 go test ./...
```

## 注意事项

- 好友列表、群列表、群成员列表属于敏感数据，不建议默认落盘或完整打印。
- `BOTMSG_ONEBOT_ACCESS_TOKEN` 不应写入日志。
- 群发或循环发送消息可能触发账号风控，业务层应自行加限速。
- WS 事件回调不要做长时间阻塞操作，复杂任务建议丢到队列或 goroutine 池。
