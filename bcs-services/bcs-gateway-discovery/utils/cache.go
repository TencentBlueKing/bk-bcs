package utils

import (
	"sync"
	"time"
)

type Cache interface {
	SetData(interface{})
	GetData() interface{}
}

type ResourceCache struct {
	sync.RWMutex
	lasteUpdateTime *time.Time
	timeout         time.Duration
	data            interface{}
}

func NewResourceCache(timeout time.Duration) Cache {
	return &ResourceCache{timeout: timeout}
}

func (rc *ResourceCache) SetData(data interface{}) {
	rc.Lock()
	defer rc.Unlock()
	rc.data = data
	now := time.Now()
	rc.lasteUpdateTime = &now
}

func (rc *ResourceCache) GetData() interface{} {
	rc.RLock()
	defer rc.RUnlock()
	if rc.needRenew() {
		return nil
	}
	return rc.data
}

func (rc *ResourceCache) needRenew() bool {
	if rc.lasteUpdateTime == nil {
		rc.data = nil
		return true
	}
	if time.Now().Sub(*rc.lasteUpdateTime) >= rc.timeout {
		rc.data = nil
		return true
	}
	return false
}
