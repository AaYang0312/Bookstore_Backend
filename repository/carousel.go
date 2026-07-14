package repository

import (
	"bookstore-manager/global"
	"bookstore-manager/model"

	"gorm.io/gorm"
)

type CarouselDAO struct {
	db *gorm.DB
}

func NewCarouselDAO() *CarouselDAO {
	return &CarouselDAO{
		db: global.GetDB(),
	}
}

func (c *CarouselDAO) GetActiveCarousels() ([]*model.Carousel, error) {
	var carousels []*model.Carousel

	err := c.db.
		Where("is_active = ?", true).
		Order("sort_order ASC").
		Find(&carousels).Error
	if err != nil {
		return nil, err
	}

	return carousels, nil
}
