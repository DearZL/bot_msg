# bot_msg

`bot_msg` 既可以作为 OneBot v11 SDK 使用，也可以作为一个基于 Gin 的 HTTP 服务运行。服务模式下，调用本服务的 HTTP API 后，本服务会通过 OneBot v11 HTTP/HTTPS 或 WS/WSS API 转发请求。

## 架构

```text
cmd/botmsg                 程序入口、启动与优雅关闭
internal/config            环境变量配置
pkg/onebot                 对外 SDK：配置、消息模型、HTTP/WS transport、事件订阅
internal/service           业务编排与返回模型
internal/server            Gin 路由、中间件、HTTP action handler
```

## 运行

```bash
go mod tidy
go run ./cmd/botmsg
```

默认监听 `:8080`。

可选环境变量：

```bash
BOTMSG_HTTP_ADDR=:8080
BOTMSG_ACCESS_TOKEN=
BOTMSG_SELF_ID=10000
BOTMSG_NICKNAME=bot_msg
BOTMSG_ONEBOT_HTTP_URL=
BOTMSG_ONEBOT_WS_URL=ws://127.0.0.1:3001
BOTMSG_ONEBOT_ACCESS_TOKEN=
BOTMSG_ONEBOT_TIMEOUT_SECONDS=10
BOTMSG_ONEBOT_HEARTBEAT_SECONDS=30
BOTMSG_ONEBOT_RECONNECT_SECONDS=5
BOTMSG_LOG_LEVEL=info
```

其中：

- `BOTMSG_ACCESS_TOKEN` 是调用本 Go 服务时需要的 token。
- `BOTMSG_ONEBOT_HTTP_URL` 是 OneBot v11 HTTP/HTTPS API 地址，例如 `http://dev-server:3000`。
- `BOTMSG_ONEBOT_WS_URL` 是 OneBot v11 WS/WSS API 地址，例如 `ws://dev-server:3001`。
- `BOTMSG_ONEBOT_ACCESS_TOKEN` 是调用 OneBot API 时使用的 token，如果服务端没配置 token 可以留空。
- `BOTMSG_ONEBOT_HEARTBEAT_SECONDS` 是 WS 心跳间隔，默认 30 秒，会通过 `get_status` 检测连接。
- `BOTMSG_ONEBOT_RECONNECT_SECONDS` 是 WS 断线后的重连间隔，默认 5 秒。

## SDK 用法

```go
client := onebot.NewClient(
    onebot.WithHTTP("http://dev-server:3000"),
    onebot.WithWebSocket("ws://dev-server:3001"),
    onebot.WithAccessToken("token"),
)

client.OnMessage(func(ctx context.Context, event onebot.MessageEvent) {
    if event.Message.PlainText() == "ping" {
        _, _ = client.SendGroupMessage(ctx, event.GroupID, onebot.TextMessage("pong"))
    }
})

client.Start(ctx)
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

接口同时支持：

- `POST /send_private_msg`
- `POST /api/send_private_msg`
- `GET /get_status`

## 请求示例

```bash
curl -X POST http://127.0.0.1:8080/send_private_msg \
  -H 'Content-Type: application/json' \
  -d '{"user_id":123456,"message":"你好 OneBot"}'
```

调用链路：

```text
调用方 -> bot_msg:8080/send_private_msg -> OneBot API -> QQ
```

响应：

```json
{
  "status": "ok",
  "retcode": 0,
  "data": {
    "message_id": 100001
  }
}
```

消息段格式也可以直接传 OneBot 数组：

```json
{
  "group_id": 123,
  "message": [
    {
      "type": "text",
      "data": {
        "text": "hello"
      }
    }
  ]
}
```
