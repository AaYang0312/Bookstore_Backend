package main

import (
	"bookstore-manager/config"
	"bookstore-manager/global"
	"bookstore-manager/mq"
	"bookstore-manager/service"
	"bookstore-manager/storage"
	"bookstore-manager/web/router"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	// 初始化配置、Mysql、Redis、MinIO
	config.InitConfig("conf/config.yaml")
	global.InitMysql()
	global.InitRedis()
	storage.InitMinIO()

	// Kafka：确保 topic 存在，初始化全局生产者（订单 + 收藏）
	if err := mq.EnsureTopic(); err != nil {
		log.Fatalln("Kafka 初始化失败：", err)
	}
	mq.InitProducer()
	mq.InitFavoriteProducer()

	// 后台任务：下单/收藏消息消费者 + 超时关单定时任务
	orderService := service.NewOrderService()
	favoriteService := service.NewFavoriteService()
	workerCtx, stopWorkers := context.WithCancel(context.Background())
	var workers sync.WaitGroup
	workers.Add(3)
	go func() {
		defer workers.Done()
		mq.StartOrderConsumer(workerCtx, orderService.HandleOrderCreate)
	}()
	go func() {
		defer workers.Done()
		mq.StartFavoriteConsumer(workerCtx, favoriteService.HandleFavoriteEvent)
	}()
	go func() {
		defer workers.Done()
		orderService.StartOrderTimeoutWorker(workerCtx)
	}()

	r := router.InitRouter()
	addr := fmt.Sprintf("%s:%d", config.AppConfig.Server.Host, config.AppConfig.Server.Port)

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
	}

	// 停止后台任务（消费者、定时关单），再关闭生产者
	stopWorkers()
	workers.Wait()
	mq.CloseProducer()
	mq.CloseFavoriteProducer()

	log.Println("服务器正常退出")
	cleanResources()
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
