package ratelimit

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Info struct {
	Limit         uint
	RateLimited   bool
	ResetTime     time.Time
	RemainingHits uint
}

type Store interface {
	Limit(key string, c *gin.Context) Info
}

type Options struct {
	ErrorHandler   func(*gin.Context, Info)
	KeyFunc        func(*gin.Context) string
	BeforeResponse func(c *gin.Context, info Info)
}

type user struct {
	ts     int64
	tokens uint
}

type inMemoryStoreType struct {
	rate  int64
	limit uint
	data  *sync.Map
	skip  func(ctx *gin.Context) bool
}

func (i *inMemoryStoreType) Limit(key string, c *gin.Context) Info {
	var u user
	m, ok := i.data.Load(key)
	if !ok {
		u = user{ts: time.Now().Unix(), tokens: i.limit}
	} else {
		u = m.(user)
	}
	if u.ts+i.rate <= time.Now().Unix() {
		u.tokens = i.limit
	}

	if i.skip != nil && i.skip(c) {
		return Info{
			Limit:         i.limit,
			RateLimited:   false,
			ResetTime:     time.Now().Add(time.Duration((i.rate - (time.Now().Unix() - u.ts)) * time.Second.Nanoseconds())),
			RemainingHits: u.tokens,
		}
	}

	if u.tokens <= 0 {
		return Info{
			Limit:         i.limit,
			RateLimited:   true,
			ResetTime:     time.Now().Add(time.Duration((i.rate - (time.Now().Unix() - u.ts)) * time.Second.Nanoseconds())),
			RemainingHits: 0,
		}
	}
	u.tokens--
	u.ts = time.Now().Unix()
	i.data.Store(key, u)

	return Info{
		Limit:         i.limit,
		RateLimited:   false,
		ResetTime:     time.Now().Add(time.Duration((i.rate - (time.Now().Unix() - u.ts)) * time.Second.Nanoseconds())),
		RemainingHits: u.tokens,
	}
}

func clearMemoryInBackground(data *sync.Map, rate int64) {
	for {
		data.Range(func(key, value any) bool {
			if value.(user).ts+rate < time.Now().Unix() {
				data.Delete(key)
			}
			return true
		})
		time.Sleep(time.Minute)
	}
}

type InMemoryOptions struct {
	Rate  time.Duration
	Limit uint
	Skip  func(*gin.Context) bool
}

func InMemoryStore(options *InMemoryOptions) Store {
	data := &sync.Map{}
	store := inMemoryStoreType{int64(options.Rate.Seconds()), options.Limit, data, options.Skip}
	go clearMemoryInBackground(data, store.rate)
	return &store
}
