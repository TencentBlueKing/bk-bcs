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

package lib

import (
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

// Manager manager
type Manager struct {
	total   int64
	avgTime int64

	subTotal   int64
	subAvgTime int64

	lt10  int64
	lt50  int64
	lt100 int64
	lt200 int64
	gt200 int64

	sep time.Duration
	sync.Mutex

	title string
}

// NewManager create manager
func NewManager(sep int64, title string) *Manager {
	return &Manager{
		sep:   time.Duration(sep),
		title: title,
	}
}

// Start start
func (m *Manager) Start() {
	for {
		m.Flush()
		time.Sleep(m.sep * time.Second)
	}
}

// Add add
func (m *Manager) Add(t time.Duration) {
	_t := int64(t)
	m.Lock()
	defer m.Unlock()
	m.subAvgTime = (m.subAvgTime*m.subTotal + _t) / (m.subTotal + 1)
	m.subTotal++

	if t < 10*time.Millisecond {
		m.lt10++
	}
	if t < 50*time.Millisecond {
		m.lt50++
	}
	if t < 100*time.Millisecond {
		m.lt100++
	}
	if t < 200*time.Millisecond {
		m.lt200++
	}
	if t >= 200*time.Millisecond {
		m.gt200++
	}
}

// Flush manager do flush
func (m *Manager) Flush() {
	m.Lock()
	defer m.Unlock()
	if m.total+m.subTotal == 0 {
		m.avgTime = 0
	} else {
		m.avgTime = (m.avgTime*m.total + m.subAvgTime*m.subTotal) / (m.total + m.subTotal)
	}
	m.total += m.subTotal

	blog.Infof(
		"\n%s | %s"+
			"\n%s | %-20s %-20s %-20s %-20s [per %d second]"+
			"\n%s | %-20d %-20f %-20d %-20f"+
			"\n%s |"+
			"\n%s | %-20s %-20s %-20s %-20s %-20s"+
			"\n%s | %-20s %-20s %-20s %-20s %-20s"+
			"\n%s | ---------------------------------------------------------------------------------------------------\n",
		m.title, time.Now().String(),
		m.title, "TOTAL", "AVG_TIME(ms)", "SUB_TOTAL", "SUB_AVG_TIME(ms)", m.sep,
		m.title, m.total, float64(m.avgTime)/1e6, m.subTotal, float64(m.subAvgTime)/1e6,
		m.title,
		m.title, "<10ms", "<50ms", "<100ms", "<200ms", ">=200ms",
		m.title, fmt.Sprintf("%.2f%%", float64(m.lt10)/float64(m.total)*100),
		fmt.Sprintf("%.2f%%", float64(m.lt50)/float64(m.total)*100),
		fmt.Sprintf("%.2f%%", float64(m.lt100)/float64(m.total)*100),
		fmt.Sprintf("%.2f%%", float64(m.lt200)/float64(m.total)*100),
		fmt.Sprintf("%.2f%%", float64(m.gt200)/float64(m.total)*100),
		m.title)

	m.subTotal = 0
	m.subAvgTime = 0
}
