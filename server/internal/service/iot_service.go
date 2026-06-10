package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/user/can-server/internal/influx"
	"github.com/user/can-server/internal/model"
)

type IoTHostRepository interface {
	GetAll() ([]model.IoTHost, error)
	Create(model.IoTHost) (*model.IoTHost, error)
	Update(id int64, host model.IoTHost) (*model.IoTHost, error)
	Delete(id int64) error
	GetOrCreateByCode(hostCode string) (*model.IoTHost, error)
}

type IoTChannelRepository interface {
	GetAll(hostID int64) ([]model.IoTChannel, error)
	Create(model.IoTChannel) (*model.IoTChannel, error)
	Update(id int64, channel model.IoTChannel) (*model.IoTChannel, error)
	Delete(id int64) error
	GetOrCreate(hostID int64, channelCode, channelName, dataType string) (*model.IoTChannel, error)
}

type IoTChannelDataRepository interface {
	Insert(model.IoTChannelData) (*model.IoTChannelData, error)
	GetAll(hostID int64, channelID int64, limit int) ([]model.IoTChannelData, error)
}

type MQTTInfluxWriter interface {
	WriteMQTT(ctx context.Context, record influx.MQTTRecord) error
}

type IoTService struct {
	hosts    IoTHostRepository
	channels IoTChannelRepository
	data     IoTChannelDataRepository
	influx   MQTTInfluxWriter
}

type IoTMQTTPayload struct {
	Channel   string `json:"channel"`
	Value     any    `json:"value"`
	Quality   int    `json:"quality"`
	Timestamp string `json:"ts"`
}

func NewIoTService(hosts IoTHostRepository, channels IoTChannelRepository, data IoTChannelDataRepository, influx MQTTInfluxWriter) *IoTService {
	return &IoTService{hosts: hosts, channels: channels, data: data, influx: influx}
}

func (s *IoTService) ListHosts() ([]model.IoTHost, error) {
	return s.hosts.GetAll()
}

func (s *IoTService) CreateHost(host model.IoTHost) (*model.IoTHost, error) {
	if err := validateIoTHost(host); err != nil {
		return nil, err
	}
	return s.hosts.Create(host)
}

func (s *IoTService) UpdateHost(id int64, host model.IoTHost) (*model.IoTHost, error) {
	if err := validateIoTHost(host); err != nil {
		return nil, err
	}
	return s.hosts.Update(id, host)
}

func (s *IoTService) DeleteHost(id int64) error {
	return s.hosts.Delete(id)
}

func (s *IoTService) ListChannels(hostID int64) ([]model.IoTChannel, error) {
	return s.channels.GetAll(hostID)
}

func (s *IoTService) CreateChannel(channel model.IoTChannel) (*model.IoTChannel, error) {
	if err := validateIoTChannel(channel); err != nil {
		return nil, err
	}
	return s.channels.Create(channel)
}

func (s *IoTService) UpdateChannel(id int64, channel model.IoTChannel) (*model.IoTChannel, error) {
	if err := validateIoTChannel(channel); err != nil {
		return nil, err
	}
	return s.channels.Update(id, channel)
}

func (s *IoTService) DeleteChannel(id int64) error {
	return s.channels.Delete(id)
}

func (s *IoTService) ListChannelData(hostID int64, channelID int64, limit int) ([]model.IoTChannelData, error) {
	return s.data.GetAll(hostID, channelID, limit)
}

func (s *IoTService) IngestMQTT(topic string, raw []byte) (*model.IoTChannelData, error) {
	hostCode, channelCode, ok := parseIoTTopic(topic)
	if !ok {
		return nil, fmt.Errorf("invalid iot mqtt topic: %s", topic)
	}

	var payload IoTMQTTPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		_ = s.writeInflux(topic, hostCode, channelCode, string(raw), parsedValue{}, 0, time.Now())
		return nil, fmt.Errorf("decode iot mqtt payload: %w", err)
	}
	value, dataType, err := parsePayloadValue(payload.Value)
	if err != nil {
		_ = s.writeInflux(topic, hostCode, channelCode, string(raw), value, defaultQuality(payload.Quality), time.Now())
		return nil, err
	}
	ts := parsePayloadTime(payload.Timestamp)
	quality := defaultQuality(payload.Quality)

	if err := s.writeInflux(topic, hostCode, channelCode, string(raw), value, quality, ts); err != nil {
		return nil, fmt.Errorf("write influxdb: %w", err)
	}

	host, err := s.hosts.GetOrCreateByCode(hostCode)
	if err != nil {
		return nil, fmt.Errorf("get or create iot host: %w", err)
	}
	channel, err := s.channels.GetOrCreate(host.ID, channelCode, payload.Channel, dataType)
	if err != nil {
		return nil, fmt.Errorf("get or create iot channel: %w", err)
	}

	return s.data.Insert(model.IoTChannelData{
		HostID:    host.ID,
		ChannelID: channel.ID,
		Value:     value.FloatValue,
		StrValue:  value.StringValue,
		BoolValue: value.BoolValue,
		Quality:   quality,
		Ts:        ts,
	})
}

func (s *IoTService) writeInflux(topic string, hostCode string, channelCode string, raw string, value parsedValue, quality int, ts time.Time) error {
	if s.influx == nil {
		return nil
	}
	return s.influx.WriteMQTT(context.Background(), influx.MQTTRecord{
		Topic:       topic,
		HostCode:    hostCode,
		ChannelCode: channelCode,
		Payload:     raw,
		Value:       value.FloatValue,
		StrValue:    value.StringValue,
		BoolValue:   value.BoolValue,
		Quality:     quality,
		ReceivedAt:  ts,
	})
}

type parsedValue struct {
	FloatValue  *float64
	StringValue *string
	BoolValue   *bool
}

func parsePayloadValue(value any) (parsedValue, string, error) {
	switch v := value.(type) {
	case nil:
		return parsedValue{}, "", fmt.Errorf("value is required")
	case float64:
		return parsedValue{FloatValue: &v}, "float", nil
	case bool:
		return parsedValue{BoolValue: &v}, "bool", nil
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return parsedValue{FloatValue: &f}, "float", nil
		}
		return parsedValue{StringValue: &v}, "string", nil
	default:
		str := fmt.Sprint(v)
		return parsedValue{StringValue: &str}, "string", nil
	}
}

func parseIoTTopic(topic string) (string, string, bool) {
	parts := strings.Split(strings.Trim(topic, "/"), "/")
	if len(parts) != 3 || parts[0] != "iot" {
		return "", "", false
	}
	hostCode := strings.TrimSpace(parts[1])
	channelCode := strings.TrimSpace(parts[2])
	return hostCode, channelCode, hostCode != "" && channelCode != ""
}

func parsePayloadTime(raw string) time.Time {
	if strings.TrimSpace(raw) == "" {
		return time.Now()
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05"} {
		if ts, err := time.Parse(layout, raw); err == nil {
			return ts
		}
	}
	return time.Now()
}

func defaultQuality(quality int) int {
	if quality == 0 {
		return 1
	}
	return quality
}

func validateIoTHost(host model.IoTHost) error {
	if strings.TrimSpace(host.HostCode) == "" {
		return fmt.Errorf("host_code is required")
	}
	if strings.TrimSpace(host.HostName) == "" {
		return fmt.Errorf("host_name is required")
	}
	return nil
}

func validateIoTChannel(channel model.IoTChannel) error {
	if channel.HostID <= 0 {
		return fmt.Errorf("host_id is required")
	}
	if strings.TrimSpace(channel.ChannelCode) == "" {
		return fmt.Errorf("channel_code is required")
	}
	if strings.TrimSpace(channel.ChannelName) == "" {
		return fmt.Errorf("channel_name is required")
	}
	return nil
}
