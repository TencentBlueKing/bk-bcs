/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"context"
	"sync"

	"bscp.io/pkg/logs"
	pbbase "bscp.io/pkg/protocol/core/base"
)

// bizsOfTS are bizs which already have default template spaces
var bizsOfTS BizsOfTmplSpace

// BizsOfTmplSpace are bizs which already have default template spaces with a lock which can be used concurrently
type BizsOfTmplSpace struct {
	sync.Mutex
	Bizs map[uint32]struct{}
}

// Set save a key in the bizs map
func (b *BizsOfTmplSpace) Set(key uint32) {
	b.Lock()
	defer b.Unlock()
	b.Bizs[key] = struct{}{}
}

// Has judge if a key in the bizs map
func (b *BizsOfTmplSpace) Has(key uint32) bool {
	b.Lock()
	defer b.Unlock()
	_, has := b.Bizs[key]
	return has
}

// initBizsOfTmplSpaces get all bizs which already have default template spaces from db
func (p *proxy) initBizsOfTmplSpaces() {
	bizsOfTS.Bizs = make(map[uint32]struct{})

	resp, err := p.cfgClient.GetAllBizsOfTemplateSpaces(context.Background(), &pbbase.EmptyReq{})
	if err != nil {
		logs.Warnf("init bizs of template spaces from db failed, err: %v", err)
		return
	}

	for _, bizID := range resp.BizIds {
		// no need to use lock for init step
		bizsOfTS.Bizs[bizID] = struct{}{}
	}
	logs.Infof("init bizs of template spaces success, len(biz):%d", len(resp.BizIds))

	return
}
