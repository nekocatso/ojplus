package models

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/go-redis/redis"
)

type Verification struct {
	Email string
	Code  string
}

var cacheKey = "auth:email:%s:verification"

func NewVerification(email, code string) *Verification {
	return &Verification{
		Email: email,
		Code:  code,
	}
}

func (v *Verification) Generate(cache *Cache, exp int) {
	if v.Code == "" {
		min := 100000
		max := 999999
		code := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(max-min+1) + min
		v.Code = fmt.Sprint(code)
	}
	key := strings.ToLower(fmt.Sprintf(cacheKey, v.Email))
	cache.Client.Set(key, v.Code, time.Duration(exp)*time.Second)

}

func (v *Verification) Verify(cache *Cache) (bool, error) {
	key := strings.ToLower(fmt.Sprintf(cacheKey, v.Email))
	trueCode, err := cache.Client.Get(key).Result()
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, err
	}
	if trueCode != v.Code {
		return false, nil
	}
	return true, nil
}
