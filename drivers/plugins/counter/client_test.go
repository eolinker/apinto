package counter

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	redis "github.com/go-redis/redis/v8"
)

type demoClient struct {
}

func (d *demoClient) Get(variables map[string]string) (int64, error) {
	return 100, nil
}

func TestLocalCounter(t *testing.T) {
	client := &demoClient{}
	lc := NewLocalCounter("test", client)
	wg := sync.WaitGroup{}
	wg.Add(100)
	for i := 0; i < 100; i++ {
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))
		go func() {
			defer wg.Done()
			count := rand.Int63n(20)
			err := lc.Lock(count)
			if err != nil {
				t.Error(err)
				return
			}
			condition := rand.Intn(100)
			switch condition % 2 {
			case 0:
				err = lc.Complete(count)
			case 1:
				err = lc.RollBack(count)
			}
			if err != nil {
				t.Error(err)
				return
			}
		}()
	}
	wg.Wait()
}

func TestRedisCounter(t *testing.T) {
	client := &demoClient{}
	key := "apinto-apiddww"
	lc := NewLocalCounter(key, client)
	redisConn := redis.NewClient(&redis.Options{
		Addr:     "172.18.65.42:6380",
		Password: "password", // 如果有密码，请填写密码
		DB:       9,          // 选择数据库，默认为0
	})
	rc := NewRedisCounter(key, redisConn, client)
	wg := sync.WaitGroup{}
	wg.Add(100)
	for i := 0; i < 100; i++ {
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))
		go func() {
			defer wg.Done()
			count := rand.Int63n(20)
			err := rc.Lock(count)
			if err != nil {
				t.Error(err)
				return
			}
			condition := rand.Intn(100)
			switch condition % 2 {
			case 0:
				err = rc.Complete(count)
			case 1:
				err = rc.RollBack(count)
			}
			if err != nil {
				t.Error(err)
				return
			}
		}()
	}
	wg.Wait()
}
