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

package api

import (
	"context"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v3"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

func TestGetMaxPods(t *testing.T) {
	nts := []*proto.NodeTemplate{
		{
			ExtraArgs: map[string]string{
				kubeletType: "xxxx;max-pods=1",
			},
		},
		{
			ExtraArgs: map[string]string{
				kubeletType: "xxxx;max-pods=2;yyyy",
			},
		},
		{
			ExtraArgs: map[string]string{
				kubeletType: "xxxxx;max-pods=3;",
			},
		},
		// 反例
		{
			ExtraArgs: map[string]string{
				kubeletType: "xxxxx;max-pods=;xxx",
			},
		},
		{
			ExtraArgs: map[string]string{
				kubeletType: "xxxxx;max-pod=4;",
			},
		},
		{
			ExtraArgs: map[string]string{
				kubeletType: "xxxxx;max-pod=5xxx;",
			},
		},
		{
			ExtraArgs: map[string]string{
				kubeletType: "xxxxx;max-pod=6!!xxx;",
			},
		},
		{
			ExtraArgs: map[string]string{
				kubeletType: "xxxxx",
			},
		},
		{
			ExtraArgs: map[string]string{},
		},
		{},
	}

	for _, nt := range nts {
		pool := new(armcontainerservice.AgentPool)
		ng := new(proto.NodeGroup)
		ng.NodeTemplate = nt
		ap := newNodeGroupToAgentPoolConverter(ng, pool)

		t.Logf("pod number:%v", ap.getMaxPods(nt.ExtraArgs))
	}
}

func buildTaint(taint string) (res *proto.Taint) {
	res = &proto.Taint{}

	v := strings.Split(taint, "=")
	if len(v) == 0 {
		return res
	}
	res.Key = v[0]
	if len(v) <= 1 {
		return res
	}

	v = strings.Split(v[1], ":")
	if len(v) == 0 {
		return res
	}
	res.Value = v[0]
	if len(v) <= 1 {
		return res
	}

	res.Effect = v[1]
	return res
}

func TestGetTaint(t *testing.T) {
	data := []string{
		"key1=value1:NoSchedule",
		// 反例
		"key1=value1:",
		"key1=value1",
		"key1=",
		"key1",
		"",
	}
	for _, s := range data {
		t.Log(buildTaint(s))
	}
}

func TestTick(t *testing.T) {
	tick := time.Tick(2 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

into:
	for {
		select {
		case <-ctx.Done():
			t.Log("stop for.")
			break into
		case <-tick:
			t.Log(time.Now().String())
		}
	}
	t.Log("end")
}

func TestBuyDataDisk(t *testing.T) {
	set := new(armcompute.VirtualMachineScaleSet)

	group := &proto.NodeGroup{
		LaunchTemplate: &proto.LaunchConfiguration{
			DataDisks: []*proto.DataDisk{
				{
					DiskSize: "50",
					DiskType: "Premium_LRS",
				},
				{
					DiskSize: "60",
					DiskType: "StandardSSD_LRS",
				},
			},
		},
	}

	BuyDataDisk(group, set)

	log.Println(set)
}
