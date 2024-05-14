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

package render

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	monitorextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/api/v1"
)

const (
	// GenerateMonitorRuleName xxx
	GenerateMonitorRuleName = "auto-generate-monitor-rule"
	// GeneratePanelName xxx
	GeneratePanelName = "auto-generate-monitor-panel"
	// GenerateNoticeGroupName xxx
	GenerateNoticeGroupName = "auto-generate-monitor-notice-group"
	// GenerateAppendNoticeGroupName xxx
	GenerateAppendNoticeGroupName = "auto-generate-monitor-append-ng"
)

// ReadScenario load scenario info from directory, and transfer to Result
// scenario支持两种配置方式 （允许同时存在）
// 1. 蓝鲸监控导出 （通过文件夹名称区分不同YAML文件， 例如grafana下认为是告警面板配置）
// 2. cr模式配置 （通过文件名称区分，例如monitorrule&mr开头文件认为是告警规则配置）
// nolint funlen
func (r *MonitorRender) ReadScenario(repoKey, scenario string) (*Result, error) {
	res := &Result{}
	repo, ok := r.repoManager.GetRepo(repoKey)
	if !ok {
		return nil, fmt.Errorf("repo[%s] not found", repoKey)
	}

	// 配置2 读取（cr模式配置）
	err := filepath.Walk(filepath.Join(repo.GetDirectory(), scenario), func(path string, info os.FileInfo,
		err error) error {
		if err != nil {
			blog.Errorf("walk through directory'%s' failed, err: %s", filepath.Join(repo.GetDirectory(), scenario),
				scenario)
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".yaml") {
			data, inErr := ioutil.ReadFile(path)
			if inErr != nil {
				blog.Errorf("read file'%s' failed, err: %v", path, inErr)
				return inErr
			}

			switch {
			case strings.HasPrefix(info.Name(), "monitorrule") || strings.HasPrefix(info.Name(), "mr"):
				blog.Infof("scenario '%s' got monitor rule: %s", scenario, info.Name())
				var monitorRule monitorextensionv1.MonitorRule
				inErr = runtime.DecodeInto(r.decoder, data, &monitorRule)
				if inErr != nil {
					blog.Errorf("unmarshal yaml file %s failed: %v", path, inErr)
					return inErr
				}

				res.MonitorRule = append(res.MonitorRule, &monitorRule)
			case strings.HasPrefix(info.Name(), "noticegroup") || strings.HasPrefix(info.Name(), "ng"):
				blog.Infof("scenario '%s' got notice group: %s", scenario, info.Name())
				var noticeGroup monitorextensionv1.NoticeGroup
				inErr = runtime.DecodeInto(r.decoder, data, &noticeGroup)
				if inErr != nil {
					blog.Errorf("unmarshal yaml file %s failed: %v", path, inErr)
					return inErr
				}

				res.NoticeGroup = append(res.NoticeGroup, &noticeGroup)
			case strings.HasPrefix(info.Name(), "panel") || strings.HasPrefix(info.Name(), "p"):
				blog.Infof("scenario '%s' got panel: %s", scenario, info.Name())
				var panel monitorextensionv1.Panel
				inErr = runtime.DecodeInto(r.decoder, data, &panel)
				if inErr != nil {
					blog.Errorf("unmarshal yaml file %s failed: %v", path, inErr)
					return inErr
				}

				res.Panel = append(res.Panel, &panel)
			case strings.HasPrefix(info.Name(), "configmap") || strings.HasPrefix(info.Name(), "cm"):
				blog.Infof("scenario '%s' got configmap: %s", scenario, info.Name())
				var configmap v1.ConfigMap
				inErr = runtime.DecodeInto(r.decoder, data, &configmap)
				if inErr != nil {
					blog.Errorf("unmarshal yaml file %s failed: %v", path, inErr)
					return inErr
				}

				res.ConfigMaps = append(res.ConfigMaps, &configmap)
			}
		}
		return nil
	})
	if err != nil {
		blog.Errorf("read scenario failed: %v", err)
		return nil, err
	}

	// 配置1读取 （蓝鲸监控导出）
	configMaps, panel, err := r.loadPanel(filepath.Join(repo.GetDirectory(), scenario, "grafana"))
	if err != nil {
		return nil, err
	}
	if len(configMaps) != 0 && panel != nil {
		res.ConfigMaps = append(res.ConfigMaps, configMaps...)
		res.Panel = append(res.Panel, panel)
	}

	mr, err := r.LoadRule(filepath.Join(repo.GetDirectory(), scenario, "rule"))
	if err != nil {
		return nil, err
	}
	if mr != nil {
		res.MonitorRule = append(res.MonitorRule, mr)
	}

	ng, err := r.loadNoticeGroup(filepath.Join(repo.GetDirectory(), scenario, "notice"))
	if err != nil {
		return nil, err
	}
	if ng != nil {
		res.NoticeGroup = append(res.NoticeGroup, ng)
	}

	return res, nil
}

// LoadRule load rule from certain path, filterFunc 根据告警名判断
func (r *MonitorRender) LoadRule(
	path string, filterFunc ...func(string) bool) (*monitorextensionv1.MonitorRule, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		blog.Infof("empty raw rule, continue...")
		return nil, nil
	}
	rules := make([]*monitorextensionv1.MonitorRuleDetail, 0)

	err := filepath.Walk(path, func(path string, info os.FileInfo,
		err error) error {
		if err != nil {
			return fmt.Errorf("walk through directory'%s' failed, err: %s", path, err.Error())
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".yaml") {
			for _, filter := range filterFunc {
				if !filter(info.Name()) {
					return nil
				}
			}
			data, inErr := ioutil.ReadFile(path)
			if inErr != nil {
				return fmt.Errorf("read file'%s' failed, err: %s", path, inErr.Error())
			}

			var rule monitorextensionv1.MonitorRuleDetail
			if inErr = yaml.Unmarshal(data, &rule); inErr != nil {
				return fmt.Errorf("unmarshal rule in'%s' failed, err: %s", info.Name(), inErr.Error())
			}
			rules = append(rules, &rule)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	mr := &monitorextensionv1.MonitorRule{
		ObjectMeta: metav1.ObjectMeta{
			Name: GenerateMonitorRuleName,
		},
		Spec: monitorextensionv1.MonitorRuleSpec{
			Rules: rules,
		},
	}
	return mr, nil
}

func (r *MonitorRender) loadPanel(path string) ([]*v1.ConfigMap, *monitorextensionv1.Panel, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		blog.Infof("empty raw grafana panel, continue...")
		return nil, nil, nil
	}
	configmaps := make([]*v1.ConfigMap, 0)
	dashBoards := make([]monitorextensionv1.DashBoardConfig, 0)

	err := filepath.Walk(path, func(path string, info os.FileInfo,
		err error) error {
		if err != nil {
			blog.Errorf("walk through directory'%s' failed, err: %s", path, err.Error())
			return err
		}

		// load grafana panel json config
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".json") {
			data, inErr := ioutil.ReadFile(path)
			if inErr != nil {
				return fmt.Errorf("read file'%s' failed, err: %s", path, inErr.Error())
			}
			dataMap := make(map[string]string)
			dataMap["panel"] = string(data)

			splits := strings.Split(info.Name(), ".")
			if len(splits) != 2 {
				return fmt.Errorf("invalid panel Name[%s], should be 'xxx.json'", info.Name())
			}

			name := splits[0]
			configmap := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Data: dataMap,
			}
			configmaps = append(configmaps, configmap)
			dashBoards = append(dashBoards, monitorextensionv1.DashBoardConfig{
				Board:     name,
				Render:    true,
				ConfigMap: name,
			})
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	panel := &monitorextensionv1.Panel{
		ObjectMeta: metav1.ObjectMeta{
			Name: GeneratePanelName,
		},
		Spec: monitorextensionv1.PanelSpec{
			DashBoard: dashBoards,
		},
	}
	return configmaps, panel, nil
}

// LoadDashBoard return DashBoard info, use filterFunc to filter file name
func (r *MonitorRender) LoadDashBoard(path string, filterFunc ...func(string) bool) ([]*DashBoard, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		blog.Infof("empty raw grafana panel, continue...")
		return nil, nil
	}
	panels := make([]*DashBoard, 0)

	err := filepath.Walk(path, func(path string, info os.FileInfo,
		err error) error {
		if err != nil {
			blog.Errorf("walk through directory'%s' failed, err: %s", path, err.Error())
			return err
		}

		// load grafana panel json config
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".json") {
			for _, filter := range filterFunc {
				if !filter(info.Name()) {
					return nil
				}
			}
			data, inErr := ioutil.ReadFile(path)
			if inErr != nil {
				return fmt.Errorf("read file'%s' failed, err: %s", path, inErr.Error())
			}
			var panel DashBoard
			if inErr = yaml.Unmarshal(data, &panel); inErr != nil {
				return fmt.Errorf("unmarshal panel in'%s' failed, err: %s", info.Name(), inErr.Error())
			}
			panels = append(panels, &panel)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return panels, nil
}

func (r *MonitorRender) loadNoticeGroup(path string) (*monitorextensionv1.NoticeGroup, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		blog.Infof("empty raw notice , continue...")
		return nil, nil
	}
	ngs := make([]*monitorextensionv1.NoticeGroupDetail, 0)

	err := filepath.Walk(path, func(path string, info os.FileInfo,
		err error) error {
		if err != nil {
			return fmt.Errorf("walk through directory'%s' failed, err: %s", path, err.Error())
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".yaml") {
			data, inErr := os.ReadFile(path)
			if inErr != nil {
				blog.Errorf("read file'%s' failed, err: %v", path, inErr)
				return inErr
			}

			var ng monitorextensionv1.NoticeGroupDetail
			if inErr = yaml.Unmarshal(data, &ng); inErr != nil {
				return fmt.Errorf("unmarshal notice in'%s' failed, err: %s", info.Name(), inErr.Error())
			}
			ngs = append(ngs, &ng)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	ng := &monitorextensionv1.NoticeGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name: GenerateNoticeGroupName,
		},
		Spec: monitorextensionv1.NoticeGroupSpec{
			Groups: ngs,
		},
	}

	return ng, nil
}

func genName(bizID, scenario, name string) string {
	return fmt.Sprintf("%s-%s-%s", name, scenario, bizID)
}
