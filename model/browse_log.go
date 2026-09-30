package model

import "time"

// BrowseLog 浏览记录：一个用户对一本书保留一条记录，重复浏览只更新时间
// （unique 索引防重，配合 service 层 Redis 去重窗口防刷）
type BrowseLog struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	UserID    int       `gorm:"uniqueIndex:uk_browse_user_book" json:"user_id"`
	BookID    int       `gorm:"uniqueIndex:uk_browse_user_book" json:"book_id"`
	ViewedAt  time.Time `json:"viewed_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联字段
	Book *Book `gorm:"foreignKey:BookID" json:"book,omitempty"`
}

func (BrowseLog) TableName() string { return "browse_logs" }
