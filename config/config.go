package config

import (
	"log"
	"os"
	"strconv"

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
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
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

	if AppConfig.Server.Host == "" {
		AppConfig.Server.Host = "localhost"
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
