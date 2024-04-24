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

package scr

type ApplyRequest struct {
	App       string      `json:"app"`
	User      string      `json:"user"`
	Follower  string      `json:"follower,omitempty"`
	Describe  string      `json:"describe,omitempty"`
	ApplyInfo []ApplyInfo `json:"applyInfo"`
}

type ApplyInfo struct {
	ClientVersion       string   `json:"clientVersion"`
	DbUser              string   `json:"dbUser"`
	DbPassword          string   `json:"dbPassword"`
	DbName              string   `json:"dbName"`
	TbName              string   `json:"tbName"`
	Grants              []string `json:"grants"`
	SourceIPInput       string   `json:"sourceIPInput"`
	TargetInstanceInput string   `json:"targetInstanceInput"`
}

type SCRResponse struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	Jobid string `json:"jobid"`
	URL   string `json:"url"`
}
