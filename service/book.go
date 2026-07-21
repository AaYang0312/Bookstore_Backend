package service

import (
	"bookstore-manager/cache"
	"bookstore-manager/model"
	"bookstore-manager/repository"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

var ErrBookCategoryNotFound = errors.New("图书分类不存在")

type BookService struct {
	BookDB    *repository.BookDAO
	BookCache *cache.BookCache
}

func NewBookService() *BookService {
	return &BookService{
		BookDB:    repository.NewBookDAO(),
		BookCache: cache.NewBookCache(),
	}
}

func (b *BookService) GetHotBooks(limit int) ([]*model.Book, error) {
	books, version, found := b.BookCache.GetHotBooks(limit)
	if found {
		return books, nil
	}
	books, err := b.BookDB.GetHotBooks(limit)
	if err != nil {
		return nil, err
	}
	b.BookCache.SetHotBooks(limit, version, books)
	return books, nil
}

func (b *BookService) GetNewBooks(limit int) ([]*model.Book, error) {
	books, version, found := b.BookCache.GetNewBooks(limit)
	if found {
		return books, nil
	}
	books, err := b.BookDB.GetNewBooks(limit)
	if err != nil {
		return nil, err
	}
	b.BookCache.SetNewBooks(limit, version, books)
	return books, nil
}

func (b *BookService) GetBooksByPage(page, pageSize int) ([]*model.Book, int64, error) {
	books, total, version, found := b.BookCache.GetBookList(page, pageSize)
	if found {
		return books, total, nil
	}
	books, total, err := b.BookDB.GetBooksByPage(page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	b.BookCache.SetBookList(page, pageSize, version, books, total)
	return books, total, nil
}

func (b *BookService) SearchBooksWithPage(keyword string, page, pageSize int) ([]*model.Book, int64, error) {
	// 搜索结果不缓存，关键词组合太多
	return b.BookDB.SearchBooksWithPage(keyword, page, pageSize)
}

func (b *BookService) GetBookDetail(id int) (*model.Book, error) {
	if book, found := b.BookCache.GetBookDetail(id); found {
		if book == nil {
			return nil, gorm.ErrRecordNotFound
		}
		return book, nil
	}

	// singleflight 合并并发请求，防止缓存击穿
	sfKey := fmt.Sprintf("book:detail:%d", id)
	val, err := b.BookCache.DoWithSingleFlight(sfKey, func() (any, error) {
		book, err := b.BookDB.GetBookDetail(id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				b.BookCache.SetBookDetail(id, nil)
			}
			return nil, err
		}
		b.BookCache.SetBookDetail(id, book)
		return book, nil
	})
	if err != nil {
		return nil, err
	}
	return val.(*model.Book), nil
}

func (b *BookService) GetBooksByCategory(categoryName string, page int, pageSize int) ([]*model.Book, int64, error) {
	return b.BookDB.GetBooksByCategory(categoryName, page, pageSize)
}

func (b *BookService) AdminGetBooks(keyword string, status *int, page, pageSize int) ([]*repository.AdminBook, int64, error) {
	return b.BookDB.GetAdminBooks(keyword, status, page, pageSize)
}

func (b *BookService) AdminCreateBook(book *model.Book, categoryName string) (*repository.AdminBook, error) {
	if err := b.resolveCategory(book, categoryName); err != nil {
		return nil, err
	}
	if book.Language == "" {
		book.Language = "中文"
	}
	if book.Format == "" {
		book.Format = "平装"
	}
	if err := b.BookDB.CreateBook(book); err != nil {
		return nil, err
	}
	b.BookCache.InvalidateBook(book.ID)
	return b.BookDB.GetAdminBookByID(book.ID)
}

func (b *BookService) AdminUpdateBook(book *model.Book, categoryName string) (*repository.AdminBook, error) {
	if _, err := b.BookDB.GetAdminBookByID(book.ID); err != nil {
		return nil, err
	}
	if err := b.resolveCategory(book, categoryName); err != nil {
		return nil, err
	}
	if book.Language == "" {
		book.Language = "中文"
	}
	if book.Format == "" {
		book.Format = "平装"
	}
	if err := b.BookDB.UpdateBook(book); err != nil {
		return nil, err
	}
	b.BookCache.InvalidateBook(book.ID)
	return b.BookDB.GetAdminBookByID(book.ID)
}

func (b *BookService) AdminUpdateBookStatus(id, status int) (*repository.AdminBook, error) {
	if err := b.BookDB.UpdateBookStatus(id, status); err != nil {
		return nil, err
	}
	b.BookCache.InvalidateBook(id)
	return b.BookDB.GetAdminBookByID(id)
}

func (b *BookService) AdminUpdateBookStock(id, stock int) (*repository.AdminBook, error) {
	if err := b.BookDB.UpdateBookStock(id, stock); err != nil {
		return nil, err
	}
	b.BookCache.InvalidateBook(id)
	return b.BookDB.GetAdminBookByID(id)
}

func (b *BookService) resolveCategory(book *model.Book, categoryName string) error {
	if categoryName == "" {
		return nil
	}
	categoryID, err := b.BookDB.GetCategoryIDByName(categoryName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrBookCategoryNotFound
		}
		return err
	}
	book.CategoryID = categoryID
	return nil
}
