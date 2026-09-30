package controller

import (
	"bookstore-manager/storage"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UploadController struct{}

func NewUploadController() *UploadController {
	return &UploadController{}
}

// AdminUploadImage 管理端图片上传：multipart 字段 file 为图片文件，
// 可选字段 type 为 covers（书籍封面，默认）/ carousel（轮播图）。
func (u *UploadController) AdminUploadImage(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "缺少上传文件"})
		return
	}
	fileURL, err := storage.UploadImage(c.PostForm("type"), fileHeader)
	if err != nil {
		if errors.Is(err, storage.ErrInvalidImageType) ||
			errors.Is(err, storage.ErrImageTooLarge) ||
			errors.Is(err, storage.ErrStorageNotReady) {
			c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "上传图片失败",
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "上传成功",
		"data":    gin.H{"url": fileURL},
	})
}
