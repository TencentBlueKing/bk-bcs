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
	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-upgrader/app/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-upgrader/upgrader"
)

// Run is to run the bcs-upgrader
func Run(op *options.UpgraderOptions) error {
	setConfig(op)

	upgrader, err := upgrader.NewUpgrader(op)
	if err != nil {
		blog.Error("fail to create upgrader server. err:%s", err.Error())
		return err
	}

	if err := common.SavePid(op.ProcessConfig); err != nil {
		blog.Warn("fail to save pid. err:%s", err.Error())
	}

	return upgrader.Start()
}

func setConfig(op *options.UpgraderOptions) {
	op.ServerCert.CertFile = op.ServerCertFile
	op.ServerCert.KeyFile = op.ServerKeyFile
	op.ServerCert.CAFile = op.CAFile

	if op.ServerCert.CertFile != "" && op.ServerCert.KeyFile != "" {
		op.ServerCert.IsSSL = true
	}
}
