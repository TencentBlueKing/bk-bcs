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

// Package kubehelm xxx
package kubehelm

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"k8s.io/klog/v2"
)

// command to use helm struct
type cmdHelm struct{}

// NewCmdHelm new cmd helm struct, the object requires helm command-line tool
func NewCmdHelm() KubeHelm {
	return &cmdHelm{}
}

// InstallChart xxx
// nolint helm install --name xxxx chart-dir --set k1=v1 --set k2=v2 --kube-apiserver=xxxx --kube-token=xxxxx --kubeconfig kubeconfig
func (h *cmdHelm) InstallChart(inf InstallFlags, glf GlobalFlags) error {
	gPara, err := glf.ParseParameters()
	if err != nil {
		return err
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(glf.Kubeconfig)

	parameters := inf.ParseParameters() + gPara
	klog.Infof("helm install%s", parameters)
	_ = os.Remove("install.sh")
	file, err := os.OpenFile("install.sh", os.O_CREATE|os.O_RDWR, 0755) // NOCC:gas/permission(设计如此)
	if err != nil {
		return err
	}
	err = file.Truncate(0)
	if err != nil {
		_ = file.Close()
		return err
	}
	_, err = file.Write([]byte(fmt.Sprintf("helm install%s", parameters)))
	_ = file.Close()
	if err != nil {
		return err
	}
	cmd := exec.Command("/bin/bash", "-c", "./install.sh") // NOCC:gas/subprocess(设计如此)
	buf := bytes.NewBuffer(make([]byte, 1024))
	cmd.Stderr = buf
	err = cmd.Run()
	if err != nil {
		klog.Errorf("helm install failed, stderr %s error %s", buf.String(), err.Error())
		return err
	}

	klog.Infof("helm install %s success", parameters)
	return nil
}
