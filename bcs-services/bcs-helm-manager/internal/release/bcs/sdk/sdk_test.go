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

package sdk

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"helm.sh/helm/v3/pkg/action"
	rspb "helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
)

func BenchmarkList(b *testing.B) {
	namespace := "bcs-system"
	clusterId := "BCS-K8S-00000"
	for i := 0; i < b.N; i++ {
		_, err := List(namespace, clusterId)
		if err != nil {
			b.Errorf("List error: %s", err)
			return
		}
	}
}

// List helm release List的模拟函数
func List(namespace, clusterId string) ([]*rspb.Release, error) {
	conf := new(action.Configuration)
	configFlag, err := getConfigFlag(namespace, clusterId)
	if err != nil {
		return nil, err
	}
	if err = conf.Init(configFlag, namespace, "", blog.Infof); err != nil {
		return nil, err
	}

	lister := action.NewList(conf)
	lister.All = true
	if len(namespace) == 0 {
		lister.AllNamespaces = true
	}

	releases, err := lister.Run()
	if err != nil {
		return nil, err
	}
	for i := range releases {
		releases[i].Config = removeValuesTemplate(releases[i].Config)
	}
	return releases, nil
}

// getConfigFlag 获取helm-client配置
func getConfigFlag(namespace, clusterID string) (*genericclioptions.ConfigFlags, error) {
	flags := genericclioptions.NewConfigFlags(false)

	apiserver := os.Getenv("APIServer")
	if apiserver == "" {
		return nil, errors.New("APIServer is null")
	}
	flagsAPIServer := fmt.Sprintf(bcsAPIGWK8SBaseURI, apiserver, clusterID)
	flagsBearerToken := os.Getenv("BearerToken")
	if flagsBearerToken == "" {
		return nil, errors.New("BearerToken is null")
	}
	flags.WrapConfigFn = func(config *rest.Config) *rest.Config {
		config.TLSClientConfig = rest.TLSClientConfig{
			Insecure: true,
		}
		config.Host = flagsAPIServer
		config.BearerToken = flagsBearerToken
		return config
	}

	flags.Namespace = common.GetStringP(namespace)
	flags.Timeout = common.GetStringP(defaultTimeout)
	return flags, nil
}
