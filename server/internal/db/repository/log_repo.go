package repository

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/user/can-server/internal/can"
	"github.com/user/can-server/internal/db"
)

type LogRepo struct{}

func (r *LogRepo) Insert(canID byte, data [8]byte, direction int) error {
	return r.recordCANMessage(canID, data, direction)
}

func (r *LogRepo) RecordReceivedFrame(canID byte, data [8]byte) error {
	return r.recordCANMessage(canID, data, int(can.DirControllerToPC))
}

func (r *LogRepo) RecordSentFrame(canID byte, data [8]byte) error {
	return r.RecordFrontendFrame(canID, data)
}

func (r *LogRepo) RecordFrontendFrame(canID byte, data [8]byte) error {
	parsed := can.Parse(canID, data)
	parsedData, err := json.Marshal(parsed.Parsed)
	if err != nil {
		return fmt.Errorf("marshal parsed can data: %w", err)
	}
	rawFrameData, err := marshalRawFrame(canID, data)
	if err != nil {
		return fmt.Errorf("marshal raw can frame: %w", err)
	}

	_, err = db.DB.Exec(
		"INSERT INTO raw_can_data (can_id, direction, raw_frame, parsed_data) VALUES ($1, $2, $3, $4)",
		[]byte{canID}, int(can.DirPCToController), rawFrameData, parsedData,
	)
	return err
}

func (r *LogRepo) recordCANMessage(canID byte, data [8]byte, direction int) error {
	_, err := db.DB.Exec(
		"INSERT INTO can_messages (can_id, data, direction) VALUES ($1, $2, $3)",
		formatCANID(canID), hex.EncodeToString(data[:]), direction,
	)
	return err
}

func marshalRawFrame(canID byte, data [8]byte) ([]byte, error) {
	rawFrame := map[string]string{
		"canid": formatCANID(canID),
		"data":  hex.EncodeToString(data[:]),
	}
	return json.Marshal(rawFrame)
}

func formatCANID(canID byte) string {
	return fmt.Sprintf("0x%02X", canID)
}
