package repository

import "github.com/user/can-server/internal/db"

type DeviceRepo struct{}

func (r *DeviceRepo) GetAll() ([]map[string]any, error) {
	rows, err := db.DB.Query("SELECT id, name, can_id FROM devices ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []map[string]any
	for rows.Next() {
		var id int
		var name string
		var canID []byte
		if err := rows.Scan(&id, &name, &canID); err != nil {
			return nil, err
		}
		devices = append(devices, map[string]any{
			"id":     id,
			"name":   name,
			"can_id": canID,
		})
	}
	return devices, nil
}
