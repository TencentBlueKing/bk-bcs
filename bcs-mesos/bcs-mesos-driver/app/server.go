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

package app

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-driver/app/options"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-driver/mesosdriver"
)

//Run the mesos driver
func Run(op *options.MesosDriverOption) error {

	blog.Info("config: %+v", op)

	setConfig(op)

	driver, err := mesosdriver.NewMesosDriverServer(op.DriverConf)
	if err != nil {
		blog.Error("fail to create mesos driver. err:%s", err.Error())
		return err
	}

	blog.Info("app begin to start mesos driver ... ")
	driver.Start()
	blog.Info("mesos driver finish ")
	return nil
}

func setConfig(op *options.MesosDriverOption) {

	if op.DriverConf.ServCert.CertFile != "" && op.DriverConf.ServCert.KeyFile != "" {
		op.DriverConf.ServCert.IsSSL = true
	}

	if op.DriverConf.ClientCert.CertFile != "" && op.DriverConf.ClientCert.KeyFile != "" {
		op.DriverConf.ClientCert.IsSSL = true
	}
}
