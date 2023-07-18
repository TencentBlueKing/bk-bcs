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

package api

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"

	as "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/as/v20180419"
	tke "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tke/v20180525"
)

const (
	TencentCloudSecretIDClusterEnv  = "InnerKey"
	TencentCloudSecretKeyClusterEnv = "InnerSecret"
)

func getASClient(region string) *ASClient {
	cli, err := NewASClient(&cloudprovider.CommonOption{
		Account: &cmproto.Account{
			SecretID:  os.Getenv(TencentCloudSecretIDClusterEnv),
			SecretKey: os.Getenv(TencentCloudSecretKeyClusterEnv),
		},
		Region: region,
	})
	if err != nil {
		panic(err)
	}
	return cli
}

func TestDescribeAutoScalingInstances(t *testing.T) {
	cli := getASClient("ap-xxx")
	ins, err := cli.DescribeAutoScalingInstances("asg-xxx")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("ins: %s, count: %d", utils.ToJSONString(ins), len(ins))
}

func TestRemoveInstances(t *testing.T) {
	cli := getASClient("ap-xxx")
	acID, err := cli.RemoveInstances("asg-xxx", []string{"ins-xxx", "ins-xxx"})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(acID)
}

func TestDetachInstances(t *testing.T) {
	cli := getASClient("ap-guangzhou")
	err := cli.DetachInstances("asg-xxx", []string{"ins-xxx"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestModifyDesiredCapacity(t *testing.T) {
	cli := getASClient("ap-guangzhou")
	err := cli.ModifyDesiredCapacity("asg-xxx", 5)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDescribeAutoScalingGroups(t *testing.T) {
	cli := getASClient("ap-guangzhou")
	asg, err := cli.DescribeAutoScalingGroups("asg-xxx")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(*asg.DefaultCooldown)
}

func TestDescribeAutoScalingActivities(t *testing.T) {
	cli := getASClient("ap-guangzhou")
	activity, err := cli.DescribeAutoScalingActivities("")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(activity.StatusCode)
}

// ResourceUnavailable.AutoScalingGroupInActivity 伸缩组正在活动中
func TestASClient_ScaleOutInstances(t *testing.T) {
	cli := getASClient("ap-guangzhou")
	activityID, err := cli.ScaleOutInstances("asg-xxx", 3)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf(activityID)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var (
		activity *as.Activity
	)
	for {
		select {
		case <-ticker.C:
		default:
			continue
		}
		activity, err = cli.DescribeAutoScalingActivities(activityID)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(*activity.StatusCode)
		if *activity.StatusCode == "SUCCESSFUL" {
			break
		}
	}

	var (
		successInstanceID []string
		failedInstanceID  []string
	)

	for _, ins := range activity.ActivityRelatedInstanceSet {
		if *ins.InstanceStatus == "SUCCESSFUL" {
			successInstanceID = append(successInstanceID, *ins.InstanceId)
		} else {
			failedInstanceID = append(failedInstanceID, *ins.InstanceId)
		}
	}

	fmt.Printf("%+v, %+v\n", successInstanceID, failedInstanceID)

	tkeCli := getClient("ap-guangzhou")

	// wait node group state to normal
	timeCtx, cancel := context.WithTimeout(context.TODO(), 10*time.Minute)
	defer cancel()

	var (
		addSucessNodes  = make([]string, 0)
		addFailureNodes = make([]string, 0)
		instances       []*tke.Instance
	)

	// wait all nodes to be ready
	err = loop.LoopDoFunc(timeCtx, func() error {
		instances, err = tkeCli.QueryTkeClusterInstances(&DescribeClusterInstances{
			ClusterID:   "cls-xxx",
			InstanceIDs: successInstanceID,
		})
		if err != nil {
			return nil
		}

		index := 0
		running, failure := make([]string, 0), make([]string, 0)
		for _, ins := range instances {
			t.Logf("checkClusterInstanceStatus instance[%s] status[%s]", *ins.InstanceId, *ins.InstanceState)
			switch *ins.InstanceState {
			case RunningInstanceTke.String():
				running = append(running, *ins.InstanceId)
				index++
			case FailedInstanceTke.String():
				failure = append(failure, *ins.InstanceId)
				index++
			default:
			}
		}

		if index == len(successInstanceID) {
			addSucessNodes = running
			addFailureNodes = failure
			return loop.EndLoop
		}

		return nil
	}, loop.LoopInterval(10*time.Second))
	// other error
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		fmt.Printf("checkClusterInstanceStatus QueryTkeClusterInstances failed: %v", err)
		return
	}

	t.Log(addSucessNodes)
	t.Log(addFailureNodes)

}

func TestASClient_ScaleOutInstances2(t *testing.T) {
	cli := getASClient("ap-guangzhou")

	var (
		activityID string
		err        error
	)

	loop.LoopDoFunc(context.Background(), func() error {
		activityID, err = cli.ScaleOutInstances("asg-xxx", 2)
		if err != nil {
			if strings.Contains(err.Error(), as.RESOURCEUNAVAILABLE_AUTOSCALINGGROUPINACTIVITY) {
				return nil
			}
			return err
		}

		fmt.Println(activityID)
		return loop.EndLoop
	}, loop.LoopInterval(1*time.Second))

	if activityID == "" {
		t.Fatal("failed")
	}
	t.Log(activityID)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
		default:
			continue
		}
		activity, err := cli.DescribeAutoScalingActivities(activityID)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(*activity.StatusCode)
		if *activity.StatusCode == "SUCCESSFUL" {
			for _, ins := range activity.ActivityRelatedInstanceSet {
				fmt.Println(*ins.InstanceId, *ins.InstanceStatus)
			}
			return
		}
	}
}

func TestASClient_ScaleInInstances(t *testing.T) {
	cli := getASClient("ap-guangzhou")
	activityID, err := cli.ScaleInInstances("asg-xxx", 1)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf(activityID)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
		default:
			continue
		}
		activity, err := cli.DescribeAutoScalingActivities(activityID)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(*activity.StatusCode)
		if *activity.StatusCode == "SUCCESSFUL" {
			return
		}
	}
}

func TestDescribeLaunchConfigurations(t *testing.T) {
	cli := getASClient("ap-guangzhou")
	asc, err := cli.DescribeLaunchConfigurations([]string{"asc-xxx"})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(utils.ToJSONString(asc))
}
