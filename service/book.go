package service

import (
	"bookstore-manager/cache"
	"bookstore-manager/model"
	"bookstore-manager/repository"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

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
	if books, ok := b.BookCache.GetHotBooks(limit); ok {
		return books, nil
	}
	books, err := b.BookDB.GetHotBooks(limit)
	if err != nil {
		return nil, err
	}
	b.BookCache.SetHotBooks(limit, books)
	return books, nil
}

func (b *BookService) GetNewBooks(limit int) ([]*model.Book, error) {
	if books, ok := b.BookCache.GetNewBooks(limit); ok {
		return books, nil
	}
	books, err := b.BookDB.GetNewBooks(limit)
	if err != nil {
		return nil, err
	}
	b.BookCache.SetNewBooks(limit, books)
	return books, nil
}

func (b *BookService) GetBooksByPage(page, pageSize int) ([]*model.Book, int64, error) {
	if books, total, ok := b.BookCache.GetBookList(page, pageSize); ok {
		return books, total, nil
	}
	books, total, err := b.BookDB.GetBooksByPage(page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	b.BookCache.SetBookList(page, pageSize, books, total)
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
