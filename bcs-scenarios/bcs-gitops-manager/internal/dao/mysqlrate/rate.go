/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package mysqlrate

import (
	"sync"

	"gorm.io/gorm"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/throttle"
)

var (
	rateLimit *RateLimit
	once      sync.Once
)

// RateInterface defines the interface the db rate limit
type RateInterface interface {
	Table(table string) *gorm.DB
}

// RateLimit defines the limit of mysql db
type RateLimit struct {
	sync.Mutex
	*gorm.DB

	limitQPS    int64
	rateLimiter throttle.RateLimiter
}

// NewRateLimit create the instance of mysql rate limit
func NewRateLimit(db *gorm.DB, limitQPS int64) RateInterface {
	once.Do(func() {
		rateLimit = &RateLimit{
			DB:          db,
			limitQPS:    limitQPS,
			rateLimiter: throttle.NewTokenBucket(limitQPS, limitQPS),
		}
	})
	return rateLimit
}

// Table add rate lock and rate unlock after executed
func (rl *RateLimit) Table(table string) *gorm.DB {
	rl.rateLimiter.Accept()
	return rl.DB.Table(table)
}
