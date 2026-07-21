package service

import (
	"bookstore-manager/model"
	"bookstore-manager/repository"
	"errors"
	"strings"
)

var ErrCategoryNameRequired = errors.New("分类名称不能为空")

type CategoryService struct {
	CategoryDB *repository.CategoryDAO
}

func NewCategoryService() *CategoryService {
	return &CategoryService{
		CategoryDB: repository.NewCategoryDAO(),
	}
}

func (c *CategoryService) GetCategoryList() ([]*model.Category, error) {
	return c.CategoryDB.GetActiveCategories()
}

func (c *CategoryService) AdminGetCategories() ([]*model.Category, error) {
	return c.CategoryDB.GetAdminCategories()
}

func (c *CategoryService) AdminCreateCategory(category *model.Category) (*model.Category, error) {
	normalizeCategory(category)
	if category.Name == "" {
		return nil, ErrCategoryNameRequired
	}
	if err := c.CategoryDB.CreateCategory(category); err != nil {
		return nil, err
	}
	return c.CategoryDB.GetCategoryByID(category.ID)
}

func (c *CategoryService) AdminUpdateCategory(category *model.Category) (*model.Category, error) {
	normalizeCategory(category)
	if category.Name == "" {
		return nil, ErrCategoryNameRequired
	}
	if err := c.CategoryDB.UpdateCategory(category); err != nil {
		return nil, err
	}
	return c.CategoryDB.GetCategoryByID(category.ID)
}

func (c *CategoryService) AdminUpdateCategoryStatus(id int, isActive bool) (*model.Category, error) {
	if err := c.CategoryDB.UpdateCategoryStatus(id, isActive); err != nil {
		return nil, err
	}
	return c.CategoryDB.GetCategoryByID(id)
}

func normalizeCategory(category *model.Category) {
	category.Name = strings.TrimSpace(category.Name)
	category.Description = strings.TrimSpace(category.Description)
	category.Icon = strings.TrimSpace(category.Icon)
	category.Color = strings.TrimSpace(category.Color)
	category.Gradient = strings.TrimSpace(category.Gradient)
}
