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

// Package bkdata xxx
package bkdata

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/esb/client"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/throttle"
)

const (
	// NewAccessDeployPlanAPI bkdata new deploy plan api
	NewAccessDeployPlanAPI = "v3/access/deploy_plan/"
	// NewCleanConfigAPI bkdata new clean strategy api
	NewCleanConfigAPI = "v3/databus/cleans/"
)

// ClientInterface specified bkdata api method
type ClientInterface interface {
	ObtainDataID(CustomAccessDeployPlanConfig) (int64, error)
	SetCleanStrategy(strategy DataCleanStrategy) error
}

// ClientCreatorInterface specified bkdata api client creator method
type ClientCreatorInterface interface {
	NewClientFromConfig(BKDataClientConfig) ClientInterface
}

// Client is implementation of ClientInterface
type Client struct {
	client *client.RESTClient
	config BKDataClientConfig
}

// ClientCreator is implementation of ClientCreatorInterface
type ClientCreator struct {
}

// NewClientCreator create bkdata client creator
func NewClientCreator() ClientCreatorInterface {
	return &ClientCreator{}
}

// ObtainDataID obtain a new dataid from bk-data
func (c *Client) ObtainDataID(conf CustomAccessDeployPlanConfig) (int64, error) {
	conf.BkAppCode = c.config.BkAppCode
	conf.BkAppSecret = c.config.BkAppSecret
	conf.BkUsername = c.config.BkUsername
	conf.BkdataAuthenticationMethod = c.config.BkdataAuthenticationMethod
	jsonstr, err := json.Marshal(conf)
	if err != nil {
		return -1, err
	}
	blog.Infof("requerst info: %s", string(jsonstr))
	var payload map[string]interface{}
	err = json.Unmarshal(jsonstr, &payload)
	if err != nil {
		return -1, err
	}
	// request bkdata api
	result := c.client.Post().
		WithEndpoints([]string{c.config.Host}).
		WithBasePath("/").
		SubPathf(NewAccessDeployPlanAPI).
		WithHeaders(http.Header{
			"Content-Type": []string{"application/json"},
		}).
		Body(payload).
		WithTimeout(time.Second * 10).
		Do()
	if result.StatusCode != 200 {
		return -1, fmt.Errorf("Obtain dataid failed: %s", result.Status)
	}
	var res map[string]interface{}
	err = json.Unmarshal(result.Body, &res)
	if err != nil {
		return -1, err
	}
	if succ := res["result"].(bool); !succ {
		return -1, fmt.Errorf("Obtain dataid failed: %+v", res["message"])
	}
	var dataid int64
	res = res["data"].(map[string]interface{})
	tmp, ok := res["raw_data_id"].(float64)
	if !ok {
		return -1, fmt.Errorf("convert return value [raw_data_id] from %+v to float64 failed: type assert failed", res)
	}
	dataid = int64(tmp)
	return dataid, nil
}

// SetCleanStrategy create clean strategy in bkdata
func (c *Client) SetCleanStrategy(strategy DataCleanStrategy) error {
	strategy.BkAppCode = c.config.BkAppCode
	strategy.BkAppSecret = c.config.BkAppSecret
	strategy.BkUsername = c.config.BkUsername
	strategy.BkdataAuthenticationMethod = c.config.BkdataAuthenticationMethod
	payload := map[string]interface{}{}
	jsonstr, err := json.Marshal(strategy)
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonstr, &payload)
	if err != nil {
		return err
	}
	// request bkdata api
	result := c.client.Post().
		WithEndpoints([]string{c.config.Host}).
		WithBasePath("/").
		SubPathf(NewCleanConfigAPI).
		WithHeaders(http.Header{
			"Content-Type": []string{"application/json"},
		}).
		Body(payload).
		WithTimeout(time.Second * 10).
		Do()
	if result.StatusCode != 200 {
		return fmt.Errorf("Set clean strategy failed: %s", result.Status)
	}
	var res map[string]interface{}
	err = json.Unmarshal(result.Body, &res)
	if err != nil {
		return err
	}
	if succ := res["result"].(bool); !succ {
		return fmt.Errorf("Set clean strategy failed: %+v", res["message"])
	}
	return nil
}

// NewClientFromConfig set config of BKDataApiClient
func (c *ClientCreator) NewClientFromConfig(conf BKDataClientConfig) ClientInterface {
	// bkgateway auth required
	conf.BkdataAuthenticationMethod = "user"

	var cli *client.RESTClient
	if conf.TLSConf != nil {
		cli = client.NewRESTClientWithTLS(conf.TLSConf).
			WithRateLimiter(throttle.NewTokenBucket(50, 50)).
			WithCredential(conf.BkAppCode, conf.BkAppSecret)
	} else {
		cli = client.NewRESTClient().
			WithRateLimiter(throttle.NewTokenBucket(50, 50)).
			WithCredential(conf.BkAppCode, conf.BkAppSecret)
	}
	return &Client{
		client: cli,
		config: conf,
	}
}
