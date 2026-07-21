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

func (c *CarouselDAO) GetAdminCarousels() ([]*model.Carousel, error) {
	var carousels []*model.Carousel
	err := c.db.Order("sort_order ASC, id ASC").Find(&carousels).Error
	return carousels, err
}

func (c *CarouselDAO) GetCarouselByID(id int) (*model.Carousel, error) {
	var carousel model.Carousel
	if err := c.db.First(&carousel, id).Error; err != nil {
		return nil, err
	}
	return &carousel, nil
}

func (c *CarouselDAO) CreateCarousel(carousel *model.Carousel) error {
	return c.db.Create(carousel).Error
}

func (c *CarouselDAO) UpdateCarousel(carousel *model.Carousel) error {
	result := c.db.Model(&model.Carousel{}).Where("id = ?", carousel.ID).Updates(map[string]any{
		"title":       carousel.Title,
		"description": carousel.Description,
		"image_url":   carousel.ImageURL,
		"link_url":    carousel.LinkURL,
		"sort_order":  carousel.SortOrder,
		"is_active":   carousel.IsActive,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return c.ensureCarouselExists(carousel.ID)
	}
	return nil
}

func (c *CarouselDAO) UpdateCarouselStatus(id int, isActive bool) error {
	result := c.db.Model(&model.Carousel{}).Where("id = ?", id).Update("is_active", isActive)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return c.ensureCarouselExists(id)
	}
	return nil
}

func (c *CarouselDAO) DeleteCarousel(id int) error {
	result := c.db.Delete(&model.Carousel{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (c *CarouselDAO) ensureCarouselExists(id int) error {
	var count int64
	if err := c.db.Model(&model.Carousel{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
