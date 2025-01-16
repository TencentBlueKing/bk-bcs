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

// Package selectui xxx
package selectui

import (
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
)

// SelectUI output select ui
func SelectUI(kubeItems []Needle, label string) (int, error) {
	s, err := selectUIRunner(kubeItems, label, nil)
	if err != nil {
		if err.Error() == "exit" {
			os.Exit(0)
		}
		return 0, errors.Wrapf(err, "prompt failed")
	}
	return s, nil
}

// Needle use for switch
type Needle struct {
	Name      string `json:"name"`
	ClusterID string `json:"clusterID"`
	Project   string `json:"project"`
}

// SelectRunner interface - For better unit testing
type SelectRunner interface {
	Run() (int, string, error)
}

// selectUIRunner
func selectUIRunner(kubeItems []Needle, label string, runner SelectRunner) (int, error) {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   ">> {{ .ClusterID | yellow }}\t{{ .Name | yellow}}",
		Inactive: "   {{ .ClusterID | cyan }}\t{{ .Name | cyan}}",
		// Selected: "\U0001F638 Select:{{ .Name | green }}",
		Details: `
--------- Info ----------
{{ "Name:" | faint }}	{{ .Name }}
{{ "Cluster:" | faint }}	{{ .ClusterID }}
{{ "Project:" | faint }}	{{ .Project }}`,
	}
	searcher := func(input string, index int) bool {
		pepper := kubeItems[index]
		input = strings.Replace(strings.ToLower(input), " ", "", -1)
		if input == "q" {
			return true
		}
		name := strings.Replace(strings.ToLower(pepper.Name), " ", "", -1)
		cluster := strings.Replace(strings.ToLower(pepper.ClusterID), " ", "", -1)
		project := strings.Replace(strings.ToLower(pepper.Project), " ", "", -1)
		return strings.Contains(name, input) || strings.Contains(cluster, input) || strings.Contains(project, input)
	}
	prompt := promptui.Select{
		Label:     label,
		Items:     kubeItems,
		Templates: templates,
		Size:      20,
		Searcher:  searcher,
	}
	if runner == nil {
		runner = &prompt
	}
	i, _, err := runner.Run()
	if err != nil {
		return 0, err
	}
	if kubeItems[i].Name == "<Exit>" {
		return 0, errors.New("exit")
	}
	return i, err
}
