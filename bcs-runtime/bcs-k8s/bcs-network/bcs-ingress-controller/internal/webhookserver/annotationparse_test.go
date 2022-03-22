/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package webhookserver

import (
	"fmt"
	"reflect"
	"testing"
)

func TestParserAnnotation(t *testing.T) {
	tests := []struct {
		name            string
		value           string
		annotationPorts []*annotationPort
		exceptedError   error
	}{
		{
			name:  "sample",
			value: "portpool-sample TCP 8000",
			annotationPorts: []*annotationPort{
				{
					poolNamespace: "",
					poolName:      "portpool-sample",
					protocol:      "TCP",
					portIntOrStr:  "8000",
					hostPort:      false,
				},
			},
		},
		{
			name:  "tcp and udp",
			value: "portpool-sample TCP_UDP 8000",
			annotationPorts: []*annotationPort{
				{
					poolNamespace: "",
					poolName:      "portpool-sample",
					protocol:      "TCP_UDP",
					portIntOrStr:  "8000",
					hostPort:      false,
				},
			},
		},
		{
			name:  "with namespace",
			value: "portpool1.ns1 TCP 8000",
			annotationPorts: []*annotationPort{
				{
					poolNamespace: "ns1",
					poolName:      "portpool1",
					protocol:      "TCP",
					portIntOrStr:  "8000",
					hostPort:      false,
				},
			},
		},
		{
			name:  "with port name",
			value: "portpool1.ns1 TCP httpport",
			annotationPorts: []*annotationPort{
				{
					poolNamespace: "ns1",
					poolName:      "portpool1",
					protocol:      "TCP",
					portIntOrStr:  "httpport",
					hostPort:      false,
				},
			},
		},
		{
			name:          "no protocol",
			value:         "portpool1.ns1 httpport",
			exceptedError: fmt.Errorf("protocol %s is invalid", ""),
		},
		{
			name:  "mutil port",
			value: "portpool1.ns1 TCP 8000;portpool1.ns1 UDP 8000;portpool2.ns1 TCP 8080",
			annotationPorts: []*annotationPort{
				{
					poolNamespace: "ns1",
					poolName:      "portpool1",
					protocol:      "TCP",
					portIntOrStr:  "8000",
					hostPort:      false,
				},
				{
					poolNamespace: "ns1",
					poolName:      "portpool1",
					protocol:      "UDP",
					portIntOrStr:  "8000",
					hostPort:      false,
				},
				{
					poolNamespace: "ns1",
					poolName:      "portpool2",
					protocol:      "TCP",
					portIntOrStr:  "8080",
					hostPort:      false,
				},
			},
		},
		{
			name: "mutil port",
			value: `portpool1.ns1 TCP 8000
			portpool1.ns1 UDP 8000
			portpool2.ns1 TCP 8080`,
			annotationPorts: []*annotationPort{
				{
					poolNamespace: "ns1",
					poolName:      "portpool1",
					protocol:      "TCP",
					portIntOrStr:  "8000",
					hostPort:      false,
				},
				{
					poolNamespace: "ns1",
					poolName:      "portpool1",
					protocol:      "UDP",
					portIntOrStr:  "8000",
					hostPort:      false,
				},
				{
					poolNamespace: "ns1",
					poolName:      "portpool2",
					protocol:      "TCP",
					portIntOrStr:  "8080",
					hostPort:      false,
				},
			},
		},
		{
			name:  "hostport",
			value: "portpool1.ns1 TCP 8080/hostport",
			annotationPorts: []*annotationPort{
				{
					poolNamespace: "ns1",
					poolName:      "portpool1",
					protocol:      "TCP",
					portIntOrStr:  "8080",
					hostPort:      true,
				},
			},
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			annotationPorts, err := parserAnnotation(s.value)
			if !reflect.DeepEqual(err, s.exceptedError) {
				t.Errorf("parserAnnotation(%s) error, excepted: %v, actual: %v", s.value, s.exceptedError, err)
			}
			if !reflect.DeepEqual(annotationPorts, s.annotationPorts) {
				t.Errorf("parserAnnotation(%s) error, excepted: %v, actual: %v", s.value, s.annotationPorts, annotationPorts)
			}
		})
	}
}
