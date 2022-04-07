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

package storage

import (
	"time"

	"github.com/patrickmn/go-cache"
)

// SlotCache :
type SlotCache struct {
	Slot              *cache.Cache
	DefaultExpiration time.Duration
}

// NewSlotCache :
func NewSlotCache() (*SlotCache, error) {
	c := SlotCache{
		Slot:              cache.New(5*time.Minute, 10*time.Minute),
		DefaultExpiration: cache.DefaultExpiration,
	}
	return &c, nil
}

// LocalCache :
var LocalCache, _ = NewSlotCache()
