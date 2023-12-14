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

// Package shutdown NOTES
package shutdown

import (
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
)

// shutdownNotifier is a notifier which is used to notify all the job/tasks
// doing shutdown operation immediately. so that jobs/tasks can exit gracefully.
// Note: do not send message to this notifier, it can only be closed to
// broadcast the shutdown message.
var shutdownNotifier chan struct{}
var shutdownFlag atomic.Value
var waitOnceFlag atomic.Value
var globalWaitGroup sync.WaitGroup
var shutdownOnce sync.Once
var shutdownSignal chan struct{}
var firstOnceLock sync.Mutex
var firstShutdown func()

func init() {
	shutdownNotifier = make(chan struct{})
	shutdownFlag.Store(false)
	waitOnceFlag.Store(false)
	globalWaitGroup = sync.WaitGroup{}
	shutdownOnce = sync.Once{}
	shutdownSignal = make(chan struct{}, 10)
	firstOnceLock = sync.Mutex{}

	go waitForShutdown()
}

func waitForShutdown() {
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGTERM)

	select {
	case sig := <-exit:
		logs.Infof("received shutdown system signal '%v', start to do shutdown process...", sig)

	case <-shutdownSignal:
	}

	shutdownOnce.Do(func() {
		shutdownFlag.Store(true)

		if firstShutdown != nil {
			firstShutdown()
		}

		// then notify other register to do shut down operation.
		close(shutdownNotifier)
	})
}

// IsShuttingDown return shutdownFlag, which describe if the service is in shutting down process.
func IsShuttingDown() bool {
	return shutdownFlag.Load().(bool)
}

// RegisterFirstShutdown is used to register the first called function when doing the shutdown
// process. Only one function can be registered, otherwise it will panic.
// Note: the registered callback function can not hang and should be return as soon as possible,
// otherwise, the shutdown process will be hanged.
func RegisterFirstShutdown(first func()) {
	firstOnceLock.Lock()
	defer firstOnceLock.Unlock()
	if firstShutdown != nil {
		panic("first shutdown function has already been registered, only one can be registered.")
	}

	firstShutdown = first
}

// Notifier is used to help the tasks/jobs to receive the shutdown signal and notify the waiter
// that it has already finished the shutdown tasks/jobs.
type Notifier struct {
	Signal <-chan struct{}
}

// Done means a shutdown task/job has already finished.
func (n *Notifier) Done() {
	if !shutdownFlag.Load().(bool) {
		logs.Errorf("not in shutdown process for now, do not need to call notifier Done method.")
		return
	}
	globalWaitGroup.Done()
}

// SignalShutdownGracefully send the signal to shut down the process.
func SignalShutdownGracefully() {

	if waitOnceFlag.Load().(bool) {
		logs.Infof("already in shutdown process for now, do not need to call SignalShutdownGracefully again.")
		return
	}

	logs.InfoDepthf(1, "received auto shutdown notify, start to do shutdown process...")

	select {
	case shutdownSignal <- struct{}{}:
	default:
	}
}

// AddNotifier return the shutdown notifier.
// When caller received message from this notifier, it should execute the shutdown
// process immediately. This process will be forced to exit after a timeout time.
func AddNotifier() *Notifier {
	globalWaitGroup.Add(1)
	return &Notifier{Signal: shutdownNotifier}
}

// WaitShutdown blocks to wait for all the jobs/tasks is shutdown.
// timeoutSeconds' min value is 5 seconds.
// 1. timeoutSeconds is <0 means no timeout limit.
// 2. timeoutSeconds is >0 means the process will be exited anyway after timeout.
// finalizer is other jobs/tasks shutdown dependent function, only can final exec.
func WaitShutdown(timeoutSeconds int, finalizer ...func()) {

	// wait for shutdown the process, then handle the finalizer process.
	<-shutdownNotifier

	if !shutdownFlag.Load().(bool) {
		logs.Errorf("not in shutdown process for now, do not need to call WaitShutdown method.")
		return
	}

	if waitOnceFlag.Load().(bool) {
		logs.Errorf("already in shutdown process for now, do not need to call WaitShutdown method again.")
		return
	}

	waitOnceFlag.Store(true)

	start := time.Now()

	// wait with no timeout limit.
	if timeoutSeconds <= 0 {
		logs.Infof("wait to shutdown all the jobs/tasks with no timout limit...")
		globalWaitGroup.Wait()
		logs.Infof("shutdown all the jobs/tasks success, cost: %s", time.Since(start).String())

		return
	}

	if timeoutSeconds < 5 {
		timeoutSeconds = 5
	}

	logs.Infof("wait to shutdown all the jobs/tasks with timout: %d seconds limit.", timeoutSeconds)

	// wait with timeout limit.
	timeoutNotifier := time.After(time.Duration(timeoutSeconds) * time.Second)
	finishedNotifier := make(chan struct{})
	go func() {
		globalWaitGroup.Wait()
		for _, f := range finalizer {
			f()
		}
		finishedNotifier <- struct{}{}
	}()

	select {
	case <-finishedNotifier:
		logs.Infof("shutdown all the jobs/tasks success. cost: %s", time.Since(start).String())

	case <-timeoutNotifier:
		logs.Infof("wait for shutdown timeout after %d seconds, force shutdown now...", timeoutSeconds)
	}

	logs.CloseLogs()

	// force exit now.
	os.Exit(1)
}
