package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/user/can-server/internal/db"
	"github.com/user/can-server/internal/model"
)

type IoTHostRepo struct{}
type IoTChannelRepo struct{}
type IoTChannelDataRepo struct{}

func (r *IoTHostRepo) GetAll() ([]model.IoTHost, error) {
	rows, err := db.DB.Query(`
		SELECT id, host_code, host_name, COALESCE(ip, ''), port, COALESCE(protocol, ''),
		       COALESCE(location, ''), COALESCE(status, 1), COALESCE(remark, ''), created_at, updated_at
		FROM iot_host
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hosts []model.IoTHost
	for rows.Next() {
		host, err := scanIoTHost(rows)
		if err != nil {
			return nil, err
		}
		hosts = append(hosts, host)
	}
	return hosts, rows.Err()
}

func (r *IoTHostRepo) Create(host model.IoTHost) (*model.IoTHost, error) {
	return scanIoTHostRow(db.DB.QueryRow(`
		INSERT INTO iot_host (host_code, host_name, ip, port, protocol, location, status, remark)
		VALUES ($1, $2, NULLIF($3, ''), $4, NULLIF($5, ''), NULLIF($6, ''), $7, NULLIF($8, ''))
		RETURNING id, host_code, host_name, COALESCE(ip, ''), port, COALESCE(protocol, ''),
		          COALESCE(location, ''), COALESCE(status, 1), COALESCE(remark, ''), created_at, updated_at
	`, host.HostCode, host.HostName, host.IP, host.Port, host.Protocol, host.Location, defaultStatus(host.Status), host.Remark))
}

func (r *IoTHostRepo) Update(id int64, host model.IoTHost) (*model.IoTHost, error) {
	saved, err := scanIoTHostRow(db.DB.QueryRow(`
		UPDATE iot_host
		SET host_code = $1,
		    host_name = $2,
		    ip = NULLIF($3, ''),
		    port = $4,
		    protocol = NULLIF($5, ''),
		    location = NULLIF($6, ''),
		    status = $7,
		    remark = NULLIF($8, ''),
		    updated_at = NOW()
		WHERE id = $9
		RETURNING id, host_code, host_name, COALESCE(ip, ''), port, COALESCE(protocol, ''),
		          COALESCE(location, ''), COALESCE(status, 1), COALESCE(remark, ''), created_at, updated_at
	`, host.HostCode, host.HostName, host.IP, host.Port, host.Protocol, host.Location, defaultStatus(host.Status), host.Remark, id))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("iot host not found: %d", id)
	}
	return saved, err
}

func (r *IoTHostRepo) Delete(id int64) error {
	result, err := db.DB.Exec("DELETE FROM iot_host WHERE id = $1", id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("iot host not found: %d", id)
	}
	return nil
}

func (r *IoTHostRepo) GetOrCreateByCode(hostCode string) (*model.IoTHost, error) {
	hostCode = strings.TrimSpace(hostCode)
	host, err := scanIoTHostRow(db.DB.QueryRow(`
		SELECT id, host_code, host_name, COALESCE(ip, ''), port, COALESCE(protocol, ''),
		       COALESCE(location, ''), COALESCE(status, 1), COALESCE(remark, ''), created_at, updated_at
		FROM iot_host
		WHERE host_code = $1
	`, hostCode))
	if err == nil {
		return host, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}
	return scanIoTHostRow(db.DB.QueryRow(`
		INSERT INTO iot_host (host_code, host_name, protocol, status)
		VALUES ($1, $1, 'mqtt', 1)
		ON CONFLICT (host_code) DO UPDATE SET updated_at = iot_host.updated_at
		RETURNING id, host_code, host_name, COALESCE(ip, ''), port, COALESCE(protocol, ''),
		          COALESCE(location, ''), COALESCE(status, 1), COALESCE(remark, ''), created_at, updated_at
	`, hostCode))
}

func (r *IoTChannelRepo) GetAll(hostID int64) ([]model.IoTChannel, error) {
	query := `
		SELECT id, host_id, channel_code, channel_name, COALESCE(data_type, ''), COALESCE(unit, ''),
		       COALESCE(accuracy, 2), min_value, max_value, COALESCE(status, 1), COALESCE(remark, ''),
		       created_at, updated_at
		FROM iot_channel
	`
	args := []any{}
	if hostID > 0 {
		query += " WHERE host_id = $1"
		args = append(args, hostID)
	}
	query += " ORDER BY id"
	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []model.IoTChannel
	for rows.Next() {
		channel, err := scanIoTChannel(rows)
		if err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}
	return channels, rows.Err()
}

func (r *IoTChannelRepo) Create(channel model.IoTChannel) (*model.IoTChannel, error) {
	return scanIoTChannelRow(db.DB.QueryRow(`
		INSERT INTO iot_channel (host_id, channel_code, channel_name, data_type, unit, accuracy, min_value, max_value, status, remark)
		VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), $6, $7, $8, $9, NULLIF($10, ''))
		RETURNING id, host_id, channel_code, channel_name, COALESCE(data_type, ''), COALESCE(unit, ''),
		          COALESCE(accuracy, 2), min_value, max_value, COALESCE(status, 1), COALESCE(remark, ''),
		          created_at, updated_at
	`, channel.HostID, channel.ChannelCode, channel.ChannelName, channel.DataType, channel.Unit,
		defaultAccuracy(channel.Accuracy), channel.MinValue, channel.MaxValue, defaultStatus(channel.Status), channel.Remark))
}

func (r *IoTChannelRepo) Update(id int64, channel model.IoTChannel) (*model.IoTChannel, error) {
	saved, err := scanIoTChannelRow(db.DB.QueryRow(`
		UPDATE iot_channel
		SET host_id = $1,
		    channel_code = $2,
		    channel_name = $3,
		    data_type = NULLIF($4, ''),
		    unit = NULLIF($5, ''),
		    accuracy = $6,
		    min_value = $7,
		    max_value = $8,
		    status = $9,
		    remark = NULLIF($10, ''),
		    updated_at = NOW()
		WHERE id = $11
		RETURNING id, host_id, channel_code, channel_name, COALESCE(data_type, ''), COALESCE(unit, ''),
		          COALESCE(accuracy, 2), min_value, max_value, COALESCE(status, 1), COALESCE(remark, ''),
		          created_at, updated_at
	`, channel.HostID, channel.ChannelCode, channel.ChannelName, channel.DataType, channel.Unit,
		defaultAccuracy(channel.Accuracy), channel.MinValue, channel.MaxValue, defaultStatus(channel.Status), channel.Remark, id))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("iot channel not found: %d", id)
	}
	return saved, err
}

func (r *IoTChannelRepo) Delete(id int64) error {
	result, err := db.DB.Exec("DELETE FROM iot_channel WHERE id = $1", id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("iot channel not found: %d", id)
	}
	return nil
}

func (r *IoTChannelRepo) GetOrCreate(hostID int64, channelCode, channelName, dataType string) (*model.IoTChannel, error) {
	channelCode = strings.TrimSpace(channelCode)
	channelName = strings.TrimSpace(channelName)
	if channelName == "" {
		channelName = channelCode
	}
	channel, err := scanIoTChannelRow(db.DB.QueryRow(`
		SELECT id, host_id, channel_code, channel_name, COALESCE(data_type, ''), COALESCE(unit, ''),
		       COALESCE(accuracy, 2), min_value, max_value, COALESCE(status, 1), COALESCE(remark, ''),
		       created_at, updated_at
		FROM iot_channel
		WHERE host_id = $1 AND channel_code = $2
	`, hostID, channelCode))
	if err == nil {
		return channel, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}
	return scanIoTChannelRow(db.DB.QueryRow(`
		INSERT INTO iot_channel (host_id, channel_code, channel_name, data_type, accuracy, status)
		VALUES ($1, $2, $3, NULLIF($4, ''), 2, 1)
		ON CONFLICT (host_id, channel_code) DO UPDATE SET updated_at = iot_channel.updated_at
		RETURNING id, host_id, channel_code, channel_name, COALESCE(data_type, ''), COALESCE(unit, ''),
		          COALESCE(accuracy, 2), min_value, max_value, COALESCE(status, 1), COALESCE(remark, ''),
		          created_at, updated_at
	`, hostID, channelCode, channelName, dataType))
}

func (r *IoTChannelDataRepo) Insert(data model.IoTChannelData) (*model.IoTChannelData, error) {
	if data.Ts.IsZero() {
		data.Ts = time.Now()
	}
	return scanIoTChannelDataRow(db.DB.QueryRow(`
		INSERT INTO iot_channel_data (host_id, channel_id, value, str_value, bool_value, quality, ts)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, host_id, channel_id, value, str_value, bool_value, COALESCE(quality, 1), ts, created_at
	`, data.HostID, data.ChannelID, data.Value, data.StrValue, data.BoolValue, defaultStatus(data.Quality), data.Ts))
}

func (r *IoTChannelDataRepo) GetAll(hostID int64, channelID int64, limit int) ([]model.IoTChannelData, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	query := `
		SELECT d.id, d.host_id, d.channel_id, d.value, d.str_value, d.bool_value, COALESCE(d.quality, 1),
		       d.ts, d.created_at, h.host_code, c.channel_code
		FROM iot_channel_data d
		JOIN iot_host h ON h.id = d.host_id
		JOIN iot_channel c ON c.id = d.channel_id
		WHERE ($1::BIGINT = 0 OR d.host_id = $1)
		  AND ($2::BIGINT = 0 OR d.channel_id = $2)
		ORDER BY d.ts DESC, d.id DESC
		LIMIT $3
	`
	rows, err := db.DB.Query(query, hostID, channelID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var values []model.IoTChannelData
	for rows.Next() {
		value, err := scanIoTChannelDataWithCodes(rows)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanIoTHost(scanner rowScanner) (model.IoTHost, error) {
	var host model.IoTHost
	var port sql.NullInt64
	err := scanner.Scan(&host.ID, &host.HostCode, &host.HostName, &host.IP, &port, &host.Protocol,
		&host.Location, &host.Status, &host.Remark, &host.CreatedAt, &host.UpdatedAt)
	if port.Valid {
		p := int(port.Int64)
		host.Port = &p
	}
	return host, err
}

func scanIoTHostRow(scanner rowScanner) (*model.IoTHost, error) {
	host, err := scanIoTHost(scanner)
	if err != nil {
		return nil, err
	}
	return &host, nil
}

func scanIoTChannel(scanner rowScanner) (model.IoTChannel, error) {
	var channel model.IoTChannel
	var minValue, maxValue sql.NullFloat64
	err := scanner.Scan(&channel.ID, &channel.HostID, &channel.ChannelCode, &channel.ChannelName,
		&channel.DataType, &channel.Unit, &channel.Accuracy, &minValue, &maxValue,
		&channel.Status, &channel.Remark, &channel.CreatedAt, &channel.UpdatedAt)
	if minValue.Valid {
		channel.MinValue = &minValue.Float64
	}
	if maxValue.Valid {
		channel.MaxValue = &maxValue.Float64
	}
	return channel, err
}

func scanIoTChannelRow(scanner rowScanner) (*model.IoTChannel, error) {
	channel, err := scanIoTChannel(scanner)
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

func scanIoTChannelData(scanner rowScanner) (*model.IoTChannelData, error) {
	var data model.IoTChannelData
	var value sql.NullFloat64
	var strValue sql.NullString
	var boolValue sql.NullBool
	err := scanner.Scan(&data.ID, &data.HostID, &data.ChannelID, &value, &strValue, &boolValue,
		&data.Quality, &data.Ts, &data.CreatedAt)
	if value.Valid {
		data.Value = &value.Float64
	}
	if strValue.Valid {
		data.StrValue = &strValue.String
	}
	if boolValue.Valid {
		data.BoolValue = &boolValue.Bool
	}
	return &data, err
}

func scanIoTChannelDataRow(scanner rowScanner) (*model.IoTChannelData, error) {
	return scanIoTChannelData(scanner)
}

func scanIoTChannelDataWithCodes(scanner rowScanner) (model.IoTChannelData, error) {
	var data model.IoTChannelData
	var value sql.NullFloat64
	var strValue sql.NullString
	var boolValue sql.NullBool
	err := scanner.Scan(&data.ID, &data.HostID, &data.ChannelID, &value, &strValue, &boolValue,
		&data.Quality, &data.Ts, &data.CreatedAt, &data.HostCode, &data.ChannelCode)
	if err != nil {
		return model.IoTChannelData{}, err
	}
	if value.Valid {
		data.Value = &value.Float64
	}
	if strValue.Valid {
		data.StrValue = &strValue.String
	}
	if boolValue.Valid {
		data.BoolValue = &boolValue.Bool
	}
	return data, nil
}

func defaultStatus(status int) int {
	if status == 0 {
		return 1
	}
	return status
}

func defaultAccuracy(accuracy int) int {
	if accuracy == 0 {
		return 2
	}
	return accuracy
}
