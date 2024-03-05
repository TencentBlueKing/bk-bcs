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
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/util/strategicpatch"

	v1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/api/v1"
)

func TestCompress(t *testing.T) {
	// ng := v1.NoticeGroup{}
	// ng.Spec.Groups = append(ng.Spec.Groups, &v1.NoticeGroupDetail{})
	// ng.Spec.Groups[0].Name = "porterlin-test-gen"
	// ng.Spec.Groups[0].Users = []string{"porterlin"}
	// ng.Spec.Groups[0].Alert = make(map[string]*v1.NoticeAlert)
	// ng.Spec.Groups[0].Alert["00:00--23:59"] = &v1.NoticeAlert{
	// 	Fatal: &v1.NoticeType{
	// 		Type: []string{"rtx"},
	// 	},
	// 	Remind: &v1.NoticeType{
	// 		Type: []string{"rtx"},
	// 	},
	// 	Warning: &v1.NoticeType{
	// 		Type: []string{"rtx"},
	// 	},
	// }
	//
	// ng.Spec.Groups[0].Action = make(map[string]*v1.NoticeAction)
	// ng.Spec.Groups[0].Action["00:00--23:59"] = &v1.NoticeAction{
	// 	Execute: &v1.NoticeType{
	// 		Type: []string{"rtx"},
	// 	},
	// 	ExecuteFailed: &v1.NoticeType{
	// 		Type: []string{"sms"},
	// 	},
	// 	ExecuteSuccess: &v1.NoticeType{
	// 		Type: []string{"mail"},
	// 	},
	// }
	//
	// // fo := &FileOperator{}
	// // outpath, err := fo.Compress(ng.Spec)
	// // if err != nil {
	// // 	println(err.Error())
	// // }
	// // println("outpath:" + outpath)
	//
	// bts, _ := json.Marshal(ng.Spec.Groups)
	// println(string(bts))
	//
	// ng2 := &v1.NoticeGroupDetail{}
	// ng2.Name = "porterlin-test-gen"
	// ng2.Users = []string{"porterlin"}
	// ng2.Alert = make(map[string]*v1.NoticeAlert)
	// ng2.Alert["00:00--23:59"] = &v1.NoticeAlert{
	// 	Fatal: &v1.NoticeType{
	// 		Type: []string{"rtx"},
	// 	},
	// 	Remind: &v1.NoticeType{
	// 		Type: []string{"rtx"},
	// 	},
	// 	Warning: &v1.NoticeType{
	// 		Type: []string{"rtx"},
	// 	},
	// }
	//
	// ng2.Action = make(map[string]*v1.NoticeAction)
	// ng2.Action["00:00--23:59"] = &v1.NoticeAction{
	// 	Execute: &v1.NoticeType{
	// 		Type: []string{"rtx", "sms"},
	// 	},
	// 	ExecuteFailed: &v1.NoticeType{
	// 		Type: []string{"sms"},
	// 	},
	// 	ExecuteSuccess: &v1.NoticeType{
	// 		Type: []string{"mail"},
	// 	},
	// }
	//
	// ng3 := &v1.NoticeGroupDetail{}
	// ng3.Name = "porterlin-test-gen"
	// ng3.Users = []string{"porterlin"}
	// ng3.Alert = make(map[string]*v1.NoticeAlert)
	// ng3.Alert["00:00--23:59"] = &v1.NoticeAlert{
	// 	Fatal: &v1.NoticeType{
	// 		Type: []string{"rtx"},
	// 	},
	// 	Remind: &v1.NoticeType{
	// 		Type: []string{"rtx"},
	// 	},
	// 	Warning: &v1.NoticeType{
	// 		Type: []string{"rtx"},
	// 	},
	// }
	//
	// ng3.Action = make(map[string]*v1.NoticeAction)
	// ng3.Action["00:00--23:59"] = &v1.NoticeAction{
	// 	Execute: &v1.NoticeType{
	// 		Type: []string{"rtx", "sms"},
	// 	},
	// 	ExecuteFailed: &v1.NoticeType{
	// 		Type: []string{"sms"},
	// 	},
	// 	ExecuteSuccess: &v1.NoticeType{
	// 		Type: []string{"mail"},
	// 	},
	// }

	mr1 := v1.MonitorRuleSpec{
		Rules: []*v1.MonitorRuleDetail{
			{
				Name:       "a",
				Enabled:    true,
				ActiveTime: "1234",
				Detect: &v1.Detect{
					Algorithm: &v1.Algorithm{
						Remind: []*v1.AlgorithmConfig{
							{
								ConfigStr: "<5",
								Type:      "Threshold",
							},
							{
								ConfigStr: "<10",
								Type:      "Threshold",
							},
						},
					},
				},
			},
		},
	}
	mr2 := v1.MonitorRuleSpec{
		Rules: []*v1.MonitorRuleDetail{
			{
				Name:       "a",
				Enabled:    true,
				ActiveTime: "1234",
				Detect: &v1.Detect{
					Algorithm: &v1.Algorithm{
						Remind: []*v1.AlgorithmConfig{
							{
								ConfigStr: "<10",
								Type:      "Threshold",
							},
							{
								ConfigStr: "<5",
								Type:      "Threshold",
							},
						},
					},
				},
				Labels: []string{"123"},
			},
		},
	}
	mr3 := v1.MonitorRuleSpec{
		Rules: []*v1.MonitorRuleDetail{
			{
				Name:       "a",
				Enabled:    true,
				ActiveTime: "1234",
				Detect: &v1.Detect{
					Algorithm: &v1.Algorithm{
						Remind: []*v1.AlgorithmConfig{
							{
								ConfigStr: "<10",
								Type:      "Threshold",
							},
							{
								ConfigStr: "<5",
								Type:      "Threshold",
							},
						},
					},
				},
			},
			{
				Name:       "b",
				Enabled:    false,
				ActiveTime: "1234",
				Detect: &v1.Detect{
					Algorithm: &v1.Algorithm{
						Remind: []*v1.AlgorithmConfig{
							{
								ConfigStr: "<10",
								Type:      "Threshold",
							},
						},
					},
				},
			},
		},
	}
	// mr1 := v1.MonitorRuleDetail{
	// 	Name:       "name",
	// 	Enabled:    true,
	// 	ActiveTime: "1234",
	// 	Detect: &v1.Detect{
	// 		Algorithm: &v1.Algorithm{
	// 			Remind: []*v1.AlgorithmConfig{
	// 				{
	// 					ConfigStr: "<10",
	// 					Type:      "Threshold",
	// 				},
	// 			},
	// 			Operator: "",
	// 		},
	// 	},
	// }
	//
	// mr2 := v1.MonitorRuleDetail{
	// 	Name:       "name",
	// 	Enabled:    true,
	// 	ActiveTime: "12345",
	// 	Detect: &v1.Detect{
	// 		Algorithm: &v1.Algorithm{
	// 			Remind: []*v1.AlgorithmConfig{
	// 				{
	// 					ConfigStr: "<15",
	// 					Type:      "Threshold",
	// 				},
	// 			},
	// 			Operator: "",
	// 		},
	// 	},
	// }
	//
	// mr3 := v1.MonitorRuleDetail{
	// 	Name:       "name",
	// 	Enabled:    false,
	// 	ActiveTime: "1234",
	// 	Detect: &v1.Detect{
	// 		Algorithm: &v1.Algorithm{
	// 			Remind: []*v1.AlgorithmConfig{
	// 				{
	// 					ConfigStr: "<10",
	// 					Type:      "Threshold",
	// 				},
	// 			},
	// 			Operator: "",
	// 		},
	// 	},
	// }

	println("deep equal :", reflect.DeepEqual(mr1.Rules[0], mr3.Rules[0]))

	original, _ := json.Marshal(mr1)
	current, _ := json.Marshal(mr2)
	modified, _ := json.Marshal(mr3)
	println("marshal", string(current))

	ty := strategicpatch.GetTagStructTypeOrDie(v1.MonitorRuleSpec{})
	println(ty.Kind().String())
	println(ty.Kind() == reflect.Struct)
	println(ty.Name())
	println(ty.String())

	mergeItemStructSchema := strategicpatch.PatchMetaFromStruct{T: ty}
	println(mergeItemStructSchema.T.Kind().String())
	// println()

	// _, _, err := mergeItemStructSchema.LookupPatchMetadataForStruct("action")
	// if err != nil {
	// 	t.Fatal(err)
	// }
	actual, err := strategicpatch.CreateThreeWayMergePatch(original, modified, current, mergeItemStructSchema, false)
	if err != nil {
		t.Fatal(err.Error())
	}

	println(string(actual))

	result, err := strategicpatch.StrategicMergePatchUsingLookupPatchMeta(original, actual, mergeItemStructSchema)
	if err != nil {
		t.Fatal(err.Error())
	}

	println(string(result))

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.Encode(json.RawMessage(result))
	print(buf.String())

	// var r v1.MonitorRuleSpec
	// err = json.Unmarshal(result, &r)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	//
	// bts, _ := json.Marshal(r)
	// print(string(bts))
}

func TestDecompress(t *testing.T) {
	fo := &FileOperator{}
	err := fo.Decompress("/tmp/1068_config.tar.gz", "/tmp/1068_config")
	if err != nil {
		t.Fatal(err)
	}
}
