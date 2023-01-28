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

package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/constants"
	storage "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/proto"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/util"
)

func TestMapToStruct(t *testing.T) {
	const (
		extraTag     = "extra"
		extraConTag  = "extra_contain"
		idTag        = "ID"
		envTag       = "env"
		kindTag      = "kind"
		levelTag     = "level"
		componentTag = "component"
		typeTag      = "type"
	)

	var conditionTagList = [...]string{idTag, envTag, kindTag, levelTag, componentTag, typeTag, constants.ClusterIDTag,
		"extraInfo.name", "extraInfo.namespace", "extraInfo.kind"}

	event := &storage.BcsStorageEvent{
		XId:       "111",
		ClusterId: "xxx",
		Env:       "test",
		Kind:      "pod",
		Level:     "111",
		Component: "ddd",
		Type:      "hhh",
		Describe:  "ssssxxxxx",
		EventTime: "35654654654",
		ExtraInfo: &storage.EventExtraInfo{
			Namespace: "bcs-system",
			Name:      "storage",
			Kind:      "ddd",
		},
	}

	data := util.StructToMap(event)
	log.Println(data)

	for _, k := range conditionTagList {
		keys := []string{k}
		if strings.Contains(k, ".") {
			keys = strings.Split(k, ".")
		}
		var temp = data
		var result interface{}
		for _, key := range keys {
			switch temp[key].(type) {
			case map[string]interface{}:
				temp = temp[key].(map[string]interface{})
			default:
				result = temp[key]
			}
		}
		log.Println(k, result)
	}
}

func TestGetStructTags(t *testing.T) {
	fmt.Println(util.GetStructTags(&storage.IPPoolStatic{}))
	fmt.Println(util.GetStructTags(&storage.IPPoolStaticDetail{}))
	fmt.Println(util.GetStructTags(&storage.Pod{}))
	fmt.Println(util.GetStructTags(&storage.ReplicaSet{}))
	fmt.Println(util.GetStructTags(&storage.DeploymentK8S{}))
	fmt.Println(util.GetStructTags(&storage.ServiceK8S{}))
	fmt.Println(util.GetStructTags(&storage.ConfigMapK8S{}))
	fmt.Println(util.GetStructTags(&storage.SecretK8S{}))
	fmt.Println(util.GetStructTags(&storage.EndpointsK8S{}))
	fmt.Println(util.GetStructTags(&storage.Ingress{}))
	fmt.Println(util.GetStructTags(&storage.Namespace{}))
	fmt.Println(util.GetStructTags(&storage.Node{}))
	fmt.Println(util.GetStructTags(&storage.DaemonSet{}))
	fmt.Println(util.GetStructTags(&storage.Job{}))
	fmt.Println(util.GetStructTags(&storage.StatefulSet{}))
}

func TestToJson(t *testing.T) {
	secret := &storage.SecretK8S{
		XId:       "xxxxwd",
		ClusterId: "xxx",
	}

	bytes, _ := json.Marshal(secret)

	fmt.Println(string(bytes))
}
