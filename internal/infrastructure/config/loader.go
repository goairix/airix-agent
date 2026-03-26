package config

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// LoadResult 配置加载结果
type LoadResult struct {
	Config      *Config
	Source      string // "local" 或配置源类型如 "etcd"
	SourcePath  string
	WatchSource Source // 仅配置中心模式下非 nil
}

// Load 两阶段加载配置
func Load(path string) (*LoadResult, error) {
	// 加载.env
	_ = godotenv.Load()

	// 阶段1：读取本地引导配置
	bootstrap := viper.New()
	bootstrap.SetConfigFile(path)
	bootstrap.AutomaticEnv()
	if err := bootstrap.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 显式绑定配置中心的环境变量（AutomaticEnv 在 Unmarshal 时不生效）
	_ = bootstrap.BindEnv("config_center.type", "CONFIG_CENTER_TYPE")
	_ = bootstrap.BindEnv("config_center.etcd.endpoints", "CONFIG_CENTER_ETCD_ENDPOINTS")
	_ = bootstrap.BindEnv("config_center.etcd.key", "CONFIG_CENTER_ETCD_KEY")
	_ = bootstrap.BindEnv("config_center.etcd.timeout", "CONFIG_CENTER_ETCD_TIMEOUT")
	_ = bootstrap.BindEnv("config_center.etcd.username", "CONFIG_CENTER_ETCD_USERNAME")
	_ = bootstrap.BindEnv("config_center.etcd.password", "CONFIG_CENTER_ETCD_PASSWORD")

	// endpoints 环境变量是逗号分隔的字符串，需要手动拆分为 []string
	if envEndpoints, ok := os.LookupEnv("CONFIG_CENTER_ETCD_ENDPOINTS"); ok {
		bootstrap.Set("config_center.etcd.endpoints", strings.Split(envEndpoints, ","))
	}

	// 解析引导配置，检查 config_center
	bootstrapCfg := &Config{}
	if err := bootstrap.Unmarshal(bootstrapCfg); err != nil {
		return nil, fmt.Errorf("解析引导配置失败: %w", err)
	}

	// 阶段2：如果配置了配置中心，尝试从远程拉取
	if source := createSource(bootstrapCfg.ConfigCenter); source != nil {
		cfg, err := loadFromSource(source)
		if err != nil {
			slog.Warn("从配置中心加载失败，回退使用本地配置",
				"source", source.Type(), "error", err)
		} else {
			GlobalConfig = cfg
			return &LoadResult{
				Config:      cfg,
				Source:      source.Type(),
				SourcePath:  sourceKey(bootstrapCfg.ConfigCenter),
				WatchSource: source,
			}, nil
		}
	}

	// 回退：使用本地配置（保留逐 section 加载逻辑以支持环境变量绑定）
	cfg, err := loadLocal(path)
	if err != nil {
		return nil, err
	}
	return &LoadResult{Config: cfg, Source: "local", SourcePath: path}, nil
}

// loadLocal 从本地文件加载配置（保留逐 section 的 BindEnv 逻辑）
func loadLocal(configPath string) (*Config, error) {
	_ = godotenv.Load()

	v := viper.New()
	v.SetConfigFile(configPath)
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var appConfig AppConfig
	app := v.Sub("app")
	appBindEnv(app)
	if err := app.Unmarshal(&appConfig); err != nil {
		return nil, err
	}

	var serverConfig Server
	server := v.Sub("server")
	serverBindEnv(server)
	if err := server.Unmarshal(&serverConfig); err != nil {
		return nil, err
	}

	var securityConfig Security
	security := v.Sub("security")
	securityBindEnv(security)
	if err := security.Unmarshal(&securityConfig); err != nil {
		return nil, err
	}

	var databaseConfig DatabaseConfig
	database := v.Sub("database")
	databaseBindEnv(database)
	if err := database.Unmarshal(&databaseConfig); err != nil {
		return nil, err
	}

	var redisConfig Redis
	redis := v.Sub("redis")
	redisBindEnv(redis)
	if err := redis.Unmarshal(&redisConfig); err != nil {
		return nil, err
	}

	var cacheConfig Cache
	cache := v.Sub("cache")
	cacheBindEnv(cache)
	if err := cache.Unmarshal(&cacheConfig); err != nil {
		return nil, err
	}

	var messageQueueConfig MessageQueue
	messageQueue := v.Sub("message_queue")
	messageQueueBindEnv(messageQueue)
	if err := messageQueue.Unmarshal(&messageQueueConfig); err != nil {
		return nil, err
	}

	var etcdConfig Etcd
	etcd := v.Sub("etcd")
	etcdBindEnv(etcd)
	if err := etcd.Unmarshal(&etcdConfig); err != nil {
		return nil, err
	}

	var storageConfig Storage
	storage := v.Sub("storage")
	storageBindEnv(storage)
	if err := storage.Unmarshal(&storageConfig); err != nil {
		return nil, err
	}

	var monitorConfig Monitor
	monitor := v.Sub("monitor")
	monitorBindEnv(monitor)
	if err := monitor.Unmarshal(&monitorConfig); err != nil {
		return nil, err
	}

	var thirdPartyConfig ThirdParty
	thirdParty := v.Sub("third_party")
	thirdPartyBindEnv(thirdParty)
	if err := thirdParty.Unmarshal(&thirdPartyConfig); err != nil {
		return nil, err
	}

	cfg := &Config{
		App:          appConfig,
		Server:       serverConfig,
		Security:     securityConfig,
		Database:     databaseConfig,
		Redis:        redisConfig,
		Cache:        cacheConfig,
		MessageQueue: messageQueueConfig,
		Etcd:         etcdConfig,
		Storage:      storageConfig,
		Monitor:      monitorConfig,
		ThirdParty:   thirdPartyConfig,
	}

	GlobalConfig = cfg

	return cfg, nil
}

// parseConfig 解析配置中心的 YAML 字节流
func parseConfig(data []byte) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}
	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("反序列化配置失败: %w", err)
	}
	return cfg, nil
}

// loadFromSource 从配置源加载并解析配置
func loadFromSource(source Source) (*Config, error) {
	data, err := source.Load()
	if err != nil {
		return nil, err
	}
	return parseConfig(data)
}

// createSource 根据配置中心配置创建配置源
func createSource(cc *ConfigCenterConfig) Source {
	if cc == nil || cc.Type == "" || cc.Type == "static" {
		return nil
	}
	switch cc.Type {
	case "etcd":
		if cc.Etcd == nil {
			return nil
		}
		return NewEtcdSource(cc.Etcd)
	default:
		slog.Warn("不支持的配置中心类型", "type", cc.Type)
		return nil
	}
}

// sourceKey 获取配置源的 key 信息
func sourceKey(cc *ConfigCenterConfig) string {
	if cc == nil {
		return ""
	}
	if cc.Type == "etcd" && cc.Etcd != nil {
		return cc.Etcd.Key
	}
	return ""
}
