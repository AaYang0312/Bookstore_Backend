package service

import (
	"bookstore-manager/cache"
	"bookstore-manager/model"
	"bookstore-manager/repository"
	"errors"
	"strings"
)

var ErrCarouselTitleRequired = errors.New("轮播图标题不能为空")

type CarouselService struct {
	CarouselDB    *repository.CarouselDAO
	CarouselCache *cache.CarouselCache
}

func NewCarouselService() *CarouselService {
	return &CarouselService{
		CarouselDB:    repository.NewCarouselDAO(),
		CarouselCache: cache.NewCarouselCache(),
	}
}

func (c *CarouselService) GetCarouselList() ([]*model.Carousel, error) {
	if carousels, ok := c.CarouselCache.GetActiveCarousels(); ok {
		return carousels, nil
	}

	val, err := c.CarouselCache.DoWithSingleFlight("carousel:active", func() (any, error) {
		carousels, err := c.CarouselDB.GetActiveCarousels()
		if err != nil {
			return nil, err
		}
		c.CarouselCache.SetActiveCarousels(carousels)
		return carousels, nil
	})
	if err != nil {
		return nil, err
	}
	return val.([]*model.Carousel), nil
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
	c.CarouselCache.InvalidateActiveCarousels()
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
	c.CarouselCache.InvalidateActiveCarousels()
	return c.CarouselDB.GetCarouselByID(carousel.ID)
}

func (c *CarouselService) AdminUpdateCarouselStatus(id int, isActive bool) (*model.Carousel, error) {
	if err := c.CarouselDB.UpdateCarouselStatus(id, isActive); err != nil {
		return nil, err
	}
	c.CarouselCache.InvalidateActiveCarousels()
	return c.CarouselDB.GetCarouselByID(id)
}

func (c *CarouselService) AdminDeleteCarousel(id int) error {
	if err := c.CarouselDB.DeleteCarousel(id); err != nil {
		return err
	}
	c.CarouselCache.InvalidateActiveCarousels()
	return nil
}

func normalizeCarousel(carousel *model.Carousel) {
	carousel.Title = strings.TrimSpace(carousel.Title)
	carousel.Description = strings.TrimSpace(carousel.Description)
	carousel.ImageURL = strings.TrimSpace(carousel.ImageURL)
	carousel.LinkURL = strings.TrimSpace(carousel.LinkURL)
}
