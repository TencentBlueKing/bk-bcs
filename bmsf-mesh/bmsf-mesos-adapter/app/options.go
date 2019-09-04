/*
Copyright (C) 2019 The BlueKing Authors. All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package app

import (
	"bk-bcs/bcs-common/common/conf"
	"fmt"
)

// Config detail configuration item
type Config struct {
	conf.FileConfig
	conf.ServiceConfig
	conf.MetricConfig
	conf.CertConfig
	conf.ZkConfig
	conf.LogConfig
	conf.ProcessConfig
	Scheme     string `json:"metric_scheme" value:"http" usage:"scheme for metric api"`
	Zookeeper  string `json:"zookeeper" value:"127.0.0.1:3181" usage:"data source for taskgroups and services"`
	Cluster    string `json:"cluster" value:"" usage:"cluster id or name"`
	KubeConfig string `json:"kubeconfig" value:"kubeconfig" usage:"configuration file for kube-apiserver"`
}

// Validate validate command line parameter
func (c *Config) Validate() error {
	if len(c.Cluster) == 0 {
		return fmt.Errorf("cluster cannot be empty")
	}
	return nil
}
