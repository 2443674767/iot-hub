package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

type config struct {
	MQTTBroker     string
	MQTTClientID   string
	MQTTTopic      string
	MQTTQOS        byte
	InfluxURL      string
	InfluxToken    string
	InfluxOrg      string
	InfluxBucket   string
	Measurement    string
	Mode           string
	QueryRange     time.Duration
	QueryLimit     int
	QueryDeviceID  string
	QueryChannel   string
	QueryCANID     string
	QueryField     string
	ConnectTimeout time.Duration
}

type framePayload struct {
	DeviceID string         `json:"device_id"`
	Channel  string         `json:"channel"`
	CANID    string         `json:"can_id"`
	Data     string         `json:"data"`
	Parsed   map[string]any `json:"parsed"`
}

type framePoint struct {
	Topic    string
	DeviceID string
	Channel  string
	CANID    string
	Device   string
	DataHex  string
	Data     [8]byte
	Parsed   map[string]any
}

func main() {
	cfg := loadConfig()
	if cfg.InfluxToken == "" {
		log.Fatal("missing INFLUXDB_TOKEN or -influx-token")
	}

	influxClient := influxdb2.NewClient(cfg.InfluxURL, cfg.InfluxToken)
	defer influxClient.Close()

	if cfg.Mode == "query" {
		if err := queryInfluxData(context.Background(), cfg, influxClient.QueryAPI(cfg.InfluxOrg)); err != nil {
			log.Fatal(err)
		}
		return
	}

	writeAPI := influxClient.WriteAPIBlocking(cfg.InfluxOrg, cfg.InfluxBucket)

	mqttClient := newMQTTClient(cfg, writeAPI)
	if token := mqttClient.Connect(); !token.WaitTimeout(cfg.ConnectTimeout) {
		log.Fatalf("connect mqtt broker timeout after %s", cfg.ConnectTimeout)
	} else if token.Error() != nil {
		log.Fatalf("connect mqtt broker %s: %v", cfg.MQTTBroker, token.Error())
	}
	defer mqttClient.Disconnect(250)

	log.Printf("mqtt connected broker=%s topic=%s", cfg.MQTTBroker, cfg.MQTTTopic)
	waitForShutdown()
}

func loadConfig() config {
	var cfg config
	flag.StringVar(&cfg.MQTTBroker, "mqtt-broker", getenv("MQTT_BROKER", "tcp://127.0.0.1:11883"), "MQTT broker address")
	flag.StringVar(&cfg.MQTTClientID, "mqtt-client-id", getenv("MQTT_CLIENT_ID", "influxdb-go-client"), "MQTT client id")
	flag.StringVar(&cfg.MQTTTopic, "mqtt-topic", getenv("MQTT_TOPIC", "can/devices/+/channels/+/frames"), "MQTT topic")
	flag.StringVar(&cfg.InfluxURL, "influx-url", getenv("INFLUXDB_URL", "http://localhost:8086"), "InfluxDB URL")
	flag.StringVar(&cfg.InfluxToken, "influx-token", getenv("INFLUXDB_TOKEN", "7OoXOhv2qw0iJ4VPScFIovb9zHbHrqd-tHuyyqmqg1n_mFN5HbA5UtuIX1QlOdcpikQj5l-JzQngNGVN3KIpcQ=="), "InfluxDB token")
	flag.StringVar(&cfg.InfluxOrg, "influx-org", getenv("INFLUXDB_ORG", "zy"), "InfluxDB org")
	flag.StringVar(&cfg.InfluxBucket, "influx-bucket", getenv("INFLUXDB_BUCKET", "iot_data"), "InfluxDB bucket")
	flag.StringVar(&cfg.Measurement, "measurement", getenv("INFLUXDB_MEASUREMENT", "can_frames"), "InfluxDB measurement")
	flag.StringVar(&cfg.Mode, "mode", getenv("MODE", "write"), "run mode: write or query")
	flag.DurationVar(&cfg.QueryRange, "query-range", getenvDuration("QUERY_RANGE", 10*time.Minute), "query time range")
	flag.IntVar(&cfg.QueryLimit, "query-limit", getenvInt("QUERY_LIMIT", 20), "query result limit")
	flag.StringVar(&cfg.QueryDeviceID, "query-device", getenv("QUERY_DEVICE_ID", ""), "query filter device_id")
	flag.StringVar(&cfg.QueryChannel, "query-channel", getenv("QUERY_CHANNEL", ""), "query filter channel")
	flag.StringVar(&cfg.QueryCANID, "query-can-id", getenv("QUERY_CAN_ID", ""), "query filter can_id")
	flag.StringVar(&cfg.QueryField, "query-field", getenv("QUERY_FIELD", "data"), "query field")
	qos := flag.Int("mqtt-qos", getenvInt("MQTT_QOS", 1), "MQTT QoS")
	flag.DurationVar(&cfg.ConnectTimeout, "connect-timeout", getenvDuration("CONNECT_TIMEOUT", 5*time.Second), "connect timeout")
	flag.Parse()

	cfg.Mode = strings.ToLower(strings.TrimSpace(cfg.Mode))
	if cfg.Mode != "write" && cfg.Mode != "query" {
		log.Fatalf("invalid mode %q, want write or query", cfg.Mode)
	}
	if *qos < 0 || *qos > 2 {
		log.Fatalf("invalid mqtt qos %d, want 0, 1, or 2", *qos)
	}
	cfg.MQTTQOS = byte(*qos)
	return cfg
}

func newMQTTClient(cfg config, writeAPI api.WriteAPIBlocking) mqtt.Client {
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.MQTTBroker).
		SetClientID(cfg.MQTTClientID).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectTimeout(cfg.ConnectTimeout)

	opts.OnConnect = func(client mqtt.Client) {
		log.Printf("subscribing mqtt topic=%s qos=%d", cfg.MQTTTopic, cfg.MQTTQOS)
		token := client.Subscribe(cfg.MQTTTopic, cfg.MQTTQOS, func(_ mqtt.Client, msg mqtt.Message) {
			handleMessage(cfg, writeAPI, msg)
		})
		if token.Wait() && token.Error() != nil {
			log.Printf("mqtt subscribe failed: %v", token.Error())
		}
	}
	opts.OnConnectionLost = func(_ mqtt.Client, err error) {
		log.Printf("mqtt connection lost: %v", err)
	}
	return mqtt.NewClient(opts)
}

func queryInfluxData(ctx context.Context, cfg config, queryAPI api.QueryAPI) error {
	query := buildFluxQuery(cfg)
	results, err := queryAPI.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("query influxdb: %w", err)
	}
	defer results.Close()

	count := 0
	fmt.Printf("%-30s %-16s %-16s %-8s %-20s %-20s\n", "time", "device_id", "channel", "can_id", "field", "value")
	for results.Next() {
		record := results.Record()
		fmt.Printf("%-30s %-16s %-16s %-8s %-20s %-20v\n",
			record.Time().Format(time.RFC3339),
			stringValue(record.ValueByKey("device_id")),
			stringValue(record.ValueByKey("channel")),
			stringValue(record.ValueByKey("can_id")),
			record.Field(),
			record.Value(),
		)
		count++
	}
	if err := results.Err(); err != nil {
		return fmt.Errorf("read query result: %w", err)
	}
	log.Printf("query completed, rows=%d", count)
	return nil
}

func buildFluxQuery(cfg config) string {
	var b strings.Builder
	start := fmt.Sprintf("-%ds", int(cfg.QueryRange.Seconds()))
	if cfg.QueryRange <= 0 {
		start = "-10m"
	}
	limit := cfg.QueryLimit
	if limit <= 0 {
		limit = 20
	}

	fmt.Fprintf(&b, "from(bucket: %q)\n", cfg.InfluxBucket)
	fmt.Fprintf(&b, "  |> range(start: %s)\n", start)
	fmt.Fprintf(&b, "  |> filter(fn: (r) => r._measurement == %q)\n", cfg.Measurement)
	if cfg.QueryField != "" {
		fmt.Fprintf(&b, "  |> filter(fn: (r) => r._field == %q)\n", cfg.QueryField)
	}
	if cfg.QueryDeviceID != "" {
		fmt.Fprintf(&b, "  |> filter(fn: (r) => r.device_id == %q)\n", cfg.QueryDeviceID)
	}
	if cfg.QueryChannel != "" {
		fmt.Fprintf(&b, "  |> filter(fn: (r) => r.channel == %q)\n", cfg.QueryChannel)
	}
	if cfg.QueryCANID != "" {
		fmt.Fprintf(&b, "  |> filter(fn: (r) => r.can_id == %q)\n", normalizeCANID(cfg.QueryCANID))
	}
	fmt.Fprintln(&b, `  |> sort(columns: ["_time"], desc: true)`)
	fmt.Fprintf(&b, "  |> limit(n: %d)\n", limit)
	return b.String()
}

func stringValue(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func handleMessage(cfg config, writeAPI api.WriteAPIBlocking, msg mqtt.Message) {
	point, err := decodeFrameMessage(msg.Topic(), msg.Payload())
	if err != nil {
		log.Printf("ignore mqtt message topic=%s err=%v", msg.Topic(), err)
		return
	}

	influxPoint := buildInfluxPoint(cfg.Measurement, point)
	if err := writeAPI.WritePoint(context.Background(), influxPoint); err != nil {
		log.Printf("write influxdb failed topic=%s err=%v", msg.Topic(), err)
		return
	}
	log.Printf("wrote influxdb measurement=%s device=%s channel=%s can_id=%s data=%s", cfg.Measurement, point.DeviceID, point.Channel, point.CANID, point.DataHex)
}

func decodeFrameMessage(topic string, raw []byte) (framePoint, error) {
	var payload framePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return framePoint{}, fmt.Errorf("decode json: %w", err)
	}

	deviceID, channel := parseDeviceChannel(topic)
	if payload.DeviceID != "" {
		deviceID = payload.DeviceID
	}
	if payload.Channel != "" {
		channel = payload.Channel
	}
	if payload.CANID == "" {
		return framePoint{}, fmt.Errorf("missing can_id")
	}
	if payload.Data == "" {
		return framePoint{}, fmt.Errorf("missing data")
	}

	data, dataHex, err := parseCANData(payload.Data)
	if err != nil {
		return framePoint{}, err
	}
	canID := normalizeCANID(payload.CANID)
	parsed := payload.Parsed
	if parsed == nil {
		parsed = parseCANFrame(canID, data)
	}

	return framePoint{
		Topic:    topic,
		DeviceID: deviceID,
		Channel:  channel,
		CANID:    canID,
		Device:   deviceName(canID),
		DataHex:  dataHex,
		Data:     data,
		Parsed:   parsed,
	}, nil
}

func buildInfluxPoint(measurement string, frame framePoint) *write.Point {
	tags := map[string]string{
		"source": "mqtt",
		"topic":  frame.Topic,
		"can_id": frame.CANID,
	}
	if frame.DeviceID != "" {
		tags["device_id"] = frame.DeviceID
	}
	if frame.Channel != "" {
		tags["channel"] = frame.Channel
	}
	if frame.Device != "" {
		tags["device"] = frame.Device
	}

	fields := map[string]any{
		"data":      frame.DataHex,
		"direction": 0,
	}
	for i, value := range frame.Data {
		fields[fmt.Sprintf("byte_%d", i)] = int(value)
	}
	for key, value := range frame.Parsed {
		fields["parsed_"+key] = value
	}

	return write.NewPoint(measurement, tags, fields, time.Now())
}

func parseDeviceChannel(topic string) (string, string) {
	parts := strings.Split(topic, "/")
	deviceID := ""
	channel := ""
	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == "devices" && i+1 < len(parts) {
			deviceID = parts[i+1]
		}
		if parts[i] == "channels" && i+1 < len(parts) {
			channel = parts[i+1]
		}
	}
	return deviceID, channel
}

func parseCANData(raw string) ([8]byte, string, error) {
	var data [8]byte
	value := strings.ReplaceAll(strings.TrimSpace(raw), " ", "")
	value = strings.TrimPrefix(strings.TrimPrefix(value, "0x"), "0X")
	decoded, err := hex.DecodeString(value)
	if err != nil {
		return data, "", fmt.Errorf("invalid data hex: %w", err)
	}
	if len(decoded) != 8 {
		return data, "", fmt.Errorf("data must be 8 bytes")
	}
	copy(data[:], decoded)
	return data, strings.ToLower(value), nil
}

func normalizeCANID(raw string) string {
	value := strings.TrimSpace(raw)
	value = strings.TrimPrefix(strings.TrimPrefix(value, "0x"), "0X")
	n, err := strconv.ParseUint(value, 16, 8)
	if err != nil {
		return raw
	}
	return fmt.Sprintf("0x%02X", n)
}

func parseCANFrame(canID string, data [8]byte) map[string]any {
	switch canID {
	case "0xA1", "0xA2":
		return map[string]any{
			"gear": int(data[0]),
			"rpm":  int(data[1])<<8 | int(data[2]),
		}
	case "0xB1":
		directions := map[byte]string{
			0x00: "中间",
			0x01: "右边",
			0x02: "左边",
		}
		direction := "未知"
		if value, ok := directions[data[0]]; ok {
			direction = value
		}
		return map[string]any{
			"direction": direction,
			"angle":     int(data[1]) | int(data[2])<<8,
		}
	default:
		return map[string]any{}
	}
}

func deviceName(canID string) string {
	switch canID {
	case "0xA1":
		return "左推进器"
	case "0xA2":
		return "右推进器"
	case "0xB1":
		return "转向舵"
	case "0xC1":
		return "三联屏"
	case "0xD2":
		return "氛围灯"
	default:
		return ""
	}
}

func waitForShutdown() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("shutdown")
}

func getenv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getenvDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}
