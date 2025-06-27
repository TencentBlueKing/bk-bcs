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

// Package bkbase bkbase
package bkbase

// BaseResp base resp
type BaseResp struct {
	Result  bool   `json:"result"`
	Message string `json:"message"`
}

// CreateDataIDBody create data id body
// NOCC:tosa/indent(设计如此)
const CreateDataIDBody = `{
	"config": [
		{
			"kind": "DataId",
			"metadata": {
				"namespace": "%s",
				"name": "%s",
				"labels": {
					"plat_project": "bk_bcs"
				},
				"annotations": {}
			},
			"spec": {
				"description": "%s",
				"alias": "%s",
				"bizId": %d,
				"maintainers": [
					"admin"
				],
				"predefined": {
					"dataId": %d,
					"channel": {
						"kind": "KafkaChannel",
						"namespace": "%s",
						"name": "%s"
					},
					"topic": "%s"
				},
				"eventType": "log"
			}
		}
	]
}`

// CreateDatabusBody create databus body
// NOCC:tosa/indent(设计如此)
const CreateDatabusBody = `{
	"config": [
		{
			"kind": "Databus",
			"metadata": {
				"namespace": "%s",
				"name": "%s",
				"labels": {
					"plat_project": "bk_bcs"
				},
				"annotations": {}
			},
			"spec": {
				"sources": [
					{
						"kind": "DataId",
						"namespace": "%s",
						"name": "%s"
					}
				],
				"sinks": [
					{
						"kind": "ChannelBinding",
						"namespace": "%s",
						"name": "%s"
					}
				],
				"transforms": [
					{
						"kind": "Clean",
						"rules": [
							{
								"input_id": "__raw_data",
								"operator": {
									"type": "json_de"
								},
								"output_id": "json_data"
							},
							{
								"input_id": "json_data",
								"operator": {
									"type": "get",
									"missing_strategy": null,
									"key_index": [
										{
											"value": "ext",
											"type": "key"
										}
									]
								},
								"output_id": "ext"
							},
							{
								"output_id": "bk_bcs_cluster_id",
								"input_id": "ext",
								"operator": {
									"type": "assign",
									"key_index": "bk_bcs_cluster_id",
									"output_type": "string"
								}
							},
							{
								"output_id": "io_kubernetes_pod",
								"input_id": "ext",
								"operator": {
									"type": "assign",
									"key_index": "io_kubernetes_pod",
									"output_type": "string"
								}
							},
							{
								"input_id": "json_data",
								"operator": {
									"type": "get",
									"missing_strategy": null,
									"key_index": [
										{
											"value": "items",
											"type": "key"
										}
									]
								},
								"output_id": "items"
							},
							{
								"input_id": "items",
								"operator": {
									"type": "iter"
								},
								"output_id": "item"
							},
							{
								"input_id": "item",
								"operator": {
									"type": "get",
									"missing_strategy": null,
									"key_index": [
										{
											"value": "data",
											"type": "key"
										}
									]
								},
								"output_id": "data"
							},
							{
								"output_id": "data",
								"input_id": "item",
								"operator": {
									"type": "assign",
									"key_index": "data",
									"output_type": "string"
								}
							},
							{
								"input_id": "data",
								"operator": {
									"type": "json_de"
								},
								"output_id": "iter_json_data"
							},
							{
								"output_id": "datetime",
								"input_id": "json_data",
								"operator": {
									"type": "assign",
									"key_index": "datetime",
									"is_time_field": true,
									"output_type": "string",
									"in_place_time_parsing": null,
									"default_value": null,
									"time_format": {
										"format": "%%Y-%%m-%%d %%H:%%M:%%S",
										"zone": 8
									}
								}
							},
							{
								"output_id": "level",
								"input_id": "iter_json_data",
								"operator": {
									"type": "assign",
									"key_index": "level",
									"output_type": "string"
								}
							}
						],
						"filter_rules": "True"
					}
				],
				"maintainers": [
					"admin"
				]
			}
		}
	]
}`
