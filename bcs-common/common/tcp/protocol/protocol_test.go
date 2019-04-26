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

package protocol

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"
)

func TestHeadLength(t *testing.T) {
	t.Logf("Protocol Head Length:%d", HeadLength())
}

func TestConvertToMsgHead(t *testing.T) {

	buffer := new(bytes.Buffer)

	msgHead := &MsgHead{}
	msgHead.Magic = ProtocolMagicNum
	msgHead.ExtID = 0x01
	msgHead.Length = 1
	msgHead.Timestamp = time.Now().UnixNano()
	msgHead.Type = 1

	// package the head
	if err := binary.Write(buffer, binary.BigEndian, msgHead); nil != err {
		t.Errorf("can not package the head, error %s", err.Error())
	}

	// package the data after the head
	head := buffer.Bytes()
	t.Logf("MsgHead bytes:%+#v", head)

	tmpHead, tmpErr := ConvertToMsgHead(head)
	if nil != tmpErr {
		t.Errorf("convert to msghead failed, error %s", tmpErr.Error())
	} else {
		t.Logf("MsgHead: %+#v", tmpHead)
	}
}

func TestPackage(t *testing.T) {

	msgHead := &MsgHead{}
	msgHead.ExtID = 0x01
	msgHead.Length = 1
	msgHead.Timestamp = time.Now().UnixNano()
	msgHead.Type = 1

	resp, respErr := Package(msgHead, []byte("hello world"))
	if nil != respErr {
		t.Errorf("convert to msghead failed, error %s", respErr.Error())
	} else {
		t.Logf("data: %+v", resp)
	}

}
