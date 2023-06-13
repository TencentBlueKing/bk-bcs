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

package loadbalance

import (
	"encoding/json"
	"fmt"
)

// NewPtrExportService create default ExportService
func NewPtrExportService() *ExportService {
	svr := new(ExportService)
	svr.Balance = "roundrobin"
	svr.SSLCert = false
	svr.MaxConn = 50000
	return svr
}

// NewExportService return ExportService default object
func NewExportService() ExportService {
	return ExportService{
		Balance: "roundrobin",
		SSLCert: false,
		MaxConn: 50000,
	}
}

// ExportPort hold port reflection info
type ExportPort struct {
	BCSVHost    string `json:"BCSVHost"`
	Protocol    string `json:"protocol"`
	ServicePort int    `json:"servicePort"`
	TargetPort  int    `json:"targetPort"`
}

// DeepCopy copy src to dst by json
func DeepCopy(src, dst *ExportService) {
	dataBytes, err := json.Marshal(src)
	if err != nil {
		fmt.Println(err.Error())
	}
	if err = json.Unmarshal(dataBytes, dst); err != nil {
		fmt.Println(err.Error())
	}
}

// ExportService info to hold export service
type ExportService struct {
	Cluster     string       `json:"cluster"`     // cluster info
	Namespace   string       `json:"namespace"`   // namespace info, for business
	ServiceName string       `json:"serviceName"` // service name
	ServicePort []ExportPort `json:"ports"`       // export ports info
	Backends    []string     `json:"backend"`     // backend ip list
	BCSGroup    []string     `json:"BCSGroup"`    // service export group
	SSLCert     bool         `json:"sslcert"`     // SSL certificate for ser
	Balance     string       `json:"balance"`     // loadbalance algorithm, default source
	MaxConn     int          `json:"maxconn"`     // max connection setting
}

// AddBackend add single backend to service Backends list
func (es *ExportService) AddBackend(ip string) {
	es.Backends = append(es.Backends, ip)
}

// EptServiceList define ExportService list implementing sorter interface
type EptServiceList []ExportService

// Len is the number of elements in the collection.
func (el EptServiceList) Len() int {
	return len(el)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (el EptServiceList) Less(i, j int) bool {
	return el[i].ServiceName < el[j].ServiceName
}

// Swap swaps the elements with indexes i and j.
func (el EptServiceList) Swap(i, j int) {
	el[i], el[j] = el[j], el[i]
}
