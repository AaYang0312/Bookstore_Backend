package service

import (
	"bookstore-manager/model"
	"bookstore-manager/repository"
)

type FavoriteService struct {
	FavoriteDAO *repository.FavoriteDAO
}

func NewFavoriteService() *FavoriteService {
	return &FavoriteService{
		FavoriteDAO: repository.NewFavoriteDAO(),
	}
}
func (f *FavoriteService) AddFavorite(userID, bookID int) error {
	return f.FavoriteDAO.AddFavorite(userID, bookID)
}
func (f *FavoriteService) DelFavorite(userID, bookID int) error {
	return f.FavoriteDAO.DelFavorite(userID, bookID)
}
func (f *FavoriteService) GetUserFavorites(userID int, page, pageSize int, timeFilter string) ([]*model.Favorite, int64, error) {
	fav, err := f.FavoriteDAO.GetUserFavorites(userID)
	if err != nil {
		return nil, 0, err
	}
	total := len(fav)
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= total {
		return []*model.Favorite{}, int64(total), nil
	}
	if end >= total {
		end = total
	}
	return fav[start:end], int64(total), nil
}
func (f *FavoriteService) GetUserFavoriteCount(userID int) (int64, error) {
	return f.FavoriteDAO.GetUserFavoriteCount(userID)
}
func (f *FavoriteService) IsFavorite(userID, bookID int) (bool, error) {
	return f.FavoriteDAO.IsFavorite(userID, bookID)
}
