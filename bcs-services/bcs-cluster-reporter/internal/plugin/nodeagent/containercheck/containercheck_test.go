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

// Package containercheck xxx
package containercheck

import (
	"fmt"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"
	"github.com/docker/docker/api/types"
)

func TestGetDockerCli(t *testing.T) {
	cli, err := GetDockerCli("/var/run/docker.sock")
	if err != nil {
		t.Errorf(err.Error())
	}

	ctx := util.GetCtx(time.Second * 10)
	containerList, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		t.Errorf(err.Error())
	}

	for _, container := range containerList {
		status, err := GetContainerPIDStatus(1)
		if err != nil {
			t.Errorf(err.Error())
		}
		fmt.Printf("%s: %s\n", container.ID, status)

	}

}
