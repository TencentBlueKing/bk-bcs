/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package modules

import (
	pb "bk-bscp/internal/protocol/tunnelserver"
)

// AgentInformation is gse agent information.
type AgentInformation struct {
	// HostIP is agent host IP.
	HostIP string `json:"ip"`

	// CloudID is agent host cloud id.
	CloudID int32 `json:"cloudid"`
}

// GSERecvProcesserMessage is gse recv processer message.
type GSERecvProcesserMessage struct {
	// MsgSeqID tunnel down stream message sequence id.
	MsgSeqID uint64

	// Agent tunnel down stream message from.
	Agent *AgentInformation

	// DownStream down stream message.
	DownStream *pb.GeneralTunnelDownStream
}

// GSESendProcesserMessage is gse send processer message.
type GSESendProcesserMessage struct {
	// MsgSeqID tunnel up stream message sequence id.
	MsgSeqID uint64

	// Agents tunnel up stream message to.
	Agents []*AgentInformation

	// UpStream up stream message.
	UpStream *pb.GeneralTunnelUpStream
}
