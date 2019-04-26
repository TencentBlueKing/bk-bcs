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

package processor

import (
	"bk-bcs/bcs-services/bcs-health/util"
	"fmt"
	"time"
)

const normalMsg string = "" +
	"服务异常告警\n" +
	"Module: %s\n" +
	"Type: %s\n" +
	"Cluster: %s\n" +
	"Message: %s\n" +
	"Time: %s"

const smsMsg string = "" +
	"Module: %s\n" +
	"Cluster: %s\n" +
	"Message: %s\n"

func formatNormalMsg(module, typer, cluster, msg string) string {
	return fmt.Sprintf(normalMsg, module, typer, cluster, msg, time.Now().Format(util.TimeFormat))
}

func formatSmsMsg(module, cluster, msg string) string {
	return fmt.Sprintf(smsMsg, module, cluster, msg)
}
