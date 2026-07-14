package repository

import (
	"bookstore-manager/global"
	"bookstore-manager/model"

	"gorm.io/gorm"
)

type CategoryDAO struct {
	db *gorm.DB
}

func NewCategoryDAO() *CategoryDAO {
	return &CategoryDAO{
		db: global.GetDB(),
	}
}

func (c *CategoryDAO) GetActiveCategories() ([]*model.Category, error) {
	var categories []*model.Category

	err := c.db.
		Where("is_active = ?", true).
		Order("sort ASC").
		Find(&categories).Error

	if err != nil {
		return nil, err
	}
	return categories, nil
}
