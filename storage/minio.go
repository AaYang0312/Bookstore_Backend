package storage

import (
	"bookstore-manager/config"
	"context"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	minioClient   *minio.Client
	bucketName    string
	publicBaseURL string
)

// 上传图片的格式与大小限制
var allowedImageExts = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".webp": "image/webp",
	".gif":  "image/gif",
}

const maxImageSize = 5 << 20 // 5MB

var (
	ErrInvalidImageType = errors.New("仅支持 jpg/jpeg/png/webp/gif 格式的图片")
	ErrImageTooLarge    = errors.New("图片大小不能超过 5MB")
	ErrStorageNotReady  = errors.New("对象存储未初始化")
)

// InitMinIO 初始化 MinIO 客户端，并确保 bucket 存在且为公共只读，
// 使 bucket 内图片可通过 PublicBaseURL 直接被浏览器访问。
func InitMinIO() {
	cfg := config.AppConfig.MinIO
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		log.Fatalln("创建 MinIO 客户端失败：", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		log.Fatalln("连接 MinIO 失败：", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			log.Fatalf("创建 bucket %s 失败: %v", cfg.Bucket, err)
		}
		log.Printf("MinIO bucket 已创建: %s", cfg.Bucket)
	}
	policy := fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {"AWS": ["*"]},
      "Action": ["s3:GetObject"],
      "Resource": ["arn:aws:s3:::%s/*"]
    }
  ]
}`, cfg.Bucket)
	if err := client.SetBucketPolicy(ctx, cfg.Bucket, policy); err != nil {
		log.Fatalf("设置 bucket %s 公共读策略失败: %v", cfg.Bucket, err)
	}

	minioClient = client
	bucketName = cfg.Bucket
	publicBaseURL = strings.TrimRight(cfg.PublicBaseURL, "/")
	log.Printf("MinIO 初始化成功: endpoint=%s bucket=%s public=%s", cfg.Endpoint, cfg.Bucket, publicBaseURL)
}

// UploadImage 校验并上传图片到 MinIO，返回可直接访问的完整 URL。
// fileType 目前为 covers（书籍封面）或 carousel（轮播图），仅用于组织对象键。
func UploadImage(fileType string, fileHeader *multipart.FileHeader) (string, error) {
	if minioClient == nil {
		return "", ErrStorageNotReady
	}
	fileType = strings.TrimSpace(fileType)
	if fileType != "carousel" {
		fileType = "covers"
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	contentType, ok := allowedImageExts[ext]
	if !ok {
		return "", ErrInvalidImageType
	}
	if fileHeader.Size > maxImageSize {
		return "", ErrImageTooLarge
	}

	src, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("读取上传文件失败: %w", err)
	}
	defer src.Close()

	objectKey := fmt.Sprintf("%s/%s/%s%s", fileType, time.Now().Format("200601"), uuid.NewString(), ext)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := minioClient.PutObject(ctx, bucketName, objectKey, src, fileHeader.Size,
		minio.PutObjectOptions{ContentType: contentType}); err != nil {
		return "", fmt.Errorf("上传到 MinIO 失败: %w", err)
	}
	return fmt.Sprintf("%s/%s/%s", publicBaseURL, bucketName, objectKey), nil
}
