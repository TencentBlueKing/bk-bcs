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
	"time"

	"gopkg.in/redis.v5"
)

// Options is redis balancer option struct.
type Options struct {
	// including the redis.Options.
	redis.Options

	// interval for redis PING check.
	interval time.Duration

	// mark the instance as up.
	rise int

	// mark the instance as down.
	fall int
}

// getInterval returns the interval of redis PING checking.
func (o *Options) getInterval() time.Duration {
	if o.interval < defaultPingInterval {
		return defaultPingInterval
	}
	return o.interval
}

// getRise returns the rise count.
func (o *Options) getRise() int {
	if o.rise < 1 {
		return 1
	}
	return o.rise
}

// getFall returns the fall count.
func (o *Options) getFall() int {
	if o.fall < 1 {
		return 1
	}
	return o.fall
}
