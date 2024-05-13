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

// Package common define common methods
package common

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// WorkerFunc woker function
type WorkerFunc func(done <-chan bool, param interface{})

// GoroutineManager 管理goroutines的启动和重启
type GoroutineManager struct {
	workers    map[string]chan bool // 用于给每个goroutine发送停止信号
	workerFunc WorkerFunc           // 被管理的goroutine执行的函数
	params     map[string]interface{}
	mu         sync.Mutex
}

// NewGoroutineManager 创建一个新的GoroutineManager实例。
// workerFunc 参数是一个WorkerFunc类型，它是分配给每个worker的任务函数。
// 返回的是指向新创建的GoroutineManager的指针。
func NewGoroutineManager(workerFunc WorkerFunc) *GoroutineManager {
	// 初始化GoroutineManager结构体实例。
	// workers 字段是一个map，用于存储worker的信道。
	// workerFunc 字段存储传入的workerFunc。
	// params 字段是一个map，用于存储任何额外参数。
	return &GoroutineManager{
		workers:    make(map[string]chan bool),
		workerFunc: workerFunc,
		params:     make(map[string]interface{}),
	}
}

// Start 启动一个新的goroutine
func (gm *GoroutineManager) Start(id string, param interface{}) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	done := make(chan bool)
	gm.workers[id] = done
	gm.params[id] = param

	go func() {
		gm.workerFunc(done, param)
		// 清理
		gm.mu.Lock()
		delete(gm.workers, id)
		delete(gm.params, id)
		gm.mu.Unlock()
	}()
}

// Restart 重启一个指定ID的goroutine
func (gm *GoroutineManager) Restart(id string, param interface{}) {
	gm.mu.Lock()
	done, exists := gm.workers[id]
	if exists {
		done <- true           // 发送停止信号
		delete(gm.workers, id) // 从map中移除
		fmt.Printf("id: %s 已停止", id)
	}
	gm.mu.Unlock()

	if exists {
		time.Sleep(100 * time.Millisecond) // 稍等goroutine优雅退出
		gm.Start(id, param)                // 重新启动goroutine
		fmt.Printf("id: %s 已重启", id)
	}
}

// List 返回所有正在运行的goroutine的列表
func (gm *GoroutineManager) List() []string {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	ids := make([]string, 0, len(gm.workers))
	for id := range gm.workers {
		ids = append(ids, id)
	}
	return ids
}

// HandleRestart HTTP处理函数用于重启指定ID的goroutine
func HandleRestart(gm *GoroutineManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取URL参数中的goroutine ID
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "缺少goroutine ID", http.StatusBadRequest)
			return
		}

		gm.Restart(id, gm.params[id])
		fmt.Fprintf(w, "goroutine %s 已重启", id)
	}
}

// HandleList creates an HTTP handler function that all Goroutine IDs managed the Goroutine.
func HandleList(gm *GoroutineManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ids := gm.List()
		for _, id := range ids {
			fmt.Fprintf(w, "Goroutine ID: %s\n", id)
		}
	}
}

// HandleWorkList 创建一个HTTP处理程序来处理特定的工作列表
func HandleWorkList(gm *GoroutineManager, workList []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, id := range workList {
			fmt.Fprintf(w, "BcsClusterID: %s\n", id)
		}
	}
}
