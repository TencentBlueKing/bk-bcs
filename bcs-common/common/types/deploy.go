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

package types

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type ConfigSet struct {
	Common map[string]interface{}            `json:"common"`
	Conf   map[string]map[string]interface{} `json:"conf"`
}

func ParseConfigSet(data interface{}) (c *ConfigSet, err error) {
	c = new(ConfigSet)
	var tmp []byte
	if tmp, err = json.Marshal(data); err != nil {
		return c, err
	}
	err = json.Unmarshal(tmp, c)
	return c, err
}

type ClusterSet struct {
	ClusterId     string    `json:"clusterId"`
	ClusterConfig ConfigSet `json:"clusterConfig"`
}

type DeployConfig struct {
	Service       string       `json:"service"`
	ServiceConfig ConfigSet    `json:"serviceConfig"`
	Clusters      []ClusterSet `json:"clusters"`
	StableVersion string       `json:"stableVersion"`
}

type RenderConfig struct {
	MesosZk          string `render:"mesosZkHost"`
	MesosZkSpace     string `render:"mesosZkHostSpace"`
	MesosZkSemicolon string `render:"mesosZkHostSemicolon"`
	MesosZkRaw       string `render:"mesosZkRaw"`
	MesosMaster      string `render:"mesosMaster"`
	MesosQuorum      string `render:"mesosQuorum"`
	Dns              string `render:"dnsUpStream"`
	ClusterId        string `render:"clusterId"`
	ClusterIdNum     string `render:"clusterIdNumber"`
	City             string `render:"city"`
	JfrogUrl         string `render:"jfrogUrl"`
	NeedNat          string `render:"needNat"`
}

var tagFormat = "${%s}"

func (rc RenderConfig) Render(s string) (r string) {
	r = s

	typeOf := reflect.TypeOf(rc)
	n := typeOf.NumField()
	i := 0
	for i < n {
		field := typeOf.Field(i)
		tag := field.Tag.Get("render")
		value := reflect.ValueOf(rc).FieldByName(field.Name).String()
		r = strings.Replace(r, fmt.Sprintf(tagFormat, tag), value, -1)
		i++
	}
	return r
}
