# MQTT 数据获取说明

本文说明后端服务从哪里获取 MQTT 数据，以及设备、通道和 CAN 帧数据如何进入后端处理流程。

## 数据来源

后端服务从本地 EMQX 容器获取 MQTT 数据。

当前本地 EMQX 容器暴露的 MQTT 端口是：

```text
127.0.0.1:11883
```

对应 MQTT broker 地址：

```text
tcp://127.0.0.1:11883
```

后端启动后会作为 MQTT 客户端连接这个 broker，并订阅设备数据主题。

## MQTT 配置

建议后端配置如下：

```yaml
mqtt:
  enabled: true
  broker: "tcp://127.0.0.1:11883"
  client_id: "can-server"
  topic: "can/devices/+/channels/+/frames"
  qos: 1
  connect_timeout_ms: 5000
```

也可以通过环境变量覆盖：

```text
MQTT_ENABLED=true
MQTT_BROKER=tcp://127.0.0.1:11883
MQTT_CLIENT_ID=can-server
MQTT_TOPIC=can/devices/+/channels/+/frames
MQTT_QOS=1
MQTT_CONNECT_TIMEOUT_MS=5000
```

## 订阅主题

后端订阅的主题格式是：

```text
can/devices/{device_id}/channels/{channel}/frames
```

其中：

```text
device_id: 设备标识
channel:   通道标识
frames:    表示该消息携带 CAN 帧数据
```

示例：

```text
can/devices/left-thruster/channels/telemetry/frames
can/devices/right-thruster/channels/telemetry/frames
can/devices/steering-rudder/channels/status/frames
```

后端通过 MQTT topic 中的 `devices/{device_id}` 获取设备信息，通过 `channels/{channel}` 获取通道信息。

## 设备和通道

当前建议的设备标识：

```text
left-thruster     左推进器
right-thruster    右推进器
steering-rudder   转向舵
triple-screen     三联屏
ambient-light     氛围灯
```

当前建议的通道标识：

```text
telemetry   遥测数据，例如转速、档位、角度
status      状态数据，例如在线状态、故障状态
raw         原始数据
command     命令通道，后续用于 MQTT 下发控制命令
```

第一阶段后端只需要订阅上行数据，推荐使用：

```text
telemetry
status
raw
```

`command` 通道可以保留给后续扩展。

## 消息 Payload

MQTT 消息体使用 JSON。

最小格式：

```json
{
  "can_id": "0xA1",
  "data": "01000A0000000000"
}
```

字段说明：

```text
can_id: CAN 报文 ID，支持 0xA1 这种十六进制字符串
data:   CAN 数据帧，固定 8 字节，使用 16 位十六进制字符串
```

也可以在 payload 中显式携带设备和通道：

```json
{
  "device_id": "left-thruster",
  "channel": "telemetry",
  "can_id": "0xA1",
  "data": "01000A0000000000"
}
```

如果 payload 中没有 `device_id` 和 `channel`，后端从 topic 中解析。

如果 payload 中有 `device_id` 和 `channel`，建议以后端 payload 中的值为准。

## 后端处理流程

MQTT 数据进入后端后的处理流程：

```text
设备或模拟器
  -> EMQX tcp://127.0.0.1:11883
  -> 后端 MQTT subscriber
  -> 解析 topic 中的 device_id 和 channel
  -> 解析 payload 中的 can_id 和 data
  -> CAN 解析器 can.Parse()
  -> 写入 can_messages 日志表
  -> 通过 /api/v1/ws/can 广播给前端
```

也就是说，MQTT 只是新增的数据入口。进入后端后，数据会复用现有 CAN 帧解析和 websocket 广播逻辑。

## WebSocket 广播数据

后端处理 MQTT 消息后，建议广播给前端的数据格式如下：

```json
{
  "type": "can_frame",
  "source": "mqtt",
  "device_id": "left-thruster",
  "channel": "telemetry",
  "can_id": "0xA1",
  "device": "左推进器",
  "data": "01000a0000000000",
  "parsed": {
    "gear": 1,
    "rpm": 10
  },
  "direction": 0
}
```

字段说明：

```text
type:      消息类型，固定为 can_frame
source:    数据来源，MQTT 数据为 mqtt
device_id: 设备标识，来自 topic 或 payload
channel:   通道标识，来自 topic 或 payload
can_id:    CAN 报文 ID
device:    根据 can_id 解析出的设备中文名称
data:      原始 CAN 数据
parsed:    根据 can_id 解析后的业务数据
direction: 数据方向，0 表示控制器到 PC 的上行数据
```

## 示例消息

左推进器遥测数据：

```text
Topic:
can/devices/left-thruster/channels/telemetry/frames
```

```json
{
  "can_id": "0xA1",
  "data": "01000A0000000000"
}
```

右推进器遥测数据：

```text
Topic:
can/devices/right-thruster/channels/telemetry/frames
```

```json
{
  "can_id": "0xA2",
  "data": "0200140000000000"
}
```

转向舵状态数据：

```text
Topic:
can/devices/steering-rudder/channels/status/frames
```

```json
{
  "can_id": "0xB1",
  "data": "010F000000000000"
}
```

## 当前阶段边界

当前阶段只做 MQTT 数据获取：

```text
后端订阅 MQTT
接收设备和通道数据
解析 CAN 帧
入库日志
广播给前端
```

暂不做：

```text
通过 MQTT 下发控制命令
新增设备状态缓存表
修改 GET /api/v1/devices/:id/data 的数据来源
```

后续如果需要让设备详情接口直接返回 MQTT 最新数据，可以在服务端增加设备状态缓存，将每条 MQTT 上行数据按 `device_id + channel` 存储为最新状态。
