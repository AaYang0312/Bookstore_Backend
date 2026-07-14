package model

import "time"

type Category struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name"`
	Description string    `json:"description"` // 描述
	Icon        string    `json:"icon"`
	Color       string    `json:"color"`
	Gradient    string    `json:"gradient"`
	Sort        int       `json:"sort"`
	IsActive    bool      `json:"is_active"`
	BookCount   int       `json:"book_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Category) TableName() string { return "categories" }
