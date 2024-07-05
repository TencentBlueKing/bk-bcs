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

// Package bkdevops xxx
package bkdevops

// createResp defines the bkdevops create response
type createResp struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    struct {
		ID string `json:"id"`
	} `json:"data"`
}

type getResp struct {
	Status int       `json:"status"`
	Data   *pipeline `json:"data"`
}

// createResp update or delete response
type updateOrDeleteResp struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    bool   `json:"data"`
}

// executeResp execute workflow response
type executeResp struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    struct {
		ID           string `json:"id"`
		ExecuteCount int    `json:"executeCount"`
		ProjectID    string `json:"projectId"`
		PipelineID   string `json:"pipelineId"`
		Num          int64  `json:"num"`
	} `json:"data"`
}

// executeStatusResp query the execute status response
type executeStatusResp struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Status    string `json:"status"`
		StartTime int64  `json:"startTime"`
		EndTime   int64  `json:"endTime"`
		Model     struct {
			Stages []struct {
				Containers []struct {
					Elements []struct {
						Status string `json:"status"`
					} `json:"elements"`
					Status string `json:"status"`
				} `json:"containers"`
				Status string `json:"status"`
			} `json:"stages"`
		} `json:"model"`
	} `json:"data"`
}

// pipeline defines the pipeline object
type pipeline struct {
	Name   string           `json:"name,omitempty"`
	Desc   string           `json:"desc,omitempty"`
	Stages []*pipelineStage `json:"stages,omitempty"`
}

// pipelineStage defines the pipeline stage object
type pipelineStage struct {
	Containers         []*container   `json:"containers,omitempty"`
	Name               string         `json:"name,omitempty"`
	FastKill           bool           `json:"fastKill,omitempty"`
	Finally            bool           `json:"finally,omitempty"`
	StageEnable        bool           `json:"stageEnable,omitempty"`
	StageControlOption *controlOption `json:"stageControlOption,omitempty"`
	CheckIn            *checkIn       `json:"checkIn,omitempty"`

	// only use for root stage
	IsTrigger bool `json:"isTrigger,omitempty"`
}

type jobType string

const (
	vmBuildJobType jobType = "vmBuild"
	normalJobType  jobType = "normal"
	triggerJobType jobType = "trigger"
)

type container struct {
	Type                jobType              `json:"@type,omitempty"`
	Name                string               `json:"name,omitempty"`
	ContainerEnable     bool                 `json:"containerEnable,omitempty"`
	BaseOS              string               `json:"baseOS,omitempty"`
	DispatchType        *dispatchType        `json:"dispatchType,omitempty"`
	JobControlOption    *controlOption       `json:"jobControlOption,omitempty"`
	MatrixGroupFlag     bool                 `json:"matrixGroupFlag,omitempty"`
	MatrixControlOption *matrixControlOption `json:"matrixControlOption,omitempty"`

	// only use for root job
	Params []customParam `json:"params,omitempty"`

	Elements []interface{} `json:"elements,omitempty"`
}

type dispatchType struct {
	BuildType    string `json:"buildType,omitempty"`
	Value        string `json:"value,omitempty"`
	ImageType    string `json:"imageType,omitempty"`
	ImageCode    string `json:"imageCode,omitempty"`
	ImageVersion string `json:"imageVersion,omitempty"`
}

type elementType string

const (
	linuxScript     elementType = "linuxScript"
	marketBuild     elementType = "marketBuild"
	marketBuildLess elementType = "marketBuildLess"

	manualTrigger elementType = "manualTrigger"
	remoteTrigger elementType = "remoteTrigger"
)

type elementManualTrigger struct {
	Type           elementType `json:"@type,omitempty"`
	Name           string      `json:"name,omitempty"`
	CanElementSkip bool        `json:"canElementSkip,omitempty"`
	AtomCode       string      `json:"atomCode,omitempty"`
	Version        string      `json:"version,omitempty"`
	ElementEnable  bool        `json:"elementEnable,omitempty"`
	ClassType      string      `json:"classType,omitempty"`
}

type elementRemoteTrigger struct {
	Type           elementType `json:"@type,omitempty"`
	Name           string      `json:"name,omitempty"`
	CanElementSkip bool        `json:"canElementSkip,omitempty"`
	RemoteToken    string      `json:"remoteToken,omitempty"`
	Version        string      `json:"version,omitempty"`
	ElementEnable  bool        `json:"elementEnable,omitempty"`
	ClassType      string      `json:"classType,omitempty"`
	AtomCode       string      `json:"atomCode,omitempty"`
}

type elementLinuxScript struct {
	Type              elementType    `json:"@type,omitempty"`
	Name              string         `json:"name,omitempty"`
	ElementEnable     bool           `json:"elementEnable,omitempty"`
	Version           string         `json:"version,omitempty"`
	AtomCode          string         `json:"atomCode,omitempty"`
	ClassType         string         `json:"classType,omitempty"`
	ScriptType        string         `json:"scriptType,omitempty"`
	Script            string         `json:"script,omitempty"`
	ContinueNoneZero  bool           `json:"continueNoneZero,omitempty"`
	AdditionalOptions *controlOption `json:"additionalOptions,omitempty"`
}

type elementMarketBuild struct {
	Type          elementType `json:"@type,omitempty"`
	Name          string      `json:"name,omitempty"`
	ElementEnable bool        `json:"elementEnable,omitempty"`
	Version       string      `json:"version,omitempty"`
	AtomCode      string      `json:"atomCode,omitempty"`
	ClassType     string      `json:"classType,omitempty"`
	ExecuteCount  int64       `json:"executeCount,omitempty"`
	Data          struct {
		Input map[string]string `json:"input,omitempty"`
	} `json:"data"`
	AdditionalOptions *controlOption `json:"additionalOptions,omitempty"`
}

type matrixControlOption struct {
	StrategyStr    string `json:"strategyStr,omitempty"`
	IncludeCaseStr string `json:"includeCaseStr,omitempty"`
	ExcludeCaseStr string `json:"excludeCaseStr,omitempty"`
	FastKill       bool   `json:"fastKill,omitempty"`
	MaxConcurrency int64  `json:"maxConcurrency,omitempty"`
}

type customParam struct {
	ID           string `json:"id,omitempty"`
	Required     bool   `json:"required,omitempty"`
	Type         string `json:"type,omitempty"`
	DefaultValue string `json:"defaultValue,omitempty"`
	Desc         string `json:"desc,omitempty"`
	ReadOnly     bool   `json:"readOnly,omitempty"`
}

type conditionType string

const (
	stageDfaultConditionType = "AFTER_LAST_FINISHED"
	jobDefaultConditionType  = "STAGE_RUNNING"

	preTaskSuccess      conditionType = "PRE_TASK_SUCCESS"
	customVariableMatch conditionType = "CUSTOM_VARIABLE_MATCH"
)

type controlOption struct {
	Enable             bool          `json:"enable,omitempty"`
	Timeout            int64         `json:"timeout,omitempty"`
	RunCondition       conditionType `json:"runCondition,omitempty"`
	CustomVariables    []variable    `json:"customVariables,omitempty"`
	ContinueWhenFailed bool          `json:"continueWhenFailed,omitempty"`
}

type checkIn struct {
	ManualTrigger bool          `json:"manualTrigger,omitempty"`
	NotifyType    []string      `json:"notifyType,omitempty"`
	ReviewGroups  []reviewGroup `json:"reviewGroups,omitempty"`
	NotifyGroup   []string      `json:"notifyGroup,omitempty"`
}

type reviewGroup struct {
	Name      string   `json:"name,omitempty"`
	Reviewers []string `json:"reviewers,omitempty"`
}

type variable struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}
