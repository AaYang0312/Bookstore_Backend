package global

import (
	"bookstore-manager/config"
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DBClient *gorm.DB
var RedisClient *redis.Client

func InitMysql() {
	cfg := config.AppConfig.Database
	// 登录Mysql数据库
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
	client, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Println("连接数据库失败：", err)
		return
	}
	DBClient = client
	log.Println("连接Mysql数据库成功")
}

func InitRedis() {
	redisConfig := config.AppConfig.Redis
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalln("redis连接失败：", err)
		return
	}
	RedisClient = client
	log.Println("Redis连接成功")
}

func GetDB() *gorm.DB {
	return DBClient
}

func CloseDB() {
	if DBClient != nil {
		sqlDB, err := DBClient.DB()
		if err == nil && sqlDB != nil {
			sqlDB.Close()
		}
	}
}
