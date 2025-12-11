package config

import (
	"fmt"
	"time"
)

// Config represents the complete application configuration
type Config struct {
	App     AppConfig     `yaml:"app"`
	Server  ServerConfig  `yaml:"server"`
	Logger  LoggerConfig  `yaml:"logger"`
	DB      DBConfig      `yaml:"db"`
	Tracing TracingConfig `yaml:"tracing"`
	Metrics MetricsConfig `yaml:"metrics"`
	Kafka   KafkaConfig   `yaml:"kafka"`
}

// AppConfig contains application-level settings
type AppConfig struct {
	Name        string `yaml:"name" env:"APP_NAME" env-default:"vibe-architecture"`
	Version     string `yaml:"version" env:"APP_VERSION" env-default:"1.0.0"`
	Environment string `yaml:"environment" env:"APP_ENVIRONMENT" env-default:"development"`
	Debug       bool   `yaml:"debug" env:"APP_DEBUG" env-default:"false"`
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
	Host            string        `yaml:"host" env:"SERVER_HOST" env-default:"0.0.0.0"`
	Port            int           `yaml:"port" env:"SERVER_PORT" env-default:"8080"`
	ReadTimeout     time.Duration `yaml:"read_timeout" env-default:"10s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" env-default:"10s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env-default:"30s"`
}

// LoggerConfig contains logging settings
type LoggerConfig struct {
	Level  string `yaml:"level" env:"LOG_LEVEL" env-default:"info"`
	Format string `yaml:"format" env:"LOG_FORMAT" env-default:"json"`
}

// DBConfig contains database connection settings
type DBConfig struct {
	Host            string        `yaml:"host" env:"DB_HOST" env-default:"localhost"`
	Port            int           `yaml:"port" env:"DB_PORT" env-default:"5432"`
	User            string        `yaml:"user" env:"DB_USER" env-default:"postgres"`
	Password        string        `yaml:"password" env:"DB_PASSWORD" env-default:"postgres"`
	Database        string        `yaml:"database" env:"DB_NAME" env-default:"vibe_architecture"`
	SSLMode         string        `yaml:"ssl_mode" env:"DB_SSL_MODE" env-default:"disable"`
	MaxOpenConns    int           `yaml:"max_open_conns" env:"DB_MAX_OPEN_CONNS" env-default:"25"`
	MaxIdleConns    int           `yaml:"max_idle_conns" env:"DB_MAX_IDLE_CONNS" env-default:"5"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" env:"DB_CONN_MAX_LIFETIME" env-default:"5m"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time" env:"DB_CONN_MAX_IDLE_TIME" env-default:"5m"`
}

// DSN returns the PostgreSQL connection string
func (c DBConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Database, c.SSLMode)
}

// TracingConfig contains OpenTelemetry tracing settings
type TracingConfig struct {
	Enabled         bool    `yaml:"enabled" env:"TRACING_ENABLED" env-default:"true"`
	ServiceName     string  `yaml:"service_name" env:"TRACING_SERVICE_NAME"`
	JaegerEndpoint  string  `yaml:"jaeger_endpoint" env:"JAEGER_ENDPOINT" env-default:"http://localhost:14268/api/traces"`
	SamplingRate    float64 `yaml:"sampling_rate" env:"TRACING_SAMPLING_RATE" env-default:"1.0"`
}

// MetricsConfig contains Prometheus metrics settings
type MetricsConfig struct {
	Enabled bool   `yaml:"enabled" env:"METRICS_ENABLED" env-default:"true"`
	Port    int    `yaml:"port" env:"METRICS_PORT" env-default:"9090"`
	Path    string `yaml:"path" env:"METRICS_PATH" env-default:"/metrics"`
}

// KafkaConfig contains Kafka settings
type KafkaConfig struct {
	Brokers         []string      `yaml:"brokers" env:"KAFKA_BROKERS" env-default:"localhost:9092"`
	ConsumerGroupID string        `yaml:"consumer_group_id" env:"KAFKA_CONSUMER_GROUP_ID" env-default:"vibe-architecture-group"`
	Topics          TopicsConfig  `yaml:"topics"`
	Producer        ProducerConfig `yaml:"producer"`
	Consumer        ConsumerConfig `yaml:"consumer"`
}

// TopicsConfig contains Kafka topic names
type TopicsConfig struct {
	TaskEvents string `yaml:"task_events" env:"KAFKA_TOPIC_TASK_EVENTS" env-default:"task.events"`
}

// ProducerConfig contains Kafka producer settings
type ProducerConfig struct {
	Compression     string        `yaml:"compression" env-default:"snappy"`
	RetryMax        int           `yaml:"retry_max" env-default:"3"`
	RetryBackoff    time.Duration `yaml:"retry_backoff" env-default:"100ms"`
	Idempotent      bool          `yaml:"idempotent" env-default:"true"`
	Timeout         time.Duration `yaml:"timeout" env-default:"10s"`
}

// ConsumerConfig contains Kafka consumer settings
type ConsumerConfig struct {
	Workers         int           `yaml:"workers" env:"KAFKA_CONSUMER_WORKERS" env-default:"3"`
	SessionTimeout  time.Duration `yaml:"session_timeout" env-default:"10s"`
	RebalanceTimeout time.Duration `yaml:"rebalance_timeout" env-default:"60s"`
}

// Validate performs validation on the configuration
func (c *Config) Validate() error {
	if c.App.Name == "" {
		return fmt.Errorf("app.name is required")
	}
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535")
	}
	if c.DB.Host == "" {
		return fmt.Errorf("db.host is required")
	}
	if c.DB.Database == "" {
		return fmt.Errorf("db.database is required")
	}
	if len(c.Kafka.Brokers) == 0 {
		return fmt.Errorf("kafka.brokers is required")
	}
	if c.Tracing.Enabled && c.Tracing.ServiceName == "" {
		c.Tracing.ServiceName = c.App.Name
	}
	return nil
}
