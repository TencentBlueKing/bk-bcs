/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package renderengine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	pbcommon "bk-bscp/internal/protocol/common"
)

const (
	// EngineGoTmplate go template engine
	EngineGoTmplate = "gotemplate"
	// EngineMako mako template engine
	EngineMako = "makotemplate.py"
)

// EngineTypeMap map for engine type translation
var EngineTypeMap = map[int32]string{
	0: EngineGoTmplate,
	1: EngineMako,
}

// RenderInInstance instance vars for renderer
type RenderInInstance struct {
	Index string                 `json:"index"`
	Vars  map[string]interface{} `json:"vars"`
}

// RenderInZone zone vars for renderer
type RenderInZone struct {
	Zone      string                 `json:"zone"`
	Vars      map[string]interface{} `json:"vars"`
	Instances []*RenderInInstance    `json:"instances"`
}

// RenderInCluster cluster vars for renderer
type RenderInCluster struct {
	Cluster       string                 `json:"cluster"`
	ClusterLabels map[string]string      `json:"clusterLabels"`
	Vars          map[string]interface{} `json:"vars"`
	Zones         []*RenderInZone        `json:"zones"`
}

// RenderInConf stdin for renderer
type RenderInConf struct {
	// Template encoded template content
	Template string                 `json:"template"`
	Vars     map[string]interface{} `json:"vars"`
	// Clusters cluster vars
	Clusters []*RenderInCluster `json:"clusters"`
	// Operator operator
	Operator string `json:"operator"`
}

// RenderOutInstance stdout for renderer
type RenderOutInstance struct {
	Cluster       string            `json:"cluster"`
	ClusterLabels map[string]string `json:"clusterLabels"`
	Zone          string            `json:"zone"`
	Index         string            `json:"index"`
	Content       string            `json:"content"`
}

// RenderOutConf stdout for renderer
type RenderOutConf struct {
	// ErrCode error code
	ErrCode pbcommon.ErrCode `json:"errCode"`
	// ErrMsg error message
	ErrMsg string `json:"errMsg"`
	// Instances stdout for instance
	Instances []*RenderOutInstance `json:"instances"`
}

// LoadStdin load stdin
func LoadStdin() (*RenderInConf, error) {
	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("read stdin failed, err %s", err.Error())
	}

	conf := &RenderInConf{}
	err = json.Unmarshal(bytes, conf)
	if err != nil {
		return nil, fmt.Errorf("decode stdin to json failed, err %s", err.Error())
	}
	return conf, nil
}

// SetStdout set output in stdout
func SetStdout(code pbcommon.ErrCode, msg string, instances []*RenderOutInstance) {
	out := &RenderOutConf{
		ErrCode:   code,
		ErrMsg:    msg,
		Instances: instances,
	}

	bytes, err := json.Marshal(out)
	if err != nil {
		fmt.Printf("encoding output %+v failed, err %s", out, err.Error())
	}
	fmt.Printf("%s", string(bytes))
}
