package service

import (
	"errors"
	"os"
)

// Mode 运行模式
type Mode = string

const (
	ModeDev  Mode = "dev"
	ModeTest Mode = "test"
	ModePre  Mode = "pre"
	ModePro  Mode = "pro"
)

// LogConf 日志配置
type LogConf struct {
	Level      string `json:"level" toml:"level"`
	Encoding   string `json:"encoding" toml:"encoding"`
	OutputPath string `json:"output_path" toml:"output_path"`
}

// TelemetryConf 遥测配置
type TelemetryConf struct {
	Name     string  `json:"name" toml:"name"`
	Endpoint string  `json:"endpoint" toml:"endpoint"`
	Sampler  float64 `json:"sampler" toml:"sampler"`
	Batcher  string  `json:"batcher" toml:"batcher"`
}

// ServiceConf 统一服务配置（参照 go-zero service.ServiceConf）
type ServiceConf struct {
	Name      string        `json:"name" toml:"name"`
	Mode      Mode          `json:"mode" toml:"mode"`
	Log       LogConf       `json:"log" toml:"log"`
	MetricsUrl string       `json:"metrics_url" toml:"metrics_url"`
	Telemetry TelemetryConf `json:"telemetry" toml:"telemetry"`
}

// Validate 验证配置
func (sc *ServiceConf) Validate() error {
	if sc.Name == "" {
		return errors.New("service name is required")
	}
	return nil
}

// SetUp 根据配置初始化基础设施
func (sc *ServiceConf) SetUp() error {
	if err := sc.Validate(); err != nil {
		return err
	}
	os.Setenv("APP_NAME", sc.Name)
	os.Setenv("APP_MODE", sc.Mode)
	return nil
}
