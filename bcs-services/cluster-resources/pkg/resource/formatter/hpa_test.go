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

var lightHPAManifest1 = map[string]interface{}{
	"spec": map[string]interface{}{
		"metrics": []interface{}{
			map[string]interface{}{
				"type": "External",
				"external": map[string]interface{}{
					"target": map[string]interface{}{
						"averageValue": "30m",
					},
				},
			},
		},
	},
}

var lightHPAManifest2 = map[string]interface{}{
	"spec": map[string]interface{}{
		"metrics": []interface{}{
			map[string]interface{}{
				"type": "External",
				"external": map[string]interface{}{
					"target": map[string]interface{}{
						"averageValue": "30m",
					},
				},
			},
		},
	},
	"status": map[string]interface{}{
		"currentMetrics": []interface{}{
			map[string]interface{}{
				"external": map[string]interface{}{
					"current": map[string]interface{}{
						"averageValue": "10m",
					},
				},
			},
		},
	},
}

var lightHPAManifest3 = map[string]interface{}{
	"spec": map[string]interface{}{
		"metrics": []interface{}{
			map[string]interface{}{
				"type": "External",
				"external": map[string]interface{}{
					"target": map[string]interface{}{
						"value": "30m",
					},
				},
			},
		},
	},
}

var lightHPAManifest4 = map[string]interface{}{
	"spec": map[string]interface{}{
		"metrics": []interface{}{
			map[string]interface{}{
				"type": "External",
				"external": map[string]interface{}{
					"target": map[string]interface{}{
						"value": "30m",
					},
				},
			},
		},
	},
	"status": map[string]interface{}{
		"currentMetrics": []interface{}{
			map[string]interface{}{
				"external": map[string]interface{}{
					"current": map[string]interface{}{
						"value": "10m",
					},
				},
			},
		},
	},
}

var lightHPAManifest5 = map[string]interface{}{
	"spec": map[string]interface{}{
		"metrics": []interface{}{
			map[string]interface{}{
				"type": "Pods",
				"pods": map[string]interface{}{
					"target": map[string]interface{}{
						"averageValue": "35m",
					},
				},
			},
		},
	},
}

var lightHPAManifest6 = map[string]interface{}{
	"spec": map[string]interface{}{
		"metrics": []interface{}{
			map[string]interface{}{
				"type": "Pods",
				"pods": map[string]interface{}{
					"target": map[string]interface{}{
						"averageValue": "35m",
					},
				},
			},
		},
	},
	"status": map[string]interface{}{
		"currentMetrics": []interface{}{
			map[string]interface{}{
				"pods": map[string]interface{}{
					"current": map[string]interface{}{
						"averageValue": "15m",
					},
				},
			},
		},
	},
}

var lightHPAManifest7 = map[string]interface{}{
	"spec": map[string]interface{}{
		"metrics": []interface{}{
			map[string]interface{}{
				"type": "Object",
				"object": map[string]interface{}{
					"target": map[string]interface{}{
						"averageValue": "40m",
					},
				},
			},
		},
	},
}

var lightHPAManifest8 = map[string]interface{}{
	"spec": map[string]interface{}{
		"metrics": []interface{}{
			map[string]interface{}{
				"type": "Object",
				"object": map[string]interface{}{
					"target": map[string]interface{}{
						"averageValue": "40m",
					},
				},
			},
		},
	},
	"status": map[string]interface{}{
		"currentMetrics": []interface{}{
			map[string]interface{}{
				"object": map[string]interface{}{
					"current": map[string]interface{}{
						"averageValue": "10m",
					},
				},
			},
		},
	},
}

var lightHPAManifest9 = map[string]interface{}{
	"spec": map[string]interface{}{
		"metrics": []interface{}{
			map[string]interface{}{
				"type": "Object",
				"object": map[string]interface{}{
					"target": map[string]interface{}{
						"value": "40m",
					},
				},
			},
		},
	},
}

var lightHPAManifest10 = map[string]interface{}{
	"spec": map[string]interface{}{
		"metrics": []interface{}{
			map[string]interface{}{
				"type": "Object",
				"object": map[string]interface{}{
					"target": map[string]interface{}{
						"value": "40m",
					},
				},
			},
		},
	},
	"status": map[string]interface{}{
		"currentMetrics": []interface{}{
			map[string]interface{}{
				"object": map[string]interface{}{
					"current": map[string]interface{}{
						"value": "10m",
					},
				},
			},
		},
	},
}

var lightHPAManifest11 = map[string]interface{}{
	"spec": map[string]interface{}{
		"metrics": []interface{}{
			map[string]interface{}{
				"type": "Resource",
				"resource": map[string]interface{}{
					"target": map[string]interface{}{
						"averageValue": "45m",
					},
				},
			},
		},
	},
}

var lightHPAManifest12 = map[string]interface{}{
	"spec": map[string]interface{}{
		"metrics": []interface{}{
			map[string]interface{}{
				"type": "Resource",
				"resource": map[string]interface{}{
					"target": map[string]interface{}{
						"averageValue": "45m",
					},
				},
			},
		},
	},
	"status": map[string]interface{}{
		"currentMetrics": []interface{}{
			map[string]interface{}{
				"resource": map[string]interface{}{
					"current": map[string]interface{}{
						"averageValue": "10m",
					},
				},
			},
		},
	},
}

var lightHPAManifest13 = map[string]interface{}{
	"spec": map[string]interface{}{
		"metrics": []interface{}{
			map[string]interface{}{
				"type": "Resource",
			},
		},
	},
}

var lightHPAManifest14 = map[string]interface{}{
	"spec": map[string]interface{}{
		"metrics": []interface{}{
			map[string]interface{}{
				"type": "Resource",
				"resource": map[string]interface{}{
					"target": map[string]interface{}{
						"averageUtilization": 45,
					},
				},
			},
		},
	},
}

var lightHPAManifest15 = map[string]interface{}{
	"spec": map[string]interface{}{
		"metrics": []interface{}{
			map[string]interface{}{
				"type": "Resource",
				"resource": map[string]interface{}{
					"target": map[string]interface{}{
						"averageUtilization": 45,
					},
				},
			},
		},
	},
	"status": map[string]interface{}{
		"currentMetrics": []interface{}{
			map[string]interface{}{
				"resource": map[string]interface{}{
					"current": map[string]interface{}{
						"averageUtilization": 10,
					},
				},
			},
		},
	},
}

var lightHPAManifest16 = map[string]interface{}{
	"spec": map[string]interface{}{
		"metrics": []interface{}{
			map[string]interface{}{
				"type": "ContainerResource",
				"containerResource": map[string]interface{}{
					"target": map[string]interface{}{
						"averageValue": "50m",
					},
				},
			},
		},
	},
}

var lightHPAManifest17 = map[string]interface{}{
	"spec": map[string]interface{}{
		"metrics": []interface{}{
			map[string]interface{}{
				"type": "ContainerResource",
				"containerResource": map[string]interface{}{
					"target": map[string]interface{}{
						"averageValue": "50m",
					},
				},
			},
		},
	},
	"status": map[string]interface{}{
		"currentMetrics": []interface{}{
			map[string]interface{}{
				"containerResource": map[string]interface{}{
					"current": map[string]interface{}{
						"averageValue": "10m",
					},
				},
			},
		},
	},
}

var lightHPAManifest18 = map[string]interface{}{
	"spec": map[string]interface{}{
		"metrics": []interface{}{
			map[string]interface{}{
				"type": "ContainerResource",
				"containerResource": map[string]interface{}{
					"target": map[string]interface{}{
						"averageUtilization": 50,
					},
				},
			},
		},
	},
}

var lightHPAManifest19 = map[string]interface{}{
	"spec": map[string]interface{}{
		"metrics": []interface{}{
			map[string]interface{}{
				"type": "ContainerResource",
				"containerResource": map[string]interface{}{
					"target": map[string]interface{}{
						"averageUtilization": 50,
					},
				},
			},
		},
	},
	"status": map[string]interface{}{
		"currentMetrics": []interface{}{
			map[string]interface{}{
				"containerResource": map[string]interface{}{
					"current": map[string]interface{}{
						"averageUtilization": 10,
					},
				},
			},
		},
	},
}

var lightHPAManifest20 = map[string]interface{}{
	"spec": map[string]interface{}{
		"metrics": []interface{}{
			map[string]interface{}{
				"type": "External",
				"external": map[string]interface{}{
					"target": map[string]interface{}{
						"averageValue": "30m",
					},
				},
			},
			map[string]interface{}{
				"type": "Pods",
				"pods": map[string]interface{}{
					"target": map[string]interface{}{
						"averageValue": "35m",
					},
				},
			},
		},
	},
}

var lightHPAManifest21 = map[string]interface{}{
	"spec": map[string]interface{}{
		"metrics": []interface{}{
			map[string]interface{}{
				"type": "Pods",
				"pods": map[string]interface{}{
					"target": map[string]interface{}{
						"averageValue": "35m",
					},
				},
			},
			map[string]interface{}{
				"type": "Object",
				"object": map[string]interface{}{
					"target": map[string]interface{}{
						"averageValue": "40m",
					},
				},
			},
			map[string]interface{}{
				"type": "Resource",
			},
			map[string]interface{}{
				"type": "ContainerResource",
			},
		},
	},
	"status": map[string]interface{}{
		"currentMetrics": []interface{}{
			map[string]interface{}{
				"pods": map[string]interface{}{
					"current": map[string]interface{}{
						"averageValue": "15m",
					},
				},
			},
		},
	},
}

func TestHPAMetricParser(t *testing.T) {
	parser := hpaTargetsParser{manifest: lightHPAManifest1}
	assert.Equal(t, "<unknown>/30m (avg)", parser.Parse())

	parser = hpaTargetsParser{manifest: lightHPAManifest2}
	assert.Equal(t, "10m/30m (avg)", parser.Parse())

	parser = hpaTargetsParser{manifest: lightHPAManifest3}
	assert.Equal(t, "<unknown>/30m", parser.Parse())

	parser = hpaTargetsParser{manifest: lightHPAManifest4}
	assert.Equal(t, "10m/30m", parser.Parse())

	parser = hpaTargetsParser{manifest: lightHPAManifest5}
	assert.Equal(t, "<unknown>/35m", parser.Parse())

	parser = hpaTargetsParser{manifest: lightHPAManifest6}
	assert.Equal(t, "15m/35m", parser.Parse())

	parser = hpaTargetsParser{manifest: lightHPAManifest7}
	assert.Equal(t, "<unknown>/40m (avg)", parser.Parse())

	parser = hpaTargetsParser{manifest: lightHPAManifest8}
	assert.Equal(t, "10m/40m (avg)", parser.Parse())

	parser = hpaTargetsParser{manifest: lightHPAManifest9}
	assert.Equal(t, "<unknown>/40m", parser.Parse())

	parser = hpaTargetsParser{manifest: lightHPAManifest10}
	assert.Equal(t, "10m/40m", parser.Parse())

	parser = hpaTargetsParser{manifest: lightHPAManifest11}
	assert.Equal(t, "<unknown>/45m", parser.Parse())

	parser = hpaTargetsParser{manifest: lightHPAManifest12}
	assert.Equal(t, "10m/45m", parser.Parse())

	parser = hpaTargetsParser{manifest: lightHPAManifest13}
	assert.Equal(t, "<unknown>/<auto>", parser.Parse())

	parser = hpaTargetsParser{manifest: lightHPAManifest14}
	assert.Equal(t, "<unknown>/45%", parser.Parse())

	parser = hpaTargetsParser{manifest: lightHPAManifest15}
	assert.Equal(t, "10%/45%", parser.Parse())

	parser = hpaTargetsParser{manifest: lightHPAManifest16}
	assert.Equal(t, "<unknown>/50m", parser.Parse())

	parser = hpaTargetsParser{manifest: lightHPAManifest17}
	assert.Equal(t, "10m/50m", parser.Parse())

	parser = hpaTargetsParser{manifest: lightHPAManifest18}
	assert.Equal(t, "<unknown>/50%", parser.Parse())

	parser = hpaTargetsParser{manifest: lightHPAManifest19}
	assert.Equal(t, "10%/50%", parser.Parse())

	parser = hpaTargetsParser{manifest: lightHPAManifest20}
	assert.Equal(t, "<unknown>/30m (avg), <unknown>/35m", parser.Parse())

	parser = hpaTargetsParser{manifest: lightHPAManifest21}
	assert.Equal(t, "15m/35m, <unknown>/40m (avg), <unknown>/<auto> + 1 more...", parser.Parse())
}
