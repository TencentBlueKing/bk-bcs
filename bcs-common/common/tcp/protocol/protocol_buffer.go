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

package protocol

import (
	"bytes"
	"fmt"
	"io"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

// protocolBuffer 网络数据解析
type protocolBuffer struct {
	head    *MsgHead
	data    bytes.Buffer
	handler HandlerIf
}

// Write Writer 接口实现
func (cli *protocolBuffer) Write(buf []byte) (int, error) {

	_, err := cli.data.Write(buf)

	// 如果为解析过协议头，那么就需要从协议头开始解析
	if nil == cli.head {
		// 未读取过协议头数据
		if cli.data.Len() >= HeadLength() {

			// 需要解析协议头了
			headBuff := cli.data.Next(HeadLength())
			head, opErr := ConvertToMsgHead(headBuff)
			if nil != opErr {
				blog.Error("convert to msg head failed, error information is %s", opErr.Error())
				return len(buf), opErr
			}

			// 检查是否有效的协议头，防止脏数据
			if ProtocolMagicNum != head.Magic {
				blog.Error("it is not a valid message head")
				return len(buf), fmt.Errorf("it is not a valid message head")
			}
			cli.head = head
		}
	}

	if nil != cli.head {

		if cli.head.Length > uint32(cli.data.Len()) {
			// 数据不够，暂不处理
			return len(buf), err
		}

		// 读取数据段的数据
		data := cli.data.Next(int(cli.head.Length))

		if nil != cli.handler {

			_, handlerErr := cli.handler.Write(cli.head, data)

			if nil != handlerErr {
				blog.Error("handle data failed, error information is %s", handlerErr.Error())
				return len(buf), handlerErr
			}

		} else {
			blog.Warn("not set the handlerif , the data will be lost")
		}

		// 数据被取走之后需要将协议头置空
		cli.head = nil
	}

	return len(buf), nil
}

// CreateProtocolBuffer 创建protocolBuffer用于接收数据, handler 接收业务数据的接口
func CreateProtocolBuffer(handler HandlerIf) io.Writer {
	return &protocolBuffer{handler: handler}
}
