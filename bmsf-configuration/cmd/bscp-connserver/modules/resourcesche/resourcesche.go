/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package resourcesche

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/spf13/viper"

	pbcommon "bk-bscp/internal/protocol/common"
	"bk-bscp/internal/structs"
	"bk-bscp/pkg/logger"
	"bk-bscp/pkg/rssche"
)

// ConnServerList is connserver resource list.
type ConnServerList []*structs.ConnServer

// sort server list base on load.
func (l ConnServerList) sort(m, n int) bool {
	return l[m].ConnCount < l[n].ConnCount
}

// ConnServerRes is connserver resource object, get resource instances from it.
type ConnServerRes struct {
	// config viper as context here.
	viper *viper.Viper

	// resource local cache.
	reses map[string]*structs.ConnServer

	// make operations on local cache in safely.
	mu sync.RWMutex
}

// NewConnServerRes creates a new ConnServerRes instance.
func NewConnServerRes(viper *viper.Viper) *ConnServerRes {
	return &ConnServerRes{
		viper: viper,
		reses: make(map[string]*structs.ConnServer),
	}
}

// Update impls the rssche.Resource interface, update local cache data base on etcd.
func (r *ConnServerRes) Update(updates []*rssche.Update) error {
	for _, update := range updates {
		switch update.Op {
		case rssche.Put:
			logger.Warn("connserver resource updating, PUT %+v", update.Metadata)

			// new resource adding now.
			st := structs.ConnServer{}
			if err := json.Unmarshal([]byte(update.Metadata), &st); err != nil {
				continue
			}

			r.mu.Lock()
			r.reses[r.resKey(&st)] = &st
			r.mu.Unlock()

		case rssche.Delete:
			logger.Warn("connserver resource updating, DELETE %+v", update.Metadata)

			// resource shutdown now.
			st := structs.ConnServer{}
			if err := json.Unmarshal([]byte(update.Metadata), &st); err != nil {
				continue
			}

			r.mu.Lock()
			delete(r.reses, r.resKey(&st))
			r.mu.Unlock()
		}
	}

	return nil
}

func (r *ConnServerRes) resKey(res *structs.ConnServer) string {
	return fmt.Sprintf("%s:%d", res.IP, res.Port)
}

// Query impls the rssche.Resource interface, query resource
// list from local cache.
func (r *ConnServerRes) Query() ([]interface{}, error) {
	reses := make([]interface{}, 0)

	r.mu.RLock()
	for _, v := range r.reses {
		reses = append(reses, v)
	}
	r.mu.RUnlock()

	return reses, nil
}

// NodeCount returns the count of access nodes.
func (r *ConnServerRes) NodeCount() int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return int64(len(r.reses))
}

// Schedule returns the resources list after sort.
func (r *ConnServerRes) Schedule() ([]*pbcommon.Endpoint, error) {
	reses, err := r.Query()
	if err != nil {
		return nil, err
	}

	connServerList := ConnServerList{}
	for _, res := range reses {
		c, ok := res.(*structs.ConnServer)
		if !ok {
			continue
		}
		connServerList = append(connServerList, c)
	}

	// sort resource list by load information.
	sort.Slice(connServerList, connServerList.sort)
	resources := []*pbcommon.Endpoint{}

	for _, connServer := range connServerList {
		resources = append(resources, &pbcommon.Endpoint{IP: connServer.IP, Port: int32(connServer.Port)})
	}

	if len(resources) == 0 {
		return nil, errors.New("can't get configserver resource to schedule")
	}

	// high/low split and random.
	midIndex := len(resources) / 2

	lowQue := resources[0:midIndex]
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(lowQue), func(i, j int) { lowQue[i], lowQue[j] = lowQue[j], lowQue[i] })

	highQue := resources[midIndex:]
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(highQue), func(i, j int) { highQue[i], highQue[j] = highQue[j], highQue[i] })

	// final resources queue.
	fQue := []*pbcommon.Endpoint{}
	fQue = append(fQue, lowQue...)
	fQue = append(fQue, highQue...)

	// check limit of length before return.
	limit := r.viper.GetInt("server.schedule.nodesLimit")

	if len(fQue) > limit {
		return fQue[0:limit], nil
	}
	return fQue, nil
}
