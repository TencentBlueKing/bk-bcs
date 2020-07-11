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

package main

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-exporter/app/config"
)

// AsyncProducer 统一接口
type AsyncProducer interface {
	Input(msg *ProducerMessage)
}
type dataHead struct {
	magicNumber uint32
	msgType     uint32
	bodylen     uint32
	dataID      uint32
	timestamp   uint32
	resv        uint32
}

// ProducerMessage 消息体
type ProducerMessage struct {
	DataID uint32
	Value  []byte
}

// client 客户端
type client struct {
	connIdx int                 // 索引
	conns   map[string]net.Conn // key is the address
	input   chan *ProducerMessage
	cert    *config.CertConfig
}

// packageData package the data into protocol data stream
func (cli *dataHead) packageData(data []byte) ([]byte, error) {

	buffer := new(bytes.Buffer)

	// package head
	if err := binary.Write(buffer, binary.BigEndian, cli); nil != err {
		return nil, fmt.Errorf("can not package the data head,%v", err)
	}

	head := buffer.Bytes()
	head = append(head, data...)

	return head, nil
}

// getConnection return a connection and it's address
func (cli *client) getConnection() (string, net.Conn) {

	if 0 == len(cli.conns) {
		return "", nil
	}

	idx := cli.connIdx % len(cli.conns)

	defer func() {
		cli.connIdx++
	}()

	for key, val := range cli.conns {

		if idx <= 0 && nil != val {
			return key, val
		}

		idx--

	}

	return "", nil
}

// validConnectionCount return the valid connection count
func (cli *client) validConnectionCount() int {
	validCnt := 0
	for _, val := range cli.conns {
		if nil != val {
			validCnt++
		}
	}
	return validCnt
}

func (cli *client) delArray(s []net.Conn, idx int) []net.Conn {
	s[len(s)-1], s[idx] = s[idx], s[len(s)-1]
	return s[:len(s)-1]
}

// connect connect to gse data pipe
func (cli *client) connect() error {

	invalidCnt := 0
	for address, conn := range cli.conns {

		if nil != conn {
			continue
		}

		var conn net.Conn
		var err error
		if cli.cert.IsSSL {
			cfg, err := ssl.ClientTslConfVerity(cli.cert.CAFile, cli.cert.CertFile, cli.cert.KeyFile, cli.cert.CertPasswd)
			if err != nil {
				blog.Error("can not connect to the remote server[%s], error is %s", address, err.Error())
				continue
			}
			conn, err = tls.Dial("tcp", address, cfg)
		} else {
			conn, err = net.Dial("tcp", address)
		}
		if err != nil {
			invalidCnt++
			blog.Error("can not connect to the remote server[%s], error is %s", address, err.Error())
			continue
		}

		cli.conns[address] = conn

	}

	if len(cli.conns) == invalidCnt {
		return errors.New("no usable connection ")
	}

	return nil
}

// disconnect clear all connections
func (cli *client) disconnect() {
	for _, conn := range cli.conns {
		if nil != conn {
			conn.Close()
		}
	}
	cli.conns = nil
}

// send send data to remote server
func (cli *client) send(dataid uint32, data []byte) error {

	// 在此处增加判断是为了启动多余的goroutine 带来不必要的同步
	if cli.validConnectionCount() < len(cli.conns)/2 {
		if err := cli.connect(); nil != err {
			blog.Error("failed to connect remote server, error info is %s", err.Error())
		}
	}

	address, conn := cli.getConnection()
	if conn == nil {
		if err := cli.connect(); err != nil {
			return err
		}

		address, conn = cli.getConnection()
		if nil == conn {
			return errors.New("no usable connection")
		}
	}

	dhead := dataHead{
		magicNumber: 0xdeadbeef, // 协议约定，如有变化需要和蓝鲸数据平台负责数据接收的同学确认
		msgType:     10000,      // 协议约定，如有变更需要和蓝鲸数据平台负责数据接收的同学确认
		dataID:      dataid,
		timestamp:   uint32(time.Now().UTC().Unix()),
		bodylen:     uint32(len(data)),
		resv:        0,
	}

	packageData, err := dhead.packageData(data)
	if nil != err {
		return err
	}

	err = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	if err != nil {
		blog.Error(" failed to set dead line: %s", err.Error())
	}

	if _, err = conn.Write(packageData); err != nil {

		blog.Error("failed to send data: %v", err)
		conn.Close()
		cli.conns[address] = nil
		return err
	}
	blog.V(3).Info("bkdata client success to send data")

	return nil
}

func (cli *client) Input(msg *ProducerMessage) {

	select {
	case cli.input <- msg:
	default:
		blog.Warn("pipe is full ")
	}
	return
}

// createClient create a new client of gse data pipe
func createClient(cfg *config.Config) (AsyncProducer, error) {
	endpoint := cfg.OutputAddress

	if 0 == len(endpoint) {
		return nil, errors.New("not set the remote server address")
	}

	cli := &client{
		conns: make(map[string]net.Conn),
		input: make(chan *ProducerMessage, 4096*2),
		cert:  cfg.ClientCert,
	}

	addressItems := strings.Split(endpoint, ",")
	for _, item := range addressItems {
		cli.conns[item] = nil
	}

	if err := cli.connect(); err != nil {
		return nil, err
	}

	go func() {

		defer func() {
			cli.disconnect()
		}()

		for {

			select {

			case msg := <-cli.input:

				if nil == msg {
					// 主动退出
					blog.Error("disconnected with the remote server : %s,", endpoint)
					return
				}

				if err := cli.send(msg.DataID, msg.Value); nil != err {
					blog.Error("can not send data to remote server : %s, error info: %v", endpoint, err)
				}

			}
		}

	}()

	return cli, nil
}
