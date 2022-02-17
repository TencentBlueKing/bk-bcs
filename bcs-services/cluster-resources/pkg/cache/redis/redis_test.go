/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package redis

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cache"
)

func TestCacheGenKey(t *testing.T) {
	c := NewCache("test", 5*time.Minute)

	assert.Equal(t, fmt.Sprintf("%s:test:abc", CacheKeyPrefix), c.genKey("abc"))
}

func TestCacheSetExistsGet(t *testing.T) {
	c := NewCache("test", 5*time.Minute)
	key := cache.NewStringKey("testKey1")

	// set
	err := c.Set(key, 1, 0)
	assert.NoError(t, err)

	// exists
	exists := c.Exists(key)
	assert.True(t, exists)

	// get
	var a int
	err = c.Get(key, &a)
	assert.NoError(t, err)
	assert.Equal(t, 1, a)
}

func TestDelete(t *testing.T) {
	c := NewCache("test", 5*time.Minute)

	key := cache.NewStringKey("testKey2")

	// do delete
	err := c.Delete(key)
	assert.NoError(t, err)

	// set
	err = c.Set(key, 1, 0)
	assert.NoError(t, err)

	// do it again
	err = c.Delete(key)
	assert.NoError(t, err)
}
