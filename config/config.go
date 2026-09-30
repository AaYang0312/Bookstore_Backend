package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

var AppConfig Config

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}
type KafkaConfig struct {
	Brokers         []string `yaml:"brokers"`
	Topic           string   `yaml:"topic"`
	GroupID         string   `yaml:"group_id"`
	FavoriteTopic   string   `yaml:"favorite_topic"`
	FavoriteGroupID string   `yaml:"favorite_group_id"`
}
type MinIOConfig struct {
	Endpoint      string `yaml:"endpoint"`
	AccessKey     string `yaml:"access_key"`
	SecretKey     string `yaml:"secret_key"`
	Bucket        string `yaml:"bucket"`
	UseSSL        bool   `yaml:"use_ssl"`
	PublicBaseURL string `yaml:"public_base_url"`
}
type InternalConfig struct {
	SyncSecret string `yaml:"sync_secret"`
}
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Kafka    KafkaConfig    `yaml:"kafka"`
	MinIO    MinIOConfig    `yaml:"minio"`
	Internal InternalConfig `yaml:"internal"`
}

func InitConfig(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalln("读取配置文件失败：", err)
	}
	if err := yaml.Unmarshal(data, &AppConfig); err != nil {
		log.Fatalln("yaml反序列化配置失败：", err)
	}
	applyEnvironmentOverrides()
	log.Println("加载配置文件成功")
}

func applyEnvironmentOverrides() {
	overrideString("BOOKSTORE_SERVER_HOST", &AppConfig.Server.Host)
	overrideInt("BOOKSTORE_SERVER_PORT", &AppConfig.Server.Port)
	overrideString("BOOKSTORE_DATABASE_HOST", &AppConfig.Database.Host)
	overrideInt("BOOKSTORE_DATABASE_PORT", &AppConfig.Database.Port)
	overrideString("BOOKSTORE_DATABASE_USER", &AppConfig.Database.User)
	overrideString("BOOKSTORE_DATABASE_PASSWORD", &AppConfig.Database.Password)
	overrideString("BOOKSTORE_DATABASE_NAME", &AppConfig.Database.Name)
	overrideString("BOOKSTORE_REDIS_HOST", &AppConfig.Redis.Host)
	overrideInt("BOOKSTORE_REDIS_PORT", &AppConfig.Redis.Port)
	overrideString("BOOKSTORE_REDIS_PASSWORD", &AppConfig.Redis.Password)
	overrideInt("BOOKSTORE_REDIS_DB", &AppConfig.Redis.DB)
	overrideStringSlice("BOOKSTORE_KAFKA_BROKERS", &AppConfig.Kafka.Brokers)
	overrideString("BOOKSTORE_KAFKA_TOPIC", &AppConfig.Kafka.Topic)
	overrideString("BOOKSTORE_KAFKA_GROUP_ID", &AppConfig.Kafka.GroupID)
	overrideString("BOOKSTORE_KAFKA_FAVORITE_TOPIC", &AppConfig.Kafka.FavoriteTopic)
	overrideString("BOOKSTORE_KAFKA_FAVORITE_GROUP_ID", &AppConfig.Kafka.FavoriteGroupID)
	overrideString("BOOKSTORE_MINIO_ENDPOINT", &AppConfig.MinIO.Endpoint)
	overrideString("BOOKSTORE_MINIO_ACCESS_KEY", &AppConfig.MinIO.AccessKey)
	overrideString("BOOKSTORE_MINIO_SECRET_KEY", &AppConfig.MinIO.SecretKey)
	overrideString("BOOKSTORE_MINIO_BUCKET", &AppConfig.MinIO.Bucket)
	overrideBool("BOOKSTORE_MINIO_USE_SSL", &AppConfig.MinIO.UseSSL)
	overrideString("BOOKSTORE_MINIO_PUBLIC_BASE_URL", &AppConfig.MinIO.PublicBaseURL)
	overrideString("BOOKSTORE_INTERNAL_SYNC_SECRET", &AppConfig.Internal.SyncSecret)

	if AppConfig.Server.Host == "" {
		AppConfig.Server.Host = "localhost"
	}
	if len(AppConfig.Kafka.Brokers) == 0 {
		AppConfig.Kafka.Brokers = []string{"127.0.0.1:9092"}
	}
	if AppConfig.Kafka.Topic == "" {
		AppConfig.Kafka.Topic = "order-create"
	}
	if AppConfig.Kafka.GroupID == "" {
		AppConfig.Kafka.GroupID = "order-service"
	}
	if AppConfig.Kafka.FavoriteTopic == "" {
		AppConfig.Kafka.FavoriteTopic = "favorite-events"
	}
	if AppConfig.Kafka.FavoriteGroupID == "" {
		AppConfig.Kafka.FavoriteGroupID = "favorite-service"
	}
	if AppConfig.MinIO.Bucket == "" {
		AppConfig.MinIO.Bucket = "bookstore"
	}
	if AppConfig.MinIO.PublicBaseURL == "" {
		AppConfig.MinIO.PublicBaseURL = "http://localhost:9000"
	}
}

func overrideString(name string, target *string) {
	if value, ok := os.LookupEnv(name); ok {
		*target = value
	}
}

func overrideInt(name string, target *int) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalf("环境变量 %s 必须是整数：%v", name, err)
	}
	*target = parsed
}

func overrideBool(name string, target *bool) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		log.Fatalf("环境变量 %s 必须是布尔值(true/false)：%v", name, err)
	}
	*target = parsed
}

func overrideStringSlice(name string, target *[]string) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	*target = result
}
