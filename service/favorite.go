package service

import "bookstore-manager/repository"

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
