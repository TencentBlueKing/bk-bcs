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

// Package cc NOTES
package cc

import (
	"sync"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
)

var runtimeOnce sync.Once

// rt is the runtime Setting which is loaded from
// config file.
// It can be called only after LoadSettings is executed successfully.
var rt *runtime

func initRuntime(s Setting) {
	runtimeOnce.Do(func() {
		rt = &runtime{
			settings: s,
		}
	})
}

type runtime struct {
	lock     sync.Mutex
	settings Setting
}

// Ready is used to test if the runtime configuration is
// initialized with load from file success and already
// ready to use.
func (r *runtime) Ready() bool {
	if r == nil {
		return false
	}

	if r.settings == nil {
		return false
	}

	return true
}

// ApiServer return api server Setting.
func ApiServer() ApiServerSetting {
	rt.lock.Lock()
	defer rt.lock.Unlock()

	if !rt.Ready() {
		logs.ErrorDepthf(1, "runtime not ready, return empty api server setting")
		return ApiServerSetting{}
	}

	s, ok := rt.settings.(*ApiServerSetting)
	if !ok {
		logs.ErrorDepthf(1, "current %s service can not get api server setting", ServiceName())
		return ApiServerSetting{}
	}

	return *s
}

// AuthServer return auth server Setting.
func AuthServer() AuthServerSetting {
	rt.lock.Lock()
	defer rt.lock.Unlock()

	if !rt.Ready() {
		logs.ErrorDepthf(1, "runtime not ready, return empty auth server setting")
		return AuthServerSetting{}
	}

	s, ok := rt.settings.(*AuthServerSetting)
	if !ok {
		logs.ErrorDepthf(1, "current %s service can not get auth server setting", ServiceName())
		return AuthServerSetting{}
	}

	return *s
}

// ConfigServer return config server Setting.
func ConfigServer() ConfigServerSetting {
	rt.lock.Lock()
	defer rt.lock.Unlock()

	if !rt.Ready() {
		logs.ErrorDepthf(1, "runtime not ready, return empty config server setting")
		return ConfigServerSetting{}
	}

	s, ok := rt.settings.(*ConfigServerSetting)
	if !ok {
		logs.ErrorDepthf(1, "current %s service can not get config server setting", ServiceName())
		return ConfigServerSetting{}
	}

	return *s
}

// FeedServer return feed server Setting.
func FeedServer() FeedServerSetting {
	rt.lock.Lock()
	defer rt.lock.Unlock()

	if !rt.Ready() {
		logs.ErrorDepthf(1, "runtime not ready, return empty feed server setting")
		return FeedServerSetting{}
	}

	s, ok := rt.settings.(*FeedServerSetting)
	if !ok {
		logs.ErrorDepthf(1, "current %s service can not get feed server setting", ServiceName())
		return FeedServerSetting{}
	}

	return *s
}

// CacheService return cache service Setting.
func CacheService() CacheServiceSetting {
	rt.lock.Lock()
	defer rt.lock.Unlock()

	if !rt.Ready() {
		logs.ErrorDepthf(1, "runtime not ready, return empty cache service setting")
		return CacheServiceSetting{}
	}

	s, ok := rt.settings.(*CacheServiceSetting)
	if !ok {
		logs.ErrorDepthf(1, "current %s service can not get cache service setting", ServiceName())
		return CacheServiceSetting{}
	}

	return *s
}

// DataService return data service Setting.
func DataService() DataServiceSetting {
	rt.lock.Lock()
	defer rt.lock.Unlock()

	if !rt.Ready() {
		logs.ErrorDepthf(1, "runtime not ready, return empty data service setting")
		return DataServiceSetting{}
	}

	s, ok := rt.settings.(*DataServiceSetting)
	if !ok {
		logs.ErrorDepthf(1, "current %s service can not get data service setting", ServiceName())
		return DataServiceSetting{}
	}

	return *s
}

// VaultServer return vault service Setting.
func VaultServer() VaultServerSetting {
	rt.lock.Lock()
	defer rt.lock.Unlock()

	if !rt.Ready() {
		logs.ErrorDepthf(1, "runtime not ready, return empty data service setting")
		return VaultServerSetting{}
	}

	s, ok := rt.settings.(*VaultServerSetting)
	if !ok {
		logs.ErrorDepthf(1, "current %s service can not get vault server setting", ServiceName())
		return VaultServerSetting{}
	}

	return *s
}

// VaultSidecar return vault sidecar service Setting.
func VaultSidecar() VaultSidecarSetting {
	rt.lock.Lock()
	defer rt.lock.Unlock()

	if !rt.Ready() {
		logs.ErrorDepthf(1, "runtime not ready, return empty data service setting")
		return VaultSidecarSetting{}
	}

	s, ok := rt.settings.(*VaultSidecarSetting)
	if !ok {
		logs.ErrorDepthf(1, "current %s service can not get vault sidecar setting", ServiceName())
		return VaultSidecarSetting{}
	}

	return *s
}
