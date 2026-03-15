package main

import (
	"bookstore-manager/config"
	"bookstore-manager/global"
	"bookstore-manager/web/router"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	// 初始化一些Mysql, 配置文件, redis
	// 配置
	config.InitConfig("conf/config.yaml")
	global.InitMysql()
	global.InitRedis()

	r := router.InitRouter()
	addr := fmt.Sprintf("%s:%d", "localhost", config.AppConfig.Server.Port)

	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("服务器启动失败...")
		os.Exit(-1)
	}
	err = server.Shutdown(context.TODO())
	if err != nil {
		log.Println("服务器错误退出", err)
		cleanResources()
	} else {
		log.Println("服务器正常退出")
		cleanResources()
	}
}
func cleanResources() {
	if global.RedisClient != nil {
		log.Println("Redis 资源清理....")
		global.CloseRedis()
	}
	if global.DBClient != nil {
		log.Println("Mysql 资源清理....")
		global.CloseDB()
	}
	time.Sleep(1 * time.Second)
	log.Println("所有资源清理完毕！")
}
