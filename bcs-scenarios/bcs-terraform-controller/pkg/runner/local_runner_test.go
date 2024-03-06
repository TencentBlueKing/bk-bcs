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
package runner

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	tfv1 "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/api/v1"
)

func MockTerraform() *tfv1.Terraform {
	return &tfv1.Terraform{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-tf",
			Namespace: "test-ns",
		},
		Spec: tfv1.TerraformSpec{
			ApprovePlan: "",
			Destroy:     false,
		},
		Status: tfv1.TerraformStatus{},
	}
}

func TestPlan(t *testing.T) {
	//tfRunner := terraformLocalRunner{
	//	execPath: "/opt/homebrew/bin/terraform",
	//}
	//ctx := context.Background()
	//instanceID := uuid.New().String()
	//mockTf := MockTerraform()
	//_, err := tfRunner.NewTerraform(ctx, &NewTerraformRequest{
	//	WorkingDir: "/Users/zup779/code/my_code/terraform-store",
	//	Terraform:  *mockTf,
	//	InstanceID: instanceID,
	//})
	//if err != nil {
	//	panic(err)
	//}
	//
	//err = tfRunner.exec.Init(ctx)
	//if err != nil {
	//	panic(err)
	//}
	//
	//reply, err := tfRunner.Plan(ctx, &PlanRequest{
	//	TfInstance: instanceID,
	//	Out:        "drift",
	//	Refresh:    false,
	//	Destroy:    false,
	//	Targets:    nil,
	//})
	//if err != nil {
	//	panic(err)
	//}
	//
	//log.Printf("reply: %s", utils.ToJsonString(reply))

}
