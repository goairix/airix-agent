package config

import "github.com/spf13/viper"

// 应用环境
const (
	Dev     = "dev"
	Prod    = "prod"
	Test    = "test"
	Staging = "staging"
)

// AppConfig 应用基本配置
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Environment string `mapstructure:"environment"`
	Debug       bool   `mapstructure:"debug"`
	Domain      string `mapstructure:"domain"`
}

// Security 安全配置
type Security struct {
	JWT struct {
		Secret string `mapstructure:"secret"`
	} `mapstructure:"jwt"`
	Crypto struct {
		AESKey string `mapstructure:"aes_key"`
		AESIV  string `mapstructure:"aes_iv"`
	} `mapstructure:"crypto"`
}

func appBindEnv(v *viper.Viper) {
	_ = v.BindEnv("name", "APP_NAME")
	_ = v.BindEnv("environment", "APP_ENV")
	_ = v.BindEnv("debug", "APP_DEBUG")
	_ = v.BindEnv("domain", "APP_DOMAIN")
	v.SetDefault("environment", Dev)
}

func securityBindEnv(v *viper.Viper) {
	_ = v.BindEnv("jwt.secret", "SECURITY_JWT_SECRET")
	_ = v.BindEnv("crypto.aes_key", "SECURITY_CRYPTO_AES_KEY")
	_ = v.BindEnv("crypto.aes_iv", "SECURITY_CRYPTO_AES_IV")
}
