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
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"sort"
	"strconv"
)

//NewPtrExportService create default ExportService
func NewPtrExportService() *ExportService {
	svr := new(ExportService)
	svr.Balance = "roundrobin"
	svr.SSLCert = false
	svr.MaxConn = 50000
	return svr
}

//NewExportService return ExportService default object
func NewExportService() ExportService {
	return ExportService{
		Balance: "roundrobin",
		SSLCert: false,
		MaxConn: 50000,
	}
}

//ExportPort hold port reflection info
type ExportPort struct {
	Name        string      `json:"name,omitempty"`
	BCSVHost    string      `json:"BCSVHost"`
	Path        string      `json:"path"`
	Protocol    string      `json:"protocol"`
	ServicePort int         `json:"servicePort"`
	Backends    BackendList `json:"backends"`
}

//AddBackend add single backend to service Backends list
func (ep *ExportPort) AddBackend(b Backend) {
	ep.Backends = append(ep.Backends, b)
}

//Backend target backend service info
type Backend struct {
	TargetIP   string   `json:"targetIP"`
	TargetPort int      `json:"targetPort"`
	Label      []string `json:"label,omitempty"`
}

//BackendList list for backend sort
type BackendList []Backend

// Len is the number of elements in the collection.
func (bl BackendList) Len() int {
	return len(bl)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (bl BackendList) Less(i, j int) bool {
	if bl[i].TargetIP < bl[j].TargetIP {
		return true
	}
	if bl[i].TargetPort < bl[j].TargetPort {
		return true
	}
	return false
}

// Swap swaps the elements with indexes i and j.
func (bl BackendList) Swap(i, j int) {
	//el[i], el[j] = el[j], el[i]

	bl[i], bl[j] = bl[j], bl[i]
}

type ExportPortList []ExportPort

// Len is the number of elements in the collection.
func (epl ExportPortList) Len() int {
	return len(epl)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (epl ExportPortList) Less(i, j int) bool {
	ikey := epl[i].Name + strconv.Itoa(epl[i].ServicePort) + epl[i].Path
	jkey := epl[j].Name + strconv.Itoa(epl[j].ServicePort) + epl[j].Path

	if ikey != jkey {
		return ikey < jkey
	}

	if len(epl[i].Backends) > 1 {
		sort.Sort(epl[i].Backends)
	}

	for index := 0; index < len(epl[i].Backends); index++ {
		ikey = ikey + epl[i].Backends[index].TargetIP + strconv.Itoa(epl[i].Backends[index].TargetPort)
	}

	if len(epl[j].Backends) > 1 {
		sort.Sort(epl[j].Backends)
	}

	for index := 0; index < len(epl[j].Backends); index++ {
		jkey = jkey + epl[j].Backends[index].TargetIP + strconv.Itoa(epl[j].Backends[index].TargetPort)
	}

	return ikey < jkey
}

// Swap swaps the elements with indexes i and j.
func (epl ExportPortList) Swap(i, j int) {
	epl[i], epl[j] = epl[j], epl[i]
}

//DeepCopy copy src to dst by json
func DeepCopy(src, dst *ExportService) {
	dataBytes, _ := json.Marshal(src)
	json.Unmarshal(dataBytes, dst)
}

//ExportService info to hold export service
type ExportService struct {
	ObjectMeta    commtypes.ObjectMeta `json:"metadata"`
	Cluster       string               `json:"cluster"`       //cluster info
	Namespace     string               `json:"namespace"`     //namespace info, for business
	ServiceName   string               `json:"serviceName"`   //service name
	ServiceWeight map[string]int       `json:"serviceWeight"` //weight for different service
	ServicePort   []ExportPort         `json:"ports"`         //export ports info
	BCSGroup      []string             `json:"BCSGroup"`      //service export group
	SSLCert       bool                 `json:"sslcert"`       //SSL certificate for ser
	Balance       string               `json:"balance"`       //loadbalance algorithm, default source
	MaxConn       int                  `json:"maxconn"`       //max connection setting
}

//EptServiceList define ExportService list implementing sorter interface
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
