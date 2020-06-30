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

package client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"net"
	"time"
)

type dataHead struct {
	msgType uint32
	dataID  uint32
	utctime uint32
	bodylen uint32
	resv    [2]uint32
}

// AsyncProducer producer for reporting
type AsyncProducer interface {
	Input(msg *ProducerMessage)
	Close()
}

// ProducerMessage message for reporting
type ProducerMessage struct {
	DataID uint32
	Value  []byte
}

// Client gse data pipe client
type client struct {

	// The domain socket address
	endpoint string
	conn     net.Conn
	input    chan *ProducerMessage
	sigstop  chan bool
}

func (dhead *dataHead) packageData(data []byte) ([]byte, error) {

	buffer := new(bytes.Buffer)

	// package head
	if err := binary.Write(buffer, binary.BigEndian, dhead); nil != err {
		return nil, fmt.Errorf("can not package the data head,%v", err)
	}

	head := buffer.Bytes()
	head = append(head, data...)

	return head, nil
}

// Connect connect to gse data pipe
func (gsec *client) connect() error {

	conn, err := net.Dial("unix", gsec.endpoint)

	if err != nil {
		return fmt.Errorf("no gse data pipe  available, maybe gseagent is not running,%s", gsec.endpoint)
	}

	blog.Info("current endpoint of gsedatapipe: %s", gsec.endpoint)
	gsec.conn = conn
	return nil
}

// Send write data to data pipe
func (gsec *client) send(dataid uint32, data []byte) error {

	if gsec.conn == nil {
		if err := gsec.connect(); err != nil {
			return err
		}
	}
	//glog.Info(string(data), len(data))
	dhead := dataHead{0xc01, dataid, 0, uint32(len(data)), [2]uint32{0, 0}}

	packageData, err := dhead.packageData(data)
	if nil != err {
		return err
	}

	err = gsec.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

	if err != nil {
		blog.Error(" SET WRITE DEAD LINE FAILED: %v", err)
	}

	if _, err = gsec.conn.Write(packageData); err != nil {

		blog.Error("SEND DATA TO DATA PIPE FAILED: %v", err)
		gsec.conn.Close()
		gsec.conn = nil
		return err
	}

	return nil
}

// Close close gse data pipe
func (gsec *client) close() error {

	gsec.sigstop <- true

	return nil
}

func (gsec *client) Close() {

	gsec.close()
}

// Input message input
func (gsec *client) Input(msg *ProducerMessage) {

	select {
	case gsec.input <- msg:
	default:
		//blog.Warn("pipe is full ")
	}
}

// New create a new client of gse data pipe
func New(endpoint string) (AsyncProducer, error) {

	cli := &client{
		endpoint: endpoint,
		input:    make(chan *ProducerMessage, 4096),
		sigstop:  make(chan bool)}

	if err := cli.connect(); nil != err {

		blog.Info("can not connect remote pipe : %s, error info: %v", cli.endpoint, err)
	}

	go func() {

		defer func() {

			cli.conn.Close()
			cli.conn = nil
		}()

		for {

			select {

			case msg := <-cli.input:

				if err := cli.send(msg.DataID, msg.Value); nil != err {

					blog.Error("can not send data to remote pipe : %s, error info: %v", cli.endpoint, err)

				}

			case <-cli.sigstop:

				blog.Error("disconnected with the remote pipe : %s,", cli.endpoint)
				return
			}
		}

	}()

	return cli, nil
}
