package service

import (
	"bookstore-manager/global"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mojocn/base64Captcha"
)

type CaptchaService struct {
	// 图片验证码缓存
	store base64Captcha.Store
}

func NewCaptchaService() *CaptchaService {
	return &CaptchaService{
		store: base64Captcha.DefaultMemStore,
	}
}

type CaptchaResponse struct {
	CaptchaID     string `json:"captcha_id"`
	CaptchaBase64 string `json:"captcha_base64"`
}

func (c *CaptchaService) GenerateCaptcha() (*CaptchaResponse, error) {
	// 1. 创建数字验证码配置
	driver := base64Captcha.NewDriverDigit(
		80,  // 高度
		240, // 宽度
		4,   // 数字长度
		0.7, // 干扰因子
		80,  // 噪音数量
	)
	// 2. 创建验证码
	captcha := base64Captcha.NewCaptcha(driver, c.store)

	// 3. 生成验证码 ID 和 Base64 图片
	id, b64s, answer, err := captcha.Generate()
	if err != nil {
		return nil, err
	}
	// 用redis存储有效期的图片验证码
	log.Println("图片验证码真实 answer: ", answer)
	// key为reidsKey，值为answer
	redisKey := fmt.Sprintf("captcha:%s", id)
	err = global.RedisClient.Set(context.Background(), redisKey, answer, 3*time.Minute).Err()
	if err != nil {
		return nil, err
	}
	return &CaptchaResponse{
		CaptchaID:     id,
		CaptchaBase64: b64s,
	}, nil
}
func (c *CaptchaService) VerifyCaptcha(captchaID, captchaAns string) bool {
	if captchaID == "" || captchaAns == "" {
		return false
	}
	// 从redis获取答案
	ctx := context.Background()
	redisKey := fmt.Sprintf("captcha:%s", captchaID)
	storedAns, err := global.RedisClient.Get(ctx, redisKey).Result()
	if err != nil {
		log.Println("从Redis获取存储的验证码错误：", err)
		return false
	}
	// 比较存储的答案和用户输入的验证码
	isValid := captchaAns == storedAns
	// 比对后删除redis中的验证码
	if isValid {
		global.RedisClient.Del(ctx, redisKey)
		return true
	}
	return false
}
