/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package formatter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util"
)

var lightDeploySpec = map[string]interface{}{
	"template": map[string]interface{}{
		"spec": map[string]interface{}{
			"containers": []interface{}{
				map[string]interface{}{"image": "nginx:1.14.2"},
				map[string]interface{}{"image": "nginx:latest"},
				map[string]interface{}{"image": "nginx:1.14.2"},
			},
		},
	},
}

var lightCronJobSpec = map[string]interface{}{
	"jobTemplate": map[string]interface{}{
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{"image": "busybox:1.0.0"},
						map[string]interface{}{"image": "busybox:1.0.0"},
						map[string]interface{}{"image": "busybox:1.2.3"},
						map[string]interface{}{"image": "busybox:latest"},
					},
				},
			},
		},
	},
}

func TestParseContainerImages(t *testing.T) {
	images := parseContainerImages(lightDeploySpec, "template.spec.containers")
	assert.Equal(t, 2, len(images))

	images = parseContainerImages(lightCronJobSpec, "jobTemplate.spec.template.spec.containers")
	assert.Equal(t, 3, len(images))
}

var failedPodManifest = map[string]interface{}{
	"status": map[string]interface{}{
		"phase": "Failed",
		"conditions": []map[string]interface{}{
			{
				"type":   "Initialized",
				"status": "True",
			},
		},
	},
}

var succeededPodManifest = map[string]interface{}{
	"status": map[string]interface{}{
		"phase": "Succeeded",
		"conditions": []map[string]interface{}{
			{
				"type":   "Initialized",
				"status": "True",
			},
		},
	},
}

var runningPodManifest1 = map[string]interface{}{
	"metadata": map[string]interface{}{
		"creationTimestamp": "2022-01-01T10:00:00Z",
	},
	"spec": map[string]interface{}{
		"containers": []interface{}{
			map[string]interface{}{
				"image": "busybox",
			},
		},
	},
	"status": map[string]interface{}{
		"phase": "Running",
		"conditions": []map[string]interface{}{
			{
				"type":   "Initialized",
				"status": "True",
			},
			{
				"type":   "Ready",
				"status": "True",
			},
		},
		"containerStatuses": []interface{}{
			map[string]interface{}{
				"ready":        true,
				"restartCount": int64(2),
			},
		},
	},
}

var pendingPodManifest = map[string]interface{}{
	"status": map[string]interface{}{
		"phase": "Pending",
		"conditions": []map[string]interface{}{
			{
				"type":   "Initialized",
				"status": "False",
			},
		},
	},
}

var terminatingPodManifest = map[string]interface{}{
	"metadata": map[string]interface{}{
		"deletionTimestamp": "2021-01-01T10:00:00Z",
	},
	"status": map[string]interface{}{
		"phase": "Running",
	},
}

var unknownPodManifest1 = map[string]interface{}{
	"metadata": map[string]interface{}{
		"deletionTimestamp": "2021-01-01T10:00:00Z",
	},
	"status": map[string]interface{}{
		"phase":  "Running",
		"reason": "NodeLost",
	},
}

var completedPodManifest = map[string]interface{}{
	"status": map[string]interface{}{
		"phase": "Succeeded",
		"containerStatuses": []interface{}{
			map[string]interface{}{
				"state": map[string]interface{}{
					"terminated": map[string]interface{}{
						"reason": "Completed",
					},
				},
			},
		},
	},
}

var waitReasonPodManifest = map[string]interface{}{
	"status": map[string]interface{}{
		"phase": "Pending",
		"containerStatuses": []interface{}{
			map[string]interface{}{
				"state": map[string]interface{}{
					"waiting": map[string]interface{}{
						"message": "Error response from daemon: No command specified",
						"reason":  "CreateContainerError",
					},
				},
			},
		},
	},
}

var signalPodManifest = map[string]interface{}{
	"status": map[string]interface{}{
		"phase": "Failed",
		"containerStatuses": []interface{}{
			map[string]interface{}{
				"state": map[string]interface{}{
					"terminated": map[string]interface{}{
						"signal": 1,
					},
				},
			},
		},
	},
}

var exitCodePodManifest = map[string]interface{}{
	"status": map[string]interface{}{
		"phase": "Failed",
		"containerStatuses": []interface{}{
			map[string]interface{}{
				"state": map[string]interface{}{
					"terminated": map[string]interface{}{
						"ExitCode": 1,
					},
				},
			},
		},
	},
}

var runningPodManifest2 = map[string]interface{}{
	"status": map[string]interface{}{
		"phase": "Completed",
		"conditions": []map[string]interface{}{
			{
				"type":   "Ready",
				"status": "True",
			},
		},
		"containerStatuses": []interface{}{
			map[string]interface{}{
				"ready": true,
				"state": map[string]interface{}{
					"running": map[string]interface{}{
						"startAt": "2022-01-01T10:00:00Z",
					},
				},
			},
		},
	},
}

var notReadyPodManifest = map[string]interface{}{
	"status": map[string]interface{}{
		"phase": "Completed",
		"conditions": []map[string]interface{}{
			{
				"type":   "Ready",
				"status": "False",
			},
		},
		"containerStatuses": []interface{}{
			map[string]interface{}{
				"ready": true,
				"state": map[string]interface{}{
					"running": map[string]interface{}{
						"startAt": "2022-01-01T10:00:00Z",
					},
				},
			},
		},
	},
}

var unknownPodManifest2 = map[string]interface{}{
	"status": map[string]interface{}{
		"phase": nil,
	},
}

var initSignalPodManifest = map[string]interface{}{
	"status": map[string]interface{}{
		"phase": "Pending",
		"initContainerStatuses": []interface{}{
			map[string]interface{}{
				"state": map[string]interface{}{
					"terminated": map[string]interface{}{
						"exitCode": 0,
					},
				},
			},
			map[string]interface{}{
				"state": map[string]interface{}{
					"terminated": map[string]interface{}{
						"exitCode": 1,
						"signal":   1,
					},
				},
			},
		},
	},
}

var initExitCodePodManifest = map[string]interface{}{
	"status": map[string]interface{}{
		"phase": "Pending",
		"initContainerStatuses": []interface{}{
			map[string]interface{}{
				"state": map[string]interface{}{
					"terminated": map[string]interface{}{
						"exitCode": 1,
					},
				},
			},
		},
	},
}

var initTermPodManifest = map[string]interface{}{
	"status": map[string]interface{}{
		"phase": "Pending",
		"initContainerStatuses": []interface{}{
			map[string]interface{}{
				"state": map[string]interface{}{
					"terminated": map[string]interface{}{
						"exitCode": 1,
						"reason":   "term init",
					},
				},
			},
		},
	},
}

var initWaitPodManifest = map[string]interface{}{
	"status": map[string]interface{}{
		"phase": "Pending",
		"initContainerStatuses": []interface{}{
			map[string]interface{}{
				"state": map[string]interface{}{
					"waiting": map[string]interface{}{
						"reason": "wait init",
					},
				},
			},
		},
	},
}

var initRunPodManifest = map[string]interface{}{
	"spec": map[string]interface{}{
		"initContainers": []interface{}{
			map[string]interface{}{
				"name": "nginx",
			},
		},
	},
	"status": map[string]interface{}{
		"phase": "Pending",
		"initContainerStatuses": []interface{}{
			map[string]interface{}{
				"state": map[string]interface{}{
					"running": map[string]interface{}{
						"startAt": "2022-01-01T10:00:00Z",
					},
				},
			},
		},
	},
}

func TestPodStatusParser(t *testing.T) {
	parser := PodStatusParser{Manifest: failedPodManifest}
	assert.Equal(t, "Failed", parser.Parse())

	parser = PodStatusParser{Manifest: succeededPodManifest}
	assert.Equal(t, "Succeeded", parser.Parse())

	parser = PodStatusParser{Manifest: runningPodManifest1}
	assert.Equal(t, "Running", parser.Parse())

	parser = PodStatusParser{Manifest: pendingPodManifest}
	assert.Equal(t, "Pending", parser.Parse())

	parser = PodStatusParser{Manifest: terminatingPodManifest}
	assert.Equal(t, "Terminating", parser.Parse())

	parser = PodStatusParser{Manifest: unknownPodManifest1}
	assert.Equal(t, "Unknown", parser.Parse())

	parser = PodStatusParser{Manifest: completedPodManifest}
	assert.Equal(t, "Completed", parser.Parse())

	parser = PodStatusParser{Manifest: waitReasonPodManifest}
	assert.Equal(t, "CreateContainerError", parser.Parse())

	parser = PodStatusParser{Manifest: signalPodManifest}
	assert.Equal(t, "Signal: 1", parser.Parse())

	parser = PodStatusParser{Manifest: exitCodePodManifest}
	assert.Equal(t, "ExitCode: 1", parser.Parse())

	parser = PodStatusParser{Manifest: runningPodManifest2}
	assert.Equal(t, "Running", parser.Parse())

	parser = PodStatusParser{Manifest: notReadyPodManifest}
	assert.Equal(t, "NotReady", parser.Parse())

	parser = PodStatusParser{Manifest: unknownPodManifest2}
	assert.Equal(t, "Unknown", parser.Parse())

	parser = PodStatusParser{Manifest: initSignalPodManifest}
	assert.Equal(t, "Init: Signal 1", parser.Parse())

	parser = PodStatusParser{Manifest: initExitCodePodManifest}
	assert.Equal(t, "Init: ExitCode 1", parser.Parse())

	parser = PodStatusParser{Manifest: initTermPodManifest}
	assert.Equal(t, "Init: term init", parser.Parse())

	parser = PodStatusParser{Manifest: initWaitPodManifest}
	assert.Equal(t, "Init: wait init", parser.Parse())

	parser = PodStatusParser{Manifest: initRunPodManifest}
	assert.Equal(t, "Init: 0/1", parser.Parse())
}

var lightDeployManifest = map[string]interface{}{
	"metadata": map[string]interface{}{
		"creationTimestamp": "2022-01-01T10:00:00Z",
	},
	"spec": lightDeploySpec,
}

func TestFormatWorkloadRes(t *testing.T) {
	images := FormatWorkloadRes(lightDeployManifest)["images"]
	assert.Equal(t, 2, len(images.([]string)))
}

var lightCJManifest = map[string]interface{}{
	"metadata": map[string]interface{}{
		"creationTimestamp": "2022-01-01T10:00:00Z",
	},
	"spec": lightCronJobSpec,
	"status": map[string]interface{}{
		"lastScheduleTime": "2022-02-02T08:00:00Z",
	},
}

func TestFormatCJ(t *testing.T) {
	ret := FormatCJ(lightCJManifest)
	assert.Equal(t, 3, len(ret["images"].([]string)))
	assert.Equal(t, 0, ret["active"])
	assert.Equal(t, util.CalcDuration("2022-02-02 08:00:00", ""), ret["lastSchedule"])
}

var lightJobManifest = map[string]interface{}{
	"metadata": map[string]interface{}{
		"creationTimestamp": "2022-01-01T10:00:00Z",
	},
	"spec": map[string]interface{}{
		"template": map[string]interface{}{
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"image": "perl",
					},
				},
			},
		},
	},
	"status": map[string]interface{}{
		"completionTime": "2022-01-01T12:33:35Z",
		"startTime":      "2022-01-01T12:30:00Z",
	},
}

func TestFormatJob(t *testing.T) {
	ret := FormatJob(lightJobManifest)
	assert.Equal(t, []string{"perl"}, ret["images"])
	assert.Equal(t, "3m35s", ret["duration"])
}

func TestFormatPo(t *testing.T) {
	ret := FormatPo(runningPodManifest1)
	assert.Equal(t, []string{"busybox"}, ret["images"])
	assert.Equal(t, 1, ret["readyCnt"])
	assert.Equal(t, 1, ret["totalCnt"])
	assert.Equal(t, int64(2), ret["restartCnt"])
	assert.Equal(t, "Running", ret["status"])
}
