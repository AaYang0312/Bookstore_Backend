package service

import (
	"bookstore-manager/model"
	"bookstore-manager/repository"
	"errors"
	"strings"
)

var ErrCarouselTitleRequired = errors.New("轮播图标题不能为空")

type CarouselService struct {
	CarouselDB *repository.CarouselDAO
}

func NewCarouselService() *CarouselService {
	return &CarouselService{
		CarouselDB: repository.NewCarouselDAO(),
	}
}

func (c *CarouselService) GetCarouselList() ([]*model.Carousel, error) {
	return c.CarouselDB.GetActiveCarousels()
}

func (c *CarouselService) AdminGetCarousels() ([]*model.Carousel, error) {
	return c.CarouselDB.GetAdminCarousels()
}

func (c *CarouselService) AdminCreateCarousel(carousel *model.Carousel) (*model.Carousel, error) {
	normalizeCarousel(carousel)
	if carousel.Title == "" {
		return nil, ErrCarouselTitleRequired
	}
	if err := c.CarouselDB.CreateCarousel(carousel); err != nil {
		return nil, err
	}
	return c.CarouselDB.GetCarouselByID(carousel.ID)
}

func (c *CarouselService) AdminUpdateCarousel(carousel *model.Carousel) (*model.Carousel, error) {
	normalizeCarousel(carousel)
	if carousel.Title == "" {
		return nil, ErrCarouselTitleRequired
	}
	if err := c.CarouselDB.UpdateCarousel(carousel); err != nil {
		return nil, err
	}
	return c.CarouselDB.GetCarouselByID(carousel.ID)
}

func (c *CarouselService) AdminUpdateCarouselStatus(id int, isActive bool) (*model.Carousel, error) {
	if err := c.CarouselDB.UpdateCarouselStatus(id, isActive); err != nil {
		return nil, err
	}
	return c.CarouselDB.GetCarouselByID(id)
}

func (c *CarouselService) AdminDeleteCarousel(id int) error {
	return c.CarouselDB.DeleteCarousel(id)
}

func normalizeCarousel(carousel *model.Carousel) {
	carousel.Title = strings.TrimSpace(carousel.Title)
	carousel.Description = strings.TrimSpace(carousel.Description)
	carousel.ImageURL = strings.TrimSpace(carousel.ImageURL)
	carousel.LinkURL = strings.TrimSpace(carousel.LinkURL)
}
