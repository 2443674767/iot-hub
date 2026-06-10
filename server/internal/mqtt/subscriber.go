package mqtt

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/user/can-server/config"
	"github.com/user/can-server/internal/service"
)

type Subscriber struct {
	cfg    config.MQTTConfig
	frames *service.CANFrameService
	iot    *service.IoTService
	client paho.Client
}

type framePayload struct {
	DeviceID string `json:"device_id"`
	Channel  string `json:"channel"`
	CANID    string `json:"can_id"`
	Data     string `json:"data"`
}

func NewSubscriber(cfg config.MQTTConfig, frames *service.CANFrameService, iot *service.IoTService) *Subscriber {
	return &Subscriber{cfg: cfg, frames: frames, iot: iot}
}

func (s *Subscriber) Start() error {
	if !s.cfg.Enabled {
		log.Println("mqtt disabled")
		return nil
	}

	opts := paho.NewClientOptions().
		AddBroker(s.cfg.Broker).
		SetClientID(s.cfg.ClientID).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectTimeout(s.cfg.ConnectTimeout())
	if s.cfg.Username != "" {
		opts.SetUsername(s.cfg.Username)
		opts.SetPassword(s.cfg.Password)
	}
	opts.OnConnect = func(client paho.Client) {
		topics := map[string]byte{s.cfg.Topic: s.cfg.QOS}
		if s.cfg.IOTTopic != "" {
			topics[s.cfg.IOTTopic] = s.cfg.QOS
		}
		log.Printf("mqtt connected, subscribing topics=%v", topics)
		if token := client.SubscribeMultiple(topics, s.handleMessage); token.Wait() && token.Error() != nil {
			log.Printf("mqtt subscribe failed: %v", token.Error())
		}
	}
	opts.OnConnectionLost = func(_ paho.Client, err error) {
		log.Printf("mqtt connection lost: %v", err)
	}

	s.client = paho.NewClient(opts)
	token := s.client.Connect()
	if !token.WaitTimeout(s.cfg.ConnectTimeout()) {
		return fmt.Errorf("connect timeout after %s", s.cfg.ConnectTimeout().Round(time.Millisecond))
	}
	if token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (s *Subscriber) handleMessage(_ paho.Client, msg paho.Message) {
	if strings.HasPrefix(msg.Topic(), "iot/") {
		if s.iot == nil {
			log.Printf("iot mqtt message ignored, service not configured topic=%s", msg.Topic())
			return
		}
		if _, err := s.iot.IngestMQTT(msg.Topic(), msg.Payload()); err != nil {
			log.Printf("iot mqtt ingest failed, topic=%s err=%v", msg.Topic(), err)
		}
		return
	}

	payload, err := decodeFramePayload(msg.Topic(), msg.Payload())
	if err != nil {
		log.Printf("mqtt message ignored, topic=%s err=%v", msg.Topic(), err)
		return
	}

	_, err = s.frames.ReceiveFrame(payload.CANID, payload.Data, "mqtt", service.CANFrameMetadata{
		DeviceID: payload.DeviceID,
		Channel:  payload.Channel,
	})
	if err != nil {
		log.Printf("mqtt frame ingest failed, topic=%s err=%v", msg.Topic(), err)
	}
}

func decodeFramePayload(topic string, raw []byte) (*framePayload, error) {
	var payload framePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}
	deviceID, channel := parseDeviceChannel(topic)
	if payload.DeviceID == "" {
		payload.DeviceID = deviceID
	}
	if payload.Channel == "" {
		payload.Channel = channel
	}
	if payload.CANID == "" {
		return nil, fmt.Errorf("missing can_id")
	}
	if payload.Data == "" {
		return nil, fmt.Errorf("missing data")
	}
	return &payload, nil
}

func parseDeviceChannel(topic string) (string, string) {
	parts := strings.Split(topic, "/")
	for i := 0; i < len(parts)-1; i++ {
		switch parts[i] {
		case "devices":
			if i+1 < len(parts) {
				deviceID := parts[i+1]
				channel := ""
				for j := i + 2; j < len(parts)-1; j++ {
					if parts[j] == "channels" && j+1 < len(parts) {
						channel = parts[j+1]
						break
					}
				}
				return deviceID, channel
			}
		}
	}
	return "", ""
}
