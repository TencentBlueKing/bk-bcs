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

// Package exporter xxx
package exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/google/uuid"
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-node-external-worker/httpsvr"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-node-external-worker/options"
)

// NodeExporter collect node external ip and exporter on configmap
type NodeExporter struct {
	Ctx        context.Context
	K8sClient  client.Client
	Opts       options.Options
	HttpClient http.Client
	HttpSvr    *httpsvr.HttpServerClient
	externalIP string
}

// Watch watch external ip
func (n *NodeExporter) Watch() {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	// nolint
	for {
		select {
		case <-ticker.C:
			externalIP, err := n.retrieveExternalIP()
			if err != nil {
				blog.Errorf("retrieveExternalIP failed, err: %v", err)
				continue
			}
			if n.externalIP == externalIP {
				blog.V(4).Infof("retrieve same externalIP'%s', continue...", externalIP)
				continue
			}

			UUID := uuid.New().String()
			n.HttpSvr.SetUUID(UUID)
			if n.isRealExternalIP(externalIP, UUID) {
				if err = n.updateExternalIPInConfigMap(externalIP); err != nil {
					blog.Errorf("updateExternalIPInConfigMap failed, err: %v", err)
					continue
				}
				n.externalIP = externalIP
			}
		}
	}
}

func (n *NodeExporter) retrieveExternalIP() (string, error) {
	resp, err := n.HttpClient.Get(n.Opts.ExternalIPWebURL)
	if err != nil {
		return "", fmt.Errorf("http get failed, err: %s", err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http status code '%d' invalid", resp.StatusCode)
	}

	// 读取响应内容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response body failed, err: %s", err.Error())
	}

	externalIP := string(body)
	blog.Infof("get external ip: %s", externalIP)
	return externalIP, nil
}

func (n *NodeExporter) isRealExternalIP(externalIP, uuid string) bool {
	// do connect check
	addr := net.JoinHostPort(externalIP, strconv.Itoa(int(n.Opts.ListenPort)))
	// nolint
	// todo use bk internal service
	resp, err := n.HttpClient.Get("http://" + addr + "/node-external-worker/api/v1/health_check")
	if err != nil {
		blog.Errorf("http get addr '%s' failed, err: %s", addr, err.Error())
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		blog.Errorf("http status code '%d' invalid", resp.StatusCode)
		return false
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		blog.Errorf("read body failed, err: %s", err)
		return false
	}
	blog.Infof("resp: %s, respCode: %d", string(respBody), resp.StatusCode)

	var apiResp httpsvr.APIRespone
	if err = json.Unmarshal(respBody, &apiResp); err != nil {
		blog.Errorf("unmarshal api response failed, err: %s", err.Error())
		return false
	}

	if apiResp.Data != uuid {
		blog.Warnf("receive invalid UUID '%s', want '%s'", apiResp.Data, uuid)
		return false
	}
	blog.Info("http check success, node external ip: %s", externalIP)
	return true
}

func (n *NodeExporter) updateExternalIPInConfigMap(externalIP string) error {
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// get configmap and modify
		configMap := &k8scorev1.ConfigMap{}
		if err := n.K8sClient.Get(context.Background(), k8stypes.NamespacedName{Namespace: n.Opts.Namespace,
			Name: n.Opts.ExternalIPConfigMapName}, configMap); err != nil {
			if !k8serrors.IsNotFound(err) {
				return fmt.Errorf("get configmap failed, err: %s", err.Error())
			}
			blog.Infof("not found configMap '%s/%s'", n.Opts.Namespace, n.Opts.ExternalIPConfigMapName)
			if err = n.createExternalIPConfigmap(); err != nil {
				return err
			}
			return nil
		}

		configMap.Data[n.Opts.NodeName] = externalIP
		return n.K8sClient.Update(n.Ctx, configMap)
	}); err != nil {
		return fmt.Errorf("get and update configmap'%s/%s' failed, %s ", n.Opts.Namespace,
			n.Opts.ExternalIPConfigMapName, err.Error())
	}
	return nil
}

// create configmap if not exist
func (n *NodeExporter) createExternalIPConfigmap() error {
	configMap := &k8scorev1.ConfigMap{}
	configMap.SetNamespace(n.Opts.Namespace)
	configMap.SetName(n.Opts.ExternalIPConfigMapName)

	configMap.Data = map[string]string{
		n.Opts.NodeName: n.externalIP,
	}

	if err := n.K8sClient.Create(n.Ctx, configMap); err != nil {
		return fmt.Errorf("create config map in namespace[%s] failed, err %s", n.Opts.Namespace, err.Error())
	}

	return nil
}
