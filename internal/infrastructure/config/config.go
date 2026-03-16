package config

import (
	"time"
)

const (
	VarPath  string = "var"
	LogPath         = VarPath + "/logs"
	TempPath        = VarPath + "/tmp"
)

var GlobalConfig *Config

// ConfigCenterConfig 配置中心配置
type ConfigCenterConfig struct {
	Type string            `mapstructure:"type"` // "etcd"
	Etcd *EtcdSourceConfig `mapstructure:"etcd,omitempty"`
}

// EtcdSourceConfig etcd 配置中心连接配置
type EtcdSourceConfig struct {
	Endpoints []string      `mapstructure:"endpoints"`
	Key       string        `mapstructure:"key"`
	Timeout   time.Duration `mapstructure:"timeout"`
	Username  string        `mapstructure:"username"`
	Password  string        `mapstructure:"password"`
}

// Config 应用配置
type Config struct {
	ConfigCenter *ConfigCenterConfig `mapstructure:"config_center,omitempty"`
	App          AppConfig           `mapstructure:"app"`
	Server       Server              `mapstructure:"server"`
	Security     Security            `mapstructure:"security"`
	Database     DatabaseConfig      `mapstructure:"database"`
	Redis        Redis               `mapstructure:"redis"`
	Cache        Cache               `mapstructure:"cache"`
	MessageQueue MessageQueue        `mapstructure:"message_queue"`
	Etcd         Etcd                `mapstructure:"etcd"`
	Storage      Storage             `mapstructure:"storage"`
	Monitor      Monitor             `mapstructure:"monitor"`
	ThirdParty   ThirdParty          `mapstructure:"third_party"`
}

// LoadConfig 加载配置（向后兼容）
func LoadConfig(configPath string) (*Config, error) {
	result, err := Load(configPath)
	if err != nil {
		return nil, err
	}
	return result.Config, nil
}
