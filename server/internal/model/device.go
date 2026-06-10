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
