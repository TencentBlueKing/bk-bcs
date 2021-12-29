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
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/golang/protobuf/proto"
)

// PlatService is GSE plat service protocol handler.
type PlatService struct {
}

// NewPlatService creates a new PlatService instance.
func NewPlatService() *PlatService {
	return new(PlatService)
}

// DecodeHeader decodes GSE plat service message header base on tcp connection.
func (f *PlatService) DecodeHeader(conn *tls.Conn) (*PlatServiceMessageHeader, []byte, error) {
	return f.frame(conn)
}

func (f *PlatService) frame(conn *tls.Conn) (*PlatServiceMessageHeader, []byte, error) {
	b := NewBuffer(conn, PlatServiceProtocolHeaderLen)

	if err := b.Read(PlatServiceProtocolHeaderLen); err != nil {
		return nil, nil, err
	}

	header := &PlatServiceMessageHeader{
		MainHeader:  &PlatServiceMessageMainHeader{},
		ExtraHeader: &PlatServiceMessageExtraHeader{},
	}

	// magic.
	magic, err := b.DecodeUint32()
	if err != nil {
		return nil, nil, err
	}
	header.MainHeader.Magic = magic

	if header.MainHeader.Magic != PlatServiceProtocolMagic {
		return nil, nil, errors.New("magic num not matched")
	}

	// message type.
	messageType, err := b.DecodeUint16()
	if err != nil {
		return nil, nil, err
	}
	header.MainHeader.MessageType = messageType

	// protocol version.
	protocolVersion, err := b.DecodeUint16()
	if err != nil {
		return nil, nil, err
	}
	header.MainHeader.ProtocolVersion = protocolVersion

	// package length.
	packageLength, err := b.DecodeUint32()
	if err != nil {
		return nil, nil, err
	}
	header.MainHeader.PackageLength = packageLength

	if header.MainHeader.PackageLength > PlatServiceProtocolMaxLen {
		return nil, nil, fmt.Errorf("message too large[%d]", header.MainHeader.PackageLength)
	}

	// service id.
	serviceID, err := b.DecodeUint32()
	if err != nil {
		return nil, nil, err
	}
	header.ExtraHeader.ServiceID = serviceID

	// transmit type.
	transmitType, err := b.DecodeUint32()
	if err != nil {
		return nil, nil, err
	}
	header.ExtraHeader.TransmitType = transmitType

	// session id.
	sessionID, err := b.DecodeUint64()
	if err != nil {
		return nil, nil, err
	}
	header.ExtraHeader.SessionID = sessionID

	// message sequence id.
	messageSequenceID, err := b.DecodeUint64()
	if err != nil {
		return nil, nil, err
	}
	header.ExtraHeader.MessageSequenceID = messageSequenceID

	// reserved0.
	reserved0, err := b.DecodeUint32()
	if err != nil {
		return nil, nil, err
	}
	header.ExtraHeader.Reserved0 = reserved0

	// reserved1.
	reserved1, err := b.DecodeUint32()
	if err != nil {
		return nil, nil, err
	}
	header.ExtraHeader.Reserved1 = reserved1

	// decode message body.
	body := make([]byte, header.MainHeader.PackageLength-PlatServiceProtocolHeaderLen)
	if _, err := io.ReadFull(conn, body); err != nil {
		return nil, nil, err
	}
	return header, body, nil
}

// DecodeJSONBody decodes GSE plat service message json body.
func (f *PlatService) DecodeJSONBody(body []byte, message interface{}) error {
	return json.Unmarshal(body, message)
}

// EncodeJSONBody encodes GSE plat service message json body.
func (f *PlatService) EncodeJSONBody(message interface{}) ([]byte, error) {
	return json.Marshal(message)
}

// DecodePBBody decodes GSE plat service message protobuf body.
func (f *PlatService) DecodePBBody(body []byte, message proto.Message) error {
	// TODO: decode protobuf body.
	return nil
}

// EncodePBBody encodes GSE plat service message protobuf body.
func (f *PlatService) EncodePBBody(message proto.Message) ([]byte, error) {
	// TODO: encode protobuf body.
	return nil, nil
}

// Encode encodes message including header and body.
func (f *PlatService) Encode(header *PlatServiceMessageHeader, body []byte) ([]byte, error) {
	// reset header magic.
	header.MainHeader.Magic = PlatServiceProtocolMagic

	// reset header protocol version.
	header.MainHeader.ProtocolVersion = PlatServiceProtocolVersion

	// set package length.
	header.MainHeader.PackageLength = uint32(PlatServiceProtocolHeaderLen + len(body))

	// serialize header.
	message, err := header.serialize()
	if err != nil {
		return nil, err
	}

	// append body and build message.
	message = append(message, body...)

	return message, nil
}
