// config/config.go
package config

import (
	"fmt"
	"time"

	"github.com/go-playground/validator"
	"github.com/spf13/viper"
)

type MQConfig struct {
	AMQPURL              string        `mapstructure:"amqp_url"     yaml:"amqp_url"`
	Exchange             string        `mapstructure:"exchange"     yaml:"exchange"`
	CertFile             string        `mapstructure:"cert_file"    yaml:"cert_file"`
	KeyFile              string        `mapstructure:"key_file"     yaml:"key_file"`
	CACertFile           string        `mapstructure:"ca_cert_file" yaml:"ca_cert_file"`
	Timeout              time.Duration `mapstructure:"timeout"      yaml:"timeout"`
	DeadLetterExchange   string        `mapstructure:"dead_letter_exchange" yaml:"dead_letter_exchange"`
	DeadLetterQueue      string        `mapstructure:"dead_letter_queue" yaml:"dead_letter_queue"`
	DeadLetterRoutingKey string        `mapstructure:"dead_letter_routing_key" yaml:"dead_letter_routing_key"`
	PrimaryExchange      string        `mapstructure:"primary_exchange" yaml:"primary_exchange"`
	PrimaryQueue         string        `mapstructure:"primaryqueue" yaml:"primaryqueue"`
}

type DBConfig struct {
	Host     string        `mapstructure:"host"     yaml:"host"`     // 原來的 db_host
	Instance string        `mapstructure:"instance" yaml:"instance"` // 原來的 db_instance
	Port     int           `mapstructure:"port"     yaml:"port"`
	User     string        `mapstructure:"user"     yaml:"user"`     // 原來的 db_user
	Password string        `mapstructure:"password" yaml:"password"` // 原來的 db_password
	Name     string        `mapstructure:"name"     yaml:"name"`     // 原來的 db_name
	Encrypt  string        `mapstructure:"encrypt"  yaml:"encrypt"`  // 原來的 db_encrypt
	Timeout  time.Duration `mapstructure:"timeout"  yaml:"timeout"`  // 原來的 db_timeout
}

type PrometheusConfig struct {
	MetricsPort int `mapstructure:"metrics_port"     yaml:"metrics_port"`
}

// Config 是整個服務的設定容器
type Config struct {
	MQ                 MQConfig         `mapstructure:"mq"`
	DB                 DBConfig         `mapstructure:"db"`
	Prometheus         PrometheusConfig `mapstructure:"prometheus"`
	ProcessDdlInterval time.Duration    `mapstructure:"process_ddl_interval" validate:"required"`
	ProcessDmlInterval time.Duration    `mapstructure:"process_dml_interval" validate:"required"`
	ProcessTimeout     time.Duration    `mapstructure:"process_timeout" validate:"required"`
}

// LoadConfig 從指定檔案路徑讀取設定，並支援 ENV 覆寫，最後進行欄位驗證
func LoadConfig(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path) // 例如 "config/config.yaml"
	v.SetConfigType("yaml")
	v.AutomaticEnv() // 支援以 ENV 覆寫
	v.SetEnvPrefix("FtyBiProducer")

	// 綁定 ENV 變數（可按需增減）
	v.BindEnv("mq.amqp_url")
	v.BindEnv("mq.exchange")
	v.BindEnv("mq.cert_file")
	v.BindEnv("mq.key_file")
	v.BindEnv("mq.ca_cert_file")
	v.BindEnv("mq.timeout")
	v.BindEnv("mq.dead_letter_exchange")
	v.BindEnv("mq.dead_letter_queue")
	v.BindEnv("mq.dead_letter_routing_key")
	v.BindEnv("mq.primary_exchange")
	v.BindEnv("mq.primaryqueue")

	v.BindEnv("db.host")
	v.BindEnv("db.instance")
	v.BindEnv("db.port")
	v.BindEnv("db.user")
	v.BindEnv("db.password")
	v.BindEnv("db.name")
	v.BindEnv("db.encrypt")
	v.BindEnv("db.timeout")

	v.BindEnv("prometheus.metrics_port")

	v.BindEnv("process_ddl_interval")
	v.BindEnv("process_dml_interval")
	v.BindEnv("process_timeout")

	// 1. 讀檔
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("讀取設定檔失敗: %w", err)
	}

	// 2. 解析到 struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析設定失敗: %w", err)
	}

	// 3. 驗證必填欄位
	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("設定驗證失敗: %w", err)
	}

	return &cfg, nil
}
