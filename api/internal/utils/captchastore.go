package utils

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mojocn/base64Captcha"
	"github.com/zeromicro/go-zero/core/logx"
	// "github.com/zeromicro/go-zero/core/stores/redis"
)

// CaptchaStore is a custom store for captchas using Redis
type CaptchaStore struct {
	redisClient *redis.Client
}

// NewCaptchaStore creates a new CaptchaStore with Redis
func NewCaptchaStore(redisAddr string, pass string) *CaptchaStore {

	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: pass,
	})

	// config := redis.RedisConf{
	// 	Host: redisAddr,
	// 	Type: "node",
	// 	Pass: pass,
	// }

	// client := redis.MustNewRedis(config)

	return &CaptchaStore{
		redisClient: client,
	}
}

// Set sets the captcha id and value in Redis
func (s *CaptchaStore) Set(id string, value string) error {
	ctx := context.Background()
	return s.redisClient.Set(ctx, id, value, 5*time.Minute).Err() // Set expiration to 10 minutes
}

// Get retrieves a captcha value by its id from Redis
func (s *CaptchaStore) Get(id string, clear bool) string {
	ctx := context.Background()
	value, err := s.redisClient.Get(ctx, id).Result()
	if err != nil {
		return ""
	}
	if clear {
		s.redisClient.Del(ctx, id)
	}
	return value
}

// Verify checks if the provided answer is correct for the given captcha id
func (s *CaptchaStore) Verify(id, answer string, clear bool) bool {
	v := s.Get(id, clear)
	return v == answer
}

var store *CaptchaStore

// InitCaptchaStore initializes the captcha store with Redis
func InitCaptchaStore(redisAddr string, pass string) {
	store = NewCaptchaStore(redisAddr, pass)
}

// GenerateCaptcha generates a new captcha
func GenerateCaptcha() (string, string, error) {
	driver := base64Captcha.NewDriverDigit(40, 120, 4, 0.4, 20)

	// driver := base64Captcha.NewDriverString(
	// 	80,  // height
	// 	240, // width
	// 	0,   // no noise
	// 	base64Captcha.OptionShowHollowLine,
	// 	5,                                     // length
	// 	"234567890abcdefghijklmnopqrstuvwxyz", // source
	// 	&color.RGBA{R: 0, G: 0, B: 0, A: 0},   // foreground color
	// 	nil,                                   // fonts
	// 	[]string{},
	// )
	captcha := base64Captcha.NewCaptcha(driver, store)

	id, b64s, ans, err := captcha.Generate()
	store.Set(id, ans)
	logx.Infof("data %s = %s", id, ans)
	return id, b64s, err
}

// VerifyCaptcha verifies the captcha answer
func VerifyCaptcha(id, answer string) bool {

	if id == "" || answer == "" {
		return false
	}

	return store.Verify(id, answer, true)
}
