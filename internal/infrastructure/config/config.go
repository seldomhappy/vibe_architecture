package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config represents application configuration
type Config struct {
	App     AppConfig     `yaml:"app"`
	Server  ServerConfig  `yaml:"server"`
	Logger  LoggerConfig  `yaml:"logger"`
	DB      DBConfig      `yaml:"db"`
	Tracing TracingConfig `yaml:"tracing"`
	Metrics MetricsConfig `yaml:"metrics"`
	Kafka   KafkaConfig   `yaml:"kafka"`
}

// AppConfig represents application settings
type AppConfig struct {
	Name        string `yaml:"name" env:"APP_NAME"`
	Version     string `yaml:"version" env:"APP_VERSION"`
	Environment string `yaml:"environment" env:"APP_ENVIRONMENT"`
	Debug       bool   `yaml:"debug" env:"APP_DEBUG"`
}

// ServerConfig represents HTTP server settings
type ServerConfig struct {
	Host         string        `yaml:"host" env:"SERVER_HOST"`
	Port         int           `yaml:"port" env:"SERVER_PORT"`
	ReadTimeout  time.Duration `yaml:"read_timeout" env:"SERVER_READ_TIMEOUT"`
	WriteTimeout time.Duration `yaml:"write_timeout" env:"SERVER_WRITE_TIMEOUT"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" env:"SERVER_IDLE_TIMEOUT"`
}

// LoggerConfig represents logger settings
type LoggerConfig struct {
	Level  string `yaml:"level" env:"LOG_LEVEL"`
	Format string `yaml:"format" env:"LOG_FORMAT"`
	Output string `yaml:"output" env:"LOG_OUTPUT"`
}

// DBConfig represents database settings
type DBConfig struct {
	Host              string        `yaml:"host" env:"DB_HOST"`
	Port              int           `yaml:"port" env:"DB_PORT"`
	User              string        `yaml:"user" env:"DB_USER"`
	Password          string        `yaml:"password" env:"DB_PASSWORD"`
	Database          string        `yaml:"database" env:"DB_NAME"`
	SSLMode           string        `yaml:"ssl_mode" env:"DB_SSL_MODE"`
	MaxConnections    int           `yaml:"max_connections" env:"DB_MAX_CONNECTIONS"`
	MinConnections    int           `yaml:"min_connections" env:"DB_MIN_CONNECTIONS"`
	MaxConnLifetime   time.Duration `yaml:"max_conn_lifetime" env:"DB_MAX_CONN_LIFETIME"`
	MaxConnIdleTime   time.Duration `yaml:"max_conn_idle_time" env:"DB_MAX_CONN_IDLE_TIME"`
	HealthCheckPeriod time.Duration `yaml:"health_check_period" env:"DB_HEALTH_CHECK_PERIOD"`
	ConnectTimeout    time.Duration `yaml:"connect_timeout" env:"DB_CONNECT_TIMEOUT"`
}

// TracingConfig represents tracing settings
type TracingConfig struct {
	Enabled        bool    `yaml:"enabled" env:"TRACING_ENABLED"`
	ServiceName    string  `yaml:"service_name" env:"TRACING_SERVICE_NAME"`
	ServiceVersion string  `yaml:"service_version" env:"TRACING_SERVICE_VERSION"`
	Environment    string  `yaml:"environment" env:"TRACING_ENVIRONMENT"`
	JaegerEndpoint string  `yaml:"jaeger_endpoint" env:"JAEGER_ENDPOINT"`
	SampleRate     float64 `yaml:"sample_rate" env:"TRACING_SAMPLE_RATE"`
}

// MetricsConfig represents metrics settings
type MetricsConfig struct {
	Enabled     bool   `yaml:"enabled" env:"METRICS_ENABLED"`
	Port        int    `yaml:"port" env:"METRICS_PORT"`
	Path        string `yaml:"path" env:"METRICS_PATH"`
	Version     string `yaml:"version" env:"METRICS_VERSION"`
	Environment string `yaml:"environment" env:"METRICS_ENVIRONMENT"`
}

// KafkaConfig represents Kafka settings
type KafkaConfig struct {
	Enabled  bool           `yaml:"enabled" env:"KAFKA_ENABLED"`
	Brokers  []string       `yaml:"brokers" env:"KAFKA_BROKERS" env-separator:","`
	Version  string         `yaml:"version" env:"KAFKA_VERSION"`
	Producer ProducerConfig `yaml:"producer"`
	Consumer ConsumerConfig `yaml:"consumer"`
	Topics   TopicsConfig   `yaml:"topics"`
}

// ProducerConfig represents Kafka producer settings
type ProducerConfig struct {
	Enabled bool `yaml:"enabled" env:"KAFKA_PRODUCER_ENABLED"`
}

// ConsumerConfig represents Kafka consumer settings
type ConsumerConfig struct {
	Enabled bool   `yaml:"enabled" env:"KAFKA_CONSUMER_ENABLED"`
	GroupID string `yaml:"group_id" env:"KAFKA_CONSUMER_GROUP_ID"`
}

// TopicsConfig represents Kafka topics
type TopicsConfig struct {
	TaskCreated   TopicConfig `yaml:"task_created"`
	TaskUpdated   TopicConfig `yaml:"task_updated"`
	TaskCompleted TopicConfig `yaml:"task_completed"`
	TaskDeleted   TopicConfig `yaml:"task_deleted"`
}

// TopicConfig represents a single Kafka topic
type TopicConfig struct {
	Name              string `yaml:"name"`
	Partitions        int    `yaml:"partitions"`
	ReplicationFactor int    `yaml:"replication_factor"`
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
