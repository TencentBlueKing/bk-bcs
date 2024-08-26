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

// Package alert xxx
package alert

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	paasclient "github.com/Tencent/bk-bcs/bcs-common/pkg/esb/client"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/throttle"
)

// Config alert saas configuration
type Config struct {
	// Hosts AuthCenter hosts, without http/https, default is http
	Hosts []string
	// reserved config for https, not used yet
	TLSConfig *tls.Config
	// AppCode comes from bk saas
	AppCode string
	// AppSecret comes from bk saas
	AppSecret string
	// LocalIP for alert source
	LocalIP string
	// default ClusterID, default BCS-K8S-00000
	ClusterID string
}

// Client bk-saas alert client definition, alert message sending details:
//
//	{
//		"startsAt": "2020-04-14T12:31:00.124Z",(required)
//		"endsAt": "2020-04-14T12:31:00.124Z",(required)
//		"annotations": {
//			"uuid": "cee84faf-7ee3-11ea-xxx",
//			"message": "this is alert"  (required)
//		},
//		"labels": {
//			"alert_type": "Error", (required)
//			"cluster_id": "BCS-K8S-00000", (required)
//			"namespace": "myns",
//			"ip": "127.0.0.11",
//			"module_name": "scheduler"
//		}
//	}
type Client interface {
	// SendServiceAlert xx
	// for bcs-servie modules
	SendServiceAlert(module string, message string) error
	// SendClusterAlert xx
	// for cluster bcs modules
	SendClusterAlert(cluster string, module string, message string) error
	SendCustomAlert(annotation, label map[string]string) error
}

// NewAlertClient create client instance
// according to app information
func NewAlertClient(config *Config) (Client, error) {
	if len(config.Hosts) == 0 {
		return nil, fmt.Errorf("Config lost Hosts item")
	}
	if len(config.AppCode) == 0 || len(config.AppSecret) == 0 {
		return nil, fmt.Errorf("Config lost BK App Info")
	}
	if len(config.LocalIP) == 0 {
		return nil, fmt.Errorf("Config lost alert source IP")
	}
	if len(config.ClusterID) == 0 {
		config.ClusterID = "BCS-K8S-00000"
	}
	aclient := &alertClient{
		config: config,
	}
	if config.TLSConfig != nil {
		aclient.client = paasclient.NewRESTClientWithTLS(config.TLSConfig).
			WithRateLimiter(throttle.NewTokenBucket(1000, 1000))
	} else {
		aclient.client = paasclient.NewRESTClient().
			WithRateLimiter(throttle.NewTokenBucket(1000, 1000))
	}
	return aclient, nil
}

// alertClient client implementation
type alertClient struct {
	config *Config
	client *paasclient.RESTClient
}

// SendServiceAlert implementation
func (cli *alertClient) SendServiceAlert(module string, message string) error {

	payload := newServiceAlert(module, message, cli.config.LocalIP)
	auth := map[string]string{
		"app_code":   cli.config.AppCode,
		"app_secret": cli.config.AppSecret,
	}
	authBytes, _ := json.Marshal(auth)
	authHeader := http.Header{}
	authHeader.Add("X-Bkapi-Authorization", string(authBytes))

	result := cli.client.Post().
		WithEndpoints(cli.config.Hosts).
		WithBasePath("/").
		WithHeaders(authHeader).
		SubPathf("/prod/api/v1/bcs/alerts").
		Body(payload).
		WithTimeout(time.Second * 3).
		Do()
	if result.StatusCode != 200 {
		return fmt.Errorf("Send Service Alert failed: %s", result.Status)
	}
	return nil
}

// SendClusterAlert implementation
func (cli *alertClient) SendClusterAlert(cluster, module string, message string) error {
	payload := newClusterAlert(cluster, module, message, cli.config.LocalIP)
	auth := map[string]string{
		"app_code":   cli.config.AppCode,
		"app_secret": cli.config.AppSecret,
	}
	authBytes, _ := json.Marshal(auth)
	authHeader := http.Header{}
	authHeader.Add("X-Bkapi-Authorization", string(authBytes))

	result := cli.client.Post().
		WithEndpoints(cli.config.Hosts).
		WithBasePath("/").
		WithHeaders(authHeader).
		SubPathf("/prod/api/v1/bcs/alerts").
		Body(payload).
		WithTimeout(time.Second * 3).
		Do()
	if result.StatusCode != 200 {
		return fmt.Errorf("Send Service Alert failed: %s", result.Status)
	}
	return nil
}

// SendCustomAlert implementation
func (cli *alertClient) SendCustomAlert(annotation, label map[string]string) error {
	if len(annotation) == 0 || len(label) == 0 {
		return fmt.Errorf("lost specified alert info")
	}
	payload := map[string]interface{}{
		startKey:      time.Now(),
		endKey:        time.Now(),
		annotationKey: annotation,
		labelKey:      label,
	}
	auth := map[string]string{
		"app_code":   cli.config.AppCode,
		"app_secret": cli.config.AppSecret,
	}
	authBytes, _ := json.Marshal(auth)
	authHeader := http.Header{}
	authHeader.Add("X-Bkapi-Authorization", string(authBytes))
	result := cli.client.Post().
		WithEndpoints(cli.config.Hosts).
		WithBasePath("/").
		WithHeaders(authHeader).
		SubPathf("/prod/api/v1/bcs/alerts").
		Body(payload).
		WithTimeout(time.Second * 3).
		Do()
	if result.StatusCode != 200 {
		return fmt.Errorf("Send Service Alert failed: %s", result.Status)
	}
	return nil
}
