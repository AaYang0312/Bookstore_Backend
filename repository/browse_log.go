package repository

import (
	"bookstore-manager/global"
	"bookstore-manager/model"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BrowseLogDAO struct {
	db *gorm.DB
}

func NewBrowseLogDAO() *BrowseLogDAO {
	return &BrowseLogDAO{db: global.GetDB()}
}

// RecordView 记录浏览：同一用户对同一本书只保留一条，重复浏览刷新 viewed_at
func (b *BrowseLogDAO) RecordView(userID, bookID int) error {
	log := &model.BrowseLog{
		UserID:   userID,
		BookID:   bookID,
		ViewedAt: time.Now(),
	}
	return b.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "user_id"}, {Name: "book_id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{"viewed_at"}),
	}).Create(log).Error
}

// BrowseRecord 浏览记录 + 图书摘要（browse-history 接口返回结构）
type BrowseRecord struct {
	ID        int       `json:"id"`
	BookID    int       `json:"book_id"`
	ViewedAt  time.Time `json:"viewed_at"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	Type      string    `json:"type"`
	Price     int       `json:"price"`
	CoverURL  string    `json:"cover_url"`
}

// GetByUser 分页查询用户浏览记录（按浏览时间倒序，只含在架图书）
func (b *BrowseLogDAO) GetByUser(userID, page, pageSize int) ([]*BrowseRecord, int64, error) {
	var total int64
	if err := b.db.Model(&model.BrowseLog{}).
		Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var records []*BrowseRecord
	offset := (page - 1) * pageSize
	err := b.db.Model(&model.BrowseLog{}).
		Select(`browse_logs.id, browse_logs.book_id, browse_logs.viewed_at,
			books.title, books.author, books.type, books.price, books.cover_url`).
		Joins("JOIN books ON books.id = browse_logs.book_id AND books.status = ?", 1).
		Where("browse_logs.user_id = ?", userID).
		Order("browse_logs.viewed_at DESC").
		Offset(offset).Limit(pageSize).
		Scan(&records).Error
	if err != nil {
		return nil, 0, err
	}
	return records, total, nil
}
