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
	parser := podStatusParser{manifest: failedPodManifest}
	assert.Equal(t, "Failed", parser.Parse())

	parser = podStatusParser{manifest: succeededPodManifest}
	assert.Equal(t, "Succeeded", parser.Parse())

	parser = podStatusParser{manifest: runningPodManifest1}
	assert.Equal(t, "Running", parser.Parse())

	parser = podStatusParser{manifest: pendingPodManifest}
	assert.Equal(t, "Pending", parser.Parse())

	parser = podStatusParser{manifest: terminatingPodManifest}
	assert.Equal(t, "Terminating", parser.Parse())

	parser = podStatusParser{manifest: unknownPodManifest1}
	assert.Equal(t, "Unknown", parser.Parse())

	parser = podStatusParser{manifest: completedPodManifest}
	assert.Equal(t, "Completed", parser.Parse())

	parser = podStatusParser{manifest: waitReasonPodManifest}
	assert.Equal(t, "CreateContainerError", parser.Parse())

	parser = podStatusParser{manifest: signalPodManifest}
	assert.Equal(t, "Signal: 1", parser.Parse())

	parser = podStatusParser{manifest: exitCodePodManifest}
	assert.Equal(t, "ExitCode: 1", parser.Parse())

	parser = podStatusParser{manifest: runningPodManifest2}
	assert.Equal(t, "Running", parser.Parse())

	parser = podStatusParser{manifest: notReadyPodManifest}
	assert.Equal(t, "NotReady", parser.Parse())

	parser = podStatusParser{manifest: unknownPodManifest2}
	assert.Equal(t, "Unknown", parser.Parse())

	parser = podStatusParser{manifest: initSignalPodManifest}
	assert.Equal(t, "Init: Signal 1", parser.Parse())

	parser = podStatusParser{manifest: initExitCodePodManifest}
	assert.Equal(t, "Init: ExitCode 1", parser.Parse())

	parser = podStatusParser{manifest: initTermPodManifest}
	assert.Equal(t, "Init: term init", parser.Parse())

	parser = podStatusParser{manifest: initWaitPodManifest}
	assert.Equal(t, "Init: wait init", parser.Parse())

	parser = podStatusParser{manifest: initRunPodManifest}
	assert.Equal(t, "Init: 0/1", parser.Parse())
}
