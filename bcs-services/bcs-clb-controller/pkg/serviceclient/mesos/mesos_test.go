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

package mesos

import (
	"reflect"
	"testing"

	schetypes "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	v2 "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/apis/bkbcs/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestCheckTaskgroupRunning test checkTaskgroupRunning function
func TestCheckTaskgroupRunning(t *testing.T) {
	var tg *v2.TaskGroup
	if checkTaskgroupRunning(tg) {
		t.Errorf("result of empty taskgroup should be false")
		return
	}
	tg = &v2.TaskGroup{
		Spec: v2.TaskGroupSpec{
			TaskGroup: schetypes.TaskGroup{
				Status: "Running",
			},
		},
	}
	if !checkTaskgroupRunning(tg) {
		t.Errorf("result of running taskgroup should be true")
		return
	}
	tg.Spec.Status = "Dead"
	if checkTaskgroupRunning(tg) {
		t.Errorf("result of dead taskgroup should be false")
		return
	}
}

// TestDecodeTaskgroupStatusData test decodeTaskgroupStatusData function
func TestDecodeTaskgroupStatusData(t *testing.T) {
	data := `{"ID":"xxxxxxxx","Name":"bcs-container-xxxxxx","Pid":1,` +
		`"StartAt":"2020-11-19T06:29:54.450062716Z","FinishAt":"0001-01-01T00:00:00Z",` +
		`"Status":"running","Healthy":true,"Hostname":"fake-node-hostname",` +
		`"NetworkMode":"host","NodeAddress":"127.0.0.1","Message":"container is running,` +
		`healthy status unkown","Resource":{"Cpus":0,"CPUSet":0,"Mem":256,"Disk":0}}`
	_, err := decodeTaskgroupStatusData(data)
	if err != nil {
		t.Error(err)
	}

	data = `[]`
	_, err = decodeTaskgroupStatusData(data)
	if err == nil {
		t.Error("should throw error")
	}
}

// Test test getIndexRcFromTaskgroupName function
func TestGetIndexRcFromTaskgroupName(t *testing.T) {
	testCases := []struct {
		name   string
		index  int
		rcName string
		hasErr bool
	}{
		{
			name:   "1.application-1.test.00001.160576739375102",
			index:  1,
			rcName: "application-1",
			hasErr: false,
		},
		{
			name:   "application-1.test.00001",
			hasErr: true,
		},
		{
			name:   "a.application-1.test.00001.xxxxx",
			hasErr: true,
		},
	}
	for _, test := range testCases {
		tmpIndex, tmpRcName, err := getIndexRcFromTaskgroupName(test.name)
		if (err != nil && !test.hasErr) ||
			(err == nil && test.hasErr) {
			t.Errorf("hasErr %v, err %v", test.hasErr, err)
			continue
		}
		if tmpIndex != test.index {
			t.Errorf("expected index %d, current index %d", test.index, tmpIndex)
			continue
		}
		if tmpRcName != test.rcName {
			t.Errorf("expected rcName %s, current rcName %s", test.rcName, tmpRcName)
			continue
		}
	}
}

// TestSortTaskgroups test sortTaskgroups
func TestSortTaskgroups(t *testing.T) {
	tgs := []*v2.TaskGroup{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "2.application-1.test.00001.160576739375102",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "0.application-1.test.00001.160576739375102",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "2.ppplication-1.test.00001.160576739375102",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "1.application-1.test.00001.160576739375102",
			},
		},
	}
	tgsAfterSort := []*v2.TaskGroup{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "0.application-1.test.00001.160576739375102",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "1.application-1.test.00001.160576739375102",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "2.application-1.test.00001.160576739375102",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "2.ppplication-1.test.00001.160576739375102",
			},
		},
	}
	sortTaskgroups(tgs)
	if !reflect.DeepEqual(tgs, tgsAfterSort) {
		t.Errorf("%+v should be equal to %+v", tgs, tgsAfterSort)
	}
}
