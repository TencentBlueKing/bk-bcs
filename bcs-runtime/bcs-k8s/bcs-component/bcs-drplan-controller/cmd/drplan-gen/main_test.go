/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"strings"
	"testing"
)

func TestHelmfileCommand_RequiresChartRepo(t *testing.T) {
	root := newRootCommand()
	root.SetArgs([]string{"helmfile", "-f", "helmfile.yaml", "-l", "name=demo"})

	err := root.Execute()
	if err == nil || !strings.Contains(err.Error(), "required flag \"chart-repo\" not set") {
		t.Fatalf("expected missing chart-repo error, got %v", err)
	}
}
