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

package cache

import (
	"sync"
)

//NewCache create cache with designated ObjectKeyFunc
func NewCache(kfunc ObjectKeyFunc) Store {
	return &Cache{
		dataMap: make(map[string]interface{}),
		keyFunc: kfunc,
	}
}

//CreateCache create Cache object
func CreateCache(kfunc ObjectKeyFunc) *Cache {
	return &Cache{
		dataMap: make(map[string]interface{}),
		keyFunc: kfunc,
	}
}

//Cache implements Store interface with a safe map
type Cache struct {
	lock    sync.RWMutex           //lock for datamap
	dataMap map[string]interface{} //map to hold data
	keyFunc ObjectKeyFunc          //func to create key
}

// Add inserts an item into the cache.
func (c *Cache) Add(obj interface{}) error {
	key, err := c.keyFunc(obj)
	if err != nil {
		return KeyError{obj, err}
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	c.dataMap[key] = obj
	return nil
}

// Update sets an item in the cache to its updated state.
func (c *Cache) Update(obj interface{}) error {
	return c.Add(obj)
}

// Delete removes an item from the cache.
func (c *Cache) Delete(obj interface{}) error {
	key, err := c.keyFunc(obj)
	if err != nil {
		return KeyError{obj, err}
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	if _, found := c.dataMap[key]; found {
		delete(c.dataMap, key)
	} else {
		return DataNoExist{obj}
	}
	return nil
}

// Get returns the requested item
func (c *Cache) Get(obj interface{}) (item interface{}, exists bool, err error) {
	key, err := c.keyFunc(obj)
	if err != nil {
		return nil, false, KeyError{obj, err}
	}
	c.lock.RLock()
	defer c.lock.RUnlock()
	item, exists = c.dataMap[key]
	return item, exists, nil
}

// List returns a list of all the items.
func (c *Cache) List() []interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()
	data := make([]interface{}, 0, len(c.dataMap))
	for _, item := range c.dataMap {
		data = append(data, item)
	}
	return data
}

// ListKeys returns a list of all the keys of the objects currently
// in the cache.
func (c *Cache) ListKeys() []string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	list := make([]string, 0, len(c.dataMap))
	for key := range c.dataMap {
		list = append(list, key)
	}
	return list
}

// GetByKey returns the request item, or exists=false.
func (c *Cache) GetByKey(key string) (item interface{}, exists bool, err error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	item, exists = c.dataMap[key]
	return item, exists, nil
}

//Num will return data counts in Store
func (c *Cache) Num() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return len(c.dataMap)
}

//Clear will drop all data in Store
func (c *Cache) Clear() {
	c.lock.Lock()
	defer c.lock.Unlock()
	for key := range c.dataMap {
		delete(c.dataMap, key)
	}
}

// Replace use new list to replace data in Cache
func (c *Cache) Replace(list []interface{}) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	for _, item := range list {
		key, err := c.keyFunc(item)
		if err != nil {
			return KeyError{item, err}
		}
		c.dataMap[key] = item
	}
	return nil
}
