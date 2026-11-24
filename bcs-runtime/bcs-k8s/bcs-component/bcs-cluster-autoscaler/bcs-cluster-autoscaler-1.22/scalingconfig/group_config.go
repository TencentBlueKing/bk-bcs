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

package scalingconfig

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
	klog "k8s.io/klog/v2"
)

// NodeGroupSet 定义一组具有相同 buffered ratio 配置的 NodeGroup
type NodeGroupSet struct {
	// SetName 组名称,用于标识和日志输出
	SetName string `json:"setName" yaml:"setName"`

	// NodeGroupIDs 该组包含的 NodeGroup ID 列表(显式配置)
	NodeGroupIDs []string `json:"nodeGroupIDs" yaml:"nodeGroupIDs"`

	// BufferedRatios 该组的缓冲资源比例配置
	BufferedRatios BufferedRatioConfig `json:"bufferedRatios" yaml:"bufferedRatios"`

	// Expander Set级别的扩容策略选择器
	// 可选值: "random", "most-pods", "least-waste", "price", "priority", "grpc"
	// 如果为空,则使用全局 expander 配置
	Expander string `json:"expander,omitempty" yaml:"expander,omitempty"`

	// ExpanderPriorities Priority expander 的优先级配置
	// 仅当 Expander="priority" 时生效
	// 列表中靠前的 NodeGroup 优先级更高
	ExpanderPriorities []string `json:"expanderPriorities,omitempty" yaml:"expanderPriorities,omitempty"`
}

// BufferedRatioConfig 缓冲资源比例配置
type BufferedRatioConfig struct {
	// CPURatio CPU 缓冲比例 (0.0-1.0)
	// 当 (CPU可用/CPU总量) < CPURatio 时触发扩容
	CPURatio float64 `json:"cpuRatio" yaml:"cpuRatio"`

	// MemRatio 内存缓冲比例 (0.0-1.0)
	// 当 (内存可用/内存总量) < MemRatio 时触发扩容
	MemRatio float64 `json:"memRatio" yaml:"memRatio"`

	// ResourceRatio 其他资源的缓冲比例 (0.0-1.0)
	// 应用于 CPU/Memory 之外的资源(如 GPU、存储等)
	ResourceRatio float64 `json:"resourceRatio" yaml:"resourceRatio"`
}

// NodeGroupSetList 包含所有 NodeGroupSet 的列表
type NodeGroupSetList struct {
	// Sets 所有的 NodeGroupSet
	Sets []NodeGroupSet `json:"sets" yaml:"sets"`
}

// LoadNodeGroupSetConfig 从 YAML 文件加载 NodeGroupSet 配置
func LoadNodeGroupSetConfig(configFile string) (*NodeGroupSetList, error) {
	if configFile == "" {
		klog.V(1).Info("NodeGroupSet config file not specified, NodeGroupSet feature disabled")
		return nil, nil
	}

	// 读取文件内容
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read NodeGroupSet config file %s: %v", configFile, err)
	}

	// 解析 YAML
	var config NodeGroupSetList
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse NodeGroupSet config file %s: %v", configFile, err)
	}

	// 验证配置
	if err := validateNodeGroupSetConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid NodeGroupSet config: %v", err)
	}

	klog.V(1).Infof("Successfully loaded NodeGroupSet config from %s: %d sets",
		configFile, len(config.Sets))
	return &config, nil
}

// validateNodeGroupSetConfig 验证 NodeGroupSet 配置的有效性
func validateNodeGroupSetConfig(config *NodeGroupSetList) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	// 验证每个 Set
	setNames := make(map[string]bool)
	nodeGroupIDs := make(map[string]string) // nodeGroupID -> setName

	for i, set := range config.Sets {
		// 检查 SetName 是否重复
		if set.SetName == "" {
			return fmt.Errorf("set[%d]: setName cannot be empty", i)
		}
		if setNames[set.SetName] {
			return fmt.Errorf("set[%d]: duplicate setName '%s'", i, set.SetName)
		}
		setNames[set.SetName] = true

		// 检查 NodeGroupIDs 是否重复
		for _, ngID := range set.NodeGroupIDs {
			if existingSet, exists := nodeGroupIDs[ngID]; exists {
				return fmt.Errorf("nodeGroup '%s' is assigned to multiple sets: '%s' and '%s'",
					ngID, existingSet, set.SetName)
			}
			nodeGroupIDs[ngID] = set.SetName
		}

		// 验证 BufferedRatios
		if err := validateBufferedRatios(&set.BufferedRatios, set.SetName); err != nil {
			return err
		}

		// 验证 Expander 配置
		if err := validateExpanderConfig(&set); err != nil {
			return err
		}
	}

	return nil
}

// validateBufferedRatios 验证缓冲比例的有效性
func validateBufferedRatios(ratios *BufferedRatioConfig, setName string) error {
	if ratios.CPURatio < 0 || ratios.CPURatio > 1 {
		return fmt.Errorf("set '%s': cpuRatio must be in range [0, 1], got %f", setName, ratios.CPURatio)
	}
	if ratios.MemRatio < 0 || ratios.MemRatio > 1 {
		return fmt.Errorf("set '%s': memRatio must be in range [0, 1], got %f", setName, ratios.MemRatio)
	}
	if ratios.ResourceRatio < 0 || ratios.ResourceRatio > 1 {
		return fmt.Errorf("set '%s': resourceRatio must be in range [0, 1], got %f", setName, ratios.ResourceRatio)
	}
	return nil
}

// validateExpanderConfig 验证 Expander 配置的有效性
func validateExpanderConfig(set *NodeGroupSet) error {
	if set.Expander == "" {
		// 未配置 expander,使用全局配置,无需验证
		return nil
	}

	// 验证 expander 类型是否支持
	validExpanders := map[string]bool{
		"random":      true,
		"most-pods":   true,
		"least-waste": true,
		"price":       true,
		"priority":    true,
		"grpc":        true,
	}

	if !validExpanders[set.Expander] {
		return fmt.Errorf("set '%s': invalid expander '%s', "+
			"must be one of: random, most-pods, least-waste, price, priority, grpc",
			set.SetName, set.Expander)
	}

	// 如果使用 priority expander,必须配置 ExpanderPriorities
	if set.Expander == "priority" {
		if len(set.ExpanderPriorities) == 0 {
			return fmt.Errorf("set '%s': expanderPriorities must be specified when using priority expander", set.SetName)
		}

		// 验证 ExpanderPriorities 中的 NodeGroupID 是否在该 Set 中
		nodeGroupMap := make(map[string]bool)
		for _, ngID := range set.NodeGroupIDs {
			nodeGroupMap[ngID] = true
		}

		for _, ngID := range set.ExpanderPriorities {
			if !nodeGroupMap[ngID] {
				klog.Warningf("Set '%s': expanderPriorities contains NodeGroup '%s' which is not in this Set's NodeGroupIDs",
					set.SetName, ngID)
			}
		}
	}

	return nil
}
