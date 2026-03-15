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
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 初始化一些 Mysql, 配置文件，redis
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

	// 启动服务器
	go func() {
		fmt.Printf("服务器启动在：%s\n", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("服务器启动失败...")
			log.Fatal(err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("正在关闭服务器...")

	// 优雅关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := server.Shutdown(ctx)
	if err != nil {
		log.Println("服务器错误退出", err)
		cleanResources()
		os.Exit(1)
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
