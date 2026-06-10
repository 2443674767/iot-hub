package model

import "time"

type Device struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	CANID       byte      `json:"can_id"`
	Enabled     bool      `json:"enabled"`
	Status      string    `json:"status"`
	CategoryID  *int64    `json:"category_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type DeviceCategory struct {
	ID        int64     `json:"id"`
	ParentID  *int64    `json:"parent_id"`
	Model     string    `json:"model"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CANMessage struct {
	ID         int64     `json:"id"`
	CANID      string    `json:"can_id"`
	Data       string    `json:"data"`
	Direction  int       `json:"direction"` // 0: 上行, 1: 下行
	ReceivedAt time.Time `json:"received_at"`
}

type RawCANData struct {
	ID         int64          `json:"id"`
	CANID      byte           `json:"can_id"`
	Direction  int            `json:"direction"`
	ReadAt     time.Time      `json:"read_at"`
	RawFrame   map[string]any `json:"raw_frame"`
	ParsedData map[string]any `json:"parsed_data"`
}

type TCPConfig struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type IoTHost struct {
	ID        int64     `json:"id"`
	HostCode  string    `json:"host_code"`
	HostName  string    `json:"host_name"`
	IP        string    `json:"ip"`
	Port      *int      `json:"port"`
	Protocol  string    `json:"protocol"`
	Location  string    `json:"location"`
	Status    int       `json:"status"`
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type IoTChannel struct {
	ID          int64     `json:"id"`
	HostID      int64     `json:"host_id"`
	ChannelCode string    `json:"channel_code"`
	ChannelName string    `json:"channel_name"`
	DataType    string    `json:"data_type"`
	Unit        string    `json:"unit"`
	Accuracy    int       `json:"accuracy"`
	MinValue    *float64  `json:"min_value"`
	MaxValue    *float64  `json:"max_value"`
	Status      int       `json:"status"`
	Remark      string    `json:"remark"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type IoTChannelData struct {
	ID          int64     `json:"id"`
	HostID      int64     `json:"host_id"`
	ChannelID   int64     `json:"channel_id"`
	Value       *float64  `json:"value"`
	StrValue    *string   `json:"str_value"`
	BoolValue   *bool     `json:"bool_value"`
	Quality     int       `json:"quality"`
	Ts          time.Time `json:"ts"`
	CreatedAt   time.Time `json:"created_at"`
	HostCode    string    `json:"host_code,omitempty"`
	ChannelCode string    `json:"channel_code,omitempty"`
}
