package etcd

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/RichardKnop/machinery/v2/config"
	"github.com/RichardKnop/machinery/v2/locks/iface"
	"github.com/RichardKnop/machinery/v2/log"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

var (
	// ErrLockFailed ..
	ErrLockFailed = errors.New("etcd lock: failed to acquire lock")
)

type etcdLock struct {
	ctx     context.Context
	client  *clientv3.Client
	retries int
}

// New ..
func New(ctx context.Context, conf *config.Config, retries int) (iface.Lock, error) {
	etcdConf := clientv3.Config{
		Endpoints:   []string{conf.Lock},
		Context:     ctx,
		DialTimeout: time.Second * 5,
		TLS:         conf.TLSConfig,
	}

	client, err := clientv3.New(etcdConf)
	if err != nil {
		return nil, err
	}

	lock := etcdLock{
		ctx:     ctx,
		client:  client,
		retries: retries,
	}

	return &lock, nil
}

// LockWithRetries ..
func (l *etcdLock) LockWithRetries(key string, unixTsToExpireNs int64) error {
	i := 0
	for ; i < l.retries; i++ {
		err := l.Lock(key, unixTsToExpireNs)
		if err == nil {
			// 成功拿到锁，返回
			return nil
		}

		log.DEBUG.Printf("acquired lock=%s failed, retries=%d, err=%s", key, i, err)
		time.Sleep(time.Millisecond * 100)
	}

	log.INFO.Printf("acquired lock=%s failed, retries=%d", key, i)
	return ErrLockFailed
}

// Lock ..
func (l *etcdLock) Lock(key string, unixTsToExpireNs int64) error {
	now := time.Now().UnixNano()
	ttl := time.Duration(unixTsToExpireNs + 1 - now)

	// 创建一个新的session
	s, err := concurrency.NewSession(l.client, concurrency.WithTTL(int(ttl.Seconds())))
	if err != nil {
		return err
	}
	defer s.Orphan()

	lockKey := fmt.Sprintf("/machinery/v2/lock/%s", strings.TrimRight(key, "/"))
	m := concurrency.NewMutex(s, lockKey)

	ctx, cancel := context.WithTimeout(l.ctx, time.Second*2)
	defer cancel()

	if err := m.Lock(ctx); err != nil {
		_ = s.Close()
		return err
	}

	log.INFO.Printf("acquired lock=%s, duration=%s", key, ttl)
	return nil
}
