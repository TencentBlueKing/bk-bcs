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
	defaultStrategyStr := []byte(`{"bk_username":"","clean_config_name":"容器标准日志清洗","description":"容器标准日志清洗","fields":[{"field_name":"hostip","field_type":"string","field_alias":"物理机IP","is_dimension":false,"field_index":1},{"field_name":"filename","field_type":"string","field_alias":"日志文件名","is_dimension":true,"field_index":2},{"field_name":"gseindex","field_type":"string","field_alias":"gseindex","is_dimension":false,"field_index":3},{"field_name":"time","field_type":"string","field_alias":"上报时间","is_dimension":true,"field_index":4},{"field_name":"log","field_type":"string","field_alias":"日志内容","is_dimension":false,"field_index":5},{"field_name":"stream","field_type":"string","field_alias":"日志来源","is_dimension":false,"field_index":6},{"field_name":"app_id","field_type":"string","field_alias":"业务id","is_dimension":false,"field_index":7},{"field_name":"cluster","field_type":"string","field_alias":"集群id","is_dimension":true,"field_index":8},{"field_name":"container_id","field_type":"string","field_alias":"容器id","is_dimension":true,"field_index":9},{"field_name":"server_name","field_type":"string","field_alias":"server name","is_dimension":false,"field_index":10},{"field_name":"workload_type","field_type":"string","field_alias":"workload类型","is_dimension":false,"field_index":11},{"field_name":"namespace","field_type":"string","field_alias":"namespace","is_dimension":true,"field_index":12}],"json_config":"{\"conf\":{\"timezone\":8,\"time_format\":\"yyyy-MM-dd HH:mm:ss\",\"encoding\":\"UTF-8\",\"output_field_name\":\"timestamp\",\"time_field_name\":\"time\",\"timestamp_len\":0},\"extract\":{\"result\":\"json_data\",\"label\":\"label6675e0\",\"type\":\"fun\",\"next\":{\"label\":null,\"type\":\"branch\",\"next\":[{\"subtype\":\"assign_obj\",\"assign\":[{\"assign_to\":\"hostip\",\"type\":\"string\",\"key\":\"ip\"},{\"assign_to\":\"filename\",\"type\":\"string\",\"key\":\"filename\"},{\"assign_to\":\"gseindex\",\"type\":\"string\",\"key\":\"gseindex\"},{\"assign_to\":\"time\",\"type\":\"string\",\"key\":\"datetime\"}],\"type\":\"assign\",\"next\":null,\"label\":\"label24c120\"},{\"subtype\":\"access_obj\",\"result\":\"log_data\",\"key\":\"data\",\"label\":\"label95241f\",\"type\":\"access\",\"next\":{\"label\":\"label5e01f1\",\"result\":\"iter\",\"type\":\"fun\",\"next\":{\"result\":\"log_json\",\"label\":\"label5a3b65\",\"type\":\"fun\",\"next\":{\"subtype\":\"assign_obj\",\"assign\":[{\"assign_to\":\"log\",\"type\":\"string\",\"key\":\"log\"},{\"assign_to\":\"stream\",\"type\":\"string\",\"key\":\"stream\"}],\"type\":\"assign\",\"next\":null,\"label\":\"label4327d4\"},\"args\":[],\"method\":\"from_json\"},\"args\":[],\"method\":\"iterate\"}},{\"subtype\":\"access_obj\",\"result\":\"ext_data\",\"key\":\"ext\",\"label\":\"labelaa0080\",\"type\":\"access\",\"next\":{\"subtype\":\"assign_obj\",\"assign\":[{\"assign_to\":\"app_id\",\"type\":\"string\",\"key\":\"io_tencent_bcs_appid\"},{\"assign_to\":\"cluster\",\"type\":\"string\",\"key\":\"io_tencent_bcs_cluster\"},{\"assign_to\":\"container_id\",\"type\":\"string\",\"key\":\"container_id\"},{\"assign_to\":\"server_name\",\"type\":\"string\",\"key\":\"io_tencent_bcs_server_name\"},{\"assign_to\":\"workload_type\",\"type\":\"string\",\"key\":\"io_tencent_bcs_type\"},{\"assign_to\":\"namespace\",\"type\":\"string\",\"key\":\"io_tencent_bcs_namespace\"}],\"type\":\"assign\",\"next\":null,\"label\":\"label18f51d\"}}],\"name\":\"\"},\"args\":[],\"method\":\"from_json\"}}","raw_data_id":0,"result_table_name":"container_stdout_log","result_table_name_alias":"容器标准日志清洗"}`)
	err := json.Unmarshal(defaultStrategyStr, &defaultStrategy)
	if err != nil {
		fmt.Println(err)
		return
	}
}

// NewDefaultCleanStrategy returns default log clean strategy
// [RawDataID BkBizID] required as "MUST HAVE"
func NewDefaultCleanStrategy() DataCleanStrategy {
	return defaultStrategy
}

// NewDefaultAccessDeployPlanConfig returns default config for obtain dataid
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
