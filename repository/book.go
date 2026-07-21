package repository

import (
	"bookstore-manager/global"
	"bookstore-manager/model"

	"gorm.io/gorm"
)

type BookDAO struct {
	db *gorm.DB
}

// AdminBook 是管理端图书列表的返回结构，在图书字段之外补充分组名称。
type AdminBook struct {
	model.Book
	CategoryName string `json:"category_name"`
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

// GetBookByID 根据ID获取书籍（只返回上架状态）
func (b *BookDAO) GetBookByID(id int) (*model.Book, error) {
	var book model.Book
	err := b.db.Where("status = ?", 1).First(&book, id).Error
	if err != nil {
	} else {
	}
	return &book, err
}

func (b *BookDAO) GetBooksByCategory(
	categoryName string,
	page int,
	pageSize int,
) ([]*model.Book, int64, error) {
	var books []*model.Book
	var total int64

	query := b.db.Model(&model.Book{}).
		Joins("JOIN categories ON categories.id = books.category_id").
		Where("books.status = ? AND categories.is_active = ? AND categories.name = ?",
			1,
			true,
			categoryName,
		)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize

	if err := query.
		Offset(offset).
		Limit(pageSize).
		Find(&books).Error; err != nil {
		return nil, 0, err
	}

	return books, total, nil
}

// GetAdminBooks 查询管理端图书列表，包含上架和下架图书。
func (b *BookDAO) GetAdminBooks(keyword string, status *int, page, pageSize int) ([]*AdminBook, int64, error) {
	var books []*AdminBook
	var total int64

	query := b.db.Model(&model.Book{}).
		Joins("LEFT JOIN categories ON categories.id = books.category_id")
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("books.title LIKE ? OR books.author LIKE ? OR books.isbn LIKE ?", like, like, like)
	}
	if status != nil {
		query = query.Where("books.status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.
		Select("books.*, categories.name AS category_name").
		Order("books.created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Scan(&books).Error; err != nil {
		return nil, 0, err
	}

	return books, total, nil
}

func (b *BookDAO) GetAdminBookByID(id int) (*AdminBook, error) {
	var book AdminBook
	err := b.db.Model(&model.Book{}).
		Select("books.*, categories.name AS category_name").
		Joins("LEFT JOIN categories ON categories.id = books.category_id").
		Where("books.id = ?", id).
		Scan(&book).Error
	if err != nil {
		return nil, err
	}
	if book.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &book, nil
}

func (b *BookDAO) GetCategoryIDByName(name string) (uint, error) {
	var category model.Category
	if err := b.db.Select("id").Where("name = ?", name).First(&category).Error; err != nil {
		return 0, err
	}
	return uint(category.ID), nil
}

func (b *BookDAO) CreateBook(book *model.Book) error {
	return b.db.Create(book).Error
}

func (b *BookDAO) UpdateBook(book *model.Book) error {
	result := b.db.Model(&model.Book{}).Where("id = ?", book.ID).Updates(map[string]any{
		"title":        book.Title,
		"author":       book.Author,
		"price":        book.Price,
		"discount":     book.Discount,
		"type":         book.Type,
		"category_id":  book.CategoryID,
		"stock":        book.Stock,
		"status":       book.Status,
		"description":  book.Description,
		"cover_url":    book.CoverURL,
		"isbn":         book.ISBN,
		"publisher":    book.Publisher,
		"publish_date": book.PublishDate,
		"pages":        book.Pages,
		"language":     book.Language,
		"format":       book.Format,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (b *BookDAO) UpdateBookStatus(id, status int) error {
	result := b.db.Model(&model.Book{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		var count int64
		if err := b.db.Model(&model.Book{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return gorm.ErrRecordNotFound
		}
	}
	return nil
}

func (b *BookDAO) UpdateBookStock(id, stock int) error {
	result := b.db.Model(&model.Book{}).Where("id = ?", id).Update("stock", stock)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		var count int64
		if err := b.db.Model(&model.Book{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return gorm.ErrRecordNotFound
		}
	}
	return nil
}
