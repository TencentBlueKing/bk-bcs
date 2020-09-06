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

package bkdata

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strings"
)

var defaultStrategy DataCleanStrategy

// BKDataClientConfig bkdata client config
type BKDataClientConfig struct {
	BkAppCode                  string
	BkUsername                 string
	BkAppSecret                string
	BkdataAuthenticationMethod string
	Host                       string
	TLSConf                    *tls.Config
}

// CustomAccessDeployPlanConfig is used to obtain dataid from bk-data
type CustomAccessDeployPlanConfig struct {
	BkAppCode                  string        `json:"bk_app_code"`
	BkUsername                 string        `json:"bk_username"`
	BkAppSecret                string        `json:"bk_app_secret"`
	BkdataAuthenticationMethod string        `json:"bkdata_authentication_method"`
	DataScenario               string        `json:"data_scenario"`
	BkBizID                    int           `json:"bk_biz_id"`
	Description                string        `json:"description"`
	Appenv                     string        `json:"appenv"`
	AccessRawData              AccessRawData `json:"access_raw_data"`
}

// AccessRawData is part of CustomAccessDeployPlanConfig
type AccessRawData struct {
	RawDataName  string `json:"raw_data_name"`
	Maintainer   string `json:"maintainer"`
	RawDataAlias string `json:"raw_data_alias"`
	DataSource   string `json:"data_source"`
	DataEncoding string `json:"data_encoding"`
	Sensitivity  string `json:"sensitivity"`
	Description  string `json:"description"`
}

// DataCleanStrategy is used to create data clean strategy
type DataCleanStrategy struct {
	BkAppCode                  string   `json:"bk_app_code"`
	BkUsername                 string   `json:"bk_username"`
	BkAppSecret                string   `json:"bk_app_secret"`
	BkdataAuthenticationMethod string   `json:"bkdata_authentication_method"`
	RawDataID                  int      `json:"raw_data_id"`
	JSONConfig                 string   `json:"json_config"`
	PeConfig                   string   `json:"pe_config"`
	BkBizID                    int      `json:"bk_biz_id"`
	CleanConfigName            string   `json:"clean_config_name"`
	ResultTableName            string   `json:"result_table_name"`
	ResultTableNameAlias       string   `json:"result_table_name_alias"`
	Description                string   `json:"description"`
	Fields                     []Fields `json:"fields"`
}

// Fields defines result table column info of log clean strategy
type Fields struct {
	FieldName   string `json:"field_name"`
	FieldAlias  string `json:"field_alias"`
	FieldType   string `json:"field_type"`
	IsDimension bool   `json:"is_dimension"`
	FieldIndex  int    `json:"field_index"`
}

func init() {
	defaultStrategyStr := `{"bk_username":"","clean_config_name":"容器标准日志清洗","description":"容器标准日志清洗","fields":[{"field_name":"log","field_type":"string","field_alias":"日志log","is_dimension":false,"field_index":1},{"field_name":"stream","field_type":"string","field_alias":"stream","is_dimension":false,"field_index":2},{"field_name":"time","field_type":"string","field_alias":"日志time","is_dimension":false,"field_index":3},{"field_name":"bcs_appid","field_type":"string","field_alias":"appid","is_dimension":false,"field_index":4},{"field_name":"bcs_cluster","field_type":"string","field_alias":"cluster","is_dimension":false,"field_index":5},{"field_name":"container_id","field_type":"string","field_alias":"container_id","is_dimension":false,"field_index":6},{"field_name":"bcs_namespace","field_type":"string","field_alias":"namespace","is_dimension":false,"field_index":7},{"field_name":"jsontest","field_type":"text","field_alias":"test","is_dimension":false,"field_index":8}],"json_config":"{\"extract\":{\"args\":[],\"label\":\"json_data\",\"result\":\"json_data\",\"next\":{\"label\":null,\"type\":\"branch\",\"name\":\"\",\"next\":[{\"label\":\"access_data\",\"subtype\":\"access_obj\",\"result\":\"access_data\",\"key\":\"_value_\",\"next\":{\"args\":[],\"label\":\"iter_data\",\"result\":\"iter_data\",\"next\":{\"args\":[],\"label\":\"log\",\"result\":\"log_data\",\"next\":{\"subtype\":\"assign_obj\",\"label\":\"label045559\",\"type\":\"assign\",\"assign\":[{\"type\":\"string\",\"assign_to\":\"log\",\"key\":\"log\"},{\"type\":\"string\",\"assign_to\":\"stream\",\"key\":\"stream\"},{\"type\":\"string\",\"assign_to\":\"time\",\"key\":\"time\"}],\"next\":null},\"type\":\"fun\",\"method\":\"from_json\"},\"type\":\"fun\",\"method\":\"iterate\"},\"type\":\"access\"},{\"label\":\"label0a99a0\",\"subtype\":\"access_obj\",\"result\":\"private_val\",\"key\":\"_private_\",\"next\":{\"subtype\":\"assign_obj\",\"label\":\"labela29258\",\"type\":\"assign\",\"assign\":[{\"type\":\"string\",\"assign_to\":\"bcs_appid\",\"key\":\"io.tencent.bcs.app.appid\"},{\"type\":\"string\",\"assign_to\":\"bcs_cluster\",\"key\":\"io.tencent.bcs.cluster\"},{\"type\":\"string\",\"assign_to\":\"container_id\",\"key\":\"container_id\"},{\"type\":\"string\",\"assign_to\":\"bcs_namespace\",\"key\":\"io.tencent.bcs.namespace\"}],\"next\":null},\"type\":\"access\"},{\"subtype\":\"assign_json\",\"label\":\"labelwnplb\",\"type\":\"assign\",\"assign\":[{\"type\":\"text\",\"assign_to\":\"jsontest\",\"key\":\"_private_\"}],\"next\":null}]},\"type\":\"fun\",\"method\":\"from_json\"},\"conf\":{\"timestamp_len\":0,\"encoding\":\"UTF-8\",\"time_format\":\"yyyy-MM-dd'T'HH:mm:ssXXX\",\"timezone\":8,\"output_field_name\":\"timestamp\",\"time_field_name\":\"time\"}}","raw_data_id":0,"result_table_name":"container_stdout_log","result_table_name_alias":"容器标准日志"}`
	err := json.NewDecoder(strings.NewReader(defaultStrategyStr)).Decode(&defaultStrategy)
	if err != nil {
		fmt.Println(err)
		return
	}
}

// NewDefaultLogCollectionDataCleanStrategy returns default log clean strategy
// [RawDataID BkBizID] required as "MUST HAVE"
func NewDefaultCleanStrategy() DataCleanStrategy {
	return defaultStrategy
}

// NewDefaultCustomAccessDeployPlanConfig returns default config for obtain dataid
// [BkBizID RawDataName RawDataAlias Maintainer] required as "MUST HAVE"
func NewDefaultAccessDeployPlanConfig() CustomAccessDeployPlanConfig {
	return CustomAccessDeployPlanConfig{
		Appenv:       "ieod",
		DataScenario: "custom",
		AccessRawData: AccessRawData{
			DataSource:   "svr",
			DataEncoding: "UTF-8",
			Sensitivity:  "private",
		},
	}
}

// DeepCopyInto deep copy method of DataCleanStrategy
func (in *DataCleanStrategy) DeepCopyInto(out *DataCleanStrategy) {
	*out = *in
	var fields []Fields
	for _, v := range in.Fields {
		fields = append(fields, v)
	}
	out.Fields = fields
}
