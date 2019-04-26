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

package metrichelper

import (
	"bk-bcs/bcs-common/common/metric"
	"sync"
)

//StringCollector counter for string
type StringCollector struct {
	name   string
	help   string
	value  string
	Locker sync.RWMutex
}

//NewStringCollector new a string collector
func NewStringCollector(name, help string) *StringCollector {
	return &StringCollector{
		name:  name,
		help:  help,
		value: "",
	}
}

// GetResult get string collector result
func (c *StringCollector) GetResult() (*metric.MetricResult, error) {
	c.Locker.RLock()
	defer c.Locker.RUnlock()
	v, err := metric.FormFloatOrString(c.value)
	if err != nil {
		return nil, err
	}
	return &metric.MetricResult{
		Value: v,
	}, nil
}

// GetMeta get string collector meta data
func (c *StringCollector) GetMeta() *metric.MetricMeta {
	return &metric.MetricMeta{
		Name: c.name,
		Help: c.help,
	}
}

// GetValue get string collector value
func (c *StringCollector) GetValue() (*metric.FloatOrString, error) {
	c.Locker.RLock()
	defer c.Locker.RUnlock()
	return metric.FormFloatOrString(c.value)
}

// Reset reset string collector value to val
func (c *StringCollector) Reset(val string) {
	c.Locker.Lock()
	defer c.Locker.Unlock()
	c.value = val
}
