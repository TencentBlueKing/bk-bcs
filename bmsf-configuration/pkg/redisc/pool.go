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

package redisc

import (
	"math/rand"
)

// pool is redis backend list wraper.
type pool []*backend

// up returns all backends that in up-status.
func (p pool) up() pool {
	cb := func(b *backend) bool {
		return b.isAlive()
	}
	return p.all(cb)
}

// rand returns a random backend from pool.
func (p pool) rand() *backend {
	if size := len(p); size > 0 {
		return p[rand.Intn(size)]
	}
	return nil
}

// at picks a redis client backend by the pos in pool.
func (p pool) at(pos int) *backend {
	n := len(p)

	if n < 1 {
		return nil
	}

	if pos %= n; pos < 0 {
		pos *= -1
	}

	return p[pos]
}

// all returns backends selected by cb func.
func (p pool) all(cb func(*backend) bool) pool {
	res := make(pool, 0, len(p))

	for _, b := range p {
		if cb(b) {
			res = append(res, b)
		}
	}
	return res
}
