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

// GetAdminCategories 返回全部分类，并实时统计每个分类的图书数量。
func (c *CategoryDAO) GetAdminCategories() ([]*model.Category, error) {
	var categories []*model.Category
	err := c.db.Model(&model.Category{}).
		Select("categories.*, COUNT(books.id) AS book_count").
		Joins("LEFT JOIN books ON books.category_id = categories.id").
		Group("categories.id").
		Order("categories.sort ASC, categories.id ASC").
		Scan(&categories).Error
	return categories, err
}

func (c *CategoryDAO) GetCategoryByID(id int) (*model.Category, error) {
	var category model.Category
	err := c.db.Model(&model.Category{}).
		Select("categories.*, COUNT(books.id) AS book_count").
		Joins("LEFT JOIN books ON books.category_id = categories.id").
		Where("categories.id = ?", id).
		Group("categories.id").
		Scan(&category).Error
	if err != nil {
		return nil, err
	}
	if category.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &category, nil
}

func (c *CategoryDAO) CreateCategory(category *model.Category) error {
	return c.db.Create(category).Error
}

func (c *CategoryDAO) UpdateCategory(category *model.Category) error {
	result := c.db.Model(&model.Category{}).Where("id = ?", category.ID).Updates(map[string]any{
		"name":        category.Name,
		"description": category.Description,
		"icon":        category.Icon,
		"color":       category.Color,
		"gradient":    category.Gradient,
		"sort":        category.Sort,
		"is_active":   category.IsActive,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return c.ensureCategoryExists(category.ID)
	}
	return nil
}

func (c *CategoryDAO) UpdateCategoryStatus(id int, isActive bool) error {
	result := c.db.Model(&model.Category{}).Where("id = ?", id).Update("is_active", isActive)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return c.ensureCategoryExists(id)
	}
	return nil
}

func (c *CategoryDAO) ensureCategoryExists(id int) error {
	var count int64
	if err := c.db.Model(&model.Category{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
