package cloudcollector

import (
	"sync"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
)

type StatusCache struct {
	cache map[string][]*cloud.BackendHealthStatus
	mutex sync.Mutex
}

func NewStatusCache() StatusCache {
	return StatusCache{
		cache: make(map[string][]*cloud.BackendHealthStatus),
	}
}

func (sc *StatusCache) UpdateCache(newData map[string][]*cloud.BackendHealthStatus) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	//clear old data
	sc.cache = make(map[string][]*cloud.BackendHealthStatus)
	//update new data
	for k, v := range newData {
		sc.cache[k] = v
	}
}

func (sc *StatusCache) Get() map[string][]*cloud.BackendHealthStatus {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	return sc.cache
}
