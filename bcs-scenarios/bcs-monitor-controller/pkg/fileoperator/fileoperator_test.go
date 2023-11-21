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

package fileoperator

import (
	"testing"

	v1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/api/v1"
)

func TestCompress(t *testing.T) {
	ng := v1.NoticeGroup{}
	ng.Spec.Groups = append(ng.Spec.Groups, v1.NoticeGroupDetail{})
	ng.Spec.Groups[0].Name = "porterlin-test-gen"
	ng.Spec.Groups[0].Users = []string{"porterlin"}
	ng.Spec.Groups[0].Alert = make(map[string]v1.NoticeAlert)
	ng.Spec.Groups[0].Alert["00:00--23:59"] = v1.NoticeAlert{
		Fatal: v1.NoticeType{
			Type: []string{"rtx"},
		},
		Remind: v1.NoticeType{
			Type: []string{"rtx"},
		},
		Warning: v1.NoticeType{
			Type: []string{"rtx"},
		},
	}

	ng.Spec.Groups[0].Action = make(map[string]v1.NoticeAction)
	ng.Spec.Groups[0].Action["00:00--23:59"] = v1.NoticeAction{
		Execute: v1.NoticeType{
			Type: []string{"rtx"},
		},
		ExecuteFailed: v1.NoticeType{
			Type: []string{"sms"},
		},
		ExecuteSuccess: v1.NoticeType{
			Type: []string{"mail"},
		},
	}

	fo := &FileOperator{}
	outpath, err := fo.Compress(ng.Spec)
	if err != nil {
		println(err.Error())
	}
	println("outpath:" + outpath)

}
