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

package metric

import (
	"encoding/json"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

var (
	logFileInfoGaugeLabelNames = []string{"clusterID", "crdName", "crdNamespace", "hostIP", "containerID",
		"podName", "podNamespace", "workloadType", "workloadName", "workloadNamespace"}
	// LogFileInfoGauge is container log collection task Gauge metric
	LogFileInfoGauge *prometheus.GaugeVec
)

// LogFileInfoType is structure viewed label value for LogFileInfoGauge
type LogFileInfoType struct {
	// ClusterID can be seen as primary key
	ClusterID         string `json:"clusterID"`
	CRDName           string `json:"crdName"`
	CRDNamespace      string `json:"crdNamespace"`
	HostIP            string `json:"hostIP"`
	ContainerID       string `json:"containerID"`
	PodName           string `json:"podName"`
	PodNamespace      string `json:"podNamespace"`
	WorkloadType      string `json:"workloadType"`
	WorkloadName      string `json:"workloadName"`
	WorkloadNamespace string `json:"workloadNamespace"`
}

func init() {
	LogFileInfoGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "logbeat_sidecar_logfile_gen_info",
			Help: "logbeat_sidecar_logfile_gen_info has labels describing the information" +
				" of logfile and gauge value 0/1 corresponds to (logfile generated/logfile does not generated)",
		},
		logFileInfoGaugeLabelNames,
	)
	prometheus.MustRegister(LogFileInfoGauge)
}

// LogFileInfoGaugeLabelCvt convert structure viewed label value to MapString
func LogFileInfoGaugeLabelCvt(info *LogFileInfoType) map[string]string {
	jsonstr, err := json.Marshal(info)
	if err != nil {
		blog.Errorf("Convert *LogFileInfoType(%+v) to json string failed: %s", *info, err.Error())
		return nil
	}
	mapstr := make(map[string]string)
	err = json.Unmarshal(jsonstr, &mapstr)
	if err != nil {
		blog.Errorf("Convert json string(%s) to json object failed: %s", string(jsonstr), err.Error())
		return nil
	}
	return mapstr
}

// Update use new label to replace old label with gauge value plus 1 when container collection task changed
func (info *LogFileInfoType) Update(newinfo *LogFileInfoType) error {
	if info == nil {
		return fmt.Errorf("Update metric failed with nil old metric")
	}
	if newinfo == nil {
		return fmt.Errorf("Update metric failed with nil new metric")
	}
	oldLabelValue := LogFileInfoGaugeLabelCvt(info)
	newLabelValue := LogFileInfoGaugeLabelCvt(newinfo)
	// are they same?
	same := true
	for key := range oldLabelValue {
		if oldLabelValue[key] != newLabelValue[key] {
			same = false
			break
		}
	}
	// get old value
	m := &dto.Metric{}
	singleGauge, err := LogFileInfoGauge.GetMetricWith(oldLabelValue)
	if err != nil {
		return err
	}
	singleGauge.Write(m) // nolint
	if err != nil {
		return err
	}
	oldValue := int64(m.GetGauge().GetValue())
	if oldValue == 0 {
		blog.Warnf("Old config metric with label (%+v) does not exist", newLabelValue)
	}
	// get new value
	singleGauge, err = LogFileInfoGauge.GetMetricWith(newLabelValue)
	if err != nil {
		return err
	}
	singleGauge.Write(m) // nolint error not checked
	if err != nil {
		return err
	}
	newValue := int64(m.GetGauge().GetValue())
	newinfo.set(oldValue+1, newLabelValue)
	//	newvalue	same	op
	//	!0			y		nothing	(new and old's label did not change)
	//	!0			n		warn & delete (new label changed, but new is exist, should warn)
	//	0			n		delete (new label changed and new does not exist, should delete old)
	if newValue != 0 {
		if !same {
			blog.Warnf("New config with metric label (%+v) already exists, it may be covered by the update operation",
				newLabelValue)
			info.delete(oldLabelValue)
		}
	} else {
		info.delete(oldLabelValue)
	}
	return nil
}

// Set set the gauge value
func (info *LogFileInfoType) Set(value int64) error {
	if info == nil {
		return fmt.Errorf("Set metric failed with nil metric info")
	}
	info.set(value, LogFileInfoGaugeLabelCvt(info))
	return nil
}

// Delete delete the gauge with label value equals to info
func (info *LogFileInfoType) Delete() error {
	if info == nil {
		return fmt.Errorf("Delete metric failed with nil metric info")
	}
	info.delete(LogFileInfoGaugeLabelCvt(info))
	return nil
}

// Renew renews the gauge
func (info *LogFileInfoType) Renew() error {
	if info == nil {
		return fmt.Errorf("Renew metric failed with nil metric info")
	}
	labelValue := LogFileInfoGaugeLabelCvt(info)
	// get old value
	m := &dto.Metric{}
	singleGauge, err := LogFileInfoGauge.GetMetricWith(labelValue)
	if err != nil {
		return err
	}
	singleGauge.Write(m) // nolint error not checked
	if err != nil {
		return err
	}
	oldValue := int64(m.GetGauge().GetValue())
	if oldValue == 0 {
		oldValue = 1
	}
	info.set(oldValue, labelValue)
	return nil
}

func (info *LogFileInfoType) set(value int64, labelValues map[string]string) {
	LogFileInfoGauge.With(labelValues).Set(float64(value))
}

func (info *LogFileInfoType) delete(labelValues map[string]string) {
	LogFileInfoGauge.Delete(labelValues)
}
