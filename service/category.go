package service

import (
	"bookstore-manager/model"
	"bookstore-manager/repository"
)

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
