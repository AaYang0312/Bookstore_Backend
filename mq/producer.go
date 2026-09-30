package mq

import (
	"bookstore-manager/config"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

// Producer 全局下单消息生产者，由 main 初始化。
var Producer *OrderProducer

func InitProducer() {
	Producer = NewOrderProducer()
}

func CloseProducer() {
	if Producer != nil {
		if err := Producer.Close(); err != nil {
			log.Println("关闭 Kafka 生产者失败：", err)
		}
	}
}

// OrderProducer 下单消息生产者。以 userID 作为消息 key，
// 同一用户的消息落入同一分区，保证该用户的下单消息有序。
type OrderProducer struct {
	writer *kafka.Writer
}

func NewOrderProducer() *OrderProducer {
	kafkaCfg := config.AppConfig.Kafka
	writer := &kafka.Writer{
		Addr:         kafka.TCP(kafkaCfg.Brokers...),
		Topic:        kafkaCfg.Topic,
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireOne,
		BatchTimeout: 10 * time.Millisecond,
	}
	return &OrderProducer{writer: writer}
}

// EnsureTopic 启动时确保全部业务 topic（下单、收藏事件）存在。
// topic 已存在时 CreateTopics 会返回错误，直接忽略。
func EnsureTopic() error {
	if err := ensureTopic(config.AppConfig.Kafka.Topic); err != nil {
		return err
	}
	return ensureTopic(config.AppConfig.Kafka.FavoriteTopic)
}

// ensureTopic 确保单个 topic 存在：3 分区、1 副本、消息保留 24 小时。
func ensureTopic(name string) error {
	kafkaCfg := config.AppConfig.Kafka
	conn, err := kafka.Dial("tcp", kafkaCfg.Brokers[0])
	if err != nil {
		return fmt.Errorf("连接 Kafka 失败: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("获取 Kafka controller 失败: %w", err)
	}
	controllerConn, err := kafka.Dial("tcp", netJoinHostPort(controller.Host, controller.Port))
	if err != nil {
		return fmt.Errorf("连接 Kafka controller 失败: %w", err)
	}
	defer controllerConn.Close()

	err = controllerConn.CreateTopics(kafka.TopicConfig{
		Topic:             name,
		NumPartitions:     3,
		ReplicationFactor: 1,
		ConfigEntries: []kafka.ConfigEntry{
			{ConfigName: "retention.ms", ConfigValue: strconv.FormatInt((24 * time.Hour).Milliseconds(), 10)},
		},
	})
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("创建 topic %s 失败: %w", name, err)
	}
	log.Printf("Kafka topic 就绪: %s (brokers: %v)", name, kafkaCfg.Brokers)
	return nil
}

func netJoinHostPort(host string, port int) string {
	return host + ":" + strconv.Itoa(port)
}

// SendOrderCreate 同步发送下单消息，3 秒超时。
func (p *OrderProducer) SendOrderCreate(ctx context.Context, msg *OrderCreateMessage) error {
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

func (p *OrderProducer) Close() error {
	return p.writer.Close()
}
