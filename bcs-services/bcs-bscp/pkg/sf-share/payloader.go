/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package sfs

// PayloadBuilder defines the feed message's payload details.
type PayloadBuilder interface {
	Serializer
	// PayloadName returns the payload's name
	PayloadName() string
	// MessageType return the payload related message type.
	MessageType() FeedMessageType
}

// MessagingPayloadBuilder defines the sidecar message's payload details.
type MessagingPayloadBuilder interface {
	Serializer
	// PayloadName returns the payload's name
	PayloadName() string
	// MessagingType return the payload related message type.
	MessagingType() MessagingType
}

// Serializer defines the operations to encode and decode the message payload.
type Serializer interface {
	Encode() ([]byte, error)
}
