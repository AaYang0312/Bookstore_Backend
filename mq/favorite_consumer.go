package mq

import (
	"bookstore-manager/config"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// FavoriteEventHandler 收藏事件处理函数。返回 error 表示系统故障，
// 消息不提交、稍后重投；业务失败应在 handler 内部消化并返回 nil。
type FavoriteEventHandler func(ctx context.Context, msg *FavoriteEventMessage) error

// StartFavoriteConsumer 阻塞运行收藏事件消费者组，直到 ctx 取消。
func StartFavoriteConsumer(ctx context.Context, handler FavoriteEventHandler) {
	kafkaCfg := config.AppConfig.Kafka
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        kafkaCfg.Brokers,
		GroupID:        kafkaCfg.FavoriteGroupID,
		Topic:          kafkaCfg.FavoriteTopic,
		MinBytes:       1,
		MaxBytes:       10 << 20,
		CommitInterval: 0, // 位点完全由处理结果手动提交
	})
	defer reader.Close()
	log.Printf("Kafka 收藏消费者启动: topic=%s group=%s", kafkaCfg.FavoriteTopic, kafkaCfg.FavoriteGroupID)

	for {
		m, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Println("Kafka 收藏消费者已退出")
				return
			}
			log.Println("Kafka 收藏消费者拉取消息失败：", err)
			time.Sleep(2 * time.Second)
			continue
		}

		var msg FavoriteEventMessage
		if err := json.Unmarshal(m.Value, &msg); err != nil {
			// 无法解析的消息重投也不会成功，记录后直接提交跳过
			log.Printf("收藏消息解析失败（已跳过）: %v, value=%s", err, string(m.Value))
			commitMessage(ctx, reader, m)
			continue
		}

		if err := handler(ctx, &msg); err != nil {
			// 系统故障：不提交位点，等待 Kafka 重投
			log.Printf("收藏消息处理失败（将重试）: key=%s error=%v", string(m.Key), err)
			time.Sleep(2 * time.Second)
			continue
		}
		commitMessage(ctx, reader, m)
	}
}
