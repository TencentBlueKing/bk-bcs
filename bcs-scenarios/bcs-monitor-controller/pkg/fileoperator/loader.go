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

package fileoperator

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	v1 "k8s.io/api/core/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

// Loader load file info
type Loader struct {
	client client.Client
}

// LoadFileFromUrl load file from url
func (l *Loader) LoadFileFromUrl(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed, status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// LoadFileFromConfigMap load file from configmap
func (l *Loader) LoadFileFromConfigMap(ns, configmapName string) ([]byte, error) {
	cm := &v1.ConfigMap{}
	if err := l.client.Get(context.TODO(), k8stypes.NamespacedName{
		Namespace: ns,
		Name:      configmapName,
	}, cm); err != nil {
		blog.Errorf("get configmap %s/%s from k8s failed, err: %s", ns, configmapName, err.Error())
		return nil, err
	}

	return []byte(cm.Data["panel"]), nil
}
