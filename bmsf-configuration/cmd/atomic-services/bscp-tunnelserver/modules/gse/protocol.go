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

package gse

import (
	"encoding/binary"
	"errors"

	"bk-bscp/cmd/atomic-services/bscp-tunnelserver/modules"
)

const (
	// PlatServiceProtocolMagic protocol magic num.
	PlatServiceProtocolMagic = 0xdeadbeef

	// PlatServiceProtocolVersion protocol version.
	PlatServiceProtocolVersion = 0

	// PlatServiceProtocolMainHeaderLen protocol main header length,
	// len(Magic) + len(MessageType) + len(ProtocolVersion) + len(PackageLength).
	PlatServiceProtocolMainHeaderLen = 12

	// PlatServiceProtocolExtraHeaderLen protocol extra header length,
	// len(ServiceID) + len(TransmitType) + len(SessionID) + len(MessageSequenceID) + len(Reserved0) + len(Reserved1).
	PlatServiceProtocolExtraHeaderLen = 32

	// PlatServiceProtocolHeaderLen protocol message header length.
	PlatServiceProtocolHeaderLen = PlatServiceProtocolMainHeaderLen + PlatServiceProtocolExtraHeaderLen // 44

	// PlatServiceProtocolMaxLen protocol max length.
	PlatServiceProtocolMaxLen = 2 * 1024 * 1024 // 2MB
)

// PlatServiceMessageMainHeader is GSE plat service message main header.
type PlatServiceMessageMainHeader struct {
	// Magic is protocol magic num(0xdeadbeef).
	Magic uint32

	// MessageType is message type for different cmd.
	MessageType uint16

	// ProtocolVersion is protocol version num.
	ProtocolVersion uint16

	// PackageLength is message package length(PackageLength = len(MessageHeader) + len(MessageBody)), unit byte.
	PackageLength uint32
}

func (h *PlatServiceMessageMainHeader) serializedSize() int {
	return PlatServiceProtocolMainHeaderLen
}

func (h *PlatServiceMessageMainHeader) serialize() []byte {
	buffer := make([]byte, h.serializedSize())

	// serialize to big endian net byte order.
	binary.BigEndian.PutUint32(buffer, h.Magic)
	binary.BigEndian.PutUint16(buffer[4:], h.MessageType)
	binary.BigEndian.PutUint16(buffer[6:], h.ProtocolVersion)
	binary.BigEndian.PutUint32(buffer[8:], h.PackageLength)

	return buffer
}

// PlatServiceMessageExtraHeader is GSE plat service message extra header.
type PlatServiceMessageExtraHeader struct {
	// ServiceID is used for different users in GSE plat message channel service.
	ServiceID uint32

	// TransmitType is message transmit type, including broadcast, random and point-to-point.
	TransmitType uint32

	// SessionID is session id in GSE plat service.
	SessionID uint64

	// MessageSequenceID is message sequence id.
	MessageSequenceID uint64

	// Reserved0 reserved field.
	Reserved0 uint32

	// Reserved1 reserved field.
	Reserved1 uint32
}

func (h *PlatServiceMessageExtraHeader) serializedSize() int {
	return PlatServiceProtocolExtraHeaderLen
}

func (h *PlatServiceMessageExtraHeader) serialize() []byte {
	buffer := make([]byte, h.serializedSize())

	// serialize to big endian net byte order.
	binary.BigEndian.PutUint32(buffer, h.ServiceID)
	binary.BigEndian.PutUint32(buffer[4:], h.TransmitType)
	binary.BigEndian.PutUint64(buffer[8:], h.SessionID)
	binary.BigEndian.PutUint64(buffer[16:], h.MessageSequenceID)
	binary.BigEndian.PutUint32(buffer[24:], h.Reserved0)
	binary.BigEndian.PutUint32(buffer[28:], h.Reserved1)

	return buffer
}

// PlatServiceMessageHeader is GSE plat service message header including main header and extra header.
type PlatServiceMessageHeader struct {
	// MainHeader is message mian header.
	MainHeader *PlatServiceMessageMainHeader

	// ExtraHeader is message extra header.
	ExtraHeader *PlatServiceMessageExtraHeader
}

func (h *PlatServiceMessageHeader) serializedSize() int {
	return PlatServiceProtocolHeaderLen
}

func (h *PlatServiceMessageHeader) serialize() ([]byte, error) {
	// serialize main header.
	mainHeaderBuffer := h.MainHeader.serialize()

	// serialize extra header.
	extraHeaderBuffer := h.ExtraHeader.serialize()

	// message header.
	mainHeaderBuffer = append(mainHeaderBuffer, extraHeaderBuffer...)

	if len(mainHeaderBuffer) != h.serializedSize() {
		return nil, errors.New("serialize failed, invalid serialized size of message header")
	}
	return mainHeaderBuffer, nil
}

const (
	// PlatServiceCMDRegisterReq register request cmd.
	PlatServiceCMDRegisterReq = 0x5001
	// PlatServiceCMDRegisterResp register response cmd.
	PlatServiceCMDRegisterResp = 0x5002

	// PlatServiceCMDPushToPluginReq push to plugin request cmd.
	PlatServiceCMDPushToPluginReq = 0x5007
	// PlatServiceCMDPushToPluginResp push to plugin response cmd.
	PlatServiceCMDPushToPluginResp = 0x5008

	// PlatServiceCMDSubscribeFromPluginReq subscribe message from plugin request cmd.
	PlatServiceCMDSubscribeFromPluginReq = 0x5009
	// PlatServiceCMDSubscribeFromPluginResp subscribe message from plugin response cmd.
	PlatServiceCMDSubscribeFromPluginResp = 0x500a
)

const (
	// PlatServiceErrCodeOK is GSE plat service error code for success result.
	PlatServiceErrCodeOK = 0

	// PlatServiceErrMsgSuccess is GSE plat service error message for success result.
	PlatServiceErrMsgSuccess = "success"
)

// PlatServiceSimpleBody is simple GSE plat service response message body, only including errcode and errmsg.
// Used for register/subscribe cmd.
type PlatServiceSimpleBody struct {
	// ErrCode GSE plat service errcode.
	ErrCode int32 `json:"errcode"`

	// ErrMsg GSE plat service errmsg.
	ErrMsg string `json:"errmsg"`
}

// PlatServicePushToPluginBody is message body for push to plugin cmd.
type PlatServicePushToPluginBody struct {
	// Agents target agents.
	Agents []*modules.AgentInformation `json:"agent"`

	// Message target message.
	Message string `json:"message"`
}

// PlatServiceSubscribeFromPluginBody is message body for recv subscribed message from plugin cmd.
type PlatServiceSubscribeFromPluginBody struct {
	// Agent origin agent.
	Agent *modules.AgentInformation `json:"agent"`

	// Message plugin message.
	Message string `json:"message"`
}

// Message is GSE plat service message struct.
type Message struct {
	// Header is GSE plat service message header.
	Header *PlatServiceMessageHeader

	// Body is GSE plat service message body.
	Body []byte
}
