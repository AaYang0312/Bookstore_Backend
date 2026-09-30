package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"bookstore-manager/global"
	"bookstore-manager/mq"
	"bookstore-manager/model"
	"bookstore-manager/repository"

	"github.com/go-redis/redis/v8"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	// favoritePendingTTL pending 标记在 Redis 中的保留时长（兜底防泄漏）
	favoritePendingTTL = 10 * time.Minute
)

// favoritePendingKey 每个用户一个 Hash：field=bookID，value=add|remove
func favoritePendingKey(userID int) string {
	return fmt.Sprintf("fav:pending:%d", userID)
}

type FavoriteService struct {
	FavoriteDAO *repository.FavoriteDAO
	BookDAO     *repository.BookDAO
}

func NewFavoriteService() *FavoriteService {
	return &FavoriteService{
		FavoriteDAO: repository.NewFavoriteDAO(),
		BookDAO:     repository.NewBookDAO(),
	}
}

// AddFavorite 受理添加收藏：发送 Kafka 消息成功后写入 Redis pending 标记，
// 真正落库由消费者异步执行。
func (f *FavoriteService) AddFavorite(ctx context.Context, userID, bookID int) error {
	return f.submitFavoriteEvent(ctx, userID, bookID, mq.FavoriteActionAdd)
}

// DelFavorite 受理取消收藏，与 AddFavorite 同为异步写路径。
func (f *FavoriteService) DelFavorite(ctx context.Context, userID, bookID int) error {
	return f.submitFavoriteEvent(ctx, userID, bookID, mq.FavoriteActionRemove)
}

// submitFavoriteEvent 校验并发送收藏事件。仅在消息发送成功后才写 pending，
// 发送失败直接返回 ErrMQUnavailable（不做同步 DB 降级，保持单一写路径）。
func (f *FavoriteService) submitFavoriteEvent(ctx context.Context, userID, bookID int, action string) error {
	if bookID <= 0 {
		return errors.New("图书ID不合法")
	}
	if mq.FavoriteProducer == nil {
		return ErrMQUnavailable
	}
	msg := &mq.FavoriteEventMessage{
		EventID:   uuid.NewString(),
		UserID:    userID,
		BookID:    bookID,
		Action:    action,
		CreatedAt: time.Now(),
	}
	if err := mq.FavoriteProducer.SendFavoriteEvent(ctx, msg); err != nil {
		log.Printf("发送收藏消息失败: user=%d book=%d action=%s event=%s error=%v",
			userID, bookID, action, msg.EventID, err)
		return ErrMQUnavailable
	}
	f.setFavoritePending(ctx, userID, bookID, action)
	return nil
}

func (f *FavoriteService) setFavoritePending(ctx context.Context, userID, bookID int, action string) {
	if global.RedisClient == nil {
		return
	}
	key := favoritePendingKey(userID)
	if err := global.RedisClient.HSet(ctx, key, strconv.Itoa(bookID), action).Err(); err != nil {
		log.Println("写入收藏 pending 标记失败：", err)
	}
	if err := global.RedisClient.Expire(ctx, key, favoritePendingTTL).Err(); err != nil {
		log.Println("刷新收藏 pending TTL 失败：", err)
	}
}

// clearFavoritePendingIfMatch 仅当 pending 中记录的动作与本条事件一致时才清除，
// 避免误删用户随后发起、仍在途的相反动作标记。
func (f *FavoriteService) clearFavoritePendingIfMatch(ctx context.Context, userID, bookID int, action string) {
	if global.RedisClient == nil {
		return
	}
	key := favoritePendingKey(userID)
	field := strconv.Itoa(bookID)
	pending, err := global.RedisClient.HGet(ctx, key, field).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Println("读取收藏 pending 标记失败：", err)
		}
		return
	}
	if pending != action {
		return
	}
	f.clearFavoritePending(ctx, userID, bookID)
}

func (f *FavoriteService) clearFavoritePending(ctx context.Context, userID, bookID int) {
	if global.RedisClient == nil {
		return
	}
	if err := global.RedisClient.HDel(ctx, favoritePendingKey(userID), strconv.Itoa(bookID)).Err(); err != nil {
		log.Println("清除收藏 pending 标记失败：", err)
	}
}

// HandleFavoriteEvent Kafka 消费者回调：异步执行收藏落库。
// 业务失败（图书不存在等）清除 pending 并返回 nil（消息正常提交）；
// 返回 error 视为系统故障，消费者不提交位点、消息重投。
func (f *FavoriteService) HandleFavoriteEvent(ctx context.Context, msg *mq.FavoriteEventMessage) error {
	switch msg.Action {
	case mq.FavoriteActionAdd:
		if _, err := f.BookDAO.GetBookByID(msg.BookID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("收藏业务失败（图书不存在，已跳过）: user=%d book=%d event=%s",
					msg.UserID, msg.BookID, msg.EventID)
				f.clearFavoritePending(ctx, msg.UserID, msg.BookID)
				return nil
			}
			return fmt.Errorf("查询图书失败: %w", err)
		}
		if err := f.FavoriteDAO.AddFavorite(msg.UserID, msg.BookID); err != nil {
			// 图书恰被删除导致外键冲突（1452）时按业务跳过处理，不做无效重投
			if isForeignKeyViolation(err) {
				log.Printf("收藏业务失败（外键冲突，已跳过）: user=%d book=%d event=%s",
					msg.UserID, msg.BookID, msg.EventID)
				f.clearFavoritePending(ctx, msg.UserID, msg.BookID)
				return nil
			}
			return fmt.Errorf("收藏落库失败: %w", err)
		}
	case mq.FavoriteActionRemove:
		if err := f.FavoriteDAO.DelFavorite(msg.UserID, msg.BookID); err != nil {
			return fmt.Errorf("取消收藏落库失败: %w", err)
		}
	default:
		log.Printf("收藏消息动作未知（已跳过）: action=%s user=%d book=%d event=%s",
			msg.Action, msg.UserID, msg.BookID, msg.EventID)
		f.clearFavoritePending(ctx, msg.UserID, msg.BookID)
		return nil
	}
	f.clearFavoritePendingIfMatch(ctx, msg.UserID, msg.BookID, msg.Action)
	log.Printf("异步收藏处理成功: action=%s user=%d book=%d event=%s",
		msg.Action, msg.UserID, msg.BookID, msg.EventID)
	return nil
}

// isForeignKeyViolation 判断是否为 MySQL 外键约束冲突（错误码 1452）
func isForeignKeyViolation(err error) bool {
	var me *mysql.MySQLError
	return errors.As(err, &me) && me.Number == 1452
}

func (f *FavoriteService) GetUserFavorites(userID int, page, pageSize int, timeFilter string) ([]*model.Favorite, int64, error) {
	fav, err := f.FavoriteDAO.GetUserFavorites(userID)
	if err != nil {
		return nil, 0, err
	}
	total := len(fav)
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= total {
		return []*model.Favorite{}, int64(total), nil
	}
	if end >= total {
		end = total
	}
	return fav[start:end], int64(total), nil
}
func (f *FavoriteService) GetUserFavoriteCount(userID int) (int64, error) {
	return f.FavoriteDAO.GetUserFavoriteCount(userID)
}

// IsFavorite 在数据库结果之上叠加 Redis pending：pending=add 返回 true，
// pending=remove 返回 false，消除"提交后立即查询仍显示旧状态"的读己之写缝隙。
func (f *FavoriteService) IsFavorite(ctx context.Context, userID, bookID int) (bool, error) {
	favorited, err := f.FavoriteDAO.IsFavorite(userID, bookID)
	if err != nil {
		return false, err
	}
	if global.RedisClient == nil {
		return favorited, nil
	}
	pending, err := global.RedisClient.HGet(ctx, favoritePendingKey(userID), strconv.Itoa(bookID)).Result()
	if err == nil {
		if pending == mq.FavoriteActionAdd {
			return true, nil
		}
		if pending == mq.FavoriteActionRemove {
			return false, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		log.Println("读取收藏 pending 标记失败：", err)
	}
	return favorited, nil
}
