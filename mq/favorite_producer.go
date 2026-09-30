package mq

import (
	"bookstore-manager/config"
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

// FavoriteProducer 全局收藏事件生产者，由 main 初始化。
var FavoriteProducer *FavoriteEventProducer

func InitFavoriteProducer() {
	FavoriteProducer = NewFavoriteEventProducer()
}

func CloseFavoriteProducer() {
	if FavoriteProducer != nil {
		if err := FavoriteProducer.Close(); err != nil {
			log.Println("关闭 Kafka 收藏生产者失败：", err)
		}
	}
}

// FavoriteEventProducer 收藏事件生产者。以 userID 作为消息 key，
// 同一用户的收藏事件落入同一分区，保证其 add/remove 有序。
type FavoriteEventProducer struct {
	writer *kafka.Writer
}

func NewFavoriteEventProducer() *FavoriteEventProducer {
	kafkaCfg := config.AppConfig.Kafka
	writer := &kafka.Writer{
		Addr:         kafka.TCP(kafkaCfg.Brokers...),
		Topic:        kafkaCfg.FavoriteTopic,
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireOne,
		BatchTimeout: 10 * time.Millisecond,
	}
	return &FavoriteEventProducer{writer: writer}
}

// SendFavoriteEvent 同步发送收藏事件，3 秒超时。
func (p *FavoriteEventProducer) SendFavoriteEvent(ctx context.Context, msg *FavoriteEventMessage) error {
	value, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	writeCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return p.writer.WriteMessages(writeCtx, kafka.Message{
		Key:   []byte(strconv.Itoa(msg.UserID)),
		Value: value,
	})
}

func (p *FavoriteEventProducer) Close() error {
	return p.writer.Close()
}
