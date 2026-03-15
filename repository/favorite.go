package repository

import (
	"bookstore-manager/global"
	"bookstore-manager/model"

	"gorm.io/gorm"
)

type FavoriteDAO struct {
	db *gorm.DB
}

func NewFavoriteDAO() *FavoriteDAO {
	return &FavoriteDAO{
		db: global.GetDB(),
	}
}
func (f *FavoriteDAO) AddFavorite(userID, bookID int) error {
	favorite := &model.Favorite{
		UserID: userID,
		BookID: bookID,
	}
	err := f.db.Debug().Create(favorite).Error
	if err != nil {
		return err
	}
	return nil
}
func (f *FavoriteDAO) DelFavorite(userID, bookID int) error {
	err := f.db.Debug().Where("user_id = ? AND book_id = ?", userID, bookID).Delete(&model.Favorite{}).Error
	if err != nil {
		return err
	}
	return nil
}
func (f *FavoriteDAO) GetUserFavorites(userID int) ([]*model.Favorite, error) {
	var fav []*model.Favorite
	err := f.db.Debug().Preload("Book").Where("user_id = ?", userID).Find(&fav).Error
	if err != nil {
		return nil, err
	}
	return fav, nil
}
