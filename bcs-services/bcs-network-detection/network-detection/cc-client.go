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

package network_detection

import (
	"encoding/json"
	"fmt"
	"net/http"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/pkg/esb"
	"bk-bcs/bcs-services/bcs-network-detection/config"
	"bk-bcs/bcs-services/bcs-network-detection/types"
)

type CmdbClient struct {
	conf *config.Config
	//esb client
	esb *esb.EsbClient
}

func NewCmdbClient(conf *config.Config) (*CmdbClient, error) {
	cmdb := &CmdbClient{
		conf: conf,
	}

	//new esb client
	var err error
	cmdb.esb, err = esb.NewEsbClient(conf.AppCode, conf.AppSecret, conf.Operator, conf.EsbUrl)
	if err != nil {
		return nil, err
	}
	blog.Infof("NewEsbClient done")
	return cmdb, nil
}

func (c *CmdbClient) updateNodeInfo(node *types.NodeInfo) error {
	payload := make(map[string]interface{})
	//init request cmdb payload info
	payload["header_on"] = 0
	payload["output_type"] = "json"
	payload["host_std_key_values"] = map[string]string{
		"InnerIP": node.Ip,
	}
	payload["exact_search"] = 1
	payload["app_id"] = c.conf.AppId
	payload["method"] = "getTopoModuleHostList"
	payload["host_std_req_column"] = []string{"IDC", "serverRack", "ModuleName"}

	//request cmdb through esb
	by, err := c.esb.RequestEsb(http.MethodPost, "/component/compapi/cc/get_query_info", payload)
	if err != nil {
		return err
	}

	//Unmarshal CmdbHostInfo
	var hosts []*types.CmdbHostInfo
	err = json.Unmarshal(by, &hosts)
	if err != nil {
		blog.Errorf("Unmarshal data(%s) to types.CmdbHostInfo failed: %s", string(by), err.Error())
		return err
	}
	if len(hosts) == 0 {
		return fmt.Errorf("node %s not found, response(%s)", node.Ip, string(by))
	}

	//update node info
	node.Idc = hosts[0].IDC
	node.Module = hosts[0].ModuleName
	return nil
}
