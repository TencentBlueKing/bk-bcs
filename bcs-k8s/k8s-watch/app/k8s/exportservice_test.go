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

package k8s

import (
	"github.com/json-iterator/go"
	"k8s.io/apimachinery/pkg/util/intstr"
	//"k8s.io/client-go/pkg/api/v1"
	"k8s.io/api/core/v1"
	"testing"
)

func TestGetServiceAndEndPoints(t *testing.T) {
	serviceInstance := &v1.Service{
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "test",
					Protocol:   "TCP",
					Port:       33333,
					TargetPort: intstr.FromInt(80),
				},
			},
		},
	}
	testString := "xxxxx"
	endPointsInstance := &v1.Endpoints{
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{
						IP:        "127.0.0.1",
						Hostname:  testString,
						NodeName:  &testString,
						TargetRef: nil,
					},
					{
						IP:        "127.0.0.2",
						Hostname:  testString,
						NodeName:  &testString,
						TargetRef: nil,
					},
					{
						IP:        "127.0.0.3",
						Hostname:  testString,
						NodeName:  &testString,
						TargetRef: nil,
					},
				},
				Ports: []v1.EndpointPort{
					{
						Name:     testString,
						Port:     80,
						Protocol: "TCP",
					},
				},
			},
		},
	}

	esc := &ExportServiceController{}
	exportPort := esc.getExportPort("", "", serviceInstance, endPointsInstance)

	outputJson := `[{"name":"xxxxx","BCSVHost":"","path":"","protocol":"tcp","servicePort":33333,"backends":[{"targetIP":"127.0.0.1","targetPort":80},{"targetIP":"127.0.0.2","targetPort":80},{"targetIP":"127.0.0.3","targetPort":80}]}]`
	output, _ := jsoniter.Marshal(exportPort)
	if string(output) != outputJson {
		t.Errorf("output: %s", output)
	}
}
