package service

import (
	"fmt"
	"strings"

	"github.com/user/can-server/internal/db/repository"
	"github.com/user/can-server/internal/model"
)

type TCPConfigService struct {
	repo *repository.TCPConfigRepo
}

func NewTCPConfigService(repo *repository.TCPConfigRepo) *TCPConfigService {
	return &TCPConfigService{repo: repo}
}

func (s *TCPConfigService) List() ([]model.TCPConfig, error) {
	return s.repo.GetAll()
}

func (s *TCPConfigService) Create(cfg model.TCPConfig) (*model.TCPConfig, error) {
	if err := validateTCPConfig(cfg); err != nil {
		return nil, err
	}
	return s.repo.Create(cfg)
}

func (s *TCPConfigService) Update(id int64, cfg model.TCPConfig) (*model.TCPConfig, error) {
	if err := validateTCPConfig(cfg); err != nil {
		return nil, err
	}
	return s.repo.Update(id, cfg)
}

func (s *TCPConfigService) Delete(id int64) error {
	return s.repo.Delete(id)
}

func validateTCPConfig(cfg model.TCPConfig) error {
	if strings.TrimSpace(cfg.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if strings.TrimSpace(cfg.Host) == "" {
		return fmt.Errorf("host is required")
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	return nil
}
