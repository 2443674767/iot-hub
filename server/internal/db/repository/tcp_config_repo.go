package repository

import (
	"database/sql"
	"fmt"

	"github.com/user/can-server/internal/db"
	"github.com/user/can-server/internal/model"
)

type TCPConfigRepo struct{}

func (r *TCPConfigRepo) GetAll() ([]model.TCPConfig, error) {
	rows, err := db.DB.Query(`
		SELECT id, name, host, port, enabled, created_at, updated_at
		FROM tcp_configs
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []model.TCPConfig
	for rows.Next() {
		var cfg model.TCPConfig
		if err := rows.Scan(&cfg.ID, &cfg.Name, &cfg.Host, &cfg.Port, &cfg.Enabled, &cfg.CreatedAt, &cfg.UpdatedAt); err != nil {
			return nil, err
		}
		configs = append(configs, cfg)
	}
	return configs, rows.Err()
}

func (r *TCPConfigRepo) GetActive() (*model.TCPConfig, error) {
	var cfg model.TCPConfig
	err := db.DB.QueryRow(`
		SELECT id, name, host, port, enabled, created_at, updated_at
		FROM tcp_configs
		WHERE enabled = TRUE
		ORDER BY updated_at DESC, id DESC
		LIMIT 1
	`).Scan(&cfg.ID, &cfg.Name, &cfg.Host, &cfg.Port, &cfg.Enabled, &cfg.CreatedAt, &cfg.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no enabled tcp config")
	}
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (r *TCPConfigRepo) Create(cfg model.TCPConfig) (*model.TCPConfig, error) {
	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if cfg.Enabled {
		if _, err := tx.Exec("UPDATE tcp_configs SET enabled = FALSE, updated_at = NOW() WHERE enabled = TRUE"); err != nil {
			return nil, err
		}
	}

	var saved model.TCPConfig
	err = tx.QueryRow(`
		INSERT INTO tcp_configs (name, host, port, enabled)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, host, port, enabled, created_at, updated_at
	`, cfg.Name, cfg.Host, cfg.Port, cfg.Enabled).Scan(
		&saved.ID, &saved.Name, &saved.Host, &saved.Port, &saved.Enabled, &saved.CreatedAt, &saved.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &saved, nil
}

func (r *TCPConfigRepo) Update(id int64, cfg model.TCPConfig) (*model.TCPConfig, error) {
	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if cfg.Enabled {
		if _, err := tx.Exec(
			"UPDATE tcp_configs SET enabled = FALSE, updated_at = NOW() WHERE enabled = TRUE AND id <> $1",
			id,
		); err != nil {
			return nil, err
		}
	}

	var saved model.TCPConfig
	err = tx.QueryRow(`
		UPDATE tcp_configs
		SET name = $1,
		    host = $2,
		    port = $3,
		    enabled = $4,
		    updated_at = NOW()
		WHERE id = $5
		RETURNING id, name, host, port, enabled, created_at, updated_at
	`, cfg.Name, cfg.Host, cfg.Port, cfg.Enabled, id).Scan(
		&saved.ID, &saved.Name, &saved.Host, &saved.Port, &saved.Enabled, &saved.CreatedAt, &saved.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tcp config not found: %d", id)
	}
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &saved, nil
}

func (r *TCPConfigRepo) Delete(id int64) error {
	result, err := db.DB.Exec("DELETE FROM tcp_configs WHERE id = $1", id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("tcp config not found: %d", id)
	}
	return nil
}
