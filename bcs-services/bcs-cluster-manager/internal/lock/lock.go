/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package lock

import (
	"crypto/tls"
	"time"
)

// Options options for lock
type Options struct {
	// Endpoints lock registry
	Endpoints []string
	// TLSConfig config for tls
	TLSConfig *tls.Config
	// Prefix lock path
	Prefix string
}

// LockOptions options for lock
type LockOptions struct {
	TTL  time.Duration
}

// LockOption option function for lock
type LockOption func(o *LockOptions)

// LockTTL sets the lock ttl
func LockTTL(t time.Duration) LockOption {
	return func(o *LockOptions) {
		o.TTL = t
	}
}

// Option function for set option
type Option func(o *Options)

// Endpoints set endpoints to use
func Endpoints(endpoints ...string) Option {
	return func(o *Options) {
		o.Endpoints = endpoints
	}
}

// TLS set tls config to use
func TLS(tlsConfig *tls.Config) Option {
	return func(o *Options) {
		o.TLSConfig = tlsConfig
	}
}

// Prefix set election path for prefix
func Prefix(pre string) Option {
	return func(o *Options) {
		o.Prefix = pre
	}
}

// DistributedLock lock interface
type DistributedLock interface {
	// Lock acquires a lock
	Lock(id string, opts ...LockOption) error
	// Unlock releases a lock
	Unlock(id string) error
}
