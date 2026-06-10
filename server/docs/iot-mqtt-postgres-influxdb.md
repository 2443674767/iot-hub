# IoT MQTT 数据入库说明

本文说明 `iot/{host_code}/{channel_code}` MQTT 数据如何进入 PostgreSQL 和 InfluxDB。

## PostgreSQL 表

新增三张表：

```text
iot_host
iot_channel
iot_channel_data
```

SQL 已追加到：

```text
internal/db/migrations/001_create_tables.sql
```

如果当前数据库还没有这三张表，需要先执行该 migration。

## MQTT Topic

后端默认订阅：

```text
iot/+/+
```

推荐 topic：

```text
iot/{host_code}/{channel_code}
```

示例：

```text
iot/host_001/temperature
iot/host_001/pressure
iot/host_002/status
```

## MQTT Payload

payload 使用 JSON：

```json
{
  "value": 23.5,
  "channel": "压力"
}
```

字段说明：

```text
value:   通道值，支持 number / string / bool
channel: 通道名称，找不到通道时用于自动创建 channel_name
quality: 可选，1 正常，0 异常，默认 1
ts:      可选，采集时间；没有时使用服务端当前时间
```

示例：

```json
{
  "value": 1.23,
  "channel": "压力",
  "quality": 1
}
```

## 自动创建规则

收到 MQTT 消息后：

```text
topic: iot/host_001/pressure
```

如果 `iot_host.host_code = host_001` 不存在，自动创建：

```text
host_code = host_001
host_name = host_001
protocol  = mqtt
status    = 1
```

如果该主机下 `iot_channel.channel_code = pressure` 不存在，自动创建：

```text
host_id      = 自动创建或查询到的 host id
channel_code = pressure
channel_name = payload.channel，如果为空则使用 pressure
data_type    = 根据 value 自动推断：float / bool / string
accuracy     = 2
status       = 1
```

解析后的实时数据写入：

```text
iot_channel_data
```

## InfluxDB 写入

所有 `iot/*` MQTT 原始数据会写入 InfluxDB。

环境变量：

```text
INFLUXDB_URL=http://localhost:8086
INFLUXDB_TOKEN=你的 token
INFLUXDB_ORG=zy
INFLUXDB_BUCKET=iot_data
INFLUXDB_MEASUREMENT=iot_mqtt_data
```

如果没有配置 `INFLUXDB_TOKEN`，后端会跳过 InfluxDB 写入，但 PostgreSQL 仍可写入。

InfluxDB tags：

```text
source=mqtt
topic=原始 MQTT topic
host_code=主机编码
channel_code=通道编码
```

InfluxDB fields：

```text
payload=原始 JSON payload
value=数值类型通道值
str_value=字符串类型通道值
bool_value=布尔类型通道值
quality=质量标记
```

## 管理接口

接口前缀：

```text
/api/v1
```

主机管理：

```text
GET    /iot/hosts
POST   /iot/hosts
PUT    /iot/hosts/:id
DELETE /iot/hosts/:id
```

通道管理：

```text
GET    /iot/channels
GET    /iot/channels?host_id=1
POST   /iot/channels
PUT    /iot/channels/:id
DELETE /iot/channels/:id
```

实时数据查询：

```text
GET /iot/channel-data
GET /iot/channel-data?host_id=1
GET /iot/channel-data?channel_id=1
GET /iot/channel-data?host_id=1&channel_id=1&limit=100
```

## 创建示例

手动创建主机：

```json
{
  "host_code": "host_001",
  "host_name": "一号主机",
  "ip": "127.0.0.1",
  "port": 11883,
  "protocol": "mqtt",
  "location": "实验室",
  "status": 1,
  "remark": "本地 MQTT 数据源"
}
```

手动创建通道：

```json
{
  "host_id": 1,
  "channel_code": "pressure",
  "channel_name": "压力",
  "data_type": "float",
  "unit": "MPa",
  "accuracy": 2,
  "status": 1
}
```
