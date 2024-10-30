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

// Package clustercheck xxx
package clustercheck

import (
	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
)

const (
	namespace = "bkmonitor-operator"
)

// Options bcs log options
type Options struct {
	Interval        int         `json:"interval" yaml:"interval"`
	TestYaml        interface{} `json:"testyaml" yaml:"testyaml"`
	Synchronization bool        `json:"synchronization" yaml:"synchronization"`
	Namespace       string      `json:"namespace" yaml:"namespace"`
}

// Validate validate options
func (o *Options) Validate() error {
	if o.Namespace == "" {
		o.Namespace = namespace
	}

	if o.TestYaml == nil {
		yamlStr := `
apiVersion: batch/v1
kind: Job
metadata:
  name: bcs-blackbox-job
  namespace: bcs-blackbox-job
spec:
  backoffLimit: 1
  template:
    metadata:
      labels:
        test-yaml: test-yaml
    spec:
      automountServiceAccountToken: false
      containers:
      - image: hub.bktencent.com/library/hello-world:latest
        imagePullPolicy: Always
        name: blackbox
        resources:
          limits:
            cpu: 100m
            memory: 100Mi
          requests:
            cpu: 100m
            memory: 100Mi
      nodeSelector:
        kubernetes.io/os: linux
        kubernetes.io/arch: amd64
      restartPolicy: Never
      tolerations:
      - effect: NoSchedule
        operator: Exists
`
		o.TestYaml = make(map[string]interface{})
		err := yaml.Unmarshal([]byte(yamlStr), o.TestYaml)
		if err != nil {
			klog.Infof(err.Error())
		}
	}

	return nil
}
