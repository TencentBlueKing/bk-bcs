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

package metric

func NewMetricController(conf Config, healthFunc HealthFunc, metrics ...*MetricContructor) error {
	return newMetricHandler(conf, healthFunc, metrics...)
}

type RunModeType string

// used when your module running with Master_Slave_Mode mode
type RoleType string

const (
	Master_Slave_Mode  RunModeType = "master-slave"
	Master_Master_Mode RunModeType = "master-master"
	MasterRole         RoleType    = "master"
	SlaveRole          RoleType    = "slave"
	UnknownRole        RoleType    = "unknown"
)

type Config struct {
	// name of your module
	ModuleName string
	// running mode of your module
	// could be one of Master_Slave_Mode or Master_Master_Mode
	RunMode RunModeType
	// ip address of this module running on
	IP string
	// port number of the metric's http handler depends on.
	MetricPort uint
	// cluster id of your module belongs to.
	ClusterID string
	// self defined info labeled on your metrics.
	// deprecated, unused now.
	Labels map[string]string
	// whether disable golang's metric, default is false.
	DisableGolangMetric bool
	// metric http server's ssl configuration
	SvrCaFile   string
	SvrCertFile string
	SvrKeyFile  string
	SvrKeyPwd   string
}

type HealthFunc func() HealthMeta

type HealthMeta struct {
	// the running role of your module when you are running with Master_Slave_Mode.
	// must be not empty. if you set with an empty value, an error will be occurred.
	// when your module is running in Master_Master_Mode,  this filed should be set
	// with value of "Slave".
	CurrentRole RoleType `json:"current_role"`
	// if this module is healthy
	IsHealthy bool `json:"healthy"`
	// messages which describes the health status
	Message string `json:"message"`
}

type MetricMeta struct {
	// metric's name
	Name string
	// metric's help info, which should be short and briefly.
	Help string
	// metric labels, which can describe the special info about this metric.
	ConstLables map[string]string
}

type GetMetaFunc func() *MetricMeta
type GetResultFunc func() (*MetricResult, error)
type MetricContructor struct {
	GetMeta   GetMetaFunc
	GetResult GetResultFunc
}

type MetricResult struct {
	Value *FloatOrString
	// variable labels means that this labels value can be changed with each call.
	VariableLabels map[string]string
}
