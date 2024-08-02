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

// Package types xxx
package types

import (
	"crypto/tls"
	"errors"
	"time"

	"go-micro.dev/v4/registry"
)

var (
	// ErrNotInited err server not init
	ErrNotInited = errors.New("server not init")
)

const (
	// DefaultTimeOut default timeout
	DefaultTimeOut = time.Second * 10
	// RetryCount default retry count
	RetryCount = 10
)

const (
	// ModuleHelmManager default discovery helmmanager module
	ModuleHelmManager = "helmmanager.bkbcs.tencent.com"
	// PubicRepo public repo
	PubicRepo = "public-repo"
)

// Options for init clusterManager
type Options struct {
	Enable bool
	// GateWay address
	GateWay         string
	Token           string
	Module          string
	EtcdRegistry    registry.Registry
	ClientTLSConfig *tls.Config
}

// Validate validate options
func (o *Options) Validate() bool {
	if o == nil {
		return false
	}
	if !o.Enable {
		return false
	}

	if o.Module == "" {
		o.Module = ModuleHelmManager
	}

	return true
}
