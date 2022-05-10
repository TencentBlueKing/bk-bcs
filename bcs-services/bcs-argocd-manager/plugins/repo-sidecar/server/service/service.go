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

package service

import (
	"context"
	"io/ioutil"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/plugins/repo-sidecar/server/plugin"
)

const (
	SvcPort = "8080"

	PluginNameKey = "BCS_PLUGIN_NAME"
)

// NewService return a new Service instance
func NewService(opt *Options) *Service {
	return &Service{
		opt: opt,
	}
}

// Service describe the repo-sidecar server instance
type Service struct {
	opt *Options
	p   *plugin.Plugin
}

// Start the server
func (s *Service) Start() error {
	http.HandleFunc("/", s.handler)

	// init the plugin handler
	p, err := plugin.New(s.opt.ServerAddress, s.opt.Instance)
	if err != nil {
		return err
	}
	s.p = p

	return http.ListenAndServe(":"+SvcPort, nil)
}

func (s *Service) handler(w http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		blog.Errorf("read from body failed, %v", err)
		s.writeResp(w, &Result{Code: ResultErrorCodeFailure, Message: err.Error()})
		return
	}

	var message Message
	if err = codec.DecJson(data, &message); err != nil {
		blog.Errorf("decode from body failed, %v, body(%s)", err, string(data))
		s.writeResp(w, &Result{Code: ResultErrorCodeFailure, Message: err.Error()})
		return
	}

	pluginName := message.GetEnv(PluginNameKey)
	svc, err := s.p.FetchService(context.TODO(), pluginName)
	if err != nil {
		blog.Errorf("fetch service for plugin %s failed, %v", pluginName, err)
		s.writeResp(w, &Result{Code: ResultErrorCodeFailure, Message: err.Error()})
		return
	}

	result, err := svc.DoRender(context.TODO(), message.Env, message.Content)
	if err != nil {
		blog.Errorf("do render for plugin %s according %s(%s) failed, %v",
			pluginName, svc.Protocol, svc.Address, err)
		s.writeResp(w, &Result{Code: ResultErrorCodeFailure, Message: err.Error()})
		return
	}

	s.writeResp(w, &Result{Code: ResultErrorCodeSuccess, Data: result})
}

func (s *Service) writeResp(w http.ResponseWriter, r *Result) {
	var result []byte
	_ = codec.EncJson(r, &result)
	_, _ = w.Write(result)
}

func (s *Service) getPlugin() {

}
