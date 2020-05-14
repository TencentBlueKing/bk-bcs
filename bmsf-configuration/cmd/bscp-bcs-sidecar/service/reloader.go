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

package service

import (
	"time"

	"github.com/spf13/viper"

	"bk-bscp/pkg/logger"
)

// ConfigSpec is config spec.
type ConfigSpec struct {
	// config set name
	Name string

	// config set fpath.
	Fpath string
}

// ReloadSpec specs how to reload.
type ReloadSpec struct {
	// business name.
	BusinessName string

	// app name.
	AppName string

	// release id.
	Releaseid string

	// multi release id .
	MultiReleaseid string

	// release name.
	ReleaseName string

	// reload type.
	ReloadType int32

	// config specs.
	Configs []ConfigSpec
}

// Reloader is configs reloader.
type Reloader struct {
	viper  *viper.Viper
	events chan *ReloadSpec
}

// NewReloader creates a new Reloader.
func NewReloader(viper *viper.Viper) *Reloader {
	return &Reloader{viper: viper, events: make(chan *ReloadSpec, viper.GetInt("instance.reloadChanSize"))}
}

// Init inits new Reloader.
func (r *Reloader) Init() {
}

// Reload handle configs reload.
func (r *Reloader) Reload(spec *ReloadSpec) {
	if spec != nil {
		go r.reload(spec)
	}
}

func (r *Reloader) reload(spec *ReloadSpec) {
	select {
	case r.events <- spec:
	case <-time.After(r.viper.GetDuration("instance.reloadChanTimeout")):
		logger.Warn("send reload spec to reload events channel timeout, spec[%+v]", spec)
	}
}

// EventChan is reload events channel.
func (r *Reloader) EventChan() chan *ReloadSpec {
	return r.events
}
