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
 */

// Package jsonq used to reserve some field from object
package jsonq

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

// ReserveField will reserve some fields from object
func ReserveField(obj interface{}, path []string) ([]byte, error) {
	bs, err := json.Marshal(obj)
	if err != nil {
		return nil, errors.Wrapf(err, "marshal object failed")
	}
	c, err := parsed(bs)
	if err != nil {
		return nil, errors.Wrapf(err, "jsonq parsed bytes failed")
	}
	bs, err = c.ReservePath(path...)
	if err != nil {
		return nil, errors.Wrapf(err, "jsonq reserve path failed")
	}
	return bs, nil
}

// container defines the jsonq traverse instance
type container struct {
	path   string
	object interface{}

	reservePath []string
}

// parsed will unmarshal the result byte to object
func parsed(bs []byte) (container, error) {
	c := container{}
	if err := json.Unmarshal(bs, &c.object); err != nil {
		return c, err
	}
	return c, nil
}

// ReservePath will reserve path from byte
func (c *container) ReservePath(path ...string) ([]byte, error) {
	c.reservePath = path
	c.traverse()
	return json.Marshal(c.object)
}

func (c *container) reservedPathMatched() bool {
	for i := range c.reservePath {
		p := c.reservePath[i]
		if strings.HasPrefix(p, c.path) || strings.HasPrefix(c.path, p) {
			return true
		}
	}
	return false
}

func (c *container) reservedPathMatchedForMetaNode() bool {
	for i := range c.reservePath {
		p := c.reservePath[i]
		if strings.HasPrefix(c.path+".", p+".") {
			return true
		}
	}
	return false
}

func (c *container) traverse() bool {
	if !c.reservedPathMatched() {
		return true
	}
	if c.object == nil {
		return false
	}
	switch reflect.TypeOf(c.object).Kind() {
	case reflect.Map:
		for k, v := range c.object.(map[string]interface{}) {
			newC := container{
				path:        c.path + k + ".",
				object:      v,
				reservePath: c.reservePath,
			}
			if newC.traverse() {
				delete(c.object.(map[string]interface{}), k)
			}
		}
	case reflect.Slice:
		for i := range c.object.([]interface{}) {
			v := c.object.([]interface{})[i]
			newC := container{
				path:        c.path,
				object:      v,
				reservePath: c.reservePath,
			}
			newC.traverse()
		}
	default:
		c.path = strings.TrimSuffix(c.path, ".")
		if !c.reservedPathMatchedForMetaNode() {
			return true
		}
		return false
	}
	return false
}
