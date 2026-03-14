package repository

import (
	"bookstore-manager/global"
	"bookstore-manager/model"

	"gorm.io/gorm"
)

type BookDAO struct {
	db *gorm.DB
}

func NewBookDAO() *BookDAO {
	return &BookDAO{
		db: global.GetDB(),
	}
}

func (b *BookDAO) GetHotBooks(limit int) ([]*model.Book, error) {
	var books []*model.Book
	err := b.db.Debug().Where("status = ?", 1).Order("sale DESC").Limit(limit).Find(&books).Error
	if err != nil {
		return nil, err
	}
	return books, nil
}

func (b *BookDAO) GetNewBooks(limit int) ([]*model.Book, error) {
	var books []*model.Book
	err := b.db.Debug().Where("status = ?", 1).Order("created_at DESC").Limit(limit).Find(&books).Error
	if err != nil {
		return nil, err
	}
	return books, nil
}
func (b *BookDAO) GetBooksByPage(page, pageSize int) ([]*model.Book, int64, error) {
	var books []*model.Book
	var total int64
	err := b.db.Debug().Model(&model.Book{}).Where("status = ?", 1).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	// 计算偏移量
	offset := (page - 1) * pageSize
	err = b.db.Where("status = ?", 1).Offset(offset).Limit(pageSize).Find(&books).Error
	if err != nil {
		return nil, 0, err
	}
	return books, total, nil
}
func (b *BookDAO) SearchBooksWithPage(keyword string, page, pageSize int) ([]*model.Book, int64, error) {
	var books []*model.Book
	var total int64
	// 搜索
	searchCondition := b.db.Debug().Where("status = ? AND (title LIKE ? OR author LIKE ? OR description LIKE ?)",
		1, "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	// 获取总数
	err := searchCondition.Model(&model.Book{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	// 根据页数偏移量查找
	offset := (page - 1) * pageSize
	err = searchCondition.Offset(offset).Limit(pageSize).Find(&books).Error
	if err != nil {
		return nil, 0, err
	}
	return books, total, nil
}
func (b *BookDAO) GetBookDetail(id int) (*model.Book, error) {
	var book *model.Book
	// 获取详细信息
	err := b.db.Debug().Where("status = ?", 1).First(&book, id).Error
	if err != nil {
		return nil, err
	}
	return book, nil
}
