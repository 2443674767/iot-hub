package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type frameMessage struct {
	DeviceID string `json:"device_id"`
	Channel  string `json:"channel"`
	CANID    string `json:"can_id"`
	Data     string `json:"data"`
}

func main() {
	var (
		broker   = flag.String("broker", getenv("MQTT_BROKER", "tcp://127.0.0.1:11883"), "MQTT broker address")
		clientID = flag.String("client-id", getenv("MQTT_CLIENT_ID", "mqtt-demo-publisher"), "MQTT client id")
		deviceID = flag.String("device", getenv("MQTT_DEVICE_ID", "left-thruster"), "device id")
		channel  = flag.String("channel", getenv("MQTT_CHANNEL", "telemetry"), "channel name")
		canID    = flag.String("can-id", getenv("MQTT_CAN_ID", "0xA1"), "CAN id")
		data     = flag.String("data", getenv("MQTT_DATA", "01000A0000000000"), "8-byte CAN frame data in hex")
		qos      = flag.Int("qos", getenvInt("MQTT_QOS", 1), "MQTT qos")
		count    = flag.Int("count", getenvInt("MQTT_COUNT", 1), "message count, 0 means forever")
		interval = flag.Duration("interval", getenvDuration("MQTT_INTERVAL", time.Second), "publish interval")
	)
	flag.Parse()

	topic := fmt.Sprintf("can/devices/%s/channels/%s/frames", sanitizeTopicPart(*deviceID), sanitizeTopicPart(*channel))
	payload, err := json.Marshal(frameMessage{
		DeviceID: *deviceID,
		Channel:  *channel,
		CANID:    *canID,
		Data:     strings.ToUpper(strings.ReplaceAll(*data, " ", "")),
	})
	if err != nil {
		log.Fatalf("marshal payload: %v", err)
	}

	opts := mqtt.NewClientOptions().
		AddBroker(*broker).
		SetClientID(*clientID).
		SetConnectTimeout(5 * time.Second)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("connect mqtt broker %s: %v", *broker, token.Error())
	}
	defer client.Disconnect(250)

	publishCount := 0
	for {
		token := client.Publish(topic, byte(*qos), false, payload)
		token.Wait()
		if token.Error() != nil {
			log.Fatalf("publish topic %s: %v", topic, token.Error())
		}

		publishCount++
		log.Printf("published #%d topic=%s payload=%s", publishCount, topic, payload)
		if *count > 0 && publishCount >= *count {
			return
		}
		time.Sleep(*interval)
	}
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
	var parsed int
	if _, err := fmt.Sscanf(value, "%d", &parsed); err != nil {
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

func sanitizeTopicPart(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, "/")
	if value == "" {
		return "unknown"
	}
	return value
}
