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

package endpoint

import (
	"context"
	"testing"
	"time"
)

func getEndpointsClient() ClusterEndpointsIP {
	k8sConfig := getK8sConfig()
	if k8sConfig == nil {
		return nil
	}

	clusterEndpointsClient, err := NewEndpointsClient(WithK8sConfig(*k8sConfig), WithDebug(true))
	if err != nil {
		return nil
	}

	return clusterEndpointsClient
}

func TestEndpoints_GetClusterEndpoints(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := getEndpointsClient()
	if client == nil {
		t.Fatalf("getEndpointsClient failed")
		return
	}

	t.Logf("%+v", client)

	go client.SyncClusterEndpoints(ctx)

	for {
		select {
		case <-ctx.Done():
			client.Stop()
			t.Logf("SyncClusterEndpoints quit: %v", ctx.Err())
			return
		case <-time.After(time.Second * 5):
		}

		endpoints, err := client.GetClusterEndpoints()
		if err != nil {
			t.Fatalf("GetClusterEndpoints failed: %v", err)
			return
		}

		t.Logf("GetClusterEndpoints %+v", endpoints)
	}
}
